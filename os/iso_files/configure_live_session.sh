#!/bin/bash
# Configure live session for Apparatus OS ISO
# Uses GNOME for the live/installer session (Hyprland is on the installed system)
set -eux

# Install Anaconda installer and GNOME terminal (brings GNOME dependencies)
dnf install -y anaconda-live libblockdev-btrfs gnome-terminal

# Create liveuser with no password
useradd -m -G wheel -s /bin/bash liveuser || true
passwd -d liveuser

# Allow wheel group sudo without password
echo '%wheel ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/wheel-nopasswd
chmod 440 /etc/sudoers.d/wheel-nopasswd

# Configure GDM auto-login for liveuser (GNOME session)
mkdir -p /etc/gdm
cat > /etc/gdm/custom.conf << 'EOF'
[daemon]
AutomaticLoginEnable=True
AutomaticLogin=liveuser

[security]

[xdmcp]

[chooser]

[debug]
EOF

# Install Anaconda launcher to applications
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

# Create simple installer command
cat > /usr/bin/install-apparatus << 'EOF'
#!/bin/bash
echo "Starting Apparatus OS Installer..."
sudo liveinst
EOF
chmod +x /usr/bin/install-apparatus

# Set ownership
chown -R liveuser:liveuser /home/liveuser

# Mask bootc services that shouldn't run in live session
systemctl mask bootloader-update.service || true
systemctl mask bootc-fetch-apply-updates.timer || true
systemctl mask bootc-fetch-apply-updates.service || true

echo "Live session configured with GNOME and Anaconda installer"
