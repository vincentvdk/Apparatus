# Apparatus

Stop wasting too much time configuring your development environments and let Apparatus handle that for you.

## Intro

Apparatus is a Fedora Silverblue-based operating system configured for Developers, DevOps Engineers, Cloud Engineers, and SREs. It features:

- **Hyprland** - A modern, tiling Wayland compositor
- **Pre-configured Distroboxes** - Isolated development environments
- **Developer tools** - Neovim, Zellij, asdf, chezmoi, and more

## Installation

Apparatus is based on [Fedora Silverblue](https://fedoraproject.org/atomic-desktops/silverblue/) and requires rebasing from a base Silverblue installation.

### Step 1: Install Fedora Silverblue

Download and install Fedora Silverblue from [here](https://fedoraproject.org/atomic-desktops/silverblue/download).

### Step 2: Rebase to Apparatus (unsigned)

Open a terminal and run:

```shell
IMAGE_PATH=ghcr.io/vincentvdk/apparatus-os
rpm-ostree rebase ostree-unverified-registry:$IMAGE_PATH:latest
systemctl reboot
```

### Step 3: Rebase to signed image

After rebooting, complete the installation with the signed image:

```shell
IMAGE_PATH=ghcr.io/vincentvdk/apparatus-os
rpm-ostree rebase ostree-image-signed:docker://$IMAGE_PATH:latest
systemctl reboot
```

## Desktop Environment

Apparatus uses **Hyprland** as its desktop environment with a Catppuccin Mocha color scheme.

### Included Components

| Component | Application |
|-----------|-------------|
| Window Manager | Hyprland |
| Status Bar | Waybar |
| Terminal | kitty, rio |
| App Launcher | walker |
| File Manager | thunar |
| Notifications | mako |
| Lock Screen | hyprlock |
| Idle Daemon | hypridle |
| Wallpaper | hyprpaper |
| Screenshots | hyprshot + satty |
| OSD | swayosd |

### Keybindings

#### Applications

| Key | Action |
|-----|--------|
| `Super + Return` | Open terminal |
| `Super + D` | App launcher |
| `Super + E` | File manager |
| `Super + Q` | Close window |
| `Super + Shift + Q` | Exit Hyprland |
| `Super + L` | Lock screen |
| `Super + F1` | Show help/keybindings |

#### Window Management

| Key | Action |
|-----|--------|
| `Super + V` | Toggle floating |
| `Super + F` | Fullscreen |
| `Super + P` | Pseudo-tile |
| `Super + J` | Toggle split |
| `Super + Arrow keys` | Move focus |
| `Super + H/J/K/L` | Move focus (vim keys) |

#### Workspaces

| Key | Action |
|-----|--------|
| `Super + 1-0` | Switch to workspace 1-10 |
| `Super + Shift + 1-0` | Move window to workspace 1-10 |
| `Super + S` | Toggle scratchpad |
| `Super + Scroll` | Cycle workspaces |

#### Screenshots

| Key | Action |
|-----|--------|
| `Print` | Screenshot (full screen) |
| `Shift + Print` | Screenshot (window) |
| `Super + Shift + S` | Screenshot (region) |

#### Media Keys

Volume, brightness, and media playback keys work out of the box with on-screen display feedback via swayosd.

## Butler

Butler is the Apparatus configuration utility. Run it from a terminal:

```bash
butler
```

Or use command-line arguments:

```bash
butler init       # First-time setup
butler configure  # Configure Hyprland settings
butler theme      # Change color theme
butler help       # Show keyboard shortcuts
```

### Features

| Command | Description |
|---------|-------------|
| Init | First-time setup: copies configs, enables Flathub, installs default apps (Firefox, Signal, Joplin, Rio) |
| Configure | Configure Hyprland settings (terminal, etc.) |
| Theme | Switch between Catppuccin dark/light themes |
| Help | Display keyboard shortcuts reference |

### Terminal Selection

Butler supports switching between terminals:

- **kitty** - GPU-accelerated, feature-rich (default)
- **rio** - Modern, minimal, Rust-based (installed via Flatpak)

Change your default terminal:

```bash
butler configure
# Select "Terminal" and choose kitty or rio
```

### Kitty Terminal Splits

Kitty is configured with smart splits that automatically detect if you're inside a distrobox. When you create a new split, it will enter the same distrobox container (or stay on the host if you're not in a container).

#### Split Keybindings

| Key | Action |
|-----|--------|
| `Ctrl+Shift+V` | Vertical split (same distrobox/host) |
| `Ctrl+Shift+S` | Horizontal split (same distrobox/host) |
| `Ctrl+Shift+Arrow` | Navigate between splits |
| `Ctrl+Shift+R` | Resize split |
| `Ctrl+Shift+W` | Close split |

This is powered by the `smart-split` script which checks the `$CONTAINER_ID` environment variable to detect the current distrobox.

### Configuration

On first login, `butler init` copies default configs to your home directory:

```
~/.config/
├── hypr/
│   ├── hyprland.conf    # Main config
│   ├── hyprpaper.conf   # Wallpaper
│   ├── hypridle.conf    # Idle behavior
│   └── hyprlock.conf    # Lock screen
├── waybar/
│   ├── config.jsonc     # Modules
│   └── style.css        # Theme
└── mako/
    └── config           # Notifications
```

Default configs are stored in `/usr/share/apparatus/` and can be used to reset:

```bash
# Reset all configs
cp /usr/share/apparatus/hypr/* ~/.config/hypr/
cp /usr/share/apparatus/waybar/* ~/.config/waybar/
cp /usr/share/apparatus/mako/* ~/.config/mako/
```

To customize, edit your configs directly or source defaults with overrides:

```bash
cat > ~/.config/hypr/hyprland.conf << 'EOF'
source = /usr/share/apparatus/hypr/hyprland.conf

# Your customizations below...
$terminal = alacritty
bind = $mainMod, B, exec, firefox
EOF
```

## Distrobox

Create isolated development environments using pre-configured Distroboxes.

### Create a new Distrobox

```shell
distrobox create -i ghcr.io/vincentvdk/apparatus-box:latest -n dev
distrobox enter dev
```

### Update a Distrobox

```shell
distrobox upgrade dev
```

## Default Tools (Distrobox)

The development container includes:

- **Neovim** - Text editor
- **Zellij** - Terminal multiplexer
- **asdf** - Version manager
- **chezmoi** - Dotfiles manager
- **zsh** - Shell with Powerlevel10k

## Shell Configuration

### ZSH Plugins

ZSH plugins are managed by [antidote](https://github.com/mattmc3/antidote). Edit the plugin list at:

```
~/.config/zsh/zsh_plugins.txt
```

### Dotfiles with Chezmoi

Manage your dotfiles with [chezmoi](https://www.chezmoi.io/):

```bash
chezmoi init --apply https://github.com/$GITHUB_USERNAME/dotfiles.git
```

See the [chezmoi documentation](https://www.chezmoi.io/user-guide/setup/) for more details.

## ASDF Version Manager

### Example: Golang

```bash
# Add plugin
asdf plugin add golang

# Install version
asdf install golang 1.21.6

# Set global version
asdf global golang 1.21.6

# Switch versions
asdf shell golang 1.21.6

# After installing new packages
asdf reshim golang
```

### Default Packages

Create `~/.default-golang-pkgs` to auto-install packages with new Go versions:

```
// allows comments
github.com/jesseduffield/lazygit
```

## Fonts

The following [Nerd Fonts](https://github.com/ryanoasis/nerd-fonts) are pre-installed:

- Hack Nerd Font
- JetBrains Mono Nerd Font

## Virtualization

The following virtualization tools are included:

- Docker
- Podman
- libvirt/QEMU/KVM
- virt-manager

## Networking

- **Tailscale** - Pre-installed for mesh VPN

## Image Credits

| Image | Photographer | Source |
|-------|--------------|--------|
| Default wallpaper | [Jr Korpa](https://unsplash.com/@jrkorpa) | [Unsplash](https://unsplash.com/photos/pink-and-black-wallpaper-9XngoIpxcEo) |

## Links

- [Fedora Silverblue](https://fedoraproject.org/atomic-desktops/silverblue/)
- [Hyprland Wiki](https://wiki.hypr.land/)
- [Universal Blue](https://universal-blue.org/)
