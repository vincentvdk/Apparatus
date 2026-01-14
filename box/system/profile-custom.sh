export PATH=$PATH:/opt/bin
export TMUX_TMPDIR=/tmp

# Detect if using shared home with host
# If shared, check user preference for config handling
if [ -n "$DISTROBOX_HOST_HOME" ] && [ "$HOME" = "$DISTROBOX_HOST_HOME" ]; then
    export APPARATUS_SHARED_HOME=1

    # Check if user has chosen a config mode
    CONFIG_MODE_FILE="${HOME}/.apparatus-config-mode"
    if [ -f "$CONFIG_MODE_FILE" ]; then
        APPARATUS_CONFIG_MODE=$(cat "$CONFIG_MODE_FILE")
    fi

    if [ "$APPARATUS_CONFIG_MODE" = "host" ]; then
        # User chose to use host configs - don't override anything
        export XDG_CONFIG_HOME="${XDG_CONFIG_HOME:-$HOME/.config}"
        export ZDOTDIR="${ZDOTDIR:-$HOME/.config/zsh}"
    else
        # Use container-specific paths (default for shared home)
        export APPARATUS_CONFIG_HOME="$HOME/.config/apparatus-box"
        export XDG_CONFIG_HOME="$APPARATUS_CONFIG_HOME"
        export ZDOTDIR="$APPARATUS_CONFIG_HOME/zsh"
        export GIT_CONFIG_GLOBAL="$APPARATUS_CONFIG_HOME/git/config"
    fi
else
    # Custom home - use standard paths
    export XDG_CONFIG_HOME="$HOME/.config"
    export ZDOTDIR="$HOME/.config/zsh"
    export APPARATUS_SHARED_HOME=0
fi

# mise shims for tool binaries (full activation happens in shell rc)
export PATH="$HOME/.local/share/mise/shims:$PATH"

# Set zsh as default shell if not already set
if command -v zsh >/dev/null 2>&1 && [ "$SHELL" != "/bin/zsh" ] && [ "$SHELL" != "/usr/bin/zsh" ]; then
    chsh -s /bin/zsh "$USER" 2>/dev/null
    export SHELL=/bin/zsh
    exec zsh
fi
