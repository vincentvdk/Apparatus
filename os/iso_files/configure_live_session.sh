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

# Configure GDM auto-login for liveuser with Hyprland session
mkdir -p /etc/gdm
cat > /etc/gdm/custom.conf << 'EOF'
[daemon]
AutomaticLoginEnable=True
AutomaticLogin=liveuser
DefaultSession=hyprland.desktop

[security]

[xdmcp]

[chooser]

[debug]
EOF

# Use AccountsService to force Hyprland session for liveuser
# This is more reliable than just GDM config when GNOME deps are installed
mkdir -p /var/lib/AccountsService/users
cat > /var/lib/AccountsService/users/liveuser << 'EOF'
[User]
Session=hyprland
XSession=hyprland
Icon=/usr/share/icons/hicolor/96x96/apps/anaconda.png
SystemAccount=false
EOF

# Copy Apparatus Hyprland configs to liveuser
mkdir -p /home/liveuser/.config/hypr
mkdir -p /home/liveuser/.config/waybar
mkdir -p /home/liveuser/.config/mako
mkdir -p /home/liveuser/.config/kitty

cp -r /usr/share/apparatus/hypr/* /home/liveuser/.config/hypr/
cp -r /usr/share/apparatus/waybar/* /home/liveuser/.config/waybar/
cp -r /usr/share/apparatus/mako/* /home/liveuser/.config/mako/
cp -r /usr/share/apparatus/kitty/* /home/liveuser/.config/kitty/

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

echo "Live session configured with Hyprland and Anaconda installer"
echo "Launch installer with: Super+I, wofi search 'Install', or run 'install-apparatus'"
