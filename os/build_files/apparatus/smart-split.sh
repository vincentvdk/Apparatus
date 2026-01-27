#!/bin/bash
# Smart split for kitty terminal
# Detects if inside distrobox and enters the same container in new split

# Query kitty for the foreground process of the neighboring (source) window
# then read CONTAINER_ID from that process's /proc/PID/environ
if command -v kitty &>/dev/null && command -v jq &>/dev/null; then
    # Get the foreground PID from the most recently active non-focused window in same tab
    source_pid=$(kitty @ ls 2>/dev/null | jq -r '
        [.[].tabs[] | select(.is_focused == true) | .windows[]] |
        map(select(.is_focused == false)) |
        .[0].foreground_processes[0].pid // empty
    ' 2>/dev/null)

    if [ -n "$source_pid" ] && [ -f "/proc/$source_pid/environ" ]; then
        # Read CONTAINER_ID from the process environment
        container_id=$(tr '\0' '\n' < "/proc/$source_pid/environ" 2>/dev/null | \
            grep '^CONTAINER_ID=' | cut -d= -f2-)

        if [ -n "$container_id" ]; then
            exec distrobox enter "$container_id"
        fi
    fi
fi

# Not in a distrobox or couldn't detect, just start a normal shell
exec "$SHELL"
