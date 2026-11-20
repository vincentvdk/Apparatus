package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)
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
	case "update-configs":
		return m.updateHostConfigsAction()
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
