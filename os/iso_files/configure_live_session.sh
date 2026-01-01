#!/usr/bin/env bash
# Configure live session for Apparatus OS ISO
# Based on Bluefin's approach - uses GNOME for the installer session
set -eoux pipefail

# Install Anaconda and GNOME essentials for live session
dnf install -y \
    anaconda-live \
    libblockdev-btrfs \
    libblockdev-lvm \
    libblockdev-dm \
    gnome-terminal \
    nautilus \
    rsync

# Create liveuser with no password
useradd -m -G wheel -s /bin/bash liveuser || true
passwd -d liveuser

# Allow wheel group sudo without password
echo '%wheel ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/wheel-nopasswd
chmod 440 /etc/sudoers.d/wheel-nopasswd

# Configure GNOME for live session
tee /usr/share/glib-2.0/schemas/zz1-apparatus-live.gschema.override <<'EOF'
[org.gnome.shell]
welcome-dialog-last-shown-version='4294967295'
favorite-apps = ['anaconda.desktop', 'org.gnome.Terminal.desktop', 'org.gnome.Nautilus.desktop']

[org.gnome.settings-daemon.plugins.power]
sleep-inactive-ac-type='nothing'
sleep-inactive-battery-type='nothing'
sleep-inactive-ac-timeout=0
sleep-inactive-battery-timeout=0

[org.gnome.desktop.session]
idle-delay=uint32 0
EOF

glib-compile-schemas /usr/share/glib-2.0/schemas

# Configure GDM auto-login for liveuser
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

# Disable services that shouldn't run in live session
systemctl disable bootloader-update.service || true
systemctl disable rpm-ostreed-automatic.timer || true
systemctl disable bootc-fetch-apply-updates.timer || true
systemctl mask bootloader-update.service || true
systemctl mask bootc-fetch-apply-updates.service || true

# Anaconda Profile for Apparatus
tee /etc/anaconda/profile.d/apparatus.conf <<'EOF'
[Profile]
profile_id = apparatus

[Profile Detection]
os_id = apparatus

[Network]
default_on_boot = FIRST_WIRED_WITH_LINK

[Bootloader]
efi_dir = fedora
menu_auto_hide = True

[Storage]
default_scheme = BTRFS
btrfs_compression = zstd:1

[User Interface]
hidden_spokes =
    NetworkSpoke
EOF

# Set branding
sed -i 's/ANACONDA_PRODUCTVERSION=.*/ANACONDA_PRODUCTVERSION=""/' /usr/{,s}bin/liveinst || true

# Set ownership and permissions
chown -R liveuser:liveuser /home/liveuser
chmod 755 /home/liveuser
restorecon -R /home/liveuser || true

echo "Live session configured with GNOME and Anaconda installer"
