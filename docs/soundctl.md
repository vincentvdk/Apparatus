# Soundctl - Audio Configuration TUI

Soundctl is a Terminal User Interface (TUI) tool for managing audio settings on Linux systems using PulseAudio or PipeWire. It provides an intuitive interface for controlling audio devices without requiring a GUI.

## Overview

Soundctl allows you to:
- View all available audio input and output devices
- Switch between devices by setting them as default
- Adjust volume levels with fine-grained control
- Toggle mute on/off for any device
- Refresh device list to detect changes
- Navigate with keyboard shortcuts

## Installation

### From Source

Soundctl is included in the Apparatus OS repository. To install it:

```bash
# Navigate to the soundctl directory
cd /path/to/apparatus/tools/soundctl/cmd/soundctl

# Build and install
go install ./...

# Or build manually
go build -o soundctl .
# Then copy to a directory in your PATH
cp soundctl ~/.local/bin/
```

### Dependencies

Soundctl requires:
- **Go 1.22** or higher
- **PulseAudio** or **PipeWire** with `pactl` command
- A terminal emulator that supports ANSI colors

Check dependencies:
```bash
# Check Go version
go version

# Check for pactl (PulseAudio/PipeWire)
which pactl
pactl --version
```

## Usage

### Launching Soundctl

```bash
# Run the TUI
soundctl

# Or run via butler
butler
# Select "Configure" > "Audio"
```

### Command Line Interface

Soundctl is primarily a TUI application. For CLI usage, use `pactl` directly:

```bash
# List all sinks (output devices)
pactl list sinks short

# List all sources (input devices)
pactl list sources short

# Set default sink
pactl set-default-sink <index or name>

# Set volume to 50%
pactl set-sink-volume <index> 50%

# Toggle mute
pactl set-sink-mute <index> toggle

# Get current default sink
pactl get-default-sink
```

## User Interface

### Main Menu

When you launch soundctl, you'll see the main menu with these options:

- **🔊 Output Devices** - View and manage playback devices (speakers, headphones)
- **🎤 Input Devices** - View and manage recording devices (microphones)
- **⚙️ Settings** - View keyboard shortcuts and help
- **🔄 Refresh** - Reload device list

Navigate with arrow keys and press `enter` to select.

### Output Devices Panel

The output devices panel shows:
- Device index and name
- Description (if available)
- Current volume percentage
- Mute status (🔇 when muted)
- Default device indicator

**Actions:**
- Navigate with `↑`/`↓` arrow keys
- Press `enter` to adjust volume with a visual slider
- Press `+` or `=` to increase volume by 5%
- Press `-` to decrease volume by 5%
- Press `m` to toggle mute
- Press `d` to set as default output device

### Input Devices Panel

Similar to output devices, but for recording devices (microphones, line-in):
- Same navigation and volume controls
- Press `d` to set as default input device

### Volume Adjustment Mode

When you press `enter` on a device, you enter volume adjustment mode:
- A visual volume bar is displayed
- Use `←`/`→` or `h`/`l` to fine-tune the volume
- Press `enter` to confirm
- Press `esc` to cancel

### Settings Panel

The settings panel displays all available keyboard shortcuts for reference.

## Keyboard Shortcuts

### Navigation

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate up/down in list |
| `tab` | Switch to next panel |
| `shift+tab` | Switch to previous panel |
| `enter` | Select item / adjust volume |
| `esc` | Return to main menu / cancel |

### Device Control

| Key | Action |
|-----|--------|
| `+` / `=` | Increase volume by 5% |
| `-` | Decrease volume by 5% |
| `m` | Toggle mute on/off |
| `d` | Set as default device |
| `r` | Refresh device list |

### General

| Key | Action |
|-----|--------|
| `q` | Quit soundctl |
| `ctrl+c` | Quit soundctl |

## Device Information

Soundctl displays several types of information for each device:

- **Index**: The numerical identifier used by PulseAudio/PipeWire
- **Name**: The device identifier (e.g., `alsa_output.pci-0000_00_1f.3.analog-stereo`)
- **Description**: Human-readable name (e.g., "Built-in Audio Analog Stereo")
- **Volume**: Current volume percentage (0-100%)
- **Muted**: Whether the device is currently muted
- **Default**: Whether this is the default device for its type

## Use Cases

### Switching Audio Output

1. Launch soundctl
2. Select "Output Devices" from main menu
3. Navigate to the device you want to use
4. Press `d` to set it as default

### Adjusting Volume

1. Launch soundctl
2. Select "Output Devices" or "Input Devices"
3. Navigate to the device
4. Press `+` or `-` to adjust volume in 5% increments
5. Or press `enter` and use `←`/`→` for fine-grained control

### Muting/Unmuting

1. Launch soundctl
2. Navigate to the device
3. Press `m` to toggle mute

### Quick Volume Adjustment

For quick adjustments without launching the TUI:

```bash
# Get current sink index
pactl get-default-sink

# Set volume to 75%
pactl set-sink-volume @DEFAULT_SINK@ 75%

# Increase by 10%
pactl set-sink-volume @DEFAULT_SINK@ +10%

# Decrease by 10%
pactl set-sink-volume @DEFAULT_SINK@ -10%
```

## Troubleshooting

### No Devices Found

If soundctl shows no devices:

```bash
# Check if PulseAudio/PipeWire is running
pactl info

# Restart PulseAudio
pulseaudio -k && pulseaudio --start

# Restart PipeWire
systemctl --user restart pipewire
```

### pactl Command Not Found

Install PulseAudio or PipeWire:

**On Fedora/RHEL:**
```bash
sudo dnf install pulseaudio pulsaudio-utils
# or for PipeWire
sudo dnf install pipewire pipewire-pulseaudio
```

**On Debian/Ubuntu:**
```bash
sudo apt install pulseaudio pavucontrol
# or for PipeWire
sudo apt install pipewire pipewire-pulse
```

**On Arch Linux:**
```bash
sudo pacman -S pulseaudio pavucontrol
# or for PipeWire
sudo pacman -S pipewire pipewire-pulse
```

### Permissions Issues

If you get permission errors:

```bash
# Ensure your user is in the audio group
sudo usermod -aG audio $USER

# Restart session or run
newgrp audio
```

### Device List Not Updating

Press `r` to refresh the device list, or restart soundctl.

### Volume Changes Not Working

Some applications may use their own volume control. Try:
```bash
# Check all sinks with detailed info
pactl list sinks

# Try setting volume on all sinks
for i in $(pactl list sinks short | awk '{print $1}'); do
  pactl set-sink-volume $i 50%
done
```

## Comparison with Other Tools

| Feature | soundctl | pavucontrol | alsamixer | amixer |
|---------|----------|-------------|-----------|--------|
| TUI Interface | ✅ | ❌ (GUI) | ✅ | ❌ (CLI) |
| Device Switching | ✅ | ✅ | ❌ | ❌ |
| Volume Control | ✅ | ✅ | ✅ | ✅ |
| Mute Toggle | ✅ | ✅ | ✅ | ✅ |
| Default Device | ✅ | ✅ | ❌ | ❌ |
| Per-App Volume | ❌ | ✅ | ❌ | ❌ |
| Keyboard Shortcuts | ✅ | ✅ | ✅ | ❌ |
| Works with Hyprland | ✅ | ✅ | ✅ | ✅ |
| No X11 Required | ✅ | ❌ | ✅ | ✅ |

## Integration with Hyprland

Soundctl works perfectly with Hyprland since both are terminal-based and don't require X11.

Add a keybinding to your `~/.config/hypr/hyprland.lua`:

```lua
-- Launch soundctl with Super+A
bind = $mod, A, exec, kitty --type normal soundctl

-- Or with a floating terminal
bind = $mod, A, exec, kitty --type overlay soundctl
```

## Configuration File

Soundctl doesn't have a configuration file yet. All settings are applied directly to the audio system.

## Future Enhancements

Potential features for future versions:
- Per-application volume control
- Audio profiles (switch between presets)
- Equalizer settings
- Balance control (left/right)
- Device renaming
- Persistent volume settings
- Config file for default preferences

## License

Soundctl is part of Apparatus OS and is released under the same license.

## Contributing

Contributions are welcome! See the main Apparatus OS repository for contribution guidelines.

## See Also

- [Butler Documentation](butler.md) - Configuration management tool
- [Configuration Guide](configuration.md) - Overall system configuration
- [pactl man page](https://www.freedesktop.org/wiki/Software/PulseAudio/Documentation/User/CLI/) - PulseAudio CLI documentation
- [PipeWire Documentation](https://pipewire.pages.freedesktop.org/wireplumber/) - PipeWire documentation
