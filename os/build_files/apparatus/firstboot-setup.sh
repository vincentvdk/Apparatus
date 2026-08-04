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
mkdir -p "$USER_HOME"/.config/walker
mkdir -p "$USER_HOME"/.config/uwsm
mkdir -p "$USER_HOME"/.config/satty
mkdir -p "$USER_HOME"/.config/apparatus
mkdir -p "$USER_HOME"/.local/state/apparatus

# Copy configs
cp /usr/share/apparatus/hypr/* "$USER_HOME"/.config/hypr/
cp /usr/share/apparatus/waybar/* "$USER_HOME"/.config/waybar/
# Include system config (updates with OS) with space for user customizations
cat > "$USER_HOME"/.config/kitty/kitty.conf << 'EOF'
# Include Apparatus defaults (updates with OS)
include /usr/share/apparatus/kitty/kitty.conf

# Include theme configuration
include theme.conf

# User customizations below

EOF
cp /usr/share/apparatus/walker/* "$USER_HOME"/.config/walker/
cp /usr/share/apparatus/uwsm/* "$USER_HOME"/.config/uwsm/
cp /usr/share/apparatus/satty/* "$USER_HOME"/.config/satty/


# Apply default theme (catppuccin-mocha)
THEME="catppuccin-mocha"
ln -sf /usr/share/apparatus/themes/$THEME/kitty.conf "$USER_HOME"/.config/kitty/theme.conf
ln -sf /usr/share/apparatus/themes/$THEME/waybar.css "$USER_HOME"/.config/waybar/theme.css
ln -sf /usr/share/apparatus/themes/$THEME/mako.conf "$USER_HOME"/.config/mako/config
ln -sf /usr/share/apparatus/themes/$THEME/hyprland.conf "$USER_HOME"/.config/hypr/theme.conf
ln -sf /usr/share/apparatus/themes/$THEME/satty/overrides.css "$USER_HOME"/.config/satty/overrides.css
echo "$THEME" > "$USER_HOME"/.config/apparatus/current-theme

# Fix ownership
chown -R "$USER_NAME:$USER_NAME" "$USER_HOME"/.config
chown -R "$USER_NAME:$USER_NAME" "$USER_HOME"/.local
chown -R "$USER_NAME:$USER_NAME" "$USER_HOME"/.var

# Mark firstboot as done
mkdir -p /var/lib/apparatus
touch /var/lib/apparatus/firstboot-done

echo "Apparatus setup complete for $USER_NAME"
