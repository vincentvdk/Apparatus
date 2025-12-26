#!/usr/bin/env bash
# Apparatus First Login
# Runs butler setup on first login

INIT_DONE="$HOME/.config/apparatus/init-done"

# Exit if already initialized
if [ -f "$INIT_DONE" ]; then
    exit 0
fi

# Run butler init in a terminal
kitty -e butler init

# Reload Hyprland if init completed successfully
if [ -f "$INIT_DONE" ]; then
    hyprctl reload
fi
