# Daily Usage Guide

This guide covers common tasks and workflows in Apparatus OS.

## Window Management

### Basic Window Operations

| Action | Shortcut |
|--------|----------|
| Open terminal | Super+Return |
| Open launcher | Super+Space |
| Open file manager | Super+E |
| Close window | Super+W |
| Toggle fullscreen | Super+F |
| Toggle floating | Super+V |
| Move focus | Super+H/J/K/L or Arrow keys |
| Move window | Super+Shift+H/J/K/L |

### Workspaces

| Action | Shortcut |
|--------|----------|
| Switch to workspace 1-10 | Super+1-0 |
| Move window to workspace 1-10 | Super+Shift+1-0 |
| Cycle workspaces | Super+Mouse Scroll |

### Scratchpad

The scratchpad is a hidden workspace for temporary windows:

| Action | Shortcut |
|--------|----------|
| Toggle scratchpad | Super+` (backtick) |
| Move window to scratchpad | Super+Shift+` |

### Floating Windows

Some applications work better as floating windows:

```bash
# Make a window floating
# Click and drag the window while holding Super

# Or via command (for specific apps)
# Edit ~/.config/hypr/hyprland.lua and add:
#   float = class:^(kitty)$
```

## Applications

### Terminal (Kitty)

Kitty is the default terminal with tmux-compatible keybindings:

| Action | Shortcut |
|--------|----------|
| New tab | Ctrl+A > c |
| Next tab | Ctrl+A > n |
| Previous tab | Ctrl+A > p |
| Switch to tab 1-9 | Ctrl+A > 1-9 |
| Vertical split | Ctrl+A > v |
| Horizontal split | Ctrl+A > s |
| Close split | Ctrl+A > x |
| Resize mode | Ctrl+A > r |

### File Manager (Thunar)

Thunar provides a simple GTK-based file manager:

- Navigate with mouse or arrow keys
- Open files with Enter or double-click
- Use Ctrl+C/Ctrl+V for copy/paste

### Screenshots (Satty)

Satty provides screenshot and annotation capabilities:

| Action | Shortcut |
|--------|----------|
| Full screen screenshot | Print |
| Active window screenshot | Shift+Print |
| Region screenshot | Super+Shift+P |

Screenshots are automatically opened in Satty for annotation and saving.

## System Controls

### Volume

| Action | Shortcut |
|--------|----------|
| Increase volume | Volume Up key |
| Decrease volume | Volume Down key |
| Mute | Mute key |

### Brightness

| Action | Shortcut |
|--------|----------|
| Increase brightness | Brightness Up key |
| Decrease brightness | Brightness Down key |

### Media

| Action | Shortcut |
|--------|----------|
| Play/Pause | Media Play/Pause key |
| Next track | Media Next key |
| Previous track | Media Previous key |

### Locking and Sessions

| Action | Shortcut |
|--------|----------|
| Lock screen | Super+Escape |
| Log out / Exit Hyprland | Super+Shift+Escape |
| Show help | Super+F1 |

## Distrobox Integration

Distrobox containers integrate seamlessly with the host:

```bash
# Create a new distrobox
butler  # Then use Distrobox menu

# Or via command line
distrobox create -n mydev -i ghcr.io/vincentvdk/apparatus-box:latest

# Enter a distrobox
distrobox enter mydev

# List all distroboxes
distrobox list
```

See [Distrobox Documentation](distrobox.md) for more details.

## Butler Configuration Tool

Butler is your central hub for system configuration:

```bash
# Launch interactive TUI
butler

# Theme management
butler theme catppuccin-mocha    # Dark theme
butler theme catppuccin-latte    # Light theme

# Font management
butler font jetbrains-mono
butler font hack-nerd-font
butler font ioskeley-mono

# Sync configs to distroboxes
butler sync

# Check version
butler version
```

## File Management

### Quick Access

| Directory | Purpose |
|-----------|---------|
| ~/Downloads | Downloaded files |
| ~/Documents | Personal documents |
| ~/Pictures | Images and screenshots |
| ~/.config | User configurations |
| /usr/share/apparatus | System configurations |

### Common Commands

```bash
# Navigate to home
cd ~

# List files
exa  # or ls -la

# Create directory
mkdir myproject

# Copy files
cp file.txt ~/backup/

# Edit files
nvim file.txt  # or kitty + nvim
```

## Development Workflow

### Using Distrobox for Development

```bash
# Create a development container
butler  # Use Distrobox > Create menu

# Enter the container
distrobox enter mydev

# Install development tools inside container
sudo dnf install git curl nodejs python3

# Your home directory is shared (if using shared home)
# Or use custom home for isolation
```

### Version Control

```bash
# Initialize git repository
cd myproject
git init

# Clone a repository
git clone https://github.com/user/repo.git

# Stage and commit
git add .
git commit -m "Initial commit"

# Push to remote
git push origin main
```

## Networking

### Checking Connection

```bash
# Check IP address
ip a

# Check internet connectivity
ping google.com

# Check active connections
ss -tulnp
```

### Sharing Files

```bash
# Start a simple HTTP server in current directory
python3 -m http.server 8000

# Then access from another machine: http://<your-ip>:8000
```

## Power Management

### Sleep/Suspend

```bash
# Suspend
systemctl suspend

# Hibernate (if configured)
systemctl hibernate
```

### Reboot/Shutdown

```bash
# Reboot
systemctl reboot

# Shutdown
systemctl poweroff
```

## Tips and Tricks

1. **Window Resizing**: Use Super+Ctrl+H/J/K/L to resize windows
2. **Quick Terminal**: Super+Return from anywhere
3. **Quick Launcher**: Super+Space to launch any application
4. **Screenshot**: Print key for full screen, Super+Shift+P for region
5. **Clipboard**: Use Ctrl+C/Ctrl+V or middle-click to paste
6. **Workspaces**: Use Super+1-0 to switch, Super+Shift+1-0 to move windows

---

*For more advanced usage, see the other documentation guides.*
