#!/bin/bash

set -ouex pipefail

RELEASE="$(rpm -E %fedora)"

## -- hyprland COPR from solopasha
dnf5 -y copr enable solopasha/hyprland
dnf5 -y install xdg-desktop-portal-hyprland hyprland hyprland-contrib hyprland-plugins hyprpaper hyprpicker hypridle hyprshot hyprlock pyprland waybar-git xdg-desktop-portal-hyprland hyprland-qtutils

## -- swayosd
dnf5 -y copr enable erikreider/swayosd
dnf5 -y install swayosd

## -- Hyprland essentials (terminal, launcher, notifications, file manager, etc.)
dnf5 -y install foot wofi mako thunar brightnessctl playerctl polkit papirus-icon-theme wl-clipboard

## -- Apparatus
cp /delivery/build_files/apparatus/butler.sh /usr/bin/butler
cp /delivery/build_files/config/distrobox.conf /etc/distrobox/distrobox.conf

#sudo desktop-file-install /tmp/Apparatus.desktop
#sudo update-desktop-database

## -- Install/remove packages
rpm-ostree install distrobox docker libvirt-daemon-kvm qemu-kvm virt-manager tailscale
rpm-ostree install https://github.com/charmbracelet/gum/releases/download/v0.14.5/gum-0.14.5-1.x86_64.rpm
rpm-ostree override remove firefox-langpacks
rpm-ostree override remove firefox

## -- Enabling Systemd Unit File
systemctl enable podman.socket
systemctl enable docker

## -- System Configuration
# Fonts
curl -OL --output-dir /tmp https://github.com/ryanoasis/nerd-fonts/releases/download/v3.2.1/Hack.zip
unzip -d /tmp/hack-font /tmp/Hack.zip
cp -r /tmp/hack-font /usr/share/fonts/
fc-cache -f -v

# distrobox

# -- Hyprland Configuration
# Create apparatus config directories
mkdir -p /usr/share/apparatus/hypr
mkdir -p /usr/share/apparatus/waybar
mkdir -p /usr/share/apparatus/mako
mkdir -p /usr/share/apparatus/wallpapers

# Copy default configs
cp /delivery/build_files/config/hypr/* /usr/share/apparatus/hypr/
cp /delivery/build_files/config/waybar/* /usr/share/apparatus/waybar/
cp /delivery/build_files/config/mako/* /usr/share/apparatus/mako/

# Copy wallpaper (if exists)
if [ -f /delivery/build_files/wallpapers/default.jpg ]; then
    cp /delivery/build_files/wallpapers/default.jpg /usr/share/apparatus/wallpapers/
fi

# Copy skeleton configs (stub files that source system defaults)
# Hypr configs use 'source' directive to inherit from /usr/share/apparatus/
# Waybar/mako are copies since they don't support sourcing
mkdir -p /etc/skel/.config
cp -r /delivery/build_files/skel/.config/* /etc/skel/.config/

# Enable swayosd service (for on-screen display)
systemctl enable swayosd-libinput-backend.service
