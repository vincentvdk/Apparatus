#!/bin/bash

set -ouex pipefail

RELEASE="$(rpm -E %fedora)"
VERSION="${APPARATUS_VERSION:-dev}"

# Tool versions
HYPRDYNAMICMONITORS_VERSION="${HYPRDYNAMICMONITORS_VERSION:-1.4.0}"
WALKER_VERSION="${WALKER_VERSION:-2.12.2}"
ELEPHANT_VERSION="${ELEPHANT_VERSION:-2.17.2}"

## -- Install dnf5 plugins (needed for COPR support)
dnf5 -y install dnf5-plugins

## -- Display Manager & Wayland base
# dejavu-sans-fonts needed for plymouth password prompt (Image.Text requires fonts in initramfs)
dnf5 -y install gdm xorg-x11-server-Xwayland xdg-user-dirs xdg-utils plymouth plymouth-plugin-script plymouth-plugin-label dejavu-sans-fonts

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
# Dracut config for graphical boot with LUKS prompt
# For bootc, config must be in /usr/lib/dracut/dracut.conf.d
mkdir -p /usr/lib/dracut/dracut.conf.d
cat > /usr/lib/dracut/dracut.conf.d/50-apparatus-plymouth.conf <<EOF
add_dracutmodules+=" plymouth "
# Include GPU driver for graphical LUKS password prompt
add_drivers+=" amdgpu "
# Include USB/HID drivers for keyboard input during boot
add_drivers+=" usbhid hid_generic xhci_hcd ehci_hcd "
# Include fonts for plymouth password prompt (Image.Text needs fonts)
install_items+=" /usr/share/fonts/dejavu-sans-fonts/DejaVuSans.ttf /usr/share/fonts/dejavu-sans-fonts "
EOF

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
dnf5 -y install kitty wofi mako thunar brightnessctl playerctl polkit papirus-icon-theme wl-clipboard gvfs gvfs-smb gvfs-fuse

## -- Bluetooth & Network
dnf5 -y install blueman network-manager-applet NetworkManager-wifi NetworkManager-tui wireguard-tools

## -- Power management (needed for hyprdynamicmonitors lid/power detection)
# tuned-ppd is Fedora 41+ replacement for power-profiles-daemon
dnf5 -y install upower tuned-ppd

## -- Hardware support (Framework AMD laptops)
dnf5 -y install fprintd iio-sensor-proxy

## -- Audio
dnf5 -y install pipewire pipewire-pulseaudio wireplumber pavucontrol

## -- Development & System tools
# Note: Virtualization (libvirt/qemu/virt-manager) and docker removed to reduce image size
# Install these in a distrobox if needed
dnf5 -y install distrobox podman git curl unzip flatpak

## -- Gum (for butler TUI)
dnf5 -y install https://github.com/charmbracelet/gum/releases/download/v0.14.5/gum-0.14.5-1.x86_64.rpm

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

## -- Fix hyprland desktop files (upstream has invalid DesktopNames key)
cp /delivery/build_files/config/wayland-sessions/*.desktop /usr/share/wayland-sessions/

## -- UWSM environment config
mkdir -p /etc/uwsm
cp /delivery/build_files/config/uwsm/env /etc/uwsm/env

## -- Set Hyprland (UWSM) as default session for new users
mkdir -p /etc/accountsservice/user-templates
cp /delivery/build_files/config/accountsservice/user-templates/standard /etc/accountsservice/user-templates/
cp /delivery/build_files/config/accountsservice/user-templates/administrator /etc/accountsservice/user-templates/

## -- Enabling Systemd services
systemctl enable gdm.service
systemctl enable podman.socket

# Bootc switch service (runs once after install to point updates to GHCR)
cp /delivery/build_files/config/systemd/apparatus-bootc-switch.service /usr/lib/systemd/system/
systemctl enable apparatus-bootc-switch.service

# Firstboot service (sets up user config on first boot)
cp /delivery/build_files/config/systemd/apparatus-firstboot.service /usr/lib/systemd/system/
systemctl enable apparatus-firstboot.service

## -- Mask services that don't work on immutable ostree systems
systemctl mask systemd-remount-fs.service

## -- System Configuration
# Fonts (download in parallel)
curl -OL --output-dir /tmp https://github.com/ryanoasis/nerd-fonts/releases/download/v3.2.1/Hack.zip &
curl -OL --output-dir /tmp https://github.com/ryanoasis/nerd-fonts/releases/download/v3.2.1/JetBrainsMono.zip &
curl -OL --output-dir /tmp https://github.com/ryanoasis/nerd-fonts/releases/download/v3.4.0/Noto.zip &
wait
unzip -d /tmp/hack-font /tmp/Hack.zip
unzip -d /tmp/jetbrains-font /tmp/JetBrainsMono.zip
unzip -d /tmp/notosans-font /tmp/Noto.zip
cp -r /tmp/hack-font /usr/share/fonts/
cp -r /tmp/jetbrains-font /usr/share/fonts/
cp -r /tmp/notosans-font /usr/share/fonts/
fc-cache -f -v

# Cleanup temp files
rm -rf /tmp/*.zip /tmp/hack-font /tmp/jetbrains-font /tmp/notosans-font

# distrobox

# -- Hyprland Configuration
# Default configs in /usr/share/apparatus/ (copied to user home by butler init)
mkdir -p /usr/share/apparatus/hypr
mkdir -p /usr/share/apparatus/waybar
mkdir -p /usr/share/apparatus/mako
mkdir -p /usr/share/apparatus/kitty
mkdir -p /usr/share/apparatus/rio
mkdir -p /usr/share/apparatus/wallpapers
mkdir -p /usr/share/apparatus/uwsm
mkdir -p /usr/share/apparatus/themes

cp /delivery/build_files/config/hypr/* /usr/share/apparatus/hypr/
cp /delivery/build_files/config/waybar/* /usr/share/apparatus/waybar/
cp /delivery/build_files/config/mako/* /usr/share/apparatus/mako/
cp /delivery/build_files/config/kitty/* /usr/share/apparatus/kitty/
cp /delivery/build_files/config/rio/* /usr/share/apparatus/rio/
cp /delivery/build_files/config/uwsm/* /usr/share/apparatus/uwsm/
cp -r /delivery/build_files/config/themes/* /usr/share/apparatus/themes/

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

# Enable swayosd service (for on-screen display)
systemctl enable swayosd-libinput-backend.service

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

## -- Rebuild initramfs with plymouth and amdgpu
# For bootc, initramfs lives in /usr/lib/modules/$kver/initramfs.img
KVER=$(ls /usr/lib/modules | head -1)
DRACUT_NO_XATTR=1 dracut -vf /usr/lib/modules/$KVER/initramfs.img "$KVER"

## -- Final cleanup to reduce image size
rm -rf /tmp/* /var/tmp/*
rm -rf /var/log/*
rm -rf /var/cache/fontconfig/*
rm -rf /root/.cache/*

