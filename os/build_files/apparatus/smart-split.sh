#!/bin/bash
# Smart split for kitty terminal
# Detects if inside distrobox and enters the same container in new split

# Query kitty for the source window and check if it's running distrobox
if command -v kitty &>/dev/null && command -v jq &>/dev/null; then
    kitty_ls=$(kitty @ ls 2>/dev/null)

    # Get the container name from "distrobox enter <name>" in foreground_processes cmdline
    # Look at non-focused windows in the current tab
    container_id=$(echo "$kitty_ls" | jq -r '
        [.[].tabs[] | select(.is_focused == true) | .windows[]] |
        map(select(.is_focused == false)) |
        .[0].foreground_processes[] |
        .cmdline | join(" ") |
        capture("distrobox enter (?<name>[^ ]+)") |
        .name // empty
    ' 2>/dev/null | head -1)

    if [ -n "$container_id" ]; then
        exec distrobox enter "$container_id"
    fi
fi

# Not in a distrobox or couldn't detect, just start a normal shell
exec "$SHELL"
