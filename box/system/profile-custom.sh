export PATH="$HOME/.local/bin:$PATH:/opt/bin"
export TMUX_TMPDIR=/tmp

# Detect if using shared home with host
if [ -n "$DISTROBOX_HOST_HOME" ] && [ "$HOME" = "$DISTROBOX_HOST_HOME" ]; then
    export APPARATUS_OS_HOME=1
    # Use container-specific paths to keep shared home clean
    export APPARATUS_CONFIG_HOME="$HOME/.config/apparatus-box"
    export XDG_CONFIG_HOME="$APPARATUS_CONFIG_HOME"
    export ZDOTDIR="$APPARATUS_CONFIG_HOME/zsh"
    export GIT_CONFIG_GLOBAL="$APPARATUS_CONFIG_HOME/git/config"
else
    # Custom home - use standard paths
    export XDG_CONFIG_HOME="$HOME/.config"
    export ZDOTDIR="$HOME/.config/zsh"
    export APPARATUS_OS_HOME=0
fi

# Auto-run init on first login
INIT_MARKER="$HOME/.local/state/apparatus/box-init-done"
if [ "$APPARATUS_OS_HOME" = "1" ] || [ ! -f "$INIT_MARKER" ]; then
    /opt/apparatus/init.sh
fi

# mise shims for tool binaries (full activation happens in shell rc)
export PATH="$HOME/.local/share/mise/shims:$PATH"

# Add neovim to the path
export PATH="$PATH:/opt/nvim-linux-x86_64/bin"

# Set zsh as default shell if not already set
if command -v zsh >/dev/null 2>&1 && [ "$SHELL" != "/usr/bin/zsh" ]; then
    sudo chsh -s /usr/bin/zsh "$USER" >/dev/null 2>&1
    export SHELL=/usr/bin/zsh
    exec zsh
fi
