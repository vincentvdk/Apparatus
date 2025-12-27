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
    "Back             - Return to main menu"
  )

  while true; do
    local OPT=$(gum choose "${CONFIG_OPTIONS[@]}" --height 15 --header "Configure:")
    local CHOICE=$(echo "$OPT" | awk -F ' {2,}' '{print $1}' | tr '[:upper:]' '[:lower:]' | sed 's/ /-/g')

    case "${CHOICE}" in
      terminal)
        configure_terminal
        ;;
      back)
        return
        ;;
    esac
  done
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
| Super + D | Application launcher (wofi) |
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

## Screenshots
| Key | Action |
|-----|--------|
| Print | Screenshot (full) |
| Shift + Print | Screenshot (window) |
| Super + Shift + S | Screenshot (region) |

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
  echo '{{ Bold "# Set theme:" }}' | gum format -t template
  local OPT=$(gum choose "catpuccin-dark" "catpuccin-light") #TODO read this from a config file
  local CHOICE=$(echo "$OPT")

  case ${CHOICE} in
    catpuccin-dark)
      ln -sf /usr/share/apparatus/alacritty/catppuccin-mocha.toml ~/.config/alacritty/theme.toml
      ;;
    catpuccin-light)
      ln -sf /usr/share/apparatus/alacritty/catppuccin-latte.toml ~/.config/alacritty/theme.toml
      ;;
  esac
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
  mkdir -p ${HOME}/.config/uwsm
  cp /usr/share/apparatus/hypr/* ~/.config/hypr/
  cp /usr/share/apparatus/waybar/* ~/.config/waybar/
  cp /usr/share/apparatus/mako/* ~/.config/mako/
  cp /usr/share/apparatus/uwsm/* ~/.config/uwsm/

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
