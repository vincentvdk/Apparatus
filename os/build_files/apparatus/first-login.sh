#!/usr/bin/env bash
# Apparatus First Login Check
# Shows a setup prompt if initial configuration hasn't been run

INIT_DONE="$HOME/.config/apparatus/init-done"

# Exit if already initialized
if [ -f "$INIT_DONE" ]; then
    exit 0
fi

# Show welcome dialog
zenity --question \
    --title="Welcome to Apparatus" \
    --text="It looks like this is your first login.\n\nWould you like to run the initial setup?\n\nThis will:\n• Configure Hyprland, Waybar, and Mako\n• Enable Flathub repository\n• Install recommended applications" \
    --ok-label="Run Setup" \
    --cancel-label="Later" \
    --width=400

if [ $? -eq 0 ]; then
    # User clicked "Run Setup" - run butler init
    foot -e butler

    # Check if init completed successfully
    if [ -f "$INIT_DONE" ]; then
        zenity --question \
            --title="Setup Complete" \
            --text="Initial setup is complete.\n\nWould you like to reload Hyprland now to apply the new configuration?" \
            --ok-label="Reload Now" \
            --cancel-label="Later" \
            --width=350

        if [ $? -eq 0 ]; then
            hyprctl reload
        fi
    fi
fi
