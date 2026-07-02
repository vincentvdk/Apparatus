#!/bin/bash
set -ouex pipefail

# Tool versions - load from apparatus.env, fallback to defaults
load_env() {
    local env_file="$1"
    if [ -f "$env_file" ]; then
        set -a
        . "$env_file"
        set +a
    else
        echo "Warning: $env_file not found."
    fi
}

load_env /delivery/build_files/apparatus.env

RELEASE="$(rpm -E %fedora)"
VERSION="${APPARATUS_VERSION:-dev}"

# Ensure essential versions are set
: "${HYPRDYNAMICMONITORS_VERSION:?HYPRDYNAMICMONITORS_VERSION must be defined in apparatus.env}"
: "${WALKER_VERSION:?WALKER_VERSION must be defined in apparatus.env}"
: "${ELEPHANT_VERSION:?ELEPHANT_VERSION must be defined in apparatus.env}"

## -- Install dnf5 plugins (needed for COPR support)
dnf5 -y install dnf5-plugins

## -- Display Manager & Wayland base
# dejavu-sans-fonts needed for plymouth password prompt (Image.Text requires fonts in initramfs)
# Using greetd + gtkgreet instead of GDM to save ~2GB of GNOME dependencies
dnf5 -y install greetd greetd-selinux gtkgreet cage xorg-x11-server-Xwayland xdg-user-dirs xdg-utils plymouth plymouth-plugin-script plymouth-plugin-label dejavu-sans-fonts

## -- Configure Plymouth for graphical boot
# Download connect theme from adi1090x/plymouth-themes
mkdir -p /usr/share/plymouth/themes/connect
THEME_BASE="https://raw.githubusercontent.com/adi1090x/plymouth-themes/master/pack_1/connect"
curl -sL "$THEME_BASE/connect.plymouth" -o /usr/share/plymouth/themes/connect/connect.plymouth
curl -sL "$THEME_BASE/connect.script" -o /usr/share/plymouth/themes/connect/connect.script
# Download progress animation frames (0-119)
for i in $(seq 0 119); do
    curl -sL "$THEME_BASE/progress-$i.png" -o /usr/share/plymouth/themes/connect/progress-$i.png &
done
wait

plymouth-set-default-theme connect
# Ensure default theme symlink exists (for bootc first-boot initramfs rebuild)
ln -sfn connect /usr/share/plymouth/themes/default
# Dracut config for graphical boot with LUKS prompt
# For bootc, config must be in /usr/lib/dracut/dracut.conf.d
mkdir -p /usr/lib/dracut/dracut.conf.d
cat > /usr/lib/dracut/dracut.conf.d/50-apparatus-plymouth.conf <<'EOF'
# Include Plymouth module for graphical boot
add_dracutmodules+=" plymouth "
# Include GPU driver for graphical LUKS password prompt
add_drivers+=" amdgpu "
# Include USB/HID drivers for keyboard input during boot
add_drivers+=" usbhid hid_generic xhci_hcd ehci_hcd "
# Include fonts for plymouth password prompt (Image.Text needs fonts)
install_items+=" /usr/share/fonts/dejavu-sans-fonts/DejaVuSans.ttf /usr/share/fonts/dejavu-sans-fonts "
# Include Plymouth theme files
install_items+=" /usr/share/plymouth/themes/connect/connect.plymouth "
install_items+=" /usr/share/plymouth/themes/connect/connect.script "
EOF
# Add progress image items to dracut config
for i in $(seq 0 119); do
    echo "install_items+=\" /usr/share/plymouth/themes/connect/progress-$i.png \"" >> /usr/lib/dracut/dracut.conf.d/50-apparatus-plymouth.conf
done

## -- Rebuild initramfs with Plymouth theme for bootc
# For bootc, we need to explicitly rebuild the initramfs during image build
# since bootc doesn't support runtime initramfs regeneration like rpm-ostree
KVER=$(ls /usr/lib/modules | head -1)
dracut --force --kver "$KVER" --no-hostonly

## -- hyprland COPR from solopasha
dnf5 -y copr enable solopasha/hyprland
dnf5 -y install xdg-desktop-portal-hyprland hyprland hyprland-contrib hyprland-plugins hyprpaper hyprpicker hypridle hyprshot hyprlock hyprpolkitagent pyprland waybar-git xdg-desktop-portal-hyprland hyprland-qtutils uwsm satty

## -- swayosd
dnf5 -y copr enable erikreider/swayosd
dnf5 -y install swayosd

## -- hyprdynamicmonitors (automatic monitor profile switching for Hyprland)
curl -L -o /tmp/hyprdynamicmonitors.tar.gz \
    "https://github.com/fiffeek/hyprdynamicmonitors/releases/download/v${HYPRDYNAMICMONITORS_VERSION}/hyprdynamicmonitors_Linux_x86_64.tar.gz"
tar -xzf /tmp/hyprdynamicmonitors.tar.gz -C /tmp
install -m 755 /tmp/hyprdynamicmonitors /usr/bin/hyprdynamicmonitors
rm -f /tmp/hyprdynamicmonitors.tar.gz /tmp/hyprdynamicmonitors

## -- walker (modern app launcher) and elephant (backend service)
dnf5 -y install gtk4-layer-shell
curl -L -o /tmp/walker.tar.gz \
    "https://github.com/abenz1267/walker/releases/download/v${WALKER_VERSION}/walker-v${WALKER_VERSION}-x86_64-unknown-linux-gnu.tar.gz"
tar -xzf /tmp/walker.tar.gz -C /tmp
install -m 755 /tmp/walker /usr/bin/walker
rm -f /tmp/walker.tar.gz /tmp/walker

## -- elephant (backend for walker - indexes apps, files, etc.)
ELEPHANT_BASE="https://github.com/abenz1267/elephant/releases/download/v${ELEPHANT_VERSION}"
curl -L -o /tmp/elephant.tar.gz "${ELEPHANT_BASE}/elephant-linux-amd64.tar.gz"
curl -L -o /tmp/elephant-desktopapplications.tar.gz "${ELEPHANT_BASE}/desktopapplications-linux-amd64.tar.gz"
tar -xzf /tmp/elephant.tar.gz -C /tmp
tar -xzf /tmp/elephant-desktopapplications.tar.gz -C /tmp
install -m 755 /tmp/elephant-linux-amd64 /usr/bin/elephant
# Providers are .so files, go in /usr/lib/elephant/providers
mkdir -p /usr/lib/elephant/providers
install -m 755 /tmp/desktopapplications-linux-amd64.so /usr/lib/elephant/providers/desktopapplications.so
rm -f /tmp/elephant*.tar.gz /tmp/elephant-linux-amd64 /tmp/desktopapplications-linux-amd64.so

# Elephant systemd user service (auto-starts with graphical session)
mkdir -p /usr/lib/systemd/user
cp /delivery/build_files/config/systemd/user/elephant.service /usr/lib/systemd/user/
# Enable by creating static symlink (works for all users without needing preset)
mkdir -p /usr/lib/systemd/user/graphical-session.target.wants
ln -sf ../elephant.service /usr/lib/systemd/user/graphical-session.target.wants/elephant.service

## -- Hyprland essentials (terminal, launcher, notifications, file manager, etc.)
dnf5 -y install kitty wofi mako thunar brightnessctl playerctl polkit wl-clipboard gvfs gvfs-smb gvfs-fuse

## -- Bluetooth & Network
dnf5 -y install blueman network-manager-applet NetworkManager-wifi NetworkManager-tui wireguard-tools

## -- Power management (needed for hyprdynamicmonitors lid/power detection)
# tuned-ppd is Fedora 41+ replacement for power-profiles-daemon
dnf5 -y install upower tuned-ppd

## -- Hardware support (Framework AMD laptops)
dnf5 -y install fprintd iio-sensor-proxy usbutils

## -- Audio
dnf5 -y install pipewire pipewire-pulseaudio wireplumber pavucontrol

## -- Development & System tools
# Note: Virtualization (libvirt/qemu/virt-manager) and docker removed to reduce image size
# Install these in a distrobox if needed
dnf5 -y install distrobox podman git curl unzip flatpak

## -- Printing
dnf5 -y install system-config-printer

## -- Gum (for butler TUI)
dnf5 -y install https://github.com/charmbracelet/gum/releases/download/v0.17.0/gum-0.17.0-1.x86_64.rpm

## -- Apparatus
cp /delivery/build_files/apparatus/butler.sh /usr/bin/butler
mkdir -p /etc/distrobox
cp /delivery/build_files/config/distrobox.conf /etc/distrobox/distrobox.conf

# Image info for ISO installer (like Bluefin)
mkdir -p /usr/share/apparatus
cat > /usr/share/apparatus/image-info.json <<EOF
{
  "image-name": "apparatus-os",
  "image-tag": "latest",
  "image-ref": "ghcr.io/vincentvdk/apparatus-os"
}
EOF


# Apparatus scripts
mkdir -p /usr/libexec/apparatus
cp /delivery/build_files/apparatus/first-login.sh /usr/libexec/apparatus/
cp /delivery/build_files/apparatus/firstboot-setup.sh /usr/libexec/apparatus/
chmod 755 /usr/libexec/apparatus/first-login.sh
chmod 755 /usr/libexec/apparatus/firstboot-setup.sh

# Smart-split script for kitty (detects distrobox and enters same container)
cp /delivery/build_files/apparatus/smart-split.sh /usr/libexec/apparatus/smart-split
chmod 755 /usr/libexec/apparatus/smart-split

# Config update checker (systemd timer + notify-send)
cp /delivery/build_files/apparatus/check-config-updates.sh /usr/libexec/apparatus/check-config-updates
chmod 755 /usr/libexec/apparatus/check-config-updates
cp /delivery/build_files/config/systemd/user/apparatus-config-check.service /usr/lib/systemd/user/
cp /delivery/build_files/config/systemd/user/apparatus-config-check.timer /usr/lib/systemd/user/
mkdir -p /usr/lib/systemd/user/timers.target.wants
ln -sf ../apparatus-config-check.timer /usr/lib/systemd/user/timers.target.wants/apparatus-config-check.timer

# Bootc update checker (systemd timer + notify-send)
cp /delivery/build_files/apparatus/check-bootc-updates.sh /usr/libexec/apparatus/check-bootc-updates
chmod 755 /usr/libexec/apparatus/check-bootc-updates
cp /delivery/build_files/config/systemd/apparatus-bootc-check.service /usr/lib/systemd/system/
cp /delivery/build_files/config/systemd/apparatus-bootc-check.timer /usr/lib/systemd/system/
mkdir -p /usr/lib/systemd/system/timers.target.wants
ln -sf ../apparatus-bootc-check.timer /usr/lib/systemd/system/timers.target.wants/apparatus-bootc-check.timer

# Profile.d scripts (sourced on login)
cp /delivery/build_files/config/profile.d/*.sh /etc/profile.d/
chmod 644 /etc/profile.d/distrobox-safe.sh

## -- Fix hyprland desktop files (upstream has invalid DesktopNames key)
cp /delivery/build_files/config/wayland-sessions/*.desktop /usr/share/wayland-sessions/

## -- UWSM environment config (populated from dotfiles repo below)

## -- greetd configuration (gtkgreet greeter running under cage)
# Create greeter user for greetd (runs the greeter process)
useradd -r -M -s /bin/false greeter || true

# SELinux: Make xdm_t (greetd's domain) permissive for authentication
# greetd runs as xdm_t per greetd-selinux package file contexts
dnf5 -y install selinux-policy-devel
mkdir -p /tmp/selinux-build
cp /delivery/build_files/config/selinux/greetd-auth.te /tmp/selinux-build/
cd /tmp/selinux-build
checkmodule -M -m -o greetd-auth.mod greetd-auth.te
semodule_package -o greetd-auth.pp -m greetd-auth.mod
semodule -i greetd-auth.pp
cd /
rm -rf /tmp/selinux-build

mkdir -p /etc/greetd
cp /delivery/build_files/config/greetd/config.toml /etc/greetd/
cp /delivery/build_files/config/greetd/gtkgreet.css /etc/greetd/
chmod 644 /etc/greetd/*.css /etc/greetd/*.toml
chmod 755 /etc/greetd

## -- Disable GDM (from fedora-bootc base image) and enable greetd
systemctl mask gdm.service
ln -sf /usr/lib/systemd/system/greetd.service /etc/systemd/system/display-manager.service
systemctl enable greetd.service
systemctl enable podman.socket

# Bootc switch service (runs once after install to point updates to GHCR)
cp /delivery/build_files/config/systemd/apparatus-bootc-switch.service /usr/lib/systemd/system/
systemctl enable apparatus-bootc-switch.service

# Firstboot service (sets up user config on first boot)
cp /delivery/build_files/config/systemd/apparatus-firstboot.service /usr/lib/systemd/system/
systemctl enable apparatus-firstboot.service

# SwayOSD server service (for on-screen display)
cp /delivery/build_files/config/systemd/swayosd-server.service /usr/lib/systemd/system/

# Podman controller delegation for rootless containers
mkdir -p /usr/lib/systemd/system/user@.service.d
cp /delivery/build_files/config/systemd/user@.service.d/delegate.conf /usr/lib/systemd/system/user@.service.d/

## -- Mask services that don't work on immutable ostree systems
systemctl mask systemd-remount-fs.service

## -- Disable automatic bootc updates (replaced by notification-only service)
systemctl mask bootc-fetch-apply-updates.timer
systemctl mask bootc-fetch-apply-updates.service

## -- System Configuration
# Nerd Fonts
curl -OL --output-dir /tmp https://github.com/ryanoasis/nerd-fonts/releases/download/v3.2.1/Hack.zip &
curl -OL --output-dir /tmp https://github.com/ryanoasis/nerd-fonts/releases/download/v3.2.1/JetBrainsMono.zip &
curl -OL --output-dir /tmp https://github.com/ryanoasis/nerd-fonts/releases/download/v3.4.0/Noto.zip &
curl -OL --output-dir /tmp https://github.com/ahatem/IoskeleyMono/releases/download/v2.0.0/IoskeleyMono-NerdFont.zip &
wait
unzip -d /tmp/hack-font /tmp/Hack.zip
unzip -d /tmp/jetbrains-font /tmp/JetBrainsMono.zip
unzip -d /tmp/notosans-font /tmp/Noto.zip
unzip -d /tmp/ioskeley-font /tmp/IoskeleyMono-NerdFont.zip
cp -r /tmp/hack-font /usr/share/fonts/
cp -r /tmp/jetbrains-font /usr/share/fonts/
cp -r /tmp/notosans-font /usr/share/fonts/
cp -r /tmp/ioskeley-font /usr/share/fonts/
fc-cache -f -v

# Cleanup temp files
rm -rf /tmp/*.zip /tmp/hack-font /tmp/jetbrains-font /tmp/notosans-font /tmp/ioskeley-font

# distrobox

# -- Clone dotfiles repo for default user-facing configs
DOTFILES_REPO="${DOTFILES_REPO:-https://github.com/vincentvdk/apparatus-dotfiles.git}"
git clone --depth=1 "$DOTFILES_REPO" /tmp/dotfiles

# Copy OS configs from chezmoi source (home/dot_config/) to /usr/share/apparatus/
cd /tmp/dotfiles/home/dot_config
for dir in hypr kitty waybar mako walker uwsm; do
    [ -d "$dir" ] && cp -r "$dir" /usr/share/apparatus/
done
cd /

# Copy uwsm env to system location as well
mkdir -p /etc/uwsm
cp /usr/share/apparatus/uwsm/env /etc/uwsm/env

# Copy themes
cp -r /tmp/dotfiles/themes /usr/share/apparatus/themes/
rm -rf /tmp/dotfiles

# Rio config and flatpaks (non-dotfiles repo configs)
mkdir -p /usr/share/apparatus/rio
mkdir -p /usr/share/apparatus/wallpapers
cp /delivery/build_files/config/rio/* /usr/share/apparatus/rio/
cp /delivery/build_files/config/flatpaks.conf /usr/share/apparatus/

# Ensure apparatus files are world-readable
chmod -R a+rX /usr/share/apparatus

# First-login via XDG autostart (runs on first graphical login)
mkdir -p /etc/xdg/autostart
cp /delivery/build_files/config/autostart/apparatus-first-login.desktop /etc/xdg/autostart/

# Copy wallpaper
if [ -f /delivery/build_files/wallpapers/default.jpg ]; then
    cp /delivery/build_files/wallpapers/default.jpg /usr/share/apparatus/wallpapers/
fi

# -- Hardware Support (Framework laptops/desktops)
mkdir -p /etc/modprobe.d
cp /delivery/build_files/config/modprobe.d/*.conf /etc/modprobe.d/

# Kernel parameters for bootc
# Must use /usr/lib/bootc/kargs.d/ with TOML format
mkdir -p /usr/lib/bootc/kargs.d
cat > /usr/lib/bootc/kargs.d/50-apparatus.toml <<EOF
kargs = ["quiet", "splash", "plymouth.enable=1", "rd.plymouth=1", "amd_pstate=active", "amdgpu.dcdebugmask=0x10", "amdgpu.abmlevel=0", "amdgpu.sg_display=0"]
EOF

# Enable swayosd services (for on-screen display)
systemctl enable swayosd-libinput-backend.service
systemctl enable swayosd-server.service

## -- Custom os-release for Apparatus (affects GRUB menu entry name)
# Keep ID=fedora for bootc-image-builder compatibility
cat > /etc/os-release <<EOF
NAME="Apparatus OS"
VERSION="${VERSION} (Based on Fedora ${RELEASE})"
ID=fedora
VERSION_ID=${RELEASE}
PLATFORM_ID="platform:f${RELEASE}"
PRETTY_NAME="Apparatus OS ${VERSION}"
ANSI_COLOR="0;38;2;60;110;180"
LOGO=fedora-logo-icon
CPE_NAME="cpe:/o:fedoraproject:fedora:${RELEASE}"
DEFAULT_HOSTNAME="apparatus"
HOME_URL="https://github.com/vincentvdk/apparatus"
SUPPORT_URL="https://github.com/vincentvdk/apparatus/issues"
BUG_REPORT_URL="https://github.com/vincentvdk/apparatus/issues"
VARIANT="Hyprland Desktop"
VARIANT_ID=hyprland
OSTREE_VERSION=${VERSION}
EOF

## -- Workaround for bootc-image-builder vendor detection issue
# See: https://github.com/osbuild/image-builder-cli/issues/421
# Create EFI vendor directories and populate with shim/grub files
mkdir -p /boot/efi/EFI/fedora
mkdir -p /boot/efi/EFI/BOOT

# Copy shim and grub files to EFI directories
if [ -f /boot/efi/EFI/fedora/shimx64.efi ]; then
    echo "Shim already exists in EFI/fedora"
elif [ -f /usr/share/shim/*/shimx64.efi ]; then
    cp /usr/share/shim/*/shimx64.efi /boot/efi/EFI/fedora/
    cp /usr/share/shim/*/shimx64.efi /boot/efi/EFI/BOOT/BOOTX64.EFI
fi

# Copy grub to EFI directory if not present
if [ -f /usr/lib/grub/x86_64-efi/grub.efi ]; then
    cp /usr/lib/grub/x86_64-efi/grub.efi /boot/efi/EFI/fedora/grubx64.efi 2>/dev/null || true
fi

# Reinstall shim and grub to ensure proper setup
dnf5 -y reinstall shim-x64 grub2-efi-x64 grub2-common 2>/dev/null || true

# Also ensure bootupd updates directory has vendor info
mkdir -p /usr/lib/bootupd/updates/EFI/fedora
mkdir -p /usr/lib/bootupd/updates/EFI/BOOT
if [ -f /usr/share/shim/*/shimx64.efi ]; then
    cp /usr/share/shim/*/shimx64.efi /usr/lib/bootupd/updates/EFI/fedora/ 2>/dev/null || true
    cp /usr/share/shim/*/shimx64.efi /usr/lib/bootupd/updates/EFI/BOOT/BOOTX64.EFI 2>/dev/null || true
fi

## -- Final cleanup to reduce image size
rm -rf /tmp/* /var/tmp/*
rm -rf /var/log/*
rm -rf /var/cache/fontconfig/*
rm -rf /root/.cache/*
