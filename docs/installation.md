# Installation Guide

## Prerequisites

- **Hardware**: x86_64 CPU, 8GB+ RAM recommended, 50GB+ disk space
- **Secure Boot**: Currently not supported (disable in BIOS/UEFI)
- **Firmware**: UEFI required

## Installation Methods

### Method 1: ISO Installer (Recommended)

1. **Download ISO**
   - Get the latest ISO from [GitHub Releases](https://github.com/vincentvdk/apparatus/releases)
   - Look for files named `apparatus-*.iso`

2. **Create Bootable Media**
   ```bash
   # Linux (using dd)
   sudo dd if=apparatus-*.iso of=/dev/sdX bs=4M status=progress oflag=sync
   
   # Or use Ventoy for multi-ISO USB
   ```

3. **Boot from USB**
   - Select the USB device in your BIOS/UEFI boot menu
   - Boot into the live environment

4. **Install to Disk**
   - Follow the graphical installer prompts
   - Select your target disk
   - Configure user account
   - Wait for installation to complete

5. **First Boot**
   - Remove installation media
   - Reboot into your new Apparatus OS installation
   - Complete the first-boot setup

### Method 2: bootc Image (Advanced)

For users familiar with bootc and container-based deployments:

```bash
# Pull the latest Apparatus OS image
podman pull ghcr.io/vincentvdk/apparatus-os:latest

# Deploy to disk (replace /dev/sdX with your target disk)
bootc install-to-disk --image-ref ghcr.io/vincentvdk/apparatus-os:latest /dev/sdX
```

## Post-Installation

### Initial Setup

On first login, the system will:

1. Create your user configuration directory
2. Apply default theme (Catppuccin Mocha)
3. Set up default applications
4. Configure distrobox if not already set up

### Verify Installation

```bash
# Check OS version
cat /etc/os-release

# Check Hyprland is running
hyprctl info

# Check distrobox
podman --version
distrobox --version
```

## Updating

Apparatus OS uses bootc for atomic updates:

```bash
# Check for updates
bootc status

# Apply updates (will reboot)
bootc update
```

### Manual Update Check

```bash
# Check bootc status
bootc status

# Fetch and apply updates
bootc update
systemctl reboot
```

## Rolling Back

If an update causes issues, you can roll back:

```bash
# List available deployments
bootc status

# Roll back to previous deployment by rebooting
# Bootc automatically manages deployment rollbacks
systemctl reboot
```

## Uninstalling

To remove Apparatus OS and return to your previous system:

1. Reinstall your previous operating system
2. Or restore from backup

---

*Need help? Check the [Troubleshooting](troubleshooting.md) guide or open an issue.*
