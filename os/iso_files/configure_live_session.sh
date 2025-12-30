#!/bin/bash
# Configure live session for Apparatus OS ISO
set -eux

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

echo "Live session configured for liveuser with Hyprland"
