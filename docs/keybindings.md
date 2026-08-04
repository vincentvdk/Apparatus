# Apparatus OS Keybindings

Optimized for vim-style navigation and productivity.

## Philosophy

- **SUPER** = Window manager (Hyprland) primary modifier
- **Ctrl+A** = Terminal prefix (Kitty) - tmux-compatible
- **h/j/k/l** = vim navigation everywhere
- **SHIFT** = "move" modifier (window movement)
- **CTRL** = "resize" modifier (window resizing)
- **ALT** = Alternative modifier (window cycling)

## Hyprland

### Window Focus
| Key | Action |
|-----|--------|
| SUPER + H | Focus left |
| SUPER + J | Focus down |
| SUPER + K | Focus up |
| SUPER + L | Focus right |
| SUPER + Arrow keys | Focus (alternative directional) |

### Window Move
| Key | Action |
|-----|--------|
| SUPER + SHIFT + H | Move window left |
| SUPER + SHIFT + J | Move window down |
| SUPER + SHIFT + K | Move window up |
| SUPER + SHIFT + L | Move window right |
| SUPER + LMB (hold) | Drag window |

### Window Resize
| Key | Action |
|-----|--------|
| SUPER + CTRL + H | Resize left |
| SUPER + CTRL + J | Resize down |
| SUPER + CTRL + K | Resize up |
| SUPER + CTRL + L | Resize right |
| SUPER + RMB (hold) | Resize window |

### Window Actions
| Key | Action |
|-----|--------|
| SUPER + Q | Close window |
| SUPER + F | Fullscreen |
| SUPER + V | Toggle floating |
| SUPER + P | Toggle pseudo-tiling |
| SUPER + T | Toggle layout |

### Window Cycling
| Key | Action |
|-----|--------|
| SUPER + Tab | Cycle to next window |
| SUPER + SHIFT + Tab | Cycle to previous window |
| ALT + Tab | Cycle to next window (alternative) |
| ALT + SHIFT + Tab | Cycle to previous window (alternative) |

### Applications
| Key | Action |
|-----|--------|
| SUPER + Return | Terminal (Kitty) |
| SUPER + D | Launcher (walker) |
| SUPER + Space | Launcher (walker) - alias |
| SUPER + E | File manager (Thunar) |

### Session
| Key | Action |
|-----|--------|
| SUPER + L | Lock screen (hyprlock) |
| SUPER + SHIFT + Q | Exit Hyprland |

### Configuration
| Key | Action |
|-----|--------|
| SUPER + SHIFT + C | Reload Hyprland config |
| SUPER + F1 | Show keybinding help (butler help) |

### Workspaces
| Key | Action |
|-----|--------|
| SUPER + 1-0 | Switch to workspace 1-10 |
| SUPER + SHIFT + 1-0 | Move window to workspace 1-10 |
| SUPER + Mouse scroll | Cycle workspaces |

### Scratchpad
| Key | Action |
|-----|--------|
| SUPER + S | Toggle scratchpad (magic workspace) |
| SUPER + SHIFT + S | Move window to scratchpad |

### Screenshots
| Key | Action |
|-----|--------|
| Print | Screenshot full screen |
| SHIFT + Print | Screenshot window |
| SUPER + SHIFT + P | Screenshot region |

### Help
| Key | Action |
|-----|--------|
| SUPER + F1 | Show keybinding help |

## Kitty Terminal

All kitty keybindings use the **Ctrl+A** prefix (tmux-style).

### Splits
| Key | Action |
|-----|--------|
| Ctrl+A > v | Vertical split |
| Ctrl+A > s | Horizontal split |
| Ctrl+A > h | Focus split left |
| Ctrl+A > j | Focus split down |
| Ctrl+A > k | Focus split up |
| Ctrl+A > l | Focus split right |
| Ctrl+A > x | Close split |
| Ctrl+A > r | Resize mode |
| Ctrl+A > z | Zoom (toggle stack layout) |

### Tabs
| Key | Action |
|-----|--------|
| Ctrl+A > c | New tab |
| Ctrl+A > n | Next tab |
| Ctrl+A > p | Previous tab |
| Ctrl+A > 1-9 | Go to tab 1-9 |
| Ctrl+A > , | Rename tab |

### Scrollback
| Key | Action |
|-----|--------|
| Ctrl+A > [ | Enter scrollback mode |

## Media Keys

| Key | Action |
|-----|--------|
| XF86AudioRaiseVolume | Volume up |
| XF86AudioLowerVolume | Volume down |
| XF86AudioMute | Toggle mute |
| XF86AudioMicMute | Toggle mic mute |
| XF86MonBrightnessUp | Brightness up |
| XF86MonBrightnessDown | Brightness down |
| XF86AudioPlay/Pause | Play/pause media |
| XF86AudioNext | Next track |
| XF86AudioPrev | Previous track |
