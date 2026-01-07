#!/usr/bin/env bash
#set -x

logo='
   _____                                          __                 
  /  _  \  ______  ______ _____  _______ _____  _/  |_  __ __  ______
 /  /_\  \ \____ \ \____ \\__  \ \_  __ \\__  \ \   __\|  |  \/  ___/
/    |    \|  |_> >|  |_> >/ __ \_|  | \/ / __ \_|  |  |  |  /\___ \ 
\____|__  /|   __/ |   __/(____  /|__|   (____  /|__|  |____//____  >
        \/ |__|    |__|        \/             \/                  \/ 

'

echo -e "$logo"

# Available options
OPTIONS=(
  "Init             - Initialize a fresh install"
  "Distrobox        - Manage distroboxes"
  "Configure        - Configure Hyprland settings"
  "Theme            - Select a theme"
  "Help             - Show keyboard shortcuts"
)

# -- Main func
main() {
  local OPT=$(gum choose "${OPTIONS[@]}" --height 15 --header "Option:")
  local CHOICE=$(echo "$OPT" | awk -F ' {2,}' '{print $1}' | tr '[:upper:]' '[:lower:]' | sed 's/ /-/g')

  case "${CHOICE}" in
    theme)
      set_theme
      ;;
    init)
      init
      ;;
    distrobox)
      distrobox_menu
      ;;
    configure)
      configure
      ;;
    help)
      show_help
      ;;
  esac
}


# -- Distrobox menu
distrobox_menu() {
  DISTROBOX_OPTIONS=(
    "Create           - Create a new distrobox"
    "List             - List existing distroboxes"
    "Back             - Return to main menu"
  )

  while true; do
    local OPT=$(gum choose "${DISTROBOX_OPTIONS[@]}" --height 15 --header "Distrobox:")
    local CHOICE=$(echo "$OPT" | awk -F ' {2,}' '{print $1}' | tr '[:upper:]' '[:lower:]' | sed 's/ /-/g')

    case "${CHOICE}" in
      create)
        distrobox_create
        ;;
      list)
        distrobox_list
        ;;
      back)
        return
        ;;
    esac
  done
}

# -- Create distrobox
distrobox_create() {
  local DISTROBOX_HOMES="$HOME/distrobox-homes"
  local NAME=""
  local USE_CUSTOM_HOME=""
  local FOLDER_NAME=""
  local HOME_ARG=""

  echo '{{ Bold "# Create new Distrobox" }}' | gum format -t template

  # Get name
  NAME=$(gum input --placeholder "Enter distrobox name" --header "Name:")
  if [ -z "$NAME" ]; then
    echo '{{ Color "1" "Name cannot be empty" }}' | gum format -t template
    sleep 2
    return
  fi

  # Ask for custom home folder
  USE_CUSTOM_HOME=$(gum choose "Default home (~)" "Custom home folder" --header "Home folder:")

  if [ "$USE_CUSTOM_HOME" = "Custom home folder" ]; then
    FOLDER_NAME=$(gum input --placeholder "$NAME" --value "$NAME" --header "Folder name (in ~/distrobox-homes/):")
    if [ -n "$FOLDER_NAME" ]; then
      mkdir -p "$DISTROBOX_HOMES/$FOLDER_NAME"
      HOME_ARG="--home $DISTROBOX_HOMES/$FOLDER_NAME"
    fi
  fi

  echo '{{ Bold "# Creating distrobox..." }}' | gum format -t template
  distrobox create -i ghcr.io/vincentvdk/apparatus-box:latest -n "$NAME" $HOME_ARG

  if [ $? -eq 0 ]; then
    echo '{{ Bold "Distrobox created successfully!" }}' | gum format -t template
    echo "Enter with: distrobox enter $NAME"
  else
    echo '{{ Color "1" "Failed to create distrobox" }}' | gum format -t template
  fi
  sleep 3
}

# -- List distroboxes
distrobox_list() {
  echo '{{ Bold "# Existing Distroboxes" }}' | gum format -t template
  echo ""
  distrobox list
  echo ""
  echo "Press any key to continue..."
  read -n 1
}

# -- Configure Hyprland
configure() {
  CONFIG_OPTIONS=(
    "Terminal         - Set default terminal"
    "Monitors         - Configure display setup"
    "Audio            - Configure audio devices"
    "AI Workload      - Configure GPU VRAM for AI/ML"
    "Back             - Return to main menu"
  )

  while true; do
    local OPT=$(gum choose "${CONFIG_OPTIONS[@]}" --height 15 --header "Configure:")
    local CHOICE=$(echo "$OPT" | awk -F ' {2,}' '{print $1}' | tr '[:upper:]' '[:lower:]' | sed 's/ /-/g')

    case "${CHOICE}" in
      terminal)
        configure_terminal
        ;;
      monitors)
        configure_monitors
        ;;
      audio)
        configure_audio
        ;;
      ai-workload)
        configure_ai_workload
        ;;
      back)
        return
        ;;
    esac
  done
}

# -- Configure monitors
configure_monitors() {
  echo '{{ Bold "# Monitor Configuration" }}' | gum format -t template
  echo ""
  echo "Launching hyprdynamicmonitors TUI..."
  echo "Use this to create and manage monitor profiles."
  echo ""
  sleep 1
  hyprdynamicmonitors tui
}

# -- Configure audio
configure_audio() {
  echo '{{ Bold "# Audio Configuration" }}' | gum format -t template
  echo ""
  echo "Launching PulseAudio Volume Control..."
  echo ""
  sleep 1
  pavucontrol &
  disown
}

# -- Configure AI workload (VRAM allocation for AMD APUs)
configure_ai_workload() {
  echo '{{ Bold "# AI Workload Configuration" }}' | gum format -t template
  echo ""
  echo "Configure GPU VRAM allocation for AI/ML workloads."
  echo "This applies to AMD Ryzen AI / Strix Halo APUs."
  echo ""
  echo '{{ Color "3" "⚠ Requires reboot to take effect" }}' | gum format -t template
  echo ""

  local CHOICE=$(gum choose \
    "16 GB VRAM" \
    "32 GB VRAM" \
    "64 GB VRAM" \
    "96 GB VRAM (max stable)" \
    "Reset to default" \
    "Cancel" \
    --header "Select VRAM allocation:")

  local PAGES=""
  case "$CHOICE" in
    "16 GB VRAM")
      PAGES="4194304"  # 16GB in 4KB pages
      ;;
    "32 GB VRAM")
      PAGES="8388608"  # 32GB
      ;;
    "64 GB VRAM")
      PAGES="16777216" # 64GB
      ;;
    "96 GB VRAM (max stable)")
      PAGES="25165824" # 96GB
      ;;
    "Reset to default")
      echo '{{ Bold "Removing VRAM configuration..." }}' | gum format -t template
      pkexec rm -f /etc/kernel/cmdline.d/99-amd-vram.conf
      echo '{{ Bold "Done! Reboot to apply changes." }}' | gum format -t template
      sleep 2
      return
      ;;
    *)
      return
      ;;
  esac

  echo '{{ Bold "Applying VRAM configuration..." }}' | gum format -t template

  # Create kernel cmdline snippet for VRAM allocation
  pkexec bash -c "cat > /etc/kernel/cmdline.d/99-amd-vram.conf <<EOF
amdttm.pages_limit=${PAGES}
amdttm.page_pool_size=${PAGES}
EOF"

  echo '{{ Bold "Done! Reboot to apply changes." }}' | gum format -t template
  sleep 2
}

# -- Configure terminal
configure_terminal() {
  echo '{{ Bold "# Select default terminal:" }}' | gum format -t template

  local CHOICE=$(gum choose "kitty" "rio")

  local TERM_CMD
  case "$CHOICE" in
    rio)
      TERM_CMD="flatpak run com.rioterm.Rio"
      ;;
    *)
      TERM_CMD="$CHOICE"
      ;;
  esac

  local HYPR_CONF="$HOME/.config/hypr/hyprland.conf"
  sed -i "s|^\\\$terminal = .*|\$terminal = $TERM_CMD|" "$HYPR_CONF"
  hyprctl reload
  echo '{{ Bold "Terminal set to: " }}'"$TERM_CMD" | gum format -t template
}

# -- Show help
show_help() {
  gum format << 'EOF' | gum pager
# Apparatus Keyboard Shortcuts

## General
| Key | Action |
|-----|--------|
| Super + Return | Open terminal (kitty) |
| Super + D | Application launcher (walker) |
| Super + E | File manager (thunar) |
| Super + Q | Close window |
| Super + Shift + Q | Exit Hyprland |
| Super + V | Toggle floating |
| Super + F | Fullscreen |
| Super + L | Lock screen |
| Super + F1 | Show this help |

## Window Navigation
| Key | Action |
|-----|--------|
| Super + ←/→/↑/↓ | Move focus |
| Super + H/J/K/L | Move focus (vim keys) |

## Workspaces
| Key | Action |
|-----|--------|
| Super + 1-9,0 | Switch to workspace |
| Super + Shift + 1-9,0 | Move window to workspace |
| Super + Mouse Scroll | Cycle workspaces |
| Super + S | Toggle scratchpad |

## Screenshots (with Satty annotation)
| Key | Action |
|-----|--------|
| Print | Screenshot (full) → Satty |
| Shift + Print | Screenshot (window) → Satty |
| Super + Shift + S | Screenshot (region) → Satty |

## Media
| Key | Action |
|-----|--------|
| Volume Keys | Adjust volume |
| Brightness Keys | Adjust brightness |
| Play/Pause/Next/Prev | Media control |

---
Press q to exit
EOF
}

# -- Set theme
set_theme() {
  local THEMES_DIR="/usr/share/apparatus/themes"
  local THEME=""

  echo '{{ Bold "# Set theme:" }}' | gum format -t template
  local CHOICE=$(gum choose "Catppuccin Mocha (dark)" "Catppuccin Latte (light)")

  case "$CHOICE" in
    "Catppuccin Mocha (dark)")
      THEME="catppuccin-mocha"
      ;;
    "Catppuccin Latte (light)")
      THEME="catppuccin-latte"
      ;;
    *)
      return
      ;;
  esac

  echo '{{ Bold "# Applying theme: " }}'"$THEME" | gum format -t template

  # Apply kitty theme
  if [ -d "$HOME/.config/kitty" ]; then
    ln -sf "$THEMES_DIR/$THEME/kitty.conf" "$HOME/.config/kitty/theme.conf"
  fi

  # Apply waybar theme
  if [ -d "$HOME/.config/waybar" ]; then
    ln -sf "$THEMES_DIR/$THEME/waybar.css" "$HOME/.config/waybar/theme.css"
  fi

  # Apply mako theme
  if [ -d "$HOME/.config/mako" ]; then
    ln -sf "$THEMES_DIR/$THEME/mako.conf" "$HOME/.config/mako/config"
  fi

  # Apply hyprland theme
  if [ -d "$HOME/.config/hypr" ]; then
    ln -sf "$THEMES_DIR/$THEME/hyprland.conf" "$HOME/.config/hypr/theme.conf"
  fi

  # Apply rio theme (update theme name in config.toml)
  if [ -f "$HOME/.config/rio/config.toml" ]; then
    sed -i "s/^theme = .*/theme = \"$THEME\"/" "$HOME/.config/rio/config.toml"
  fi

  # Apply GTK theme (dark/light based on theme)
  if [[ "$THEME" == *"mocha"* ]] || [[ "$THEME" == *"dark"* ]]; then
    gsettings set org.gnome.desktop.interface color-scheme 'prefer-dark'
  else
    gsettings set org.gnome.desktop.interface color-scheme 'prefer-light'
  fi

  # Save current theme
  mkdir -p "$HOME/.config/apparatus"
  echo "$THEME" > "$HOME/.config/apparatus/current-theme"

  # Reload services
  hyprctl reload
  pkill -SIGUSR1 kitty || true
  makoctl reload || true

  echo '{{ Bold "Theme applied!" }}' | gum format -t template
  sleep 2
}

# -- Init new install
init() {
  # Check if init already ran
  if test -e "$HOME"/.config/apparatus/init-done; then
    echo '{{ Bold "System already initialized.."}}' | gum format -t template
    sleep 2
    return
  fi

  echo '{{ Bold "# Configuring Hyprland" }}' | gum format -t template
  mkdir -p ${HOME}/.config/hypr
  mkdir -p ${HOME}/.config/waybar
  mkdir -p ${HOME}/.config/mako
  mkdir -p ${HOME}/.config/kitty
  mkdir -p ${HOME}/.config/rio/themes
  mkdir -p ${HOME}/.config/uwsm
  mkdir -p ${HOME}/.config/apparatus

  # Copy configs
  cp /usr/share/apparatus/hypr/* ~/.config/hypr/
  cp /usr/share/apparatus/waybar/* ~/.config/waybar/
  cp /usr/share/apparatus/kitty/* ~/.config/kitty/
  cp /usr/share/apparatus/rio/config.toml ~/.config/rio/
  cp /usr/share/apparatus/uwsm/* ~/.config/uwsm/

  # Symlink rio themes from central themes folder
  ln -sf /usr/share/apparatus/themes/catppuccin-mocha/rio.toml ~/.config/rio/themes/catppuccin-mocha.toml
  ln -sf /usr/share/apparatus/themes/catppuccin-latte/rio.toml ~/.config/rio/themes/catppuccin-latte.toml

  # Symlink rio config for flatpak (rio looks in ~/.var/app/com.rioterm.Rio/config/rio/)
  mkdir -p ${HOME}/.var/app/com.rioterm.Rio/config
  ln -sfn ${HOME}/.config/rio ${HOME}/.var/app/com.rioterm.Rio/config/rio

  echo '{{ Bold "# Applying theme" }}' | gum format -t template
  local CURRENT_THEME="catppuccin-mocha"
  ln -sf /usr/share/apparatus/themes/$CURRENT_THEME/kitty.conf ~/.config/kitty/theme.conf
  ln -sf /usr/share/apparatus/themes/$CURRENT_THEME/waybar.css ~/.config/waybar/theme.css
  ln -sf /usr/share/apparatus/themes/$CURRENT_THEME/mako.conf ~/.config/mako/config
  ln -sf /usr/share/apparatus/themes/$CURRENT_THEME/hyprland.conf ~/.config/hypr/theme.conf
  gsettings set org.gnome.desktop.interface color-scheme 'prefer-dark'
  echo "$CURRENT_THEME" > ~/.config/apparatus/current-theme

  echo '{{ Bold "# Enable Flathub Repository" }}' | gum format -t template
  /usr/bin/flatpak remote-add --user --if-not-exists flathub https://flathub.org/repo/flathub.flatpakrepo || exit 1

  echo '{{ Bold "# Installing Firefox" }}' | gum format -t template
  /usr/bin/flatpak install --user --noninteractive flathub org.mozilla.firefox || exit 1

  echo '{{ Bold "# Installing Signal" }}' | gum format -t template
  /usr/bin/flatpak install --user --noninteractive flathub org.signal.Signal || exit 1

  echo '{{ Bold "# Installing Joplin" }}' | gum format -t template
  /usr/bin/flatpak install --user --noninteractive flathub net.cozic.joplin_desktop || exit 1

  echo '{{ Bold "# Installing Rio Terminal" }}' | gum format -t template
  /usr/bin/flatpak install --user --noninteractive flathub com.rioterm.Rio || exit 1

  echo '{{ Bold "# Enabling Walker services" }}' | gum format -t template
  elephant service enable
  systemctl --user start elephant.service

  echo '{{ Bold "# Setup complete!" }}' | gum format -t template
  mkdir -p "$HOME"/.config/apparatus
  touch "$HOME"/.config/apparatus/init-done
}

# Main - handle command line arguments or show menu
case "${1:-}" in
  init)
    init
    ;;
  distrobox)
    distrobox_menu
    ;;
  configure)
    configure
    ;;
  theme)
    set_theme
    ;;
  help)
    show_help
    ;;
  *)
    main
    ;;
esac
