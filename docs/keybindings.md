# Apparatus OS Keybindings

Optimized for Ultimate Hacking Keyboard (UHK) and vim-style navigation.

## Philosophy

- **SUPER** = Window manager (Hyprland)
- **Ctrl+A** = Terminal prefix (Kitty) - tmux-compatible
- **h/j/k/l** = vim navigation everywhere
- **SHIFT** = "move" modifier
- **CTRL** = "resize" modifier

## Hyprland

### Window Focus
| Key | Action |
|-----|--------|
| SUPER + H | Focus left |
| SUPER + J | Focus down |
| SUPER + K | Focus up |
| SUPER + L | Focus right |
| SUPER + Arrow keys | Focus (alternative) |

### Window Move
| Key | Action |
|-----|--------|
| SUPER + SHIFT + H | Move window left |
| SUPER + SHIFT + J | Move window down |
| SUPER + SHIFT + K | Move window up |
| SUPER + SHIFT + L | Move window right |

### Window Resize
| Key | Action |
|-----|--------|
| SUPER + CTRL + H | Resize left |
| SUPER + CTRL + J | Resize down |
| SUPER + CTRL + K | Resize up |
| SUPER + CTRL + L | Resize right |

### Window Actions
| Key | Action |
|-----|--------|
| SUPER + W | Close window |
| SUPER + F | Fullscreen |
| SUPER + V | Toggle floating |

### Applications
| Key | Action |
|-----|--------|
| SUPER + Return | Terminal |
| SUPER + Space | Launcher (walker) |
| SUPER + E | File manager |

### Session
| Key | Action |
|-----|--------|
| SUPER + Escape | Lock screen |
| SUPER + SHIFT + Escape | Exit Hyprland |

### Workspaces
| Key | Action |
|-----|--------|
| SUPER + 1-0 | Switch to workspace 1-10 |
| SUPER + SHIFT + 1-0 | Move window to workspace 1-10 |
| SUPER + Mouse scroll | Cycle workspaces |

### Scratchpad
| Key | Action |
|-----|--------|
| SUPER + ` | Toggle scratchpad |
| SUPER + SHIFT + ` | Move window to scratchpad |

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
