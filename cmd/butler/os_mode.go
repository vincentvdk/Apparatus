package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
)

// OS Mode model
type osModel struct {
	width, height int
	quitting      bool
	showPopup     bool
	popupMode     int
	popupTitle    string
	popupContent  string
	popupList     list.Model

	// Distrobox state
	selectedDistrobox string
	createName        string
	createCustomHome  string

	// Text input for create flow
	textInput textinput.Model

	// Pending command
	pendingCmd tea.Cmd

	running       bool
	spinner       spinner.Model
	runningText   string

	// Main list
	mainList list.Model
}

func newOSModel() *osModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	m := &osModel{
		mainList: makeMainList(),
		spinner:  s,
	}
	return m
}

func makeMainList() list.Model {
	items := []list.Item{
		osItem{title: "📦 Distrobox", description: "Manage distrobox containers", action: "distrobox"},
		osItem{title: "⚙️  Configure", description: "Hyprland, terminal, monitors, audio, AI", action: "configure"},
		osItem{title: "🎨 Theme", description: "Switch color theme (Catppuccin)", action: "theme"},
		osItem{title: "🔤 Font", description: "Switch system font", action: "font"},
		osItem{title: "⌨️  Help", description: "Keyboard shortcuts and keybindings", action: "help"},
	}

	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(1)

	l := list.New(items, delegate, 0, 0)
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
	// Handle pending command
	if m.pendingCmd != nil {
		cmd := m.pendingCmd
		m.pendingCmd = nil
		// Execute the command and show result
		result := cmd()
		if cmdMsg, ok := result.(commandDoneMsg); ok {
			m.showPopup = true
			m.popupMode = popupModeOutput
			if cmdMsg.err != nil {
				m.popupTitle = "Error"
				m.popupContent = fmt.Sprintf("Error: %v\n\n%s", cmdMsg.err, cmdMsg.output)
			} else {
				m.popupTitle = "Output"
				m.popupContent = cmdMsg.output
			}
		}
		return m, nil
	}

	// Handle popup mode
	if m.showPopup {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			// Handle popup list
			if m.popupMode == popupModeDistrobox || m.popupMode == popupModeConfigure || m.popupMode == popupModeDistroboxActions || m.popupMode == popupModeUpgrade || m.popupMode == popupModeTheme || m.popupMode == popupModeFont {
				switch msg.String() {
				case "enter":
					if item, ok := m.popupList.SelectedItem().(osItem); ok {
						m.handlePopupAction(item)
						return m, nil
					}
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
						m.showPopup = false
						m.pendingCmd = m.executeCreateDistrobox("custom")
						return m, nil
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
					m.showPopup = false
					m.pendingCmd = m.executeCreateDistrobox("default")
					return m, nil
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
				m.handleMainAction(item)
				return m, nil
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

func (m *osModel) handleMainAction(item osItem) {
	switch item.action {
	case "distrobox":
		m.showDistroboxMenu()
	case "configure":
		m.showConfigureMenu()
	case "theme":
		m.showThemePopup()
	case "font":
		m.showFontPopup()
	case "help":
		m.showHelp()
	}
}

func (m *osModel) applyTheme(themeName string) {
	applyTheme(themeName)
	m.showPopup = true
	m.popupMode = popupModeOutput
	m.popupTitle = "Theme Applied"
	m.popupContent = fmt.Sprintf("Theme applied: %s\n\nReloaded: hyprctl, kitty, mako", themeName)
}

func (m *osModel) applyFontAction(fontName string) {
	fontMap := map[string]string{
		"ioskeley-mono":    "IoskeleyMono Nerd Font",
		"jetbrains-mono":   "JetBrainsMono Nerd Font",
		"hack-nerd-font":   "Hack Nerd Font",
	}
	fontFamily, ok := fontMap[fontName]
	if !ok {
		m.showPopup = true
		m.popupMode = popupModeOutput
		m.popupTitle = "Font Error"
		m.popupContent = fmt.Sprintf("Unknown font: %s", fontName)
		return
	}
	applyFont(fontFamily, fontName)
	m.showPopup = true
	m.popupMode = popupModeOutput
	m.popupTitle = "Font Applied"
	m.popupContent = fmt.Sprintf("Font applied: %s\n\nReloaded: kitty, waybar, mako", fontName)
}

func (m *osModel) showThemePopup() {
	themesDir := "/usr/share/apparatus/themes"
	entries, err := os.ReadDir(themesDir)
	if err != nil {
		m.showPopup = true
		m.popupMode = popupModeOutput
		m.popupTitle = "Theme Selection"
		m.popupContent = fmt.Sprintf("Could not read themes directory: %v", err)
		return
	}

	var items []list.Item
	for _, entry := range entries {
		if entry.IsDir() {
			label := entry.Name()
			desc := "light"
			if strings.Contains(label, "mocha") {
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
		m.showPopup = true
		m.popupMode = popupModeOutput
		m.popupTitle = "Theme Selection"
		m.popupContent = "No themes found in " + themesDir
		return
	}

	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(1)
	m.popupList = list.New(items, delegate, 0, 0)
	m.popupList.Title = "Themes"
	m.popupList.SetShowStatusBar(true)
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

	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(1)
	m.popupList = list.New(fontItems, delegate, 0, 0)
	m.popupList.Title = "Fonts"
	m.popupList.SetShowStatusBar(true)
	m.popupList.SetFilteringEnabled(true)
	m.popupList.SetShowHelp(true)
	m.popupList.SetSize(m.width/2, m.height/2)
	m.popupMode = popupModeFont
	m.showPopup = true
}

func (m *osModel) showTerminalPopup() {
	m.showPopup = true
	m.popupMode = popupModeOutput
	m.popupTitle = "Terminal"
	m.popupContent = "1. kitty\n2. rio\n\nRun: butler --os terminal <name>\n\nOr edit: ~/.config/hypr/hyprland.conf"
}

func (m *osModel) launchMonitors() {
	_, err := exec.LookPath("hyprdynamicmonitors")
	if err != nil {
		m.showPopup = true
		m.popupMode = popupModeOutput
		m.popupTitle = "Monitors"
		m.popupContent = "hyprdynamicmonitors not found.\n\nInstall it first, then run:\n  hyprdynamicmonitors tui"
		return
	}
	// Launch TUI directly - takes over terminal
	cmd := exec.Command("hyprdynamicmonitors", "tui")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if runErr := cmd.Run(); runErr != nil {
		m.showPopup = true
		m.popupMode = popupModeOutput
		m.popupTitle = "Monitors"
		m.popupContent = fmt.Sprintf("hyprdynamicmonitors exited with error:\n%v", runErr)
	}
}

func (m *osModel) showAudioPopup() {
	m.showPopup = true
	m.popupMode = popupModeOutput
	m.popupTitle = "Audio"
	m.popupContent = "Run pavucontrol to configure audio devices.\n\nThis will launch in the background."
}

func (m *osModel) showAIWorkloadPopup() {
	m.showPopup = true
	m.popupMode = popupModeOutput
	m.popupTitle = "AI Workload"
	m.popupContent = "VRAM Allocation\n\nSelect from the menu:\n  16 GB, 32 GB, 64 GB, 96 GB (max stable)\n  Reset to default\n\nConfigures amdttm.kernel_cmdline for AMD Ryzen AI APUs.\nRequires reboot to take effect."
}

func (m *osModel) showDistroboxResult(cmd tea.Cmd) {
	result := cmd()
	if cmdMsg, ok := result.(commandDoneMsg); ok {
		m.showPopup = true
		m.popupMode = popupModeOutput
		if cmdMsg.err != nil {
			m.popupTitle = "Error"
			m.popupContent = fmt.Sprintf("Error: %v\n\n%s", cmdMsg.err, cmdMsg.output)
		} else {
			m.popupTitle = "Output"
			m.popupContent = cmdMsg.output
		}
	}
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

	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(1)
	m.popupList = list.New(items, delegate, 0, 0)
	m.popupList.Title = "Distroboxes"
	m.popupList.SetShowStatusBar(true)
	m.popupList.SetFilteringEnabled(true)
	m.popupList.SetShowHelp(true)
	m.popupList.SetSize(m.width/2, m.height/2)
	m.popupMode = popupModeDistrobox
	m.showPopup = true
}

func (m *osModel) showConfigureMenu() {
	items := []list.Item{
		osItem{title: "Terminal", description: "Set default terminal (kitty/rio)", action: "terminal"},
		osItem{title: "Monitors", description: "Configure display setup", action: "monitors"},
		osItem{title: "Audio", description: "Configure audio devices", action: "audio"},
		osItem{title: "AI Workload", description: "GPU VRAM allocation for AI/ML", action: "ai-workload"},
	}

	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(1)
	m.popupList = list.New(items, delegate, 0, 0)
	m.popupList.Title = "Configure"
	m.popupList.SetShowStatusBar(true)
	m.popupList.SetFilteringEnabled(true)
	m.popupList.SetShowHelp(true)
	m.popupList.SetSize(m.width/2, m.height/2)
	m.popupMode = popupModeConfigure
	m.showPopup = true
}

func (m *osModel) handlePopupAction(item osItem) {
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
			m.showAudioPopup()
		case "ai-workload":
			m.showAIWorkloadPopup()
		}
	case popupModeDistroboxActions:
		switch item.action {
		case "enter":
			m.showDistroboxResult(m.executeDistroboxAction("enter"))
		case "upgrade":
			m.showDistroboxResult(m.executeDistroboxAction("upgrade"))
		case "stop":
			m.showDistroboxResult(m.executeDistroboxAction("stop"))
		case "remove":
			m.showDistroboxResult(m.executeDistroboxAction("remove"))
		}
	case popupModeUpgrade:
		switch item.action {
		case "upgrade-packages":
			m.showDistroboxResult(m.executeUpgradeDistrobox("upgrade-packages"))
		case "upgrade-image":
			m.showDistroboxResult(m.executeUpgradeDistrobox("upgrade-image"))
		}
	case popupModeTheme:
		m.applyTheme(item.action)
	case popupModeFont:
		m.applyFontAction(item.action)
	}
}

func (m *osModel) showUpgradeMenu() {
	items := []list.Item{
		osItem{title: "packages", description: "Update packages inside container", action: "upgrade-packages"},
		osItem{title: "image", description: "Recreate with latest image (keeps home dir)", action: "upgrade-image"},
	}

	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(1)
	m.popupList = list.New(items, delegate, 0, 0)
	m.popupList.Title = "Upgrade Type"
	m.popupList.SetShowStatusBar(true)
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

	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(1)
	m.popupList = list.New(items, delegate, 0, 0)
	m.popupList.Title = "Actions"
	m.popupList.SetShowStatusBar(true)
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
	m.popupList = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
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
			output, err = runCommand("distrobox", "create", "-i", "ghcr.io/vinintvdk/apparatus-box:latest", "-n", name, "--home", homeDir)
		} else {
			output, err = runCommand("distrobox", "create", "-i", "ghcr.io/vinintvdk/apparatus-box:latest", "-n", name)
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
			homeDir := ""
			homeOutput, homeErr := runCommand("podman", "inspect", name, "--format", "{{json .Args}}")
			if homeErr == nil {
				lines := strings.Split(strings.TrimSpace(homeOutput), "\n")
				for i, line := range lines {
					if strings.Contains(line, "--home") && i+1 < len(lines) {
						homeDir = strings.TrimSpace(strings.Trim(lines[i+1], `"`))
						break
					}
				}
			}

			image := "ghcr.io/vinintvdk/apparatus-box:latest"
			var homeArg string
			if homeDir != "" {
				homeArg = "--home " + homeDir
			}

			pullOutput, pullErr := runCommand("podman", "pull", image)
			if pullErr != nil {
				return commandDoneMsg{
					output: fmt.Sprintf("Failed to pull latest image:\n%s\nError: %v", pullOutput, pullErr),
					err:    pullErr,
					showOutput: true,
				}
			}

			_, _ = runCommand("distrobox", "stop", name, "--yes")
			_, _ = runCommand("distrobox", "rm", name, "--force")

			createArgs := []string{"distrobox", "create", "-n", name, "-i", image}
			if homeArg != "" {
				homeParts := strings.Fields(homeArg)
				createArgs = append(createArgs, homeParts...)
			}
			createCmd := exec.Command(createArgs[0], createArgs[1:]...)
			createOutput, createErr := createCmd.CombinedOutput()
			output = string(createOutput)
			err = createErr

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
		BorderForeground(lipgloss.Color("626262")).
		Padding(0, 1)

	mainView := appStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			titleStyle.Render("🔧 Butler") + "  " + modeStyle.Render(modeString(ModeOS)),
			mainListStyle.Render(m.mainList.View()),
			statusStyle.Width(m.width-4).Render("enter: select • q: quit"),
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

		popupHeader := popupTitleStyle.Render(m.popupTitle)
		var popupBody string

		// Render list if in popup list mode
		if m.popupMode == popupModeDistrobox || m.popupMode == popupModeConfigure || m.popupMode == popupModeDistroboxActions || m.popupMode == popupModeUpgrade || m.popupMode == popupModeTheme || m.popupMode == popupModeFont {
			popupBody = lipgloss.JoinVertical(lipgloss.Left, popupHeader, m.popupList.View())
		} else if m.popupMode == popupModeCreateName || m.popupMode == popupModeCreateCustom {
			popupBody = lipgloss.JoinVertical(lipgloss.Left, popupHeader, m.popupContent, "\n", m.textInput.View(), "\n[enter to confirm • esc to cancel]")
		} else {
			popupFooter := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666")).
				Render("\n[esc/enter • q to close]")
			popupBody = lipgloss.JoinVertical(lipgloss.Left, popupHeader, m.popupContent, popupFooter)
		}

		popupContent := popupStyle.
			Width(popupWidth).
			Height(popupHeight).
			Render(popupBody)

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, popupContent)
	}

	return mainView
}

// applyTheme applies a theme across all config files
func applyTheme(themeName string) {
	themesDir := "/usr/share/apparatus/themes"

	// Apply kitty theme
	kittyTheme := filepath.Join("$HOME", ".config", "kitty", "theme.conf")
	os.MkdirAll(filepath.Dir(kittyTheme), 0755)
	os.Remove(kittyTheme)
	os.Symlink(filepath.Join(themesDir, themeName, "kitty.conf"), kittyTheme)

	// Apply waybar theme
	waybarCSS := filepath.Join("$HOME", ".config", "waybar", "theme.css")
	os.MkdirAll(filepath.Dir(waybarCSS), 0755)
	os.Remove(waybarCSS)
	os.Symlink(filepath.Join(themesDir, themeName, "waybar.css"), waybarCSS)

	// Apply mako theme
	makoConf := filepath.Join("$HOME", ".config", "mako", "config")
	os.MkdirAll(filepath.Dir(makoConf), 0755)
	os.Remove(makoConf)
	os.Symlink(filepath.Join(themesDir, themeName, "mako.conf"), makoConf)

	// Apply hyprland theme
	hyprTheme := filepath.Join("$HOME", ".config", "hypr", "theme.conf")
	os.MkdirAll(filepath.Dir(hyprTheme), 0755)
	os.Remove(hyprTheme)
	os.Symlink(filepath.Join(themesDir, themeName, "hyprland.conf"), hyprTheme)

	// Apply rio theme
	rioConf := filepath.Join("$HOME", ".config", "rio", "config.toml")
	if _, err := os.Stat(rioConf); err == nil {
		content, _ := os.ReadFile(rioConf)
		content = []byte(strings.Replace(string(content), `theme = "`, `theme = "`+themeName+`"`, 1))
		os.WriteFile(rioConf, content, 0644)
	}

	// Apply GTK theme
	isDark := strings.Contains(themeName, "mocha") || strings.Contains(themeName, "dark")
	if isDark {
		exec.Command("gsettings", "set", "org.gnome.desktop.interface", "color-scheme", "prefer-dark").Run()
	} else {
		exec.Command("gsettings", "set", "org.gnome.desktop.interface", "color-scheme", "prefer-light").Run()
	}

	// Save current theme
	os.MkdirAll(filepath.Join("$HOME", ".config", "apparatus"), 0755)
	os.WriteFile(filepath.Join("$HOME", ".config", "apparatus", "current-theme"), []byte(themeName), 0644)

	// Reload services
	exec.Command("hyprctl", "reload").Run()
	exec.Command("pkill", "-SIGUSR1", "kitty").Run()
	exec.Command("makoctl", "reload").Run()
}

// applyFont applies a font across all config files
func applyFont(fontFamily, fontName string) {
	// Update kitty font
	kittyConf := filepath.Join("$HOME", ".config", "kitty", "kitty.conf")
	if data, err := os.ReadFile(kittyConf); err == nil {
		content := string(data)
		content = strings.ReplaceAll(content, "font_family      "+"IoskeleyMono Nerd Font", "font_family      "+fontFamily)
		content = strings.ReplaceAll(content, "font_family      "+"JetBrainsMono Nerd Font", "font_family      "+fontFamily)
		content = strings.ReplaceAll(content, "font_family      "+"Hack Nerd Font", "font_family      "+fontFamily)
		os.WriteFile(kittyConf, []byte(content), 0644)
	}

	// Update waybar font
	waybarCSS := filepath.Join("$HOME", ".config", "waybar", "style.css")
	if data, err := os.ReadFile(waybarCSS); err == nil {
		content := string(data)
		content = strings.ReplaceAll(content, "\"Hack Nerd Font\"", "\""+fontFamily+"\"")
		content = strings.ReplaceAll(content, "\"JetBrainsMono Nerd Font\"", "\""+fontFamily+"\"")
		content = strings.ReplaceAll(content, "\"IoskeleyMono Nerd Font\"", "\""+fontFamily+"\"")
		os.WriteFile(waybarCSS, []byte(content), 0644)
	}

	// Update mako font
	makoConf := filepath.Join("$HOME", ".config", "mako", "config")
	if data, err := os.ReadFile(makoConf); err == nil {
		content := string(data)
		content = strings.ReplaceAll(content, "Hack Nerd Font", fontFamily)
		content = strings.ReplaceAll(content, "JetBrainsMono Nerd Font", fontFamily)
		content = strings.ReplaceAll(content, "IoskeleyMono Nerd Font", fontFamily)
		os.WriteFile(makoConf, []byte(content), 0644)
	}

	// Save current font
	os.MkdirAll(filepath.Join("$HOME", ".config", "apparatus"), 0755)
	os.WriteFile(filepath.Join("$HOME", ".config", "apparatus", "current-font"), []byte(fontName), 0644)

	// Reload services
	exec.Command("pkill", "-SIGUSR1", "kitty").Run()
	exec.Command("hyprctl", "reload").Run()
	exec.Command("makoctl", "reload").Run()
}
