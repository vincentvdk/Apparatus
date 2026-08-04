# Butler - Configuration Management Tool

Butler is the unified configuration management tool for Apparatus OS. It provides a Terminal User Interface (TUI) and Command Line Interface (CLI) for managing your system configuration, theme settings, fonts, and distrobox configurations.

## Overview

Butler runs in two modes:

- **OS Mode**: For managing the host system (Hyprland, themes, fonts, distroboxes)
- **Box Mode**: For managing development tools inside distrobox containers

Butler automatically detects which mode to use, or you can specify it explicitly.

## Launching Butler

### Interactive TUI

```bash
# Auto-detect mode
butler

# Force OS mode
butler os

# Force box mode
butler box
```

### Via Environment Variable

```bash
# Override mode via environment
APPARATUS_MODE=os butler
APPARATUS_MODE=box butler
```

## OS Mode Features

When running in OS mode, Butler provides these menu options:

| Menu Item | Description |
|-----------|-------------|
| 📦 Distrobox | Create, upgrade, and manage distrobox containers |
| ⚙️ Configure | Configure Hyprland settings, monitors, audio, AI workload |
| 🎨 Theme | Switch between Catppuccin themes |
| 🔤 Font | Change system font |
| 🔄 Sync Configs | Sync configurations to all distroboxes |
| 🚫 Skip List | Manage which distroboxes to skip during config sync |
| ⌨️ Help | View keyboard shortcuts |

## Box Mode Features

When running inside a distrobox container, Butler provides:

- Tool management by category
- Version management for each tool
- LSP (Language Server Protocol) configuration

## Command Line Interface

Butler provides CLI commands for scripting and quick access:

### Theme Management

```bash
# List available themes
butler theme

# Apply a theme
butler theme catppuccin-mocha    # Dark theme
butler theme catppuccin-latte    # Light theme

# Usage
Usage: butler theme <name>
Available themes: catppuccin-mocha, catppuccin-latte
```

### Font Management

```bash
# List available fonts
butler font

# Apply a font
butler font jetbrains-mono
butler font ioskeley-mono
butler font hack-nerd-font

# Usage
Usage: butler font <name>
Available fonts: ioskeley-mono, jetbrains-mono, hack-nerd-font
```

### Config Sync

```bash
# Sync configs to all distroboxes
butler sync

# Sync configs to a specific distrobox
butler sync mybox

# Expected output
Synced mybox
Synced dev

Summary:
  Synced: 2 distroboxes
    mybox, dev
  Failed: 0 distroboxes
```

## Distrobox Management

The Distrobox menu provides:

| Action | Description |
|--------|-------------|
| Create | Create a new distrobox container |
| Upgrade | Upgrade distrobox packages or image |
| List | View existing distroboxes |

### Creating a Distrobox

1. Select "Distrobox" from the main menu
2. Choose "Create"
3. Enter a name for your distrobox
4. Choose home directory option:
   - Default home (~) - Uses your user's home directory
   - Custom home folder - Creates a dedicated home directory
5. The distrobox will be created and ready to use

### Upgrading a Distrobox

1. Select "Distrobox" from the main menu
2. Choose "Upgrade"
3. Select the distrobox to upgrade
4. Choose upgrade type:
   - **packages**: Update packages inside the container
   - **image**: Recreate with latest image (preserves home directory)

## Configuration Sync

### What Gets Synced

Butler syncs these configurations from the host to distroboxes:

| Configuration | Description |
|--------------|-------------|
| hypr/ | Hyprland window manager configuration |
| kitty/ | Kitty terminal configuration |
| waybar/ | Waybar status bar configuration |
| mako/ | Mako notification configuration |
| walker/ | Walker application launcher configuration |
| uwsm/ | UWSM window management configuration |
| satty/ | Satty screenshot tool configuration |
| atuin/ | Atuin shell history configuration |
| nvim/ | Neovim editor configuration |
| zsh/ | Zsh shell configuration |
| themes/ | Color theme configurations |

### How Sync Works

1. **Source**: Configurations are sourced from `/usr/share/apparatus/` on the host
2. **Destination**: Configurations are copied to each distrobox's `~/.config/` directory
3. **Skip List**: Distroboxes in the skip list are not synced
4. **No Overwrite**: Existing files are replaced (full sync, not merge)

### Managing the Skip List

The skip list prevents specific distroboxes from receiving config updates:

```bash
# Launch Butler and navigate to Skip List
butler
# Select "🚫 Skip List" from the menu
# Toggle checkmarks to add/remove distroboxes from skip list
```

Or edit the skip list file directly:

```bash
# Edit the skip list
nvim ~/.config/apparatus/skip-distroboxes

# File format: one distrobox name per line
my-test-box
special-config-box

# Lines starting with # are comments
# This is a comment
```

## Theme Management

Apparatus OS includes two Catppuccin themes:

| Theme | Type | Best For |
|-------|------|----------|
| catppuccin-mocha | Dark | Night use, coding, dark environments |
| catppuccin-latte | Light | Day use, bright environments |

### Applying a Theme

```bash
# Via CLI
butler theme catppuccin-mocha

# Via TUI
butler
# Select "🎨 Theme" from the menu
# Choose your theme
```

### Theme Files

Themes include configurations for:

- Kitty terminal colors
- Waybar status bar styling
- Mako notification appearance
- Hyprland window decorations
- Satty screenshot tool UI

## Font Management

Apparatus OS supports three monospace fonts optimized for development:

| Font | Description |
|------|-------------|
| ioskeley-mono | Berkeley Mono alternative, clean and readable |
| jetbrains-mono | JetBrains' popular development font (default) |
| hack-nerd-font | Classic Hack with Nerd Font icons |

### Applying a Font

```bash
# Via CLI
butler font ioskeley-mono

# Via TUI
butler
# Select "🔤 Font" from the menu
# Choose your font
```

## Configuration Management

Butler allows configuration of various system aspects:

| Option | Description |
|--------|-------------|
| Terminal | Set default terminal emulator |
| Monitors | Configure multi-monitor setup via hyprdynamicmonitors |
| Audio | Launch PulseAudio Volume Control (pavucontrol) |
| AI Workload | Configure GPU VRAM allocation for AI/ML workloads |

## Help / Keyboard Shortcuts

Press `Super+F1` at any time to see the keyboard shortcuts overlay, or launch Butler and select "Help" to see all shortcuts in the TUI.

## Environment Detection

Butler automatically detects whether it's running in:

- **Host OS**: Directly on the Apparatus OS installation
- **Distrobox**: Inside a distrobox container

Detection is based on:
- Presence of `hyprctl` command
- Presence of `/run/.containerenv` file
- Presence of `mise` command (used in distroboxes)

You can override detection with:

```bash
APPARATUS_MODE=os butler    # Force OS mode
APPARATUS_MODE=box butler   # Force box mode
```

## Troubleshooting

### "Butler not found"

The golang butler is installed at `/usr/local/bin/butler`. Ensure this is in your PATH.

### "No distroboxes found"

Create a distrobox first using the Distrobox menu in Butler.

### Config sync fails

Ensure the distrobox is not running, or check the home directory permissions.

### Theme/Font not applying

Try reloading the affected services:

```bash
# Reload Hyprland
hyprctl reload

# Reload Kitty
pkill -SIGUSR1 kitty

# Reload Mako
makoctl reload
```

---

*For more information, see the main [Usage Guide](usage.md).*
