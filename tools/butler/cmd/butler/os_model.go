package main

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)
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
		osItem{title: "⬇ Update Configs", description: "Apply latest apparatus-dotfiles to host ~/.config", action: "update-configs"},
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

