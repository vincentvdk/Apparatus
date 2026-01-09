#!/usr/bin/env bash
# Apparatus First Login
# Installs flatpaks on first graphical login
# (Config setup is done by apparatus-firstboot.service)

INIT_DONE="$HOME/.local/state/apparatus/init-done"
LOG_FILE="$HOME/.local/state/apparatus/first-login.log"

# Create state directory
mkdir -p "$HOME/.local/state/apparatus"

# Log function (writes to file only, not terminal)
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" >> "$LOG_FILE"
}

log "=== Apparatus First Login started ==="

# Exit if already initialized
if [ -f "$INIT_DONE" ]; then
    log "Already initialized, exiting"
    exit 0
fi

# Ensure PATH includes common locations
export PATH="/usr/bin:/usr/local/bin:$PATH"
log "PATH: $PATH"

# Logo
echo '
   _____                                          __
  /  _  \  ______  ______ _____  _______ _____  _/  |_  __ __  ______
 /  /_\  \ \____ \ \____ \\__  \ \_  __ \\__  \ \   __\|  |  \/  ___/
/    |    \|  |_> >|  |_> >/ __ \_|  | \/ / __ \_|  |  |  |  /\___ \
\____|__  /|   __/ |   __/(____  /|__|   (____  /|__|  |____//____  >
        \/ |__|    |__|        \/             \/                  \/
'

/usr/bin/gum style --foreground 212 --bold "Welcome to Apparatus OS!"
echo ""
echo "Installing applications..."
echo ""

# Enable Flathub repository
/usr/bin/gum style --foreground 39 --bold "# Enabling Flathub"
log "Adding Flathub remote"
/usr/bin/flatpak remote-add --user --if-not-exists flathub https://flathub.org/repo/flathub.flatpakrepo >> "$LOG_FILE" 2>&1

# Install flatpaks with spinners (redirect output to log)
/usr/bin/gum style --foreground 39 --bold "# Installing applications"
echo ""

install_flatpak() {
    local name="$1"
    local app_id="$2"
    log "Installing $name ($app_id)"
    /usr/bin/gum spin --spinner dot --title "Installing $name..." -- \
        bash -c "/usr/bin/flatpak install --user --noninteractive flathub $app_id >> '$LOG_FILE' 2>&1"
    log "Finished installing $name"
}

install_flatpak "Firefox" "org.mozilla.firefox"
install_flatpak "Signal" "org.signal.Signal"
install_flatpak "Joplin" "net.cozic.joplin_desktop"
install_flatpak "Rio Terminal" "com.rioterm.Rio"

# Mark init as done
touch "$INIT_DONE"
log "First login setup complete"

echo ""
/usr/bin/gum style --foreground 76 --bold "# Setup complete!"
echo ""
echo "Press Super+D to open the application launcher."
echo "Press Super+F1 for keyboard shortcuts."
echo ""
read -rp "Press Enter to close..."
