package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// OS Mode items
type osItem struct {
	title       string
	description string
	action      string
}

func (i osItem) Title() string       { return i.title }
func (i osItem) Description() string { return i.description }
func (i osItem) FilterValue() string { return i.title }

// Popup modes
const (
	popupModeNone = iota
	popupModeDistrobox
	popupModeConfigure
	popupModeDistroboxActions
	popupModeUpgrade
	popupModeTheme
	popupModeFont
	popupModeCreateName
	popupModeCreateHome
	popupModeCreateCustom
	popupModeHelp
	popupModeOutput
	popupModeSync
	popupModeSyncResults
)

// OS Mode model
type osModel struct {
	width, height int
	quitting      bool
	showPopup     bool
	popupMode     int
	popupTitle    string
	popupContent  string
	popupIsError  bool
	popupList     list.Model

	// Distrobox state
	selectedDistrobox string
	createName        string
	createCustomHome  string

	// Text input for create flow
	textInput textinput.Model

	running     bool
	spinner     spinner.Model
	runningText string

	// Main list
	mainList list.Model

	// Current theme/font, for the status line -- read once at startup and
	// refreshed whenever applyTheme()/applyFontAction() succeed.
	currentTheme string
	currentFont  string
}

func newOSModel() *osModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = getSpinnerStyle()

	m := &osModel{
		mainList:     makeMainList(),
		spinner:      s,
		currentTheme: currentThemeName(),
		currentFont:  currentFontName(),
	}
	return m
}

func makeMainList() list.Model {
	items := []list.Item{
		osItem{title: "📦 Distrobox", description: "Manage distrobox containers", action: "distrobox"},
		osItem{title: "⚙️  Configure", description: "Hyprland, terminal, monitors, audio, AI", action: "configure"},
		osItem{title: "🎨 Theme", description: "Switch color theme (Catppuccin)", action: "theme"},
		osItem{title: "🔤 Font", description: "Switch system font", action: "font"},
		osItem{title: "🔄 Sync Configs", description: "Sync configs from host to all distroboxes", action: "sync-configs"},
		osItem{title: "⬆ Push Configs", description: "Push config changes from distroboxes back to host (chezmoi)", action: "push-configs"},
		osItem{title: "🚫 Skip List", description: "Manage distrobox skip list for config sync", action: "skip-list"},
		osItem{title: "⌨️  Help", description: "Keyboard shortcuts and keybindings", action: "help"},
	}

	delegate := getListDelegate()
	delegate.SetHeight(1)

	l := newList(items, delegate, 0, 0)
	l.Title = "Butler Menu"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	return l
}

func (m *osModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *osModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle popup mode
	if m.showPopup {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			// Handle popup list
			if m.popupMode == popupModeDistrobox || m.popupMode == popupModeConfigure || m.popupMode == popupModeDistroboxActions || m.popupMode == popupModeUpgrade || m.popupMode == popupModeTheme || m.popupMode == popupModeFont {
				switch msg.String() {
				case "enter":
					if m.popupList.SelectedItem() != nil {
						if item, ok := m.popupList.SelectedItem().(osItem); ok {
							return m, m.handlePopupAction(item)
						}
					}
					// Fallback: close popup if no valid item selected
					m.closePopup()
					return m, nil
				case "esc", "q":
					m.closePopup()
					return m, nil
				}
				var cmd tea.Cmd
				m.popupList, cmd = m.popupList.Update(msg)
				return m, cmd
			}

			// Handle text input in popup
			if m.popupMode == popupModeCreateName || m.popupMode == popupModeCreateCustom {
				var cmd tea.Cmd
				m.textInput, cmd = m.textInput.Update(msg)

				// Handle enter
				if msg.String() == "enter" {
					if m.popupMode == popupModeCreateCustom {
						m.createCustomHome = m.textInput.Value()
						if m.createCustomHome == "" {
							m.createCustomHome = m.createName
						}
						m.textInput.Blur()
						return m, m.startCommand(fmt.Sprintf("creating distrobox: %s", m.createName), m.executeCreateDistrobox("custom"))
					} else if m.popupMode == popupModeCreateName {
						m.createName = m.textInput.Value()
						if m.createName == "" {
							m.closePopup()
							return m, nil
						}
						m.showCreateHomePopup()
						return m, nil
					}
				}

				// Handle esc
				if msg.String() == "esc" || msg.String() == "q" {
					m.textInput.Blur()
					m.closePopup()
					return m, nil
				}

				return m, cmd
			}

			// Handle number keys for create-home popup
			if m.popupMode == popupModeCreateHome {
				switch msg.String() {
				case "1":
					return m, m.startCommand(fmt.Sprintf("creating distrobox: %s", m.createName), m.executeCreateDistrobox("default"))
				case "2":
					m.showCreateCustomPopup()
					return m, nil
				case "esc", "q":
					m.closePopup()
					return m, nil
				}
			}

			// Handle esc/q for output/help popups
			if msg.String() == "esc" || msg.String() == "q" {
				m.closePopup()
				return m, nil
			}
		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height
			if m.popupMode == popupModeDistrobox || m.popupMode == popupModeConfigure || m.popupMode == popupModeDistroboxActions || m.popupMode == popupModeUpgrade || m.popupMode == popupModeTheme || m.popupMode == popupModeFont {
				m.popupList.SetSize(msg.Width/2, msg.Height/2)
			}
			return m, nil
		}
		return m, nil
	}

	// Handle command completion (see startCommand -- commands run as real
	// tea.Cmd values so the spinner keeps animating and the UI never blocks)
	if msg, ok := msg.(commandDoneMsg); ok {
		m.running = false
		m.runningText = ""
		if msg.err != nil {
			m.showErrorPopup("Error", fmt.Sprintf("Error: %v\n\n%s", msg.err, msg.output))
		} else {
			m.showInfoPopup("Output", msg.output)
		}
		return m, nil
	}

	// If running a command, handle spinner only
	if m.running {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if win, ok := msg.(tea.WindowSizeMsg); ok {
			m.width = win.Width
			m.height = win.Height
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if item, ok := m.mainList.SelectedItem().(osItem); ok {
				return m, m.handleMainAction(item)
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.mainList.SetSize(msg.Width-4, msg.Height-10)
	}

	var cmd tea.Cmd
	m.mainList, cmd = m.mainList.Update(msg)
	return m, cmd
}

// showInfoPopup and showErrorPopup are the two ways to display an output
// popup. Keeping them as separate, named entry points (rather than a raw
// m.showPopup/m.popupTitle/... assignment at each call site) makes the
// error/success intent explicit everywhere and drives the popup border
// color (see getPopupStyle).
func (m *osModel) showInfoPopup(title, content string) {
	m.closePopup()
	m.showPopup = true
	m.popupMode = popupModeOutput
	m.popupTitle = title
	m.popupContent = content
	m.popupIsError = false
}

func (m *osModel) showErrorPopup(title, content string) {
	m.closePopup()
	m.showPopup = true
	m.popupMode = popupModeOutput
	m.popupTitle = title
	m.popupContent = content
	m.popupIsError = true
}

// startCommand closes any open popup, switches to the running/spinner state,
// and returns cmd batched with the spinner tick so bubbletea executes it as
// a real background command instead of blocking Update(). Completion is
// picked up by the commandDoneMsg case in Update().
func (m *osModel) startCommand(label string, cmd tea.Cmd) tea.Cmd {
	m.closePopup()
	m.running = true
	m.runningText = label
	return tea.Batch(cmd, m.spinner.Tick)
}

func (m *osModel) handleMainAction(item osItem) tea.Cmd {
	switch item.action {
	case "distrobox":
		m.showDistroboxMenu()
	case "configure":
		m.showConfigureMenu()
	case "theme":
		m.showThemePopup()
	case "font":
		m.showFontPopup()
	case "sync-configs":
		return m.syncDistroboxConfigs()
	case "push-configs":
		return m.pushDistroboxConfigs()
	case "skip-list":
		m.showSkipList()
	case "help":
		m.showHelp()
	}
	return nil
}

func (m *osModel) applyTheme(themeName string) {
	// Apply the theme first
	err := applyTheme(themeName)

	// Close the theme selection popup
	m.showPopup = false
	m.popupMode = popupModeNone
	m.popupList = newList([]list.Item{}, getListDelegate(), 0, 0)

	if err != nil {
		m.showErrorPopup("Theme", fmt.Sprintf("Failed to apply theme '%s': %v", themeName, err))
		return
	}
	m.currentTheme = themeName
}

func (m *osModel) applyFontAction(fontName string) {
	fontMap := map[string]string{
		"ioskeley-mono":    "IoskeleyMono Nerd Font",
		"jetbrains-mono":   "JetBrainsMono Nerd Font",
		"hack-nerd-font":   "Hack Nerd Font",
	}
	fontFamily, ok := fontMap[fontName]
	if !ok {
		// Close font selection popup
		m.showPopup = false
		m.popupMode = popupModeNone
		m.popupList = newList([]list.Item{}, getListDelegate(), 0, 0)
		return
	}
	applyFont(fontFamily, fontName)
	m.currentFont = fontName
	// Close the font selection popup
	m.showPopup = false
	m.popupMode = popupModeNone
	m.popupList = newList([]list.Item{}, getListDelegate(), 0, 0)
}

func (m *osModel) showThemePopup() {
	themesDir := getThemesDir()
	entries, err := os.ReadDir(themesDir)
	if err != nil {
		m.showErrorPopup("Theme Selection", fmt.Sprintf("Could not read themes directory: %v", err))
		return
	}

	var items []list.Item
	for _, entry := range entries {
		if entry.IsDir() {
			label := entry.Name()
			desc := "light"
			if themeIsDark(label) {
				desc = "dark"
			}
			items = append(items, osItem{
				title:       fmt.Sprintf("🎨 %s", label),
				description: desc,
				action:      label,
			})
		}
	}

	if len(items) == 0 {
		m.showErrorPopup("Theme Selection", "No themes found in "+themesDir)
		return
	}

	delegate := getListDelegate()
	delegate.SetHeight(1)
	m.popupList = newList(items, delegate, 0, 0)
	m.popupList.Title = "Themes"
	m.popupList.SetShowStatusBar(false)
	m.popupList.SetFilteringEnabled(true)
	m.popupList.SetShowHelp(true)
	m.popupList.SetSize(m.width/2, m.height/2)
	m.popupMode = popupModeTheme
	m.showPopup = true
}

func (m *osModel) showFontPopup() {
	var fontItems []list.Item
	fontItems = append(fontItems, osItem{title: "🔤 Ioskeley Mono", description: "Berkeley Mono alternative", action: "ioskeley-mono"})
	fontItems = append(fontItems, osItem{title: "🔤 JetBrains Mono", description: "Default font", action: "jetbrains-mono"})
	fontItems = append(fontItems, osItem{title: "🔤 Hack Nerd Font", description: "Classic monospace", action: "hack-nerd-font"})

	delegate := getListDelegate()
	delegate.SetHeight(1)
	m.popupList = newList(fontItems, delegate, 0, 0)
	m.popupList.Title = "Fonts"
	m.popupList.SetShowStatusBar(false)
	m.popupList.SetFilteringEnabled(true)
	m.popupList.SetShowHelp(true)
	m.popupList.SetSize(m.width/2, m.height/2)
	m.popupMode = popupModeFont
	m.showPopup = true
}

func (m *osModel) showTerminalPopup() {
	m.showInfoPopup("Terminal", "1. kitty\n\nRun: butler --os terminal <name>\n\nOr edit: ~/.config/hypr/hyprland.lua")
}

func (m *osModel) launchMonitors() {
	m.closePopup()
	_, err := exec.LookPath("hyprdynamicmonitors")
	if err != nil {
		m.showErrorPopup("Monitors", "hyprdynamicmonitors not found.\n\nInstall it first, then run:\n  hyprdynamicmonitors tui")
		return
	}
	// Launch TUI directly - takes over terminal
	cmd := exec.Command("hyprdynamicmonitors", "tui")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if runErr := cmd.Run(); runErr != nil {
		m.showErrorPopup("Monitors", fmt.Sprintf("hyprdynamicmonitors exited with error:\n%v", runErr))
	}
}

func (m *osModel) launchSoundctl() {
	m.closePopup()
	// Launch soundctl TUI directly
	cmd := exec.Command("soundctl")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if runErr := cmd.Run(); runErr != nil {
		m.showErrorPopup("Sound Control", fmt.Sprintf("soundctl exited with error:\n%v", runErr))
	}
}

func (m *osModel) showAIWorkloadPopup() {
	m.showInfoPopup("AI Workload", "VRAM Allocation\n\nSelect from the menu:\n  16 GB, 32 GB, 64 GB, 96 GB (max stable)\n  Reset to default\n\nConfigures amdttm.kernel_cmdline for AMD Ryzen AI APUs.\nRequires reboot to take effect.")
}

func (m *osModel) showDistroboxMenu() {
	output, err := runCommand("distrobox", "list", "--no-color")

	var items []list.Item
	if err != nil {
		items = []list.Item{
			osItem{title: "Create", description: "Create a new distrobox", action: "create"},
			osItem{title: "Upgrade", description: "Upgrade distrobox [BETA]", action: "upgrade"},
			osItem{title: "Error", description: fmt.Sprintf("Failed to list distroboxes: %v", err), action: "none"},
		}
	} else {
		// Add action items at the top
		items = append(items,
			osItem{title: "Create", description: "Create a new distrobox", action: "create"},
			osItem{title: "Upgrade", description: "Upgrade distrobox [BETA]", action: "upgrade"},
		)

		lines := strings.Split(strings.TrimSpace(output), "\n")
		for _, line := range lines {
			if line == "" || strings.Contains(line, "ID") {
				continue
			}
			parts := strings.Split(line, "|")
			if len(parts) >= 4 {
				name := strings.TrimSpace(parts[1])
				image := strings.TrimSpace(parts[3])
				items = append(items, osItem{
					title:       fmt.Sprintf("📦 %s", name),
					description: image,
					action:      name,
				})
			}
		}

		if len(items) == 2 { // Only action items, no distroboxes
			items = append(items, osItem{title: "(no distroboxes found)", description: "Create one with the action above", action: "none"})
		}
	}

	delegate := getListDelegate()
	delegate.SetHeight(1)
	m.popupList = newList(items, delegate, 0, 0)
	m.popupList.Title = "Distroboxes"
	m.popupList.SetShowStatusBar(false)
	m.popupList.SetFilteringEnabled(true)
	m.popupList.SetShowHelp(true)
	m.popupList.SetSize(m.width/2, m.height/2)
	m.popupMode = popupModeDistrobox
	m.showPopup = true
}

func (m *osModel) showConfigureMenu() {
	items := []list.Item{
		osItem{title: "Terminal", description: "Set default terminal", action: "terminal"},
		osItem{title: "Monitors", description: "Configure display setup", action: "monitors"},
		osItem{title: "Audio", description: "Configure audio devices", action: "audio"},
		osItem{title: "AI Workload", description: "GPU VRAM allocation for AI/ML", action: "ai-workload"},
	}

	delegate := getListDelegate()
	delegate.SetHeight(1)
	m.popupList = newList(items, delegate, 0, 0)
	m.popupList.Title = "Configure"
	m.popupList.SetShowStatusBar(false)
	m.popupList.SetFilteringEnabled(true)
	m.popupList.SetShowHelp(true)
	m.popupList.SetSize(m.width/2, m.height/2)
	m.popupMode = popupModeConfigure
	m.showPopup = true
}

func (m *osModel) handlePopupAction(item osItem) tea.Cmd {
	switch m.popupMode {
	case popupModeDistrobox:
		switch item.action {
		case "create":
			m.showCreateNamePopup()
		case "upgrade":
			m.showUpgradeMenu()
		default:
			m.selectedDistrobox = item.action
			m.showDistroboxActions()
		}
	case popupModeConfigure:
		switch item.action {
		case "terminal":
			m.showTerminalPopup()
		case "monitors":
			m.launchMonitors()
		case "audio":
			m.launchSoundctl()
		case "ai-workload":
			m.showAIWorkloadPopup()
		}
	case popupModeDistroboxActions:
		// item.action is one of "enter"/"upgrade"/"stop"/"remove", which is
		// exactly what executeDistroboxAction expects.
		return m.startCommand(fmt.Sprintf("%s: %s", item.action, m.selectedDistrobox), m.executeDistroboxAction(item.action))
	case popupModeUpgrade:
		// item.action is "upgrade-packages" or "upgrade-image".
		return m.startCommand(fmt.Sprintf("%s: %s", item.action, m.selectedDistrobox), m.executeUpgradeDistrobox(item.action))
	case popupModeTheme:
		m.applyTheme(item.action)
		// Ensure popup is closed
		m.showPopup = false
		m.popupMode = popupModeNone
	case popupModeFont:
		m.applyFontAction(item.action)
		// Ensure popup is closed
		m.showPopup = false
		m.popupMode = popupModeNone
	case popupModeSync:
		// Skip list: toggle the selected distrobox
		if item.action != "" && item.action != "none" {
			m.handleSkipListToggle(item.action)
		}
	}
	return nil
}

func (m *osModel) showUpgradeMenu() {
	items := []list.Item{
		osItem{title: "packages", description: "Update packages inside container", action: "upgrade-packages"},
		osItem{title: "image", description: "Recreate with latest image (keeps home dir)", action: "upgrade-image"},
	}

	delegate := getListDelegate()
	delegate.SetHeight(1)
	m.popupList = newList(items, delegate, 0, 0)
	m.popupList.Title = "Upgrade Type"
	m.popupList.SetShowStatusBar(false)
	m.popupList.SetFilteringEnabled(true)
	m.popupList.SetShowHelp(true)
	m.popupList.SetSize(m.width/2, m.height/2)
	m.popupMode = popupModeUpgrade
}

func (m *osModel) showDistroboxActions() {
	items := []list.Item{
		osItem{title: "enter", description: "Enter the distrobox", action: "enter"},
		osItem{title: "upgrade", description: "Upgrade packages inside container", action: "upgrade"},
		osItem{title: "stop", description: "Stop the distrobox", action: "stop"},
		osItem{title: "remove", description: "Remove the distrobox", action: "remove"},
	}

	delegate := getListDelegate()
	delegate.SetHeight(1)
	m.popupList = newList(items, delegate, 0, 0)
	m.popupList.Title = "Actions"
	m.popupList.SetShowStatusBar(false)
	m.popupList.SetFilteringEnabled(true)
	m.popupList.SetShowHelp(true)
	m.popupList.SetSize(m.width/2, m.height/2)
	m.popupMode = popupModeDistroboxActions
}

func (m *osModel) showCreateNamePopup() {
	m.createName = ""
	m.createCustomHome = ""
	m.textInput = textinput.New()
	m.textInput.Placeholder = "distrobox-name"
	m.textInput.Focus()
	m.textInput.CharLimit = 64
	m.popupTitle = "Create Distrobox"
	m.popupContent = "Enter name for new distrobox:"
	m.popupMode = popupModeCreateName
	m.showPopup = true
}

func (m *osModel) showCreateHomePopup() {
	m.popupMode = popupModeCreateHome
	m.popupTitle = "Home Folder"
	m.popupContent = fmt.Sprintf("Distrobox: %s\n\n1. Default home (~)\n2. Custom home folder\n\n[1/2 • Esc to cancel]", m.createName)
	m.showPopup = true
}

func (m *osModel) showCreateCustomPopup() {
	m.createCustomHome = ""
	m.textInput = textinput.New()
	m.textInput.Placeholder = "folder-name"
	m.textInput.Focus()
	m.textInput.CharLimit = 64
	m.popupTitle = "Custom Home Folder"
	m.popupContent = fmt.Sprintf("Enter folder name (in ~/distrobox-homes/)\n\nDefault: %s", m.createName)
	m.popupMode = popupModeCreateCustom
	m.showPopup = true
}

func (m *osModel) showHelp() {
	m.showPopup = true
	m.popupMode = popupModeHelp
	m.popupTitle = "Keyboard Shortcuts"
	m.popupContent = `Apparatus Keyboard Shortcuts

General
  Super+Return   Open terminal (kitty)
  Super+D        Application launcher (walker)
  Super+E        File manager (thunar)
  Super+Q        Close window
  Super+Shift+Q  Exit Hyprland
  Super+V        Toggle floating
  Super+F        Fullscreen
  Super+L        Lock screen
  Super+F1       Show this help

Window Navigation
  Super+←/→/↑/↓  Move focus
  Super+H/J/K/L  Move focus (vim keys)

Workspaces
  Super+1-9,0    Switch workspace
  Super+Shift+1-9,0  Move window to workspace
  Super+Scroll   Cycle workspaces
  Super+S        Toggle scratchpad

Screenshots (with Satty)
  Print          Full screenshot → Satty
  Shift+Print    Window screenshot → Satty
  Super+Shift+S  Region screenshot → Satty

---
Press q/esc/enter to close
`
}

func (m *osModel) closePopup() {
	m.showPopup = false
	m.popupMode = popupModeNone
	m.popupList = newList([]list.Item{}, getListDelegate(), 0, 0)
	m.textInput.Blur()
}

func (m *osModel) executeDistroboxAction(action string) tea.Cmd {
	name := m.selectedDistrobox
	if name == "" {
		return func() tea.Msg {
			return commandDoneMsg{output: "No distrobox selected.", err: fmt.Errorf("no distrobox selected"), showOutput: true}
		}
	}

	return func() tea.Msg {
		var output string
		var err error

		switch action {
		case "upgrade":
			output, err = runCommand("distrobox", "upgrade", name)
		case "enter":
			output, err = runCommand("distrobox", "enter", name)
		case "stop":
			output, err = runCommand("distrobox", "stop", name, "--yes")
		case "remove":
			output, err = runCommand("distrobox", "rm", name, "--force")
		}

		if err != nil {
			output = fmt.Sprintf("%s\nError: %v", output, err)
		}

		return commandDoneMsg{output: output, err: err, showOutput: true}
	}
}

// assembleIniPath returns the user's distrobox-assemble manifest
func assembleIniPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "distrobox.ini"
	}
	return filepath.Join(home, ".config", "distrobox", "distrobox.ini")
}

// assembleCreateBox adds a section for name to the user's distrobox.ini (if
// missing) and runs distrobox-assemble to create the box. homeDir may be
// empty for a shared-home box.
func assembleCreateBox(name, homeDir string) (string, error) {
	ini := assembleIniPath()
	if err := os.MkdirAll(filepath.Dir(ini), 0755); err != nil {
		return "", err
	}
	data, _ := os.ReadFile(ini)
	if !strings.Contains(string(data), "["+name+"]") {
		section := fmt.Sprintf("\n[%s]\nimage = ghcr.io/vincentvdk/apparatus-box:latest\npull = true\ninit_hooks = distrobox-host-exec butler sync %s\n", name, name)
		if homeDir != "" {
			section += "home = " + homeDir + "\n"
		}
		f, err := os.OpenFile(ini, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return "", err
		}
		if _, err := f.WriteString(section); err != nil {
			f.Close()
			return "", err
		}
		f.Close()
	}
	return runCommand("distrobox-assemble", "create", "--file", ini)
}

func (m *osModel) executeCreateDistrobox(homeMode string) tea.Cmd {
	name := m.createName
	if name == "" {
		return func() tea.Msg {
			return commandDoneMsg{output: "No name provided.", err: fmt.Errorf("no name provided"), showOutput: true}
		}
	}

	return func() tea.Msg {
		var output string
		var err error

		if homeMode == "custom" {
			customHome := m.createCustomHome
			if customHome == "" {
				customHome = name
			}
			homeDir := fmt.Sprintf("%s/distrobox-homes/%s", os.Getenv("HOME"), customHome)
			output, err = assembleCreateBox(name, homeDir)
		} else {
			output, err = assembleCreateBox(name, "")
		}

		if err != nil {
			output = fmt.Sprintf("%s\nError: %v", output, err)
		} else {
			output = fmt.Sprintf("Distrobox '%s' created successfully!\n\nEnter with: distrobox enter %s", name, name)
		}

		return commandDoneMsg{output: output, err: err, showOutput: true}
	}
}

func (m *osModel) executeUpgradeDistrobox(action string) tea.Cmd {
	name := m.selectedDistrobox
	if name == "" {
		return func() tea.Msg {
			return commandDoneMsg{output: "No distrobox selected.", err: fmt.Errorf("no distrobox selected"), showOutput: true}
		}
	}

	return func() tea.Msg {
		var output string
		var err error

		switch action {
		case "upgrade-packages":
			output, err = runCommand("distrobox", "upgrade", name)
			if err == nil {
				output = fmt.Sprintf("Package upgrade complete for '%s'.", name)
			}
		case "upgrade-image":
			homeDir, _ := getDistroboxHomeDir(name)

			_, _ = runCommand("distrobox", "stop", name, "--yes")
			_, _ = runCommand("distrobox", "rm", name, "--force")

			// Recreate from the assemble manifest (pull = true fetches the
			// latest image; the init hook re-syncs configs)
			output, err = assembleCreateBox(name, homeDir)

			if err == nil {
				output = fmt.Sprintf("Image upgrade complete for '%s'.\n\nEnter with: distrobox enter %s", name, name)
			}
		}

		if err != nil {
			output = fmt.Sprintf("%s\nError: %v", output, err)
		}

		return commandDoneMsg{output: output, err: err, showOutput: true}
	}
}

func (m *osModel) executeMonitorAction() tea.Cmd {
	return func() tea.Msg {
		_, err := exec.LookPath("hyprdynamicmonitors")
		if err != nil {
			return commandDoneMsg{
				output: fmt.Sprintf("hyprdynamicmonitors not found.\n\nInstall it first, then run:\n  hyprdynamicmonitors tui"),
				err:    fmt.Errorf("hyprdynamicmonitors not found in PATH"),
			}
		}

		cmd := exec.Command("hyprdynamicmonitors", "tui")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return commandDoneMsg{
				output: fmt.Sprintf("hyprdynamicmonitors exited with error:\n%v", err),
				err:    err,
			}
		}
		return commandDoneMsg{output: "", err: nil, showOutput: false}
	}
}



func (m *osModel) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	// Main view
	mainListStyle := lipgloss.NewStyle().
		Width(m.width - 4).
		Height(m.height - 8).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(getAccentColor()).
		Padding(0, 1)

	var statusText string
	if m.running {
		statusText = fmt.Sprintf("%s %s", m.spinner.View(), m.runningText)
	} else {
		statusText = "enter: select • q: quit"
		if m.currentTheme != "" {
			statusText += "   theme: " + m.currentTheme
		}
		if m.currentFont != "" {
			statusText += "   font: " + m.currentFont
		}
	}

	mainView := appStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			getTitleStyle().Render("🔧 Butler") + "  " + getModeStyle().Render(modeString(ModeOS)),
			mainListStyle.Render(m.mainList.View()),
			getStatusStyle().Width(m.width-4).Render(statusText),
		),
	)

	// Popup overlay
	if m.showPopup {
		popupWidth := m.width / 2
		popupHeight := m.height / 2
		if popupWidth < 40 {
			popupWidth = 40
		}
		if popupHeight < 10 {
			popupHeight = 10
		}

		popupHeader := getPopupTitleStyle(m.popupIsError).Render(m.popupTitle)
		var popupBody string

		// Render list if in popup list mode
		if m.popupMode == popupModeDistrobox || m.popupMode == popupModeConfigure || m.popupMode == popupModeDistroboxActions || m.popupMode == popupModeUpgrade || m.popupMode == popupModeTheme || m.popupMode == popupModeFont {
			popupBody = lipgloss.JoinVertical(lipgloss.Left, popupHeader, m.popupList.View())
		} else if m.popupMode == popupModeCreateName || m.popupMode == popupModeCreateCustom {
			popupBody = lipgloss.JoinVertical(lipgloss.Left, popupHeader, m.popupContent, "\n", m.textInput.View(), "\n[enter to confirm • esc to cancel]")
		} else {
			popupFooter := lipgloss.NewStyle().
				Faint(true).
				Render("\n[esc/enter • q to close]")
			popupBody = lipgloss.JoinVertical(lipgloss.Left, popupHeader, m.popupContent, popupFooter)
		}

		popupContent := getPopupStyle(m.popupIsError).
			Width(popupWidth).
			Height(popupHeight).
			Render(popupBody)

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, popupContent)
	}

	return mainView
}

// Distrobox config sync types
type distroboxInfo struct {
	Name    string
	HomeDir string
}

// syncDistroboxConfigs initiates config sync to all distroboxes
func (m *osModel) syncDistroboxConfigs() tea.Cmd {
	return m.startCommand("Fetching distroboxes...", m.executeSyncConfigs())
}

func (m *osModel) executeSyncConfigs() tea.Cmd {
	return func() tea.Msg {
		output, err := syncConfigsToAllDistroboxes()
		return commandDoneMsg{
			output:     output,
			err:        err,
			showOutput: true,
		}
	}
}

// pushDistroboxConfigs initiates pushing configs from all distroboxes
func (m *osModel) pushDistroboxConfigs() tea.Cmd {
	return m.startCommand("Fetching distroboxes...", m.executePushConfigs())
}

func (m *osModel) executePushConfigs() tea.Cmd {
	return func() tea.Msg {
		output, err := pushConfigsFromAllDistroboxes()
		return commandDoneMsg{
			output:     output,
			err:        err,
			showOutput: true,
		}
	}
}

// showSkipList displays the skip list management UI
func (m *osModel) showSkipList() {
	distroboxes, err := getAllDistroboxes()
	if err != nil {
		m.showErrorPopup("Skip List", fmt.Sprintf("Error loading distroboxes: %v", err))
		return
	}
	
	// Load current skip list
	skipped, _ := loadSkipList()
	
	// Create items with checkbox indicator
	var items []list.Item
	for _, db := range distroboxes {
		skipIndicator := "[ ]"
		if skipped[db.Name] {
			skipIndicator = "[x]"
		}
		items = append(items, osItem{
			title:       fmt.Sprintf("%s %s", skipIndicator, db.Name),
			description: db.HomeDir,
			action:      db.Name,
		})
	}
	
	if len(items) == 0 {
		items = append(items, osItem{
			title:       "(no distroboxes found)",
			description: "Create a distrobox first",
			action:      "none",
		})
	}
	
	delegate := getListDelegate()
	delegate.SetHeight(1)
	m.popupList = newList(items, delegate, 0, 0)
	m.popupList.Title = "Skip List - Select to toggle"
	m.popupList.SetShowStatusBar(false)
	m.popupList.SetFilteringEnabled(true)
	m.popupList.SetShowHelp(true)
	m.popupList.SetSize(m.width/2, m.height/2)
	m.popupMode = popupModeSync
	m.showPopup = true
}

// handleSkipListToggle toggles a distrobox in the skip list
func (m *osModel) handleSkipListToggle(name string) {
	// Load current skip list
	skipList, err := loadSkipList()
	if err != nil {
		m.showErrorPopup("Error", fmt.Sprintf("Failed to load skip list: %v", err))
		return
	}
	
	// Toggle the skip status
	if skipList[name] {
		delete(skipList, name)
	} else {
		skipList[name] = true
	}
	
	// Save the updated skip list
	if err := saveSkipList(skipList); err != nil {
		m.showErrorPopup("Error", fmt.Sprintf("Failed to save skip list: %v", err))
		return
	}
	
	// Refresh the skip list display
	m.showSkipList()
}

// saveSkipList saves the skip list to file
func saveSkipList(skipList map[string]bool) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	
	// Ensure directory exists
	skipDir := filepath.Join(homeDir, ".config", "apparatus")
	if err := os.MkdirAll(skipDir, 0755); err != nil {
		return err
	}
	
	skipFile := filepath.Join(skipDir, "skip-distroboxes")
	
	// Build file content
	var content strings.Builder
	for name := range skipList {
		content.WriteString(name + "\n")
	}
	
	return os.WriteFile(skipFile, []byte(content.String()), 0644)
}

// syncConfigsToAllDistroboxes syncs configs from /usr/share/apparatus/ to all distroboxes
func syncConfigsToAllDistroboxes() (string, error) {
	var output strings.Builder
	
	// Get all distroboxes
	distroboxes, err := getAllDistroboxes()
	if err != nil {
		return "", fmt.Errorf("failed to get distroboxes: %v", err)
	}
	
	if len(distroboxes) == 0 {
		return "No distroboxes found to sync.", nil
	}
	
	// Load skip list
	skipped, err := loadSkipList()
	if err != nil {
		output.WriteString(fmt.Sprintf("Warning: Could not load skip list: %v\n\n", err))
	}
	
	var synced []string
	var failed []string
	
	// Sync to each distrobox
	for _, db := range distroboxes {
		if skipped[db.Name] {
			output.WriteString(fmt.Sprintf("Skipping %s (in skip list)\n", db.Name))
			continue
		}
		
		detail, err := syncConfigToDistrobox(db)
		if err != nil {
			failed = append(failed, db.Name)
			output.WriteString(fmt.Sprintf("Failed to sync %s: %v\n", db.Name, err))
		} else {
			synced = append(synced, db.Name)
			output.WriteString(fmt.Sprintf("Synced %s\n%s", db.Name, detail))
		}
	}
	
	// Summary
	output.WriteString(fmt.Sprintf("\nSummary:\n"))
	output.WriteString(fmt.Sprintf("  Synced: %d distroboxes\n", len(synced)))
	if len(synced) > 0 {
		output.WriteString(fmt.Sprintf("    %s\n", strings.Join(synced, ", ")))
	}
	output.WriteString(fmt.Sprintf("  Failed: %d distroboxes\n", len(failed)))
	if len(failed) > 0 {
		output.WriteString(fmt.Sprintf("    %s\n", strings.Join(failed, ", ")))
	}
	
	if len(failed) > 0 {
		return output.String(), fmt.Errorf("%d distroboxes failed to sync", len(failed))
	}
	
	return output.String(), nil
}

// getAllDistroboxes returns list of all distroboxes with their home directories
func getAllDistroboxes() ([]distroboxInfo, error) {
	var distroboxes []distroboxInfo
	
	// Get list of distroboxes
	output, err := runCommand("distrobox", "list", "--no-color")
	if err != nil {
		return nil, fmt.Errorf("failed to list distroboxes: %v", err)
	}
	
	// Parse output (skip header line)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for i, line := range lines {
		// Skip header
		if i == 0 || strings.Contains(line, "ID") || strings.Contains(line, "NAME") {
			continue
		}
		
		// Parse line: ID | NAME | STATUS | IMAGE
		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue
		}
		
		name := strings.TrimSpace(parts[1])
		if name == "" {
			continue
		}
		
		// Get home directory
		homeDir, err := getDistroboxHomeDir(name)
		if err != nil {
			// If we can't get home dir, skip this distrobox
			continue
		}
		
		distroboxes = append(distroboxes, distroboxInfo{
			Name:    name,
			HomeDir: homeDir,
		})
	}
	
	return distroboxes, nil
}

// getDistroboxHomeDir returns the home directory for a distrobox.
// An empty home dir with a nil error means the box shares the host home
// (no --home argument was passed to distrobox create).
func getDistroboxHomeDir(name string) (string, error) {
	cmd := exec.Command("podman", "inspect", name, "--format", "{{json .Args}}")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("podman inspect %s: %v", name, err)
	}

	// Parse JSON output to find --home argument
	var args []string
	if err := json.Unmarshal(output, &args); err != nil {
		return "", fmt.Errorf("parsing podman inspect output for %s: %v", name, err)
	}
	for i, arg := range args {
		if arg == "--home" && i+1 < len(args) {
			return args[i+1], nil
		}
	}

	// No --home argument: shared home with the host
	return "", nil
}

// loadSkipList loads the list of distroboxes to skip
func loadSkipList() (map[string]bool, error) {
	skipList := make(map[string]bool)
	
	// Default skip list file location
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return skipList, nil
	}
	
	skipFile := filepath.Join(homeDir, ".config", "apparatus", "skip-distroboxes")
	
	data, err := os.ReadFile(skipFile)
	if err != nil {
		// File doesn't exist, return empty skip list
		if os.IsNotExist(err) {
			return skipList, nil
		}
		return nil, fmt.Errorf("failed to read skip file: %v", err)
	}
	
	// Parse skip file (one distrobox name per line)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			skipList[line] = true
		}
	}
	
	return skipList, nil
}

// defaultSyncConfigs is the config directories synced into distrobox homes
// by default. "hypr" is deliberately excluded: Hyprland is the host
// compositor and never runs inside a distrobox, so syncing it there only
// leaves behind a broken, unusable hyprland.lua that other tooling inside
// the box can trip over.
var defaultSyncConfigs = []string{"kitty", "waybar", "mako", "walker", "uwsm", "satty", "atuin", "nvim", "zsh", "rio"}

// syncConfigDirs returns the list of config directories to sync into
// distrobox homes. If ~/.config/apparatus/sync-configs exists, its lines
// (one directory name per line, '#' comments allowed, same format as
// skip-distroboxes) replace the default list entirely -- this is the
// escape hatch for anyone who deliberately wants different sync behavior
// (a custom dotfiles setup, testing, etc.) without patching butler itself.
func syncConfigDirs() []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return defaultSyncConfigs
	}

	data, err := os.ReadFile(filepath.Join(homeDir, ".config", "apparatus", "sync-configs"))
	if err != nil {
		return defaultSyncConfigs
	}

	var dirs []string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			dirs = append(dirs, line)
		}
	}
	if len(dirs) == 0 {
		return defaultSyncConfigs
	}
	return dirs
}

// syncConfigToDistrobox syncs configs to a single distrobox. Config dirs are
// rsync'd (preserving modes and symlinks); themes are never copied -- the box
// gets a symlink to /usr/share/apparatus/themes and every themed tool is
// linked to the host's currently active theme.
func syncConfigToDistrobox(db distroboxInfo) (string, error) {
	hostHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get host home dir: %v", err)
	}

	// Source per config dir: the effective host config (~/.config/<dir>,
	// chezmoi-managed) when present, otherwise the system default baked into
	// /usr/share/apparatus at OS build time.
	srcFor := func(config string) string {
		hostPath := filepath.Join(hostHome, ".config", config)
		if _, err := os.Stat(hostPath); err == nil {
			return hostPath
		}
		return filepath.Join("/usr/share/apparatus", config)
	}

	// Destination: custom-home boxes have their own home dir; shared-home
	// boxes keep configs in ~/.config/apparatus-box on the host home (see
	// box/system/profile-custom.sh).
	var dst string
	if db.HomeDir == "" {
		dst = filepath.Join(hostHome, ".config", "apparatus-box")
	} else {
		dst = filepath.Join(db.HomeDir, ".config")
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return "", fmt.Errorf("failed to create config dir: %v", err)
	}

	var output strings.Builder

	for _, config := range syncConfigDirs() {
		srcPath := srcFor(config)
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue
		}
		dstPath := filepath.Join(dst, config)
		if out, err := runCommand("rsync", "-a", "--delete", srcPath+"/", dstPath); err != nil {
			return output.String(), fmt.Errorf("failed to sync %s: %v: %s", config, err, out)
		}
		output.WriteString(fmt.Sprintf("Synced %s\n", config))
	}

	// Themes: single source of truth, linked not copied. Inside a distrobox
	// the host root is mounted at /run/host, so links point at
	// /run/host/usr/share/apparatus/themes and resolve to the host's live
	// theme files. (They dangle if the box home is viewed from the host --
	// only the container reads these configs.)
	themesSrc := "/usr/share/apparatus/themes" // host-side path, used for validation
	themesLink := filepath.Join(dst, "apparatus", "themes")
	themesLinkTarget := "/run/host" + themesSrc
	if err := os.MkdirAll(filepath.Dir(themesLink), 0755); err != nil {
		return output.String(), fmt.Errorf("failed to create apparatus config dir: %v", err)
	}
	os.RemoveAll(themesLink) // drop a real dir left behind by older syncs
	if err := os.Symlink(themesLinkTarget, themesLink); err != nil {
		return output.String(), fmt.Errorf("failed to link themes dir: %v", err)
	}

	// Link every themed tool to the host's active theme. Links are staged as
	// ".new" siblings and committed with atomic renames, so a failure never
	// leaves the box half on the old theme.
	themeName := currentThemeName()
	if themeName == "" {
		entries, err := os.ReadDir(themesSrc)
		if err != nil || len(entries) == 0 {
			return output.String(), fmt.Errorf("no active theme and no themes in %s", themesSrc)
		}
		themeName = entries[0].Name()
	}
	themeDir := filepath.Join(themesSrc, themeName)
	if _, err := os.Stat(themeDir); err != nil {
		return output.String(), fmt.Errorf("active theme %s not found in %s", themeName, themesSrc)
	}

	// box config path -> theme file within the theme dir
	themeLinks := []struct{ target, file string }{
		{filepath.Join(dst, "kitty", "theme.conf"), "kitty.conf"},
		{filepath.Join(dst, "waybar", "theme.css"), "waybar.css"},
		{filepath.Join(dst, "mako", "config"), "mako.conf"},
		{filepath.Join(dst, "hypr", "theme.conf"), "hyprland.conf"},
		{filepath.Join(dst, "satty", "overrides.css"), filepath.Join("satty", "overrides.css")},
		{filepath.Join(dst, "nvim", "lua", "config", "theme.lua"), "nvim.lua"},
		{filepath.Join(dst, "rio", "themes", themeName+".toml"), "rio.toml"},
	}

	var pending []stagedSwap
	rollback := func() {
		for _, s := range pending {
			os.Remove(s.staged)
		}
	}
	for _, l := range themeLinks {
		// skip tools whose config dir is not synced into this box (hypr is
		// host-only by default) or whose theme file does not exist
		if _, err := os.Stat(filepath.Dir(l.target)); err != nil {
			continue
		}
		if _, err := os.Stat(filepath.Join(themeDir, l.file)); err != nil {
			continue
		}
		staged := l.target + ".new"
		os.Remove(staged) // clean up any leftover from a prior failed sync
		linkTarget := filepath.Join(themesLinkTarget, themeName, l.file)
		if err := os.Symlink(linkTarget, staged); err != nil {
			rollback()
			return output.String(), fmt.Errorf("%s: %v", filepath.Base(l.target), err)
		}
		pending = append(pending, stagedSwap{staged: staged, target: l.target})
	}
	for _, s := range pending {
		if err := os.Rename(s.staged, s.target); err != nil {
			rollback()
			return output.String(), fmt.Errorf("committing %s: %v", filepath.Base(s.target), err)
		}
	}
	output.WriteString(fmt.Sprintf("Linked theme %s\n", themeName))

	return output.String(), nil
}

// pushExcludes lists theme-link files per config dir that are owned by
// applyTheme and must never be pushed back from a box -- they are symlinks
// into the host's theme dir and would corrupt the chezmoi source if committed.
var pushExcludes = map[string][]string{
	"kitty":  {"theme.conf"},
	"waybar": {"theme.css"},
	"mako":   {"config"},
	"hypr":   {"theme.conf"},
	"satty":  {"overrides.css"},
	"nvim":   {"lua/config/theme.lua"},
	"rio":    {"themes/*.toml"},
}

// pushConfigFromDistrobox copies configs from a box back to the host's
// ~/.config (the chezmoi-managed dir), so changes made inside the box can
// trickle upstream: review with `chezmoi diff`, commit and push the
// apparatus-dotfiles repo, then `butler sync` distributes to all boxes.
func pushConfigFromDistrobox(db distroboxInfo) (string, error) {
	hostHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get host home dir: %v", err)
	}

	// Source: shared-home boxes keep configs in ~/.config/apparatus-box on
	// the host home; custom-home boxes in <box home>/.config.
	var src string
	if db.HomeDir == "" {
		src = filepath.Join(hostHome, ".config", "apparatus-box")
	} else {
		src = filepath.Join(db.HomeDir, ".config")
	}
	dst := filepath.Join(hostHome, ".config")

	var output strings.Builder
	for _, config := range syncConfigDirs() {
		srcPath := filepath.Join(src, config)
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue
		}
		args := []string{"-a"}
		for _, ex := range pushExcludes[config] {
			args = append(args, "--exclude", ex)
		}
		args = append(args, srcPath+"/", filepath.Join(dst, config))
		if out, err := runCommand("rsync", args...); err != nil {
			return output.String(), fmt.Errorf("failed to push %s: %v: %s", config, err, out)
		}
		output.WriteString(fmt.Sprintf("Pushed %s\n", config))
	}
	output.WriteString("\nReview upstream changes with: chezmoi diff\n")
	return output.String(), nil
}

// pushConfigsFromAllDistroboxes pushes configs from all non-skipped
// distroboxes back to the host
func pushConfigsFromAllDistroboxes() (string, error) {
	var output strings.Builder

	distroboxes, err := getAllDistroboxes()
	if err != nil {
		return "", fmt.Errorf("failed to get distroboxes: %v", err)
	}
	if len(distroboxes) == 0 {
		return "No distroboxes found to push from.", nil
	}

	skipped, err := loadSkipList()
	if err != nil {
		output.WriteString(fmt.Sprintf("Warning: Could not load skip list: %v\n\n", err))
	}

	var pushed []string
	var failed []string

	for _, db := range distroboxes {
		if skipped[db.Name] {
			output.WriteString(fmt.Sprintf("Skipping %s (in skip list)\n", db.Name))
			continue
		}
		detail, err := pushConfigFromDistrobox(db)
		if err != nil {
			failed = append(failed, db.Name)
			output.WriteString(fmt.Sprintf("Failed to push from %s: %v\n", db.Name, err))
		} else {
			pushed = append(pushed, db.Name)
			output.WriteString(fmt.Sprintf("Pushed from %s\n%s", db.Name, detail))
		}
	}

	output.WriteString(fmt.Sprintf("\nSummary:\n"))
	output.WriteString(fmt.Sprintf("  Pushed: %d distroboxes\n", len(pushed)))
	if len(pushed) > 0 {
		output.WriteString(fmt.Sprintf("    %s\n", strings.Join(pushed, ", ")))
	}
	output.WriteString(fmt.Sprintf("  Failed: %d distroboxes\n", len(failed)))
	if len(failed) > 0 {
		output.WriteString(fmt.Sprintf("    %s\n", strings.Join(failed, ", ")))
	}

	if len(failed) > 0 {
		return output.String(), fmt.Errorf("%d distroboxes failed to push", len(failed))
	}
	return output.String(), nil
}

// stagedSwap is a config target whose replacement has been written to a
// sibling ".new" path, ready to be committed with a single atomic rename.
type stagedSwap struct {
	staged string // the ".new" path holding the new content
	target string // the real path it will replace
}

// stageSymlink prepares target's replacement as a symlink to source at a
// ".new" sibling path, without touching target itself. Returns the staged
// swap to commit later, or an error if the source is missing or the
// symlink couldn't be created.
func stageSymlink(source, target string) (stagedSwap, error) {
	if _, err := os.Stat(source); err != nil {
		return stagedSwap{}, fmt.Errorf("%s: %w", filepath.Base(target), err)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return stagedSwap{}, fmt.Errorf("%s: %w", filepath.Base(target), err)
	}
	staged := target + ".new"
	os.Remove(staged) // clean up any leftover from a prior failed apply
	if err := os.Symlink(source, staged); err != nil {
		return stagedSwap{}, fmt.Errorf("%s: %w", filepath.Base(target), err)
	}
	return stagedSwap{staged: staged, target: target}, nil
}

// stageFile is like stageSymlink but copies source's content instead of
// symlinking to it (used for nvim.lua, which is written directly rather
// than symlinked so it survives being synced into a distrobox).
func stageFile(source, target string) (stagedSwap, error) {
	data, err := os.ReadFile(source)
	if err != nil {
		return stagedSwap{}, fmt.Errorf("%s: %w", filepath.Base(target), err)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return stagedSwap{}, fmt.Errorf("%s: %w", filepath.Base(target), err)
	}
	staged := target + ".new"
	if err := os.WriteFile(staged, data, 0644); err != nil {
		return stagedSwap{}, fmt.Errorf("%s: %w", filepath.Base(target), err)
	}
	return stagedSwap{staged: staged, target: target}, nil
}

// applyTheme applies a theme across all config files. Every app's new
// config is staged as a sibling ".new" file first; only once every app has
// staged successfully do we commit by renaming each one over its live
// target (an atomic operation on the same filesystem). If staging fails
// partway through, everything staged so far is rolled back and the old
// theme's config is left completely untouched -- so a failure never leaves
// some apps on the new theme and others broken or stuck on the old one.
func applyTheme(themeName string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	themesDir := getThemesDir()
	if _, err := os.Stat(themesDir); err != nil {
		return fmt.Errorf("themes directory not available: %s (inside a distrobox this should be the host's /run/host mount)", themesDir)
	}
	themeDir := filepath.Join(themesDir, themeName)

	var pending []stagedSwap
	rollback := func() {
		for _, s := range pending {
			os.Remove(s.staged)
		}
	}

	stage := func(source, target string) error {
		s, err := stageSymlink(source, target)
		if err != nil {
			rollback()
			return err
		}
		pending = append(pending, s)
		return nil
	}

	if err := stage(filepath.Join(themeDir, "kitty.conf"), filepath.Join(homeDir, ".config", "kitty", "theme.conf")); err != nil {
		return err
	}
	if err := stage(filepath.Join(themeDir, "waybar.css"), filepath.Join(homeDir, ".config", "waybar", "theme.css")); err != nil {
		return err
	}
	if err := stage(filepath.Join(themeDir, "mako.conf"), filepath.Join(homeDir, ".config", "mako", "config")); err != nil {
		return err
	}
	if err := stage(filepath.Join(themeDir, "hyprland.conf"), filepath.Join(homeDir, ".config", "hypr", "theme.conf")); err != nil {
		return err
	}
	if err := stage(filepath.Join(themeDir, "satty", "overrides.css"), filepath.Join(homeDir, ".config", "satty", "overrides.css")); err != nil {
		return err
	}

	// rio theme (optional -- only themed when the theme ships rio.toml)
	rioSource := filepath.Join(themeDir, "rio.toml")
	if _, err := os.Stat(rioSource); err == nil {
		if err := stage(rioSource, filepath.Join(homeDir, ".config", "rio", "themes", themeName+".toml")); err != nil {
			return err
		}
	}

	// nvim.lua is optional -- not every theme ships one
	nvimSource := filepath.Join(themeDir, "nvim.lua")
	if _, err := os.Stat(nvimSource); err == nil {
		nvimTarget := filepath.Join(homeDir, ".config", "nvim", "lua", "config", "theme.lua")
		s, err := stageFile(nvimSource, nvimTarget)
		if err != nil {
			rollback()
			return err
		}
		pending = append(pending, s)
	}

	// Commit: atomically rename every staged file over its live target.
	for _, s := range pending {
		if err := os.Rename(s.staged, s.target); err != nil {
			return fmt.Errorf("committing %s: %w", filepath.Base(s.target), err)
		}
	}

	// Point Rio at the new theme (best-effort, outside the atomic swap)
	setRioTheme(homeDir, themeName)

	// Apply GTK theme (best-effort, outside the atomic config swap)
	isDark := themeIsDark(themeName)
	if isDark {
		exec.Command("gsettings", "set", "org.gnome.desktop.interface", "color-scheme", "prefer-dark").Run()
		exec.Command("gsettings", "set", "org.gnome.desktop.interface", "gtk-theme", "Adwaita-dark").Run()
	} else {
		exec.Command("gsettings", "set", "org.gnome.desktop.interface", "color-scheme", "prefer-light").Run()
		exec.Command("gsettings", "set", "org.gnome.desktop.interface", "gtk-theme", "Adwaita").Run()
	}

	// Save current theme marker -- only after every app's config has been
	// committed, so it can never claim a theme that isn't actually applied.
	appConfigDir := filepath.Join(homeDir, ".config", "apparatus")
	os.MkdirAll(appConfigDir, 0755)
	currentThemeFile := filepath.Join(appConfigDir, "current-theme")
	tmpThemeFile := currentThemeFile + ".new"
	if err := os.WriteFile(tmpThemeFile, []byte(themeName), 0644); err == nil {
		os.Rename(tmpThemeFile, currentThemeFile)
	}

	// Reload services
	exec.Command("hyprctl", "reload").Run()
	exec.Command("pkill", "-SIGUSR1", "kitty").Run()
	exec.Command("makoctl", "reload").Run()

	// Reload waybar using SIGUSR2 signal
	exec.Command("killall", "-SIGUSR2", "waybar").Run()

	// Reload satty (must be restarted as it doesn't support signal-based reload)
	exec.Command("pkill", "satty").Run()

	return nil
}

// setRioTheme points Rio's config.toml at the given theme (best-effort;
// no-op when Rio isn't configured)
func setRioTheme(homeDir, themeName string) {
	cfg := filepath.Join(homeDir, ".config", "rio", "config.toml")
	data, err := os.ReadFile(cfg)
	if err != nil {
		return
	}
	content := string(data)
	re := regexp.MustCompile(`(?m)^theme\s*=.*$`)
	if re.MatchString(content) {
		content = re.ReplaceAllString(content, `theme = "`+themeName+`"`)
	} else {
		content += "\ntheme = \"" + themeName + "\"\n"
	}
	os.WriteFile(cfg, []byte(content), 0644)
}

// applyFont applies a font across all config files
func applyFont(fontFamily, fontName string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		return
	}

	// Update kitty font
	kittyConf := filepath.Join(homeDir, ".config", "kitty", "kitty.conf")
	if data, err := os.ReadFile(kittyConf); err == nil {
		content := string(data)
		content = strings.ReplaceAll(content, "font_family      "+"IoskeleyMono Nerd Font", "font_family      "+fontFamily)
		content = strings.ReplaceAll(content, "font_family      "+"JetBrainsMono Nerd Font", "font_family      "+fontFamily)
		content = strings.ReplaceAll(content, "font_family      "+"Hack Nerd Font", "font_family      "+fontFamily)
		os.WriteFile(kittyConf, []byte(content), 0644)
	}

	// Update waybar font
	waybarCSS := filepath.Join(homeDir, ".config", "waybar", "style.css")
	if data, err := os.ReadFile(waybarCSS); err == nil {
		content := string(data)
		content = strings.ReplaceAll(content, "\"Hack Nerd Font\"", "\""+fontFamily+"\"")
		content = strings.ReplaceAll(content, "\"JetBrainsMono Nerd Font\"", "\""+fontFamily+"\"")
		content = strings.ReplaceAll(content, "\"IoskeleyMono Nerd Font\"", "\""+fontFamily+"\"")
		os.WriteFile(waybarCSS, []byte(content), 0644)
	}

	// Update mako font
	makoConf := filepath.Join(homeDir, ".config", "mako", "config")
	if data, err := os.ReadFile(makoConf); err == nil {
		content := string(data)
		content = strings.ReplaceAll(content, "Hack Nerd Font", fontFamily)
		content = strings.ReplaceAll(content, "JetBrainsMono Nerd Font", fontFamily)
		content = strings.ReplaceAll(content, "IoskeleyMono Nerd Font", fontFamily)
		os.WriteFile(makoConf, []byte(content), 0644)
	}

	// Save current font
	appConfigDir := filepath.Join(homeDir, ".config", "apparatus")
	os.MkdirAll(appConfigDir, 0755)
	os.WriteFile(filepath.Join(appConfigDir, "current-font"), []byte(fontName), 0644)

	// Reload services
	exec.Command("pkill", "-SIGUSR1", "kitty").Run()
	exec.Command("hyprctl", "reload").Run()
	exec.Command("makoctl", "reload").Run()

	// Reload waybar using SIGUSR2 signal
	exec.Command("killall", "-SIGUSR2", "waybar").Run()
}
