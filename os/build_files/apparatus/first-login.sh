#!/usr/bin/env bash
# Apparatus First Login
# Interactive flatpak installation on first graphical login

INIT_DONE="$HOME/.local/state/apparatus/init-done"
LOG_FILE="$HOME/.local/state/apparatus/first-login.log"
LOCK_FILE="$HOME/.local/state/apparatus/first-login.lock"
FLATPAK_CONFIG="/usr/share/apparatus/flatpaks.conf"

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

# Prevent concurrent runs with lock file
exec 200>"$LOCK_FILE"
if ! flock -n 200; then
    log "Another instance is running, exiting"
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

# Enable Flathub repository first
/usr/bin/gum style --foreground 39 --bold "# Enabling Flathub"
log "Adding Flathub remote"
/usr/bin/flatpak remote-add --user --if-not-exists flathub https://flathub.org/repo/flathub.flatpakrepo >> "$LOG_FILE" 2>&1

echo ""
/usr/bin/gum style --foreground 39 --bold "# Select applications to install"
echo ""

# Build arrays from config file
declare -a app_ids
declare -a app_names
declare -a app_labels
declare -a preselected

while IFS='|' read -r app_id name description selected; do
    # Skip comments and empty lines
    [[ "$app_id" =~ ^#.*$ || -z "$app_id" ]] && continue

    app_ids+=("$app_id")
    app_names+=("$name")
    app_labels+=("$name - $description")

    if [[ "$selected" == "true" ]]; then
        preselected+=("$name - $description")
    fi
done < "$FLATPAK_CONFIG"

# Build preselected args for gum choose
preselect_args=()
for item in "${preselected[@]}"; do
    preselect_args+=(--selected "$item")
done

# Show interactive selection
selected_apps=$(/usr/bin/gum choose --no-limit --height 15 \
    --header "Space to select, Enter to confirm" \
    "${preselect_args[@]}" \
    "${app_labels[@]}")

# Check if user cancelled or selected nothing
if [ -z "$selected_apps" ]; then
    /usr/bin/gum style --foreground 214 "No applications selected. Skipping installation."
    log "User selected no applications"
else
    echo ""
    /usr/bin/gum style --foreground 39 --bold "# Installing selected applications"
    echo ""

    # Install selected flatpaks
    while IFS= read -r selection; do
        # Find the matching app_id
        for i in "${!app_labels[@]}"; do
            if [[ "${app_labels[$i]}" == "$selection" ]]; then
                app_id="${app_ids[$i]}"
                name="${app_names[$i]}"

                log "Installing $name ($app_id)"
                /usr/bin/gum spin --spinner dot --title "Installing $name..." -- \
                    bash -c "/usr/bin/flatpak install --user --noninteractive flathub $app_id >> '$LOG_FILE' 2>&1"
                log "Finished installing $name"
                break
            fi
        done
    done <<< "$selected_apps"
fi

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
