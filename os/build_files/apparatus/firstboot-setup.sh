#!/usr/bin/env bash
# Apparatus First Boot Setup
# Copies configs to the first user's home directory

set -e

# Find the first regular user (UID >= 1000)
USER_NAME=$(getent passwd | awk -F: '$3 >= 1000 && $3 < 65534 {print $1; exit}')
USER_HOME=$(getent passwd "$USER_NAME" | cut -d: -f6)

if [ -z "$USER_NAME" ] || [ -z "$USER_HOME" ]; then
    echo "No regular user found, skipping setup"
    exit 0
fi

echo "Setting up Apparatus for user: $USER_NAME"

# Create config directories
mkdir -p "$USER_HOME"/.config/hypr
mkdir -p "$USER_HOME"/.config/waybar
mkdir -p "$USER_HOME"/.config/mako
mkdir -p "$USER_HOME"/.config/kitty
mkdir -p "$USER_HOME"/.config/rio/themes
mkdir -p "$USER_HOME"/.config/uwsm
mkdir -p "$USER_HOME"/.config/apparatus
mkdir -p "$USER_HOME"/.local/state/apparatus

# Copy configs
cp /usr/share/apparatus/hypr/* "$USER_HOME"/.config/hypr/
cp /usr/share/apparatus/waybar/* "$USER_HOME"/.config/waybar/
cp /usr/share/apparatus/kitty/* "$USER_HOME"/.config/kitty/
cp /usr/share/apparatus/rio/config.toml "$USER_HOME"/.config/rio/
cp /usr/share/apparatus/uwsm/* "$USER_HOME"/.config/uwsm/

# Symlink rio themes
ln -sf /usr/share/apparatus/themes/catppuccin-mocha/rio.toml "$USER_HOME"/.config/rio/themes/catppuccin-mocha.toml
ln -sf /usr/share/apparatus/themes/catppuccin-latte/rio.toml "$USER_HOME"/.config/rio/themes/catppuccin-latte.toml

# Symlink rio config for flatpak
mkdir -p "$USER_HOME"/.var/app/com.rioterm.Rio/config
ln -sfn "$USER_HOME"/.config/rio "$USER_HOME"/.var/app/com.rioterm.Rio/config/rio

# Apply default theme (catppuccin-mocha)
THEME="catppuccin-mocha"
ln -sf /usr/share/apparatus/themes/$THEME/kitty.conf "$USER_HOME"/.config/kitty/theme.conf
ln -sf /usr/share/apparatus/themes/$THEME/waybar.css "$USER_HOME"/.config/waybar/theme.conf
ln -sf /usr/share/apparatus/themes/$THEME/mako.conf "$USER_HOME"/.config/mako/config
ln -sf /usr/share/apparatus/themes/$THEME/hyprland.conf "$USER_HOME"/.config/hypr/theme.conf
echo "$THEME" > "$USER_HOME"/.config/apparatus/current-theme

# Fix ownership
chown -R "$USER_NAME:$USER_NAME" "$USER_HOME"/.config
chown -R "$USER_NAME:$USER_NAME" "$USER_HOME"/.local
chown -R "$USER_NAME:$USER_NAME" "$USER_HOME"/.var

# Mark firstboot as done
mkdir -p /var/lib/apparatus
touch /var/lib/apparatus/firstboot-done

echo "Apparatus setup complete for $USER_NAME"
