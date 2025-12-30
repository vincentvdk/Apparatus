#!/bin/bash
# Configure live session for Apparatus OS ISO
set -eux

# Install Anaconda installer
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

# Create desktop shortcut for Anaconda installer
mkdir -p /home/liveuser/Desktop
cat > /home/liveuser/Desktop/install-apparatus.desktop << 'EOF'
[Desktop Entry]
Name=Install Apparatus OS
Comment=Install the operating system to disk
Exec=sudo liveinst
Icon=anaconda
Terminal=false
Type=Application
Categories=System;
EOF
chmod +x /home/liveuser/Desktop/install-apparatus.desktop
chown -R liveuser:liveuser /home/liveuser/Desktop

echo "Live session configured with Anaconda installer"
