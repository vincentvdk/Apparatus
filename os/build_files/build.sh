#!/bin/bash

set -ouex pipefail

RELEASE="$(rpm -E %fedora)"

## -- Install dnf5 plugins (needed for COPR support)
dnf5 -y install dnf5-plugins

## -- Display Manager & Wayland base
# gnome-shell included for titanoboa live session support (installed system uses Hyprland)
dnf5 -y install gdm gnome-shell xorg-x11-server-Xwayland xdg-user-dirs xdg-utils

## -- hyprland COPR from solopasha
dnf5 -y copr enable solopasha/hyprland
dnf5 -y install xdg-desktop-portal-hyprland hyprland hyprland-contrib hyprland-plugins hyprpaper hyprpicker hypridle hyprshot hyprlock hyprpolkitagent pyprland waybar-git xdg-desktop-portal-hyprland hyprland-qtutils uwsm

## -- swayosd
dnf5 -y copr enable erikreider/swayosd
dnf5 -y install swayosd

## -- Hyprland essentials (terminal, launcher, notifications, file manager, etc.)
dnf5 -y install kitty wofi mako thunar brightnessctl playerctl polkit papirus-icon-theme wl-clipboard

## -- Bluetooth & Network
dnf5 -y install blueman network-manager-applet NetworkManager-wifi

## -- Audio
dnf5 -y install pipewire pipewire-pulseaudio wireplumber

## -- Development & System tools
# Note: Virtualization (libvirt/qemu/virt-manager) and docker removed to reduce image size
# Install these in a distrobox if needed
dnf5 -y install distrobox podman git curl unzip

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

# First-login setup script
mkdir -p /usr/libexec/apparatus
cp /delivery/build_files/apparatus/first-login.sh /usr/libexec/apparatus/
chmod +x /usr/libexec/apparatus/first-login.sh

# Systemd user service for first-login prompt (runs on graphical session)
mkdir -p /usr/lib/systemd/user
cp /delivery/build_files/config/systemd/apparatus-first-login.service /usr/lib/systemd/user/
systemctl --global enable apparatus-first-login.service

## -- Enabling Systemd services
systemctl enable gdm.service
systemctl enable podman.socket

## -- Ensure GNOME is found first by titanoboa for live session
# Copy gnome desktop file with name that sorts before 'h' (hyprland)
# find -type f only finds regular files, not symlinks
cp /usr/share/wayland-sessions/gnome-wayland.desktop /usr/share/wayland-sessions/aaa-gnome.desktop

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

# Copy wallpaper
if [ -f /delivery/build_files/wallpapers/default.jpg ]; then
    cp /delivery/build_files/wallpapers/default.jpg /usr/share/apparatus/wallpapers/
fi

# -- Hardware Support (Framework laptops)
mkdir -p /etc/modprobe.d
cp /delivery/build_files/config/modprobe.d/*.conf /etc/modprobe.d/

# Enable swayosd service (for on-screen display)
systemctl enable swayosd-libinput-backend.service

## -- Final cleanup to reduce image size
rm -rf /tmp/* /var/tmp/*
rm -rf /var/log/*
rm -rf /var/cache/fontconfig/*
rm -rf /root/.cache/*

