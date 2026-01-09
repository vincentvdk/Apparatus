#!/usr/bin/env bash
# Apparatus First Login
# Installs flatpaks on first graphical login
# (Config setup is done by apparatus-firstboot.service)

set -e

INIT_DONE="$HOME/.local/state/apparatus/init-done"

# Exit if already initialized
if [ -f "$INIT_DONE" ]; then
    exit 0
fi

# Logo
echo '
   _____                                          __
  /  _  \  ______  ______ _____  _______ _____  _/  |_  __ __  ______
 /  /_\  \ \____ \ \____ \\__  \ \_  __ \\__  \ \   __\|  |  \/  ___/
/    |    \|  |_> >|  |_> >/ __ \_|  | \/ / __ \_|  |  |  |  /\___ \
\____|__  /|   __/ |   __/(____  /|__|   (____  /|__|  |____//____  >
        \/ |__|    |__|        \/             \/                  \/
'

echo '{{ Bold "Welcome to Apparatus OS!" }}' | gum format -t template
echo ""
echo "Installing applications..."
echo ""

# Enable Flathub repository
echo '{{ Bold "# Enabling Flathub" }}' | gum format -t template
/usr/bin/flatpak remote-add --user --if-not-exists flathub https://flathub.org/repo/flathub.flatpakrepo

# Install flatpaks with spinners
echo '{{ Bold "# Installing applications" }}' | gum format -t template
echo ""

gum spin --spinner dot --title "Installing Firefox..." -- \
    /usr/bin/flatpak install --user --noninteractive flathub org.mozilla.firefox

gum spin --spinner dot --title "Installing Signal..." -- \
    /usr/bin/flatpak install --user --noninteractive flathub org.signal.Signal

gum spin --spinner dot --title "Installing Joplin..." -- \
    /usr/bin/flatpak install --user --noninteractive flathub net.cozic.joplin_desktop

gum spin --spinner dot --title "Installing Rio Terminal..." -- \
    /usr/bin/flatpak install --user --noninteractive flathub com.rioterm.Rio

# Mark init as done
mkdir -p "$HOME"/.local/state/apparatus
touch "$INIT_DONE"

echo ""
echo '{{ Bold "# Setup complete!" }}' | gum format -t template
echo ""
echo "Press Super+D to open the application launcher."
echo "Press Super+F1 for keyboard shortcuts."
echo ""
echo "This window will close in 5 seconds..."
sleep 5
