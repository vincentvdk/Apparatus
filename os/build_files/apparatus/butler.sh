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
  "Theme            - Select a theme"
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
  esac
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
    exit 0
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
  theme)
    set_theme
    ;;
  *)
    main
    ;;
esac
