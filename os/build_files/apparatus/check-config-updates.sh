#!/bin/bash
REPO_URL_FILE="$HOME/.config/apparatus/chezmoi-dotfiles"
DEFAULT_REPO="https://github.com/vincentvdk/apparatus-dotfiles.git"
STATE_DIR="$HOME/.local/state/apparatus"
HASH_FILE="$STATE_DIR/config-update-hash"

# Determine repo URL
if [ -f "$REPO_URL_FILE" ]; then
    REPO=$(cat "$REPO_URL_FILE")
    [[ "$REPO" != http* ]] && REPO="https://github.com/$REPO.git"
else
    REPO="$DEFAULT_REPO"
fi

# Get latest remote HEAD hash (lightweight, no fetch/clone)
REMOTE_HASH=$(git ls-remote "$REPO" HEAD 2>/dev/null | cut -f1)
[ -z "$REMOTE_HASH" ] && exit 0

mkdir -p "$STATE_DIR"

# Compare with stored hash (first run just stores, no notification)
if [ -f "$HASH_FILE" ]; then
    LOCAL_HASH=$(cat "$HASH_FILE")
    if [ "$REMOTE_HASH" != "$LOCAL_HASH" ]; then
        notify-send -a "Apparatus" "Config Updates Available" \
            "New configuration updates are available.\nRun butler to apply." \
            -i dialog-information
    fi
fi

echo "$REMOTE_HASH" > "$HASH_FILE"
