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
  "Theme            - Select a theme"
  #"Manage Tools     - Install and manage tools"
)

# -- Main func
main() {
  local OPT=$(gum choose "${OPTIONS[@]}" --height 15 --header "Option:")
  local CHOICE=$(echo "$OPT" | awk -F ' {2,}' '{print $1}' | tr '[:upper:]' '[:lower:]' | sed 's/ /-/g')

  case "${CHOICE}" in
    theme)
      set_theme
      ;;
    manage-tools)
      get_binary
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
      distrobox-host-exec gsettings set org.gnome.Ptyxis interface-style dark
      ;;
    catpuccin-light)
      distrobox-host-exec gsettings set org.gnome.Ptyxis interface-style light
      ;;
  esac
}

# Main
main
