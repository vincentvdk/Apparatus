#!/bin/bash
# Hook script for titanoboa - runs in container rootfs before squashing
# Create sway.desktop so titanoboa uses sway livesys scripts (Hyprland is Sway-based)

set -x

# Remove hyprland.desktop and create sway.desktop symlink
# titanoboa supports sway* but not hyprland*
rm -f /usr/share/wayland-sessions/hyprland.desktop
ln -sf gnome-wayland.desktop /usr/share/wayland-sessions/sway.desktop

# Show what's available
ls -la /usr/share/wayland-sessions/
