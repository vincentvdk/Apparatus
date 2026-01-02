#!/bin/bash
# Hook script for titanoboa - runs in container rootfs before squashing
# Remove Hyprland session so GNOME is detected for live session

set -x

# Remove hyprland.desktop so titanoboa finds gnome.desktop for livesys
rm -f /usr/share/wayland-sessions/hyprland.desktop

# Verify GNOME session exists
ls -la /usr/share/wayland-sessions/
