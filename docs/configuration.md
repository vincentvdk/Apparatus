# Configuration Guide

Apparatus OS provides extensive configuration options for customizing your desktop experience. This guide covers how to configure the various components of your system.

## Configuration Overview

Apparatus OS uses a two-level configuration system:

### System-Level Configuration

Stored in `/usr/share/apparatus/`:
- Default configurations for all applications
- Theme definitions
- System-wide settings

These files are managed by the Apparatus OS build process and updates.

### User-Level Configuration

Stored in `~/.config/`:
- Your personal customizations
- Theme symlinks (pointing to system themes)
- Application-specific overrides

## Configuration Files Structure

```
.config/
├── apparatus/                    # Apparatus-specific
│   ├── current-theme             # Currently active theme name
│   ├── current-font              # Currently active font name
│   └── skip-distroboxes          # Distroboxes to skip during sync
├── hypr/                        # Hyprland window manager
│   ├── hyprland.lua              # Main configuration
│   └── theme.conf                # Theme overrides (symlink)
├── kitty/                       # Kitty terminal
│   ├── kitty.conf                # Main configuration
│   └── theme.conf                # Theme (symlink)
├── waybar/                      # Waybar status bar
│   ├── config                    # Main configuration
│   ├── style.css                # Custom styles
│   └── theme.css                # Theme (symlink)
├── mako/                        # Mako notifications
│   └── config                    # Configuration (symlink)
├── satty/                       # Satty screenshot tool
│   ├── config.toml               # Main configuration
│   └── overrides.css             # Theme overrides (symlink)
├── uwsm/                        # UWSM window management
│   └── env                        # Environment configuration
├── walker/                      # Walker application launcher
│   └── config.toml               # Configuration
└── nvim/                        # Neovim editor
    ├── init.lua                  # Main configuration
    └── lsp/                      # LSP server configurations
```

## Using Butler for Configuration

Butler is the recommended way to manage your configuration. It provides both a TUI and CLI interface.

### Theme Configuration

```bash
# Switch to dark theme
butler theme catppuccin-mocha

# Switch to light theme
butler theme catppuccin-latte

# Via TUI
butler
# Select "Theme" from the menu
```

### Font Configuration

```bash
# Switch to Ioskeley Mono
butler font ioskeley-mono

# Switch to JetBrains Mono
butler font jetbrains-mono

# Switch to Hack Nerd Font
butler font hack-nerd-font
```

### Config Sync

Sync configurations from host to all distroboxes:

```bash
# Sync all
butler sync

# Sync specific distrobox
butler sync mybox
```

## Manual Configuration

### Hyprland

The main Hyprland configuration file is at `~/.config/hypr/hyprland.lua`:

```lua
-- Example: Change default terminal
$terminal = kitty

-- Example: Change mod key
$mod = SUPER

-- Example: Add keybinding
bind = $mod, P, exec, walker

-- Example: Set window rules
windowrulev2 = float, class:^(pavucontrol)$
```

After making changes, reload Hyprland:

```bash
hyprctl reload
```

Or use the keybinding: Super+Shift+R

### Kitty Terminal

Edit `~/.config/kitty/kitty.conf`:

```ini
# Font family
font_family IoskeleyMono Nerd Font

# Font size
font_size 12.0

# Window padding
window_padding_width 1.0

# Color scheme (included from theme)
include theme.conf
```

Reload Kitty after changes:

```bash
pkill -SIGUSR1 kitty
```

### Waybar

Waybar configuration consists of multiple files:

- **`~/.config/waybar/config`**: JSON configuration defining modules and layout
- **`~/.config/waybar/style.css`**: Custom CSS styles
- **`~/.config/waybar/theme.css`**: Theme-specific styles (symlink to system theme)

Edit the config file to change modules:

```json
{
  "modules-left": ["workspaces", "bluetooth", "network"],
  "modules-center": ["date", "time"],
  "modules-right": ["tray", "clock", "battery"]
}
```

Reload Waybar after changes:

```bash
pkill waybar
waybar &
```

### Mako Notifications

Edit `~/.config/mako/config`:

```ini
# Position
position=top-right

# Width
width=400

# Height
height=100
```

### Satty Screenshot Tool

Edit `~/.config/satty/config.toml`:

```toml
# Output directory
output_dir = "~/Pictures/Screenshots"

# File format
format = "png"

# Include cursor
include_cursor = true
```

### Neovim

The Neovim configuration is Lua-based. Edit `~/.config/nvim/init.lua`:

```lua
-- Basic settings
vim.opt.number = true
vim.opt.relativenumber = true
vim.opt.tabstop = 4
vim.opt.shiftwidth = 4

-- Plugins loaded via lazy.nvim
require('plugins')
```

### Zsh

Edit `~/.config/zsh/.zshrc`:

```bash
# Aliases
alias ll='exa -la --git'
alias grep='grep --color=auto'

# Environment variables
export EDITOR=nvim
```

## Theme System

Apparatus OS uses the Catppuccin color palette with two variants:

### Available Themes

| Theme | Type | Description |
|-------|------|-------------|
| catppuccin-mocha | Dark | Dark background, high contrast |
| catppuccin-latte | Light | Light background, soft contrast |

### Theme Files

Themes are stored in `/usr/share/apparatus/themes/`:

```
/usr/share/apparatus/themes/
├── catppuccin-mocha/
│   ├── kitty.conf
│   ├── waybar.css
│   ├── mako.conf
│   ├── hyprland.conf
│   └── satty/
│       └── overrides.css
└── catppuccin-latte/
    ├── kitty.conf
    ├── waybar.css
    ├── mako.conf
    ├── hyprland.conf
    └── satty/
        └── overrides.css
```

### Theme Application

Butler creates symlinks from your config directory to the theme files:

```bash
# These are symlinks
~/.config/kitty/theme.conf -> /usr/share/apparatus/themes/catppuccin-mocha/kitty.conf
~/.config/waybar/theme.css -> /usr/share/apparatus/themes/catppuccin-mocha/waybar.css
~/.config/mako/config -> /usr/share/apparatus/themes/catppuccin-mocha/mako.conf
```

### GTK Theme

Butler also sets the GTK theme preference:

```bash
# For dark theme
gsettings set org.gnome.desktop.interface color-scheme 'prefer-dark'

# For light theme
gsettings set org.gnome.desktop.interface color-scheme 'prefer-light'
```

## Font Configuration

Apparatus OS includes three Nerd Font variants:

| Font | Description |
|------|-------------|
| JetBrainsMono Nerd Font | Default, clean and professional |
| IoskeleyMono Nerd Font | Berkeley Mono alternative |
| Hack Nerd Font | Classic with icons |

Butler updates font settings in:

- **Kitty**: `font_family` in kitty.conf
- **Waybar**: Font references in style.css
- **Mako**: Font references in config

## Creating Custom Configurations

### For Host System

1. Edit files in `~/.config/` directly
2. Or copy from `/usr/share/apparatus/` and customize:

```bash
cp -r /usr/share/apparatus/hypr ~/.config/
# Edit ~/.config/hypr/hyprland.lua
```

### For Distroboxes

1. Sync from host:
   ```bash
   butler sync
   ```

2. Or edit directly in the distrobox:
   ```bash
   distrobox enter mybox
   nvim ~/.config/hypr/hyprland.lua
   ```

3. Add to skip list if you want to maintain custom configs:
   ```bash
   echo "mybox" >> ~/.config/apparatus/skip-distroboxes
   ```

## Configuration Backup

### Backup Your Configs

```bash
# Create backup directory
mkdir ~/config-backups

# Backup all configs
tar -czf ~/config-backups/apparatus-configs-$(date +%Y%m%d).tar.gz ~/.config

# Or backup specific configs
tar -czf ~/config-backups/hypr-backup.tar.gz ~/.config/hypr
```

### Restore Your Configs

```bash
# Extract backup
tar -xzf ~/config-backups/apparatus-configs.tar.gz -C ~

# Reload services
hyprctl reload
pkill -SIGUSR1 kitty
makoctl reload
```

## Version Control for Configs

For advanced users, consider using Git to track configuration changes:

```bash
# Initialize git in config directory
cd ~/.config
git init

# Add all configs
git add .

# Commit changes
git commit -m "My custom configurations"

# Push to remote (optional)
git remote add origin git@github.com:yourname/dotfiles.git
git push -u origin main
```

## Resetting to Defaults

To restore default configurations:

```bash
# Remove your custom configs
rm -rf ~/.config/hypr ~/.config/kitty ~/.config/waybar ~/.config/mako

# Copy defaults from system
cp -r /usr/share/apparatus/hypr ~/.config/
cp -r /usr/share/apparatus/kitty ~/.config/
cp -r /usr/share/apparatus/waybar ~/.config/
cp -r /usr/share/apparatus/mako ~/.config/

# Apply default theme
butler theme catppuccin-mocha
```

## Advanced Configuration

### Hyprland Monitors

Configure monitors via `hyprctl` or the TUI:

```bash
# Launch dynamic monitor configuration
butler
# Select "Configure" > "Monitors"

# Or use hyprdynamicmonitors
hyprdynamicmonitors tui
```

### Hyprland Window Rules

Add window-specific rules in `~/.config/hypr/hyprland.lua`:

```lua
-- Float specific applications
windowrulev2 = float, class:^(pavucontrol)$
windowrulev2 = float, title:^(File Operation)$

-- Move specific apps to specific workspaces
windowrulev2 = move center, class:^(firefox)$
windowrulev2 = move center, class:^(.*)$  -- default for others

-- Size rules
windowrulev2 = size 80% 60%, class:^(pavucontrol)$
```

### Kitty Themes

Customize Kitty appearance in `~/.config/kitty/kitty.conf`:

```ini
# Enable opacity
background_opacity 0.95

# Custom colors
background #1e1e2e
foreground #cdd6f4

# Cursor style
cursor_shape beam
cursor_blink_rate 0.5
```

### Waybar Customization

Add custom modules to Waybar in `~/.config/waybar/config`:

```json
{
  "modules-left": ["workspaces", "bluetooth", "network", "memory"],
  "modules-center": ["date", "time"],
  "modules-right": ["tray", "cpu", "temperature", "battery", "clock"]
}
```

Add custom CSS in `~/.config/waybar/style.css`:

```css
/* Custom workspace colors */
.workspace.active {
    background-color: #cba6f7;
    color: #1e1e2e;
}
```

---

*For more information, see the [Butler Documentation](butler.md) for configuration management.*
