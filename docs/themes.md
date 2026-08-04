# Themes

Apparatus OS features a comprehensive theming system based on the Catppuccin color palette. Themes are applied consistently across all applications for a cohesive visual experience.

## Available Themes

| Theme | Type | Palette | Best For |
|-------|------|---------|----------|
| catppuccin-mocha | Dark | Mocha | Night coding, dark environments, reduced eye strain |
| catppuccin-latte | Light | Latte | Day use, bright environments, readability |

## Catppuccin Color Palette

### Mocha (Dark)

| Color | Hex | Usage |
|-------|-----|-------|
| Rosewater | #f5e0dc | Accents, highlights |
| Flamingo | #f2cdcd | Selection, subtle accents |
| Pink | #f5c2e7 | Syntax: strings, comments |
| Mauve | #cba6f7 | Syntax: keywords, operators |
| Red | #f38ba8 | Errors, deletions |
| Maroon | #eba0ac | Syntax: tags, special |
| Peach | #fab387 | Syntax: numbers, booleans |
| Yellow | #f9e2af | Syntax: functions, classes |
| Green | #a6e3a1 | Success, additions, syntax: types |
| Teal | #94e2d5 | Syntax: decorators, constants |
| Sky | #89d185 | Syntax: attributes, builtins |
| Sapphire | #89b4fa | Syntax: literals, imports |
| Blue | #74c7ec | Accents, information |
| Lavender | #b4befe | Primary, emphasis |
| Text | #cdd6f4 | Main text |
| Subtext 1 | #bac2de | Secondary text |
| Subtext 0 | #a6adc8 | Tertiary text |
| Overlay 2 | #9399b2 | UI elements |
| Overlay 1 | #7f849c | Borders, dividers |
| Overlay 0 | #6c7086 | Background accents |
| Surface 2 | #585b70 | Surface variants |
| Surface 1 | #45475a | Elevated surfaces |
| Surface 0 | #313244 | Surfaces |
| Base | #1e1e2e | Base UI color |
| Mantle | #181825 | Mantle/container |
| Crust | #11111b | Deep background |

### Latte (Light)

| Color | Hex | Usage |
|-------|-----|-------|
| Rosewater | #dc8a78 | Accents, highlights |
| Flamingo | #dd7878 | Selection, subtle accents |
| Pink | #ea76cb | Syntax: strings, comments |
| Mauve | #8839ef | Syntax: keywords, operators |
| Red | #d20f39 | Errors, deletions |
| Maroon | #e64553 | Syntax: tags, special |
| Peach | #fe640b | Syntax: numbers, booleans |
| Yellow | #df8e1d | Syntax: functions, classes |
| Green | #40a02b | Success, additions, syntax: types |
| Teal | #179299 | Syntax: decorators, constants |
| Sky | #04a5e5 | Syntax: attributes, builtins |
| Sapphire | #209fb5 | Syntax: literals, imports |
| Blue | #1e66f5 | Accents, information |
| Lavender | #7287fd | Primary, emphasis |
| Text | #4c4f69 | Main text |
| Subtext 1 | #6c6f85 | Secondary text |
| Subtext 0 | #8c8fa1 | Tertiary text |
| Overlay 2 | #acb0be | UI elements |
| Overlay 1 | #cacdcf | Borders, dividers |
| Overlay 0 | #e6e9ef | Background accents |
| Surface 2 | #f2f5f8 | Surface variants |
| Surface 1 | #ffffff | Elevated surfaces |
| Surface 0 | #f5f5f5 | Surfaces |
| Base | #eff1f5 | Base UI color |
| Mantle | #e6e9ef | Mantle/container |
| Crust | #dce0e8 | Deep background |

## Switching Themes

### Using Butler (Recommended)

```bash
# Switch to dark theme
butler theme catppuccin-mocha

# Switch to light theme
butler theme catppuccin-latte
```

Or via the TUI:

```bash
butler
# Select "Theme" from the menu
# Choose catppuccin-mocha or catppuccin-latte
```

### Manual Theme Switching

If you prefer to switch themes manually, edit the current theme file:

```bash
# Set dark theme
echo "catppuccin-mocha" > ~/.config/apparatus/current-theme

# Set light theme
echo "catppuccin-latte" > ~/.config/apparatus/current-theme

# Then reload all services
butler theme catppuccin-mocha  # This reloads automatically
# Or manually:
hyprctl reload
pkill -SIGUSR1 kitty
makoctl reload
```

## Theme File Locations

### System Themes

All theme files are stored in `/usr/share/apparatus/themes/`:

```
/usr/share/apparatus/themes/
├── catppuccin-mocha/
│   ├── kitty.conf          # Kitty terminal theme
│   ├── waybar.css         # Waybar theme
│   ├── mako.conf          # Mako notification theme
│   ├── hyprland.conf      # Hyprland theme
│   └── satty/
│       └── overrides.css  # Satty theme
└── catppuccin-latte/
    ├── kitty.conf
    ├── waybar.css
    ├── mako.conf
    ├── hyprland.conf
    └── satty/
        └── overrides.css
```

### User Theme Symlinks

Butler creates symlinks in your home directory:

```
~/.config/
├── kitty/
│   └── theme.conf -> /usr/share/apparatus/themes/catppuccin-mocha/kitty.conf
├── waybar/
│   └── theme.css -> /usr/share/apparatus/themes/catppuccin-mocha/waybar.css
├── mako/
│   └── config -> /usr/share/apparatus/themes/catppuccin-mocha/mako.conf
├── hypr/
│   └── theme.conf -> /usr/share/apparatus/themes/catppuccin-mocha/hyprland.conf
└── satty/
    └── overrides.css -> /usr/share/apparatus/themes/catppuccin-mocha/satty/overrides.css
```

## Theme Contents

### Kitty Theme

The Kitty theme file (`kitty.conf`) includes:

```ini
# Catppuccin Mocha colors
background #1e1e2e
foreground #cdd6f4
cursor #f5e0dc
cursor_text_color #1e1e2e

# Selection colors
selection_background #f5e0dc
selection_foreground #1e1e2e

# URL colors
url_color #89b4fa

# Tab bar colors
active_tab_background #cba6f7
active_tab_foreground #1e1e2e
inactive_tab_background #313244
inactive_tab_foreground #cdd6f4
```

### Waybar Theme

The Waybar theme file (`waybar.css`) includes:

```css
/* Background */
#workspaces {
    background-color: #1e1e2e;
}

/* Active workspace */
#workspaces button.active {
    background-color: #cba6f7;
    color: #1e1e2e;
}

/* Clock */
#clock {
    color: #cdd6f4;
}

/* Tray icons */
#tray {
    color: #cdd6f4;
}
```

### Mako Theme

The Mako notification theme file (`mako.conf`) includes:

```ini
background-color=#1e1e2e
text-color=#cdd6f4
border-color=#313244

urgent-background-color=#f38ba8
urgent-text-color=#1e1e2e
urgent-border-color=#f38ba8
```

### Hyprland Theme

The Hyprland theme file (`hyprland.conf`) includes:

```ini
# Border colors
col.active_border = rgba(203, 166, 247, 1.0)  # cba6f7
col.inactive_border = rgba(49, 50, 68, 1.0)    # 313244

# Window colors
col.window_bg = rgba(30, 30, 46, 1.0)         # 1e1e2e

# Title bar
col.title_text = rgba(205, 214, 244, 1.0)    # cdd6f4
```

### Satty Theme

The Satty theme file (`overrides.css`) includes:

```css
/* Background colors */
.satty-main {
    background-color: #1e1e2e;
}

/* Button colors */
.satty-button {
    background-color: #313244;
    color: #cdd6f4;
}

/* Accent colors */
.satty-button.primary {
    background-color: #cba6f7;
    color: #1e1e2e;
}
```

## Creating Custom Themes

To create a custom theme:

1. **Create theme directory**
   ```bash
   sudo mkdir -p /usr/share/apparatus/themes/my-custom-theme
   ```

2. **Add theme files**
   ```bash
   # Copy from existing theme as starting point
   sudo cp /usr/share/apparatus/themes/catppuccin-mocha/* /usr/share/apparatus/themes/my-custom-theme/
   ```

3. **Edit theme files**
   ```bash
   sudo nvim /usr/share/apparatus/themes/my-custom-theme/kitty.conf
   # Update colors to your preference
   ```

4. **Create subdirectory for satty**
   ```bash
   sudo mkdir -p /usr/share/apparatus/themes/my-custom-theme/satty
   sudo cp /usr/share/apparatus/themes/catppuccin-mocha/satty/overrides.css /usr/share/apparatus/themes/my-custom-theme/satty/
   ```

5. **Set permissions**
   ```bash
   sudo chmod -R 755 /usr/share/apparatus/themes/my-custom-theme
   sudo chown -R root:root /usr/share/apparatus/themes/my-custom-theme
   ```

6. **Apply your theme**
   ```bash
   butler theme my-custom-theme
   ```

## GTK Theme Integration

Butler automatically sets the GTK theme preference to match your selected theme:

```bash
# For dark themes (mocha)
gsettings set org.gnome.desktop.interface color-scheme 'prefer-dark'
gsettings set org.gnome.desktop.interface gtk-theme 'Adwaita-dark'

# For light themes (latte)
gsettings set org.gnome.desktop.interface color-scheme 'prefer-light'
gsettings set org.gnome.desktop.interface gtk-theme 'Adwaita'
```

### Manual GTK Theme Setup

If GTK themes aren't applying correctly:

```bash
# List available GTK themes
gsettings get org.gnome.desktop.interface gtk-theme

# Set specific GTK theme
gsettings set org.gnome.desktop.interface gtk-theme 'Adwaita-dark'

# Set icon theme
gsettings set org.gnome.desktop.interface icon-theme 'Adwaita'

# Set cursor theme
gsettings set org.gnome.desktop.interface cursor-theme 'Adwaita'
```

## Theme Sync Across Distroboxes

When you switch themes on the host using Butler, the theme is automatically synced to all distroboxes (except those in the skip list):

```bash
# Switch theme (automatically syncs to distroboxes)
butler theme catppuccin-latte

# Or manually sync
butler sync
```

## Theme Best Practices

1. **Consistency**: Use the same theme across all applications for a cohesive look
2. **Readability**: Ensure text is readable against background colors
3. **Accessibility**: Maintain sufficient color contrast
4. **Performance**: Dark themes (mocha) are generally better for battery life on OLED displays
5. **Eye Strain**: Dark themes reduce eye strain in low-light conditions

## Theme Switching and Applications

Some applications may need to be restarted after theme changes:

| Application | Restart Command |
|-------------|-----------------|
| Hyprland | `hyprctl reload` |
| Kitty | `pkill -SIGUSR1 kitty` |
| Waybar | `pkill waybar && waybar &` |
| Mako | `makoctl reload` |
| Satty | Restart the application |
| GTK apps | Usually pick up changes automatically |

## Troubleshooting

### Theme not applying

1. Check the current theme:
   ```bash
   cat ~/.config/apparatus/current-theme
   ```

2. Verify symlinks exist:
   ```bash
   ls -la ~/.config/kitty/theme.conf
   ls -la ~/.config/waybar/theme.css
   ```

3. Reload the application:
   ```bash
   hyprctl reload
   pkill -SIGUSR1 kitty
   ```

### Colors look wrong

1. Check if the theme file exists:
   ```bash
   ls /usr/share/apparatus/themes/catppuccin-mocha/kitty.conf
   ```

2. Verify the symlink points to the correct file:
   ```bash
   readlink ~/.config/kitty/theme.conf
   ```

3. Check for syntax errors in the theme file

### GTK apps don't follow theme

1. Verify GTK theme is set:
   ```bash
   gsettings get org.gnome.desktop.interface gtk-theme
   ```

2. Ensure the theme exists on your system:
   ```bash
   ls /usr/share/themes/ | grep -i adwaita
   ```

3. Try setting it manually:
   ```bash
   gsettings set org.gnome.desktop.interface gtk-theme 'Adwaita-dark'
   ```

---

*For more information about configuration, see the [Configuration Guide](configuration.md).*
