# Distrobox - Development Containers

Distrobox provides a seamless way to create and manage containerized development environments on Apparatus OS. These containers integrate with your host system while maintaining isolation.

## Overview

Distrobox containers:

- Run as rootless Podman containers
- Share your home directory by default (or use custom home)
- Have access to host devices (GPU, audio, etc.)
- Appear as native applications in your app launcher
- Can be customized per-project

## Creating a Distrobox

### Using Butler (Recommended)

1. Launch Butler:
   ```bash
   butler
   ```
2. Select "📦 Distrobox" from the menu
3. Choose "Create"
4. Enter a name (e.g., `dev`, `python-project`, `web-dev`)
5. Choose home directory option:
   - **Default home (~)**: Shares your user's home directory with the container
   - **Custom home folder**: Creates a dedicated home directory in `~/distrobox-homes/`

### Using Command Line

```bash
# Using default image
distrobox create -n mybox

# Using specific image
distrobox create -n python-dev -i ghcr.io/vincentvdk/apparatus-box:latest

# Using custom home directory
distrobox create -n isolated-dev -i ghcr.io/vincentvdk/apparatus-box:latest --home ~/distrobox-homes/isolated-dev
```

## Available Container Images

| Image | Description |
|-------|-------------|
| `ghcr.io/vincentvdk/apparatus-box:latest` | Default Apparatus distrobox image with development tools |
| `docker.io/fedora:43` | Vanilla Fedora 43 |
| `docker.io/ublue-os/fedora:43` | UBlue Fedora base |
| Any OCI-compatible image | Use any Podman-compatible image |

## Entering a Distrobox

```bash
# Enter a running container
distrobox enter mybox

# Enter with specific command
distrobox enter mybox -- bash
```

## Managing Distroboxes

### List All Containers

```bash
# List all distroboxes
distrobox list

# With no color output (for scripts)
distrobox list --no-color
```

### Stop a Container

```bash
# Stop a running container
distrobox stop mybox

# Force stop
distrobox stop mybox --yes
```

### Remove a Container

```bash
# Remove a container (keeps home directory)
distrobox rm mybox

# Force remove
distrobox rm mybox --force
```

### Upgrade a Container

```bash
# Update packages inside container
distrobox upgrade mybox

# Full image upgrade (recreates container, preserves home)
distrobox upgrade --image mybox
```

## Home Directory Options

### Shared Home (APPARATUS_OS_HOME=1)

- Container shares your host's home directory
- Files in `~/.config/` are shared between host and container
- Useful for: Keeping all configs in one place
- Environment: `APPARATUS_OS_HOME=1`

**Config location**: `$HOME/.config/apparatus-box/`

### Custom Home (APPARATUS_OS_HOME=0)

- Container has its own home directory
- Files are isolated from the host
- Useful for: Testing, different configurations, project isolation
- Environment: `APPARATUS_OS_HOME=0`

**Config location**: `$HOME/.config/` inside the container

## Configuration Sync

Apparatus OS can sync configurations from the host to all distroboxes:

```bash
# Sync all configs to all distroboxes
butler sync

# Sync to specific distrobox
butler sync mybox
```

### What Gets Synced

- Terminal configurations (Kitty, Alacritty)
- Window manager configurations (Hyprland)
- Status bar configurations (Waybar)
- Notification configurations (Mako)
- Theme configurations
- Shell configurations (Zsh)
- Editor configurations (Neovim)

### Skip List

Prevent specific distroboxes from receiving config updates:

```bash
# Add to skip list
butler  # Navigate to Skip List menu

# Or edit manually
nvim ~/.config/apparatus/skip-distroboxes
# Add one distrobox per line
my-test-box
```

## Networking

Distrobox containers share the host's network by default:

```bash
# Check network in container
ip a

# Ping works directly
ping google.com

# SSH connections work normally
ssh user@remote-host
```

## File Sharing

### With Shared Home

All files in your home directory are automatically shared:

```bash
# On host
cd ~/projects/myapp

# In distrobox (shared home)
cd ~/projects/myapp  # Same directory
```

### With Custom Home

Home directory is separate. Use `/host` to access host filesystem:

```bash
# In distrobox (custom home)
ls /host/home/youruser/projects
```

### Explicit File Copying

```bash
# Copy from host to container
distrobox enter mybox -- cp /host/path/to/file ./

# Copy from container to host
distrobox enter mybox -- cp ./file /host/path/to/destination
```

## Package Management

Each distrobox is a Fedora container with full package management:

```bash
# Enter container
distrobox enter mybox

# Install packages
sudo dnf install python3 nodejs git

# Update all packages
sudo dnf upgrade

# Search for packages
sudo dnf search <package>
```

## Development Workflows

### Python Development

```bash
# Create a Python distrobox
distrobox create -n python-dev -i ghcr.io/vincentvdk/apparatus-box:latest

# Enter and set up
distrobox enter python-dev
sudo dnf install python3 python3-pip python3-devel
pip install virtualenv

# Create project
cd ~/projects/myapp
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

### Node.js Development

```bash
# Create a Node.js distrobox
distrobox create -n node-dev -i ghcr.io/vincentvdk/apparatus-box:latest

# Enter and install Node.js
distrobox enter node-dev
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo bash -
sudo dnf install -y nodejs

# Verify
node --version
npm --version
```

### Rust Development

```bash
# Create a Rust distrobox
distrobox create -n rust-dev -i ghcr.io/vincentvdk/apparatus-box:latest

# Enter and install Rust
distrobox enter rust-dev
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source ~/.cargo/env

# Verify
rustc --version
cargo --version
```

## Application Integration

### GUI Applications

GUI applications in distroboxes can display on the host:

```bash
# Install GUI application
distrobox enter mybox
dnf install gimp

# Run GUI application (will appear on host)
gimp
```

### Audio

Audio works seamlessly between containers and host:

```bash
# Play audio in container
distrobox enter mybox
aplay /usr/share/sounds/test.wav
# Sound plays on host speakers
```

### Wayland

Wayland applications work with host's display server:

```bash
# Run Wayland app
distrobox enter mybox
weston --width=1024 --height=768 &
```

## Persistence

Distrobox home directories persist across container recreations:

- **Shared home**: Files persist in your host home directory
- **Custom home**: Files persist in the specified home directory

To completely remove all traces of a distrobox:

```bash
# Stop container
distrobox stop mybox

# Remove container
distrobox rm mybox --force

# Remove home directory (if using custom home)
rm -rf ~/distrobox-homes/mybox
```

## Backup

### Backup a Distrobox

```bash
# Export container to archive
distrobox export mybox -f mybox-backup.tar

# Or backup just the home directory
rsync -a ~/distrobox-homes/mybox/ ~/backups/mybox-home/
```

### Restore a Distrobox

```bash
# Import from archive
distrobox import -n mybox-restored -f mybox-backup.tar

# Or restore home directory
rsync -a ~/backups/mybox-home/ ~/distrobox-homes/mybox-restored/
```

## Troubleshooting

### Container won't start

```bash
# Check container status
podman ps -a

# Check container logs
podman logs mybox

# Try to restart
podman restart mybox
```

### GUI apps don't show

Ensure the container has proper XDG and Wayland permissions:

```bash
# Enter container with proper environment
distrobox enter mybox --export DISPLAY=$DISPLAY --export WAYLAND_DISPLAY=$WAYLAND_DISPLAY
```

### Audio not working

Check if PipeWire is running on the host:

```bash
# On host
systemctl --user status pipewire
```

### Home directory not found

If using custom home, ensure the directory exists:

```bash
mkdir -p ~/distrobox-homes/mybox
```

---

*For more information, see the [Usage Guide](usage.md) or [Butler Documentation](butler.md).*
