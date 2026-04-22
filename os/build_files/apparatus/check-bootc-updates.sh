#!/bin/bash
# Check for bootc updates and notify user without auto-applying

# Staged status file to avoid duplicate notifications
STATUS_FILE="/run/apparatus/bootc-update-available"
mkdir -p /run/apparatus

# Check if bootc has updates staged
if bootc status --output json 2>/dev/null | grep -q '"Staged": true'; then
    # Only notify if we haven't already
    if [ ! -f "$STATUS_FILE" ]; then
        notify-send -a "Apparatus" "OS Update Available" \
            "A new OS update is ready to install.\n\nRun 'bootc upgrade --apply' to update." \
            -i system-software-update
        touch "$STATUS_FILE"
    fi
else
    # Updates applied, clear notification flag
    rm -f "$STATUS_FILE"
fi
