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
    echo '{{ Bold "Already initialized.."}}' | gun format -t template
    exit 0
  fi
  (
  # Alacritty config
  mkdir -p ${HOME}/.config/alacritty
  cp /usr/share/apparatus/alacritty/* ~/.config/alacritty/

  echo "# Enable Flathub Repository"
  /usr/bin/flatpak remote-add --user --if-not-exists flathub https://flathub.org/repo/flathub.flatpakrepo
  if [ "$?" != 0 ] ; then
    zenity --error \
          --text="Adding Flathub Repo Failed"
    exit 1
  fi
  echo "5"


  # FLATPAK - Install Firefox
  echo "# Installing Firefox"
  /usr/bin/flatpak install --user --noninteractive flathub org.mozilla.firefox
  if [ "$?" != 0 ] ; then
        zenity --error \
          --text="Installing Firefox Failed"
        exit 1
  fi
  echo "10"

  echo "# Installing Signal"
  /usr/bin/flatpak install --user --noninteractive flathub org.signal.Signal
  if [ "$?" != 0 ] ; then
        zenity --error \
          --text="Installing Signal Failed"
        exit 1
  fi
  echo "50"

  echo "# Installing Joplin"
  /usr/bin/flatpak install --user --noninteractive flathub net.cozic.joplin_desktop
  if [ "$?" != 0 ] ; then
        zenity --error \
          --text="Installing Joplin Failed"
        exit 1
  fi
  echo "100"

  echo "# Create init-done file"
  mkdir -p "$HOME"/.config/ublue/
  touch "$HOME"/.config/apparatus/init-done


  ) |
    zenity --progress --title="uBlue Desktop Firstboot" --percentage=0 --auto-close --no-cancel --width=300
}

# Main
main
