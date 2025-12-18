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
| Terminal | foot |
| App Launcher | wofi |
| File Manager | thunar |
| Notifications | mako |
| Lock Screen | hyprlock |
| Idle Daemon | hypridle |
| Wallpaper | hyprpaper |
| Screenshots | hyprshot |
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

### Configuration Files

Default configs are installed to `~/.config/` on first login:

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

System defaults are also available at `/usr/share/apparatus/` for reference.

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

[Hack Nerd Font](https://github.com/ryanoasis/nerd-fonts) is pre-installed for terminal and UI use.

## Virtualization

The following virtualization tools are included:

- Docker
- Podman
- libvirt/QEMU/KVM
- virt-manager

## Networking

- **Tailscale** - Pre-installed for mesh VPN

## Links

- [Fedora Silverblue](https://fedoraproject.org/atomic-desktops/silverblue/)
- [Hyprland Wiki](https://wiki.hypr.land/)
- [Universal Blue](https://universal-blue.org/)
