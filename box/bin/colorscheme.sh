#!/bin/sh


COLOR=$1
ALACRITTY_PATH="~/.config/alacritty/alacritty.yml"

configure_alacritty() {
  sed -r "s/(^colors:)(.*)/\1 $COLOR/g" $ALACRITTY_PATH
}

#configure_vim() {
#  echo $1 > ${dotfiles}/.vim/color.vim
#}

case $COLOR in
  nightfox)
    configure_alacritty
    #configure_vim 'NightFox'
    ;;
  dayfox)
    configure_alacritty
    #configure_vim 'Dayfox'
    ;;
  dawnfox)
    configure_alacritty
    #configure_vim 'Dawnfox'
    ;;
  *)
    echo "Supported colorschemes: nightfox, dayfox, dawnfox"
    exit 1
    ;;
esac
