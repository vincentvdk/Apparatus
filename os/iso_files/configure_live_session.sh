#!/bin/bash
# Configure live session for Apparatus OS ISO
set -eux

# Install Anaconda installer (this pulls in GNOME deps, we'll override session later)
dnf install -y anaconda-live libblockdev-btrfs

# Create liveuser with no password
useradd -m -G wheel -s /bin/bash liveuser || true
passwd -d liveuser

# Allow wheel group sudo without password
echo '%wheel ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/wheel-nopasswd
chmod 440 /etc/sudoers.d/wheel-nopasswd

# Configure GDM (auto-login disabled for debugging)
mkdir -p /etc/gdm
cat > /etc/gdm/custom.conf << 'EOF'
[daemon]
AutomaticLoginEnable=False

[security]

[xdmcp]

[chooser]

[debug]
EOF

# Use AccountsService to force Hyprland session for liveuser
# Use 'hyprland' not 'hyprland-uwsm' - direct launch is more reliable with GDM
mkdir -p /var/lib/AccountsService/users
cat > /var/lib/AccountsService/users/liveuser << 'EOF'
[User]
Session=hyprland
XSession=
Icon=/usr/share/icons/hicolor/96x96/apps/anaconda.png
SystemAccount=false
EOF

# Ensure hyprland.desktop exists and is the default (not uwsm version)
# GDM has known issues with Hyprland - ensure we're using the simple session
if [ -f /usr/share/wayland-sessions/hyprland-uwsm.desktop ]; then
    # Rename uwsm version so GDM picks the regular hyprland.desktop
    mv /usr/share/wayland-sessions/hyprland-uwsm.desktop /usr/share/wayland-sessions/hyprland-uwsm.desktop.bak || true
fi

# Copy Apparatus Hyprland configs to liveuser
mkdir -p /home/liveuser/.config/hypr
mkdir -p /home/liveuser/.config/waybar
mkdir -p /home/liveuser/.config/mako
mkdir -p /home/liveuser/.config/kitty

cp -r /usr/share/apparatus/hypr/* /home/liveuser/.config/hypr/
cp -r /usr/share/apparatus/waybar/* /home/liveuser/.config/waybar/
cp -r /usr/share/apparatus/mako/* /home/liveuser/.config/mako/
cp -r /usr/share/apparatus/kitty/* /home/liveuser/.config/kitty/

# Apply default theme (catppuccin-mocha) - configs expect theme.conf files
cp /usr/share/apparatus/themes/catppuccin-mocha/hyprland.conf /home/liveuser/.config/hypr/theme.conf
cp /usr/share/apparatus/themes/catppuccin-mocha/kitty.conf /home/liveuser/.config/kitty/theme.conf
cp /usr/share/apparatus/themes/catppuccin-mocha/waybar.css /home/liveuser/.config/waybar/theme.css
cp /usr/share/apparatus/themes/catppuccin-mocha/mako.conf /home/liveuser/.config/mako/theme.conf

# Install Anaconda launcher to applications (for wofi launcher)
cat > /usr/share/applications/install-apparatus.desktop << 'EOF'
[Desktop Entry]
Name=Install Apparatus OS
Comment=Install the operating system to disk
Exec=sudo liveinst
Icon=anaconda
Terminal=false
Type=Application
Categories=System;
Keywords=install;installer;anaconda;
EOF

# Create a simple terminal-based installer script as backup
cat > /usr/bin/install-apparatus << 'EOF'
#!/bin/bash
echo "Starting Apparatus OS Installer..."
sudo liveinst
EOF
chmod +x /usr/bin/install-apparatus

# Add keybind hint and installer to Hyprland config for live session
cat >> /home/liveuser/.config/hypr/hyprland.conf << 'EOF'

# Live Session - Install Apparatus OS with Super+I
bind = $mainMod, I, exec, sudo liveinst
EOF

# Set ownership
chown -R liveuser:liveuser /home/liveuser/.config

# Mask bootc/ostree services that shouldn't run in live session
systemctl mask bootloader-update.service || true
systemctl mask bootc-fetch-apply-updates.timer || true
systemctl mask bootc-fetch-apply-updates.service || true
systemctl mask ostree-remount.service || true

echo "Live session configured with Hyprland and Anaconda installer"
echo "Launch installer with: Super+I, wofi search 'Install', or run 'install-apparatus'"
