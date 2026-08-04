# Getting Started

Welcome to Apparatus OS! This guide will help you get oriented with your new system.

## First Login

When you log in for the first time:

1. **greetd** displays a graphical login screen (gtkgreet)
2. Select your user and enter your password
3. Hyprland starts with the default configuration

## Desktop Environment

### Window Management (Hyprland)

Hyprland is a dynamic tiling Wayland compositor. Key concepts:

- **Workspaces**: Virtual desktops (1-10 by default)
- **Windows**: Tiled automatically, can be made floating
- **Monitors**: Multi-monitor support with individual configurations

### Default Applications

| Category | Application | Launch Key |
|----------|-------------|------------|
| Terminal | Kitty | Super+Return |
| Launcher | Walker | Super+Space |
| File Manager | Thunar | Super+E |
| Lock Screen | hyprlock | Super+Escape |
| Screenshot | Satty | Print, Shift+Print, Super+Shift+P |

## Configuration Files

Your personal configuration is stored in `~/.config/`:

```
.config/
├── apparatus/          # Apparatus-specific configs
│   ├── current-theme   # Currently active theme
│   └── current-font    # Currently active font
├── hypr/              # Hyprland configuration
│   ├── hyprland.lua    # Main config
│   └── theme.conf     # Theme overrides (symlink)
├── kitty/             # Terminal configuration
│   ├── kitty.conf      # Main config
│   └── theme.conf     # Theme (symlink)
├── waybar/            # Status bar
│   ├── config          # Main config
│   └── theme.css      # Theme (symlink)
├── mako/              # Notifications
│   └── config          # Config (symlink)
└── satty/             # Screenshot tool
    └── overrides.css   # Theme (symlink)
```

System-wide configurations are in `/usr/share/apparatus/`.

## Using Butler

**Butler** is the main configuration tool for Apparatus OS. Launch it from a terminal:

```bash
butler
```

Or access specific features directly:

```bash
butler theme catppuccin-latte   # Switch to light theme
butler font jetbrains-mono     # Change font
butler sync                    # Sync configs to distroboxes
```

See [Butler Documentation](butler.md) for full details.

## Connecting to Networks

### WiFi

1. Click the network icon in the waybar (status bar)
2. Select your WiFi network
3. Enter password if required

### Ethernet

Ethernet connections are automatic via NetworkManager.

### VPN

Install and configure your preferred VPN client:

```bash
# Install WireGuard
sudo dnf install wireguard-tools

# Or OpenVPN
sudo dnf install openvpn
```

## Audio

Audio is managed by PipeWire:

```bash
# Launch volume control GUI
pavucontrol

# Command line volume control
pactl set-sink-volume @DEFAULT_SINK@ +5%   # Increase
pactl set-sink-volume @DEFAULT_SINK@ -5%   # Decrease
pactl set-sink-mute @DEFAULT_SINK@ toggle   # Toggle mute
```

## Bluetooth

```bash
# Launch Bluetooth manager GUI
blueman-manager

# Command line (if needed)
bluetoothctl
```

## Next Steps

- [Usage Guide](usage.md) - Learn daily usage patterns
- [Configuration](configuration.md) - Customize your system
- [Butler](butler.md) - Manage your configuration
- [Distrobox](distrobox.md) - Set up development containers

---

*Tip: Press Super+F1 to see keyboard shortcuts at any time.*
