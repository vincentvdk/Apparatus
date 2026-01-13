export PATH=$PATH:/opt/bin
export ZDOTDIR=$HOME/.config/zsh
export TMUX_TMPDIR=/tmp
export XDG_CONFIG_HOME="$HOME/.config"

# mise shims for tool binaries (full activation happens in shell rc)
export PATH="$HOME/.local/share/mise/shims:$PATH"

# Set zsh as default shell if not already set
if command -v zsh >/dev/null 2>&1 && [ "$SHELL" != "/bin/zsh" ] && [ "$SHELL" != "/usr/bin/zsh" ]; then
    chsh -s /bin/zsh "$USER" 2>/dev/null
    export SHELL=/bin/zsh
    exec zsh
fi
