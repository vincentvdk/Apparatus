# Troubleshooting

This guide covers common issues you might encounter with Apparatus OS and how to resolve them.

## System Issues

### System Won't Boot

**Symptom**: System fails to boot after an update.

**Solutions**:

1. **Roll back to previous deployment**:
   ```bash
   # Bootc manages deployments automatically
   # Reboot to roll back to the previous deployment
   systemctl reboot
   ```

2. **Check bootc status**:
   ```bash
   bootc status
   bootc fetch-apply-updates
   ```

3. **Check journal logs**:
   ```bash
   journalctl -xb
   ```

### No Display / Black Screen

**Symptom**: System boots but display is black or frozen.

**Solutions**:

1. **Wait and retry**: Sometimes Hyprland takes a moment to start
2. **Check display manager**:
   ```bash
   systemctl status greetd
   ```
3. **Switch to TTY**:
   - Press Ctrl+Alt+F2 to switch to TTY2
   - Login and check logs:
     ```bash
     journalctl -u greetd -b
     journalctl -u hyprland -b
     ```
4. **Reinstall Hyprland**:
   ```bash
   sudo dnf reinstall hyprland
   ```

### Hyprland Not Starting

**Symptom**: Hyprland fails to start, returns to greetd.

**Solutions**:

1. **Check Hyprland logs**:
   ```bash
   journalctl -u hyprland-service -b
   ```

2. **Test Hyprland manually**:
   ```bash
   Hyprland
   ```
   Check the output for errors.

3. **Check configuration syntax**:
   ```bash
   hyprctl config
   ```
   Look for syntax errors.

4. **Reset configuration**:
   ```bash
   mv ~/.config/hypr ~/.config/hypr.bak
   cp -r /usr/share/apparatus/hypr ~/.config/
   hyprctl reload
   ```

### Keyboard/Input Not Working

**Symptom**: Keyboard or mouse not responding.

**Solutions**:

1. **Check input devices**:
   ```bash
   ls /dev/input/
   ```

2. **Restart input services**:
   ```bash
   systemctl restart libinput
   ```

3. **Check Wayland permissions**:
   ```bash
   echo $XDG_SESSION_TYPE
   # Should be "wayland"
   ```

4. **Try a different TTY**:
   - Press Ctrl+Alt+F2 to switch to TTY
   - Test if keyboard works there

### Audio Not Working

**Symptom**: No sound output or input.

**Solutions**:

1. **Check PipeWire status**:
   ```bash
   systemctl --user status pipewire
   systemctl --user status wireplumber
   ```

2. **Restart audio services**:
   ```bash
   systemctl --user restart pipewire wireplumber
   ```

3. **Check volume**:
   ```bash
   pactl get-sink-volume @DEFAULT_SINK@
   pactl get-sink-mute @DEFAULT_SINK@
   ```

4. **Launch volume control**:
   ```bash
   pavucontrol
   ```

5. **Check ALSA**:
   ```bash
   alsamixer
   ```
   Ensure channels are not muted (MM = muted, OO = unmuted)

### Network Not Working

**Symptom**: No internet connectivity.

**Solutions**:

1. **Check NetworkManager**:
   ```bash
   systemctl status NetworkManager
   ```

2. **Restart NetworkManager**:
   ```bash
   sudo systemctl restart NetworkManager
   ```

3. **Check connection**:
   ```bash
   ip a
   ping google.com
   ```

4. **WiFi troubleshooting**:
   ```bash
   # List WiFi networks
   nmcli device wifi list
   
   # Connect to network
   nmcli device wifi connect "SSID" password "password"
   ```

5. **Check DNS**:
   ```bash
   cat /etc/resolv.conf
   nslookup google.com
   ```

## Display and Graphics

### Multiple Monitors Not Working

**Symptom**: Second monitor not detected or configured incorrectly.

**Solutions**:

1. **Launch monitor configuration**:
   ```bash
   butler
   # Select "Configure" > "Monitors"
   ```

2. **Use hyprdynamicmonitors**:
   ```bash
   hyprdynamicmonitors tui
   ```

3. **Check connected monitors**:
   ```bash
   hyprctl monitors
   ```

4. **Manual configuration**:
   Edit `~/.config/hypr/hyprland.lua`:
   ```lua
   monitor = DP-1, prefered, auto, auto
   monitor = HDMI-A-1, right, DP-1, auto, auto
   ```
   Then reload:
   ```bash
   hyprctl reload
   ```

### Screen Tearing

**Symptom**: Visual tearing when scrolling or moving windows.

**Solutions**:

1. **Enable VSYNC in Hyprland**:
   ```lua
   # In ~/.config/hypr/hyprland.lua
   vsync = true
   ```

2. **Check for hardware acceleration**:
   ```bash
   glxinfo | grep -i renderer
   ```

3. **Enable AMD GPU settings** (if applicable):
   ```bash
   # In ~/.config/hypr/hyprland.lua
   env = AMD_VULKAN_ICD,loader
   env = MESA_LOADER_DRIVER_OVERRIDE,zink
   ```

### Resolution Issues

**Symptom**: Wrong resolution or scaling.

**Solutions**:

1. **Check available modes**:
   ```bash
   hyprctl monitors
   ```

2. **Set resolution manually**:
   ```lua
   # In ~/.config/hypr/hyprland.lua
   monitor = eDP-1, 1920x1080@144, auto, auto
   ```

3. **Check EDID**:
   ```bash
   sudo cat /sys/class/drm/card0-*/edid | edid-decode
   ```

## Application Issues

### Kitty Terminal Issues

**Symptom**: Kitty not starting or displaying incorrectly.

**Solutions**:

1. **Check Kitty logs**:
   ```bash
   kitty --debug-config
   ```

2. **Reset configuration**:
   ```bash
   mv ~/.config/kitty ~/.config/kitty.bak
   cp -r /usr/share/apparatus/kitty ~/.config/
   ```

3. **Check font availability**:
   ```bash
   fc-list | grep -i "jetbrains"
   ```

4. **Test with default config**:
   ```bash
   kitty --config /dev/null
   ```

### Waybar Not Showing

**Symptom**: Waybar status bar is missing.

**Solutions**:

1. **Check if Waybar is running**:
   ```bash
   pgrep waybar
   ```

2. **Restart Waybar**:
   ```bash
   pkill waybar
   waybar &
   ```

3. **Check configuration syntax**:
   ```bash
   waybar --validate-config
   ```

4. **Check logs**:
   ```bash
   journalctl -u waybar -b
   ```

5. **Reset configuration**:
   ```bash
   mv ~/.config/waybar ~/.config/waybar.bak
   cp -r /usr/share/apparatus/waybar ~/.config/
   ```

### Mako Notifications Not Working

**Symptom**: Notifications not appearing.

**Solutions**:

1. **Check if Mako is running**:
   ```bash
   pgrep mako
   ```

2. **Restart Mako**:
   ```bash
   makoctl close-all
   makoctl reload
   ```

3. **Test notification**:
   ```bash
   notify-send "Test" "This is a test notification"
   ```

4. **Check configuration**:
   ```bash
   cat ~/.config/mako/config
   ```

### Satty Screenshots Not Working

**Symptom**: Satty not launching or screenshots not saving.

**Solutions**:

1. **Check if Satty is installed**:
   ```bash
   which satty
   ```

2. **Check Satty logs**:
   ```bash
   satty --debug
   ```

3. **Check output directory**:
   ```bash
   cat ~/.config/satty/config.toml | grep output_dir
   ls -la ~/Pictures/Screenshots/
   ```

4. **Test with default config**:
   ```bash
   mv ~/.config/satty ~/.config/satty.bak
   cp -r /usr/share/apparatus/satty ~/.config/
   ```

## Distrobox Issues

### Distrobox Not Creating Containers

**Symptom**: `distrobox create` fails.

**Solutions**:

1. **Check Podman**:
   ```bash
   podman --version
   systemctl --user status podman.socket
   ```

2. **Start Podman socket**:
   ```bash
   systemctl --user start podman.socket
   systemctl --user enable podman.socket
   ```

3. **Check SELinux**:
   ```bash
   setenforce 0  # Temporarily disable (for testing)
   ```
   If this fixes it, you may need to adjust SELinux policies.

4. **Check storage**:
   ```bash
   df -h
   podman info | grep storage
   ```

### Distrobox Enter Fails

**Symptom**: `distrobox enter` fails with errors.

**Solutions**:

1. **Check container status**:
   ```bash
   distrobox list
   podman ps -a
   ```

2. **Start the container first**:
   ```bash
   distrobox start mybox
   ```

3. **Check for typos**:
   Apparatus has typo protection. If you made a typo:
   ```bash
   distrobox enter mybox  # Correct
   # If you type "myboc", it will show available containers
   ```

4. **Check container logs**:
   ```bash
   podman logs mybox
   ```

### Distrobox GUI Apps Not Displaying

**Symptom**: GUI applications from distrobox don't appear on host.

**Solutions**:

1. **Check XDG environment**:
   ```bash
   echo $XDG_RUNTIME_DIR
   echo $WAYLAND_DISPLAY
   echo $DISPLAY
   ```

2. **Enter with proper environment**:
   ```bash
   distrobox enter mybox --export DISPLAY=$DISPLAY --export WAYLAND_DISPLAY=$WAYLAND_DISPLAY
   ```

3. **Install required packages in container**:
   ```bash
   distrobox enter mybox
   sudo dnf install xorg-x11-server-Xwayland
   ```

### Distrobox Audio Not Working

**Symptom**: Audio from distrobox not playing on host speakers.

**Solutions**:

1. **Check PipeWire in container**:
   ```bash
   distrobox enter mybox
   pactl info
   ```

2. **Check host PipeWire**:
   ```bash
   systemctl --user status pipewire
   ```

3. **Restart PipeWire on host**:
   ```bash
   systemctl --user restart pipewire wireplumber
   ```

## Configuration Issues

### Theme Not Applying

**Symptom**: Theme changes not taking effect.

**Solutions**:

1. **Check current theme**:
   ```bash
   cat ~/.config/apparatus/current-theme
   ```

2. **Verify symlinks**:
   ```bash
   ls -la ~/.config/kitty/theme.conf
   readlink ~/.config/kitty/theme.conf
   ```

3. **Reload applications**:
   ```bash
   hyprctl reload
   pkill -SIGUSR1 kitty
   makoctl reload
   ```

4. **Reapply theme via Butler**:
   ```bash
   butler theme catppuccin-mocha
   ```

### Config Sync Not Working

**Symptom**: `butler sync` fails or doesn't update distroboxes.

**Solutions**:

1. **Check distroboxes exist**:
   ```bash
   distrobox list
   ```

2. **Check home directory permissions**:
   ```bash
   ls -la ~/distrobox-homes/
   ```

3. **Check source configs exist**:
   ```bash
   ls /usr/share/apparatus/hypr/
   ```

4. **Run with debug output**:
   ```bash
   butler sync 2>&1 | tee sync-debug.log
   ```

5. **Check skip list**:
   ```bash
   cat ~/.config/apparatus/skip-distroboxes
   ```

### Font Not Changing

**Symptom**: Font changes not applying in applications.

**Solutions**:

1. **Check current font**:
   ```bash
   cat ~/.config/apparatus/current-font
   ```

2. **Verify font is installed**:
   ```bash
   fc-list | grep -i "jetbrains"
   ```

3. **Reload applications**:
   ```bash
   pkill -SIGUSR1 kitty
   hyprctl reload
   makoctl reload
   ```

4. **Reapply font via Butler**:
   ```bash
   butler font jetbrains-mono
   ```

## Butler Issues

### Butler Not Found

**Symptom**: `butler` command not found.

**Solutions**:

1. **Check installation path**:
   ```bash
   ls /usr/local/bin/butler
   ```

2. **Check PATH**:
   ```bash
   echo $PATH
   which butler
   ```

3. **Add to PATH**:
   ```bash
   export PATH=$PATH:/usr/local/bin
   ```
   Add to `~/.bashrc` or `~/.zshrc` for persistence.

### Butler TUI Not Working

**Symptom**: Butler TUI crashes or doesn't display properly.

**Solutions**:

1. **Check for required packages**:
   ```bash
   which gum
   gum --version
   ```

2. **Install gum**:
   ```bash
   sudo dnf install https://github.com/charmbracelet/gum/releases/download/v0.17.0/gum-0.17.0-1.x86_64.rpm
   ```

3. **Try CLI mode**:
   ```bash
   butler theme catppuccin-mocha
   ```

### Butler Auto-Detection Fails

**Symptom**: Butler can't detect OS or box mode.

**Solutions**:

1. **Set mode explicitly**:
   ```bash
   APPARATUS_MODE=os butler
   APPARATUS_MODE=box butler
   ```

2. **Check environment markers**:
   ```bash
   echo $APPARATUS_OS_HOME
   cat /run/.containerenv 2>/dev/null
   which mise 2>/dev/null
   which hyprctl 2>/dev/null
   ```

## Hardware Issues

### Bluetooth Not Working

**Symptom**: Bluetooth devices not detected.

**Solutions**:

1. **Check Bluetooth service**:
   ```bash
   systemctl status blueman
   ```

2. **Start Bluetooth**:
   ```bash
   sudo systemctl start bluetooth
   sudo systemctl enable bluetooth
   ```

3. **Check Bluetooth status**:
   ```bash
   bluetoothctl status
   ```

4. **Enable Bluetooth**:
   ```bash
   bluetoothctl power on
   bluetoothctl agent on
   ```

### Touchpad Not Working

**Symptom**: Touchpad not responding.

**Solutions**:

1. **Check libinput**:
   ```bash
   libinput list-devices
   ```

2. **Check touchpad detection**:
   ```bash
   xinput list
   ```

3. **Enable touchpad**:
   ```bash
   # Find touchpad ID
   xinput list | grep -i touchpad
   
   # Enable
   xinput enable <device-id>
   ```

4. **Check Wayland touchpad**:
   ```bash
   libinput debug-events --show-keycodes
   ```

## Logging and Debugging

### System Logs

```bash
# View system journal
journalctl -b

# View specific service
journalctl -u greetd -b
journalctl -u hyprland -b

# Follow logs in real-time
journalctl -f
```

### Application Logs

```bash
# Hyprland logs
cat ~/.cache/hypr/hyprland.log

# Waybar logs
journalctl -u waybar -b

# Podman logs
journalctl -u podman -b
```

### Debug Mode

Run applications in debug mode:

```bash
# Hyprland debug
Hyprland --debug

# Kitty debug
kitty --debug-config --debug-log

# Waybar debug
waybar --debug
```

## Getting Help

If you've tried the above solutions and still have issues:

1. **Search existing issues**: Check the GitHub Issues for similar problems
2. **Create a new issue**: Include:
   - Your system information (`neofetch` or `inxi -F`)
   - Steps to reproduce the issue
   - Relevant log output
   - What you've tried so far
3. **Ask in discussions**: Join the GitHub Discussions for questions and help
4. **Check updates**: Ensure you're running the latest version

```bash
# Check Apparatus OS version
cat /etc/os-release | grep VERSION

# Check for updates
bootc status
```

---

*For more information, see the main documentation guides.*
