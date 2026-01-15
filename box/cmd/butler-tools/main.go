package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

// Styles
var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#6C50FF")).
			Padding(0, 1)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6C50FF")).
			Padding(0, 1)

	activePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#04B575")).
				Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#333333")).
			Padding(0, 1)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	popupStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#04B575")).
			Padding(1, 2)

	popupTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)
)

// YAML config structures
type ToolConfig struct {
	Categories map[string]CategoryConfig `yaml:"categories"`
}

type CategoryConfig struct {
	Icon  string       `yaml:"icon"`
	Tools []ToolEntry  `yaml:"tools"`
}

type ToolEntry struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// Tool represents a tool for the list
type Tool struct {
	Name     string
	Desc     string
	Category string
}

func (t Tool) Title() string       { return t.Name }
func (t Tool) Description() string { return t.Desc }
func (t Tool) FilterValue() string { return t.Name }

// Category represents a tool category
type Category struct {
	Name  string
	Icon  string
	Count int
}

func (c Category) Title() string       { return fmt.Sprintf("%s %s", c.Icon, c.Name) }
func (c Category) Description() string { return fmt.Sprintf("%d tools", c.Count) }
func (c Category) FilterValue() string { return c.Name }

// Panel enum
type Panel int

const (
	CategoryPanel Panel = iota
	ToolPanel
	ActionPanel
)

// Action represents a tool action
type Action struct {
	Name       string
	Desc       string
	Command    string
	ShowOutput bool
}

func (a Action) Title() string       { return a.Name }
func (a Action) Description() string { return a.Desc }
func (a Action) FilterValue() string { return a.Name }

var defaultActions = []list.Item{
	Action{Name: "latest", Desc: "Install latest version", Command: "latest", ShowOutput: true},
	Action{Name: "show current", Desc: "Show current version", Command: "current", ShowOutput: true},
	Action{Name: "show available", Desc: "Show available versions", Command: "available", ShowOutput: true},
	Action{Name: "set version", Desc: "Set active version", Command: "setversion", ShowOutput: true},
}

// Command completion message
type commandDoneMsg struct {
	output     string
	err        error
	showOutput bool
}

// Model is the main application model
type Model struct {
	categories    list.Model
	tools         list.Model
	actions       list.Model
	allTools      []Tool
	activePanel   Panel
	selectedTool  *Tool
	width         int
	height        int
	quitting      bool
	toolManagerScript string

	// Spinner state
	spinner     spinner.Model
	running     bool
	runningText string

	// Popup state
	showPopup    bool
	popupTitle   string
	popupContent viewport.Model
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle popup mode
	if m.showPopup {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc", "q", "enter":
				m.showPopup = false
				return m, nil
			}
			// Allow scrolling in popup
			var cmd tea.Cmd
			m.popupContent, cmd = m.popupContent.Update(msg)
			return m, cmd

		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height
			m.updateListSizes()
			m.updatePopupSize()
			return m, nil
		}
		return m, nil
	}

	// Handle command completion
	if msg, ok := msg.(commandDoneMsg); ok {
		m.running = false
		m.runningText = ""
		if msg.showOutput {
			m.showPopup = true
			if msg.err != nil {
				m.popupContent.SetContent(fmt.Sprintf("Error: %v\n\n%s", msg.err, msg.output))
			} else {
				m.popupContent.SetContent(msg.output)
			}
			m.updatePopupSize()
		}
		return m, nil
	}

	// If running a command, handle spinner and window resize only
	if m.running {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)

		if msg, ok := msg.(tea.WindowSizeMsg); ok {
			m.width = msg.Width
			m.height = msg.Height
			m.updateListSizes()
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "tab":
			if m.activePanel == CategoryPanel {
				m.activePanel = ToolPanel
			} else if m.activePanel == ToolPanel && m.selectedTool != nil {
				m.activePanel = ActionPanel
			} else {
				m.activePanel = CategoryPanel
			}
			return m, nil

		case "shift+tab":
			if m.activePanel == ActionPanel {
				m.activePanel = ToolPanel
			} else if m.activePanel == ToolPanel {
				m.activePanel = CategoryPanel
			}
			return m, nil

		case "esc":
			if m.activePanel == ActionPanel {
				m.activePanel = ToolPanel
				m.selectedTool = nil
			} else if m.activePanel == ToolPanel {
				m.activePanel = CategoryPanel
			}
			return m, nil

		case "enter":
			return m.handleEnter()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateListSizes()
		return m, nil
	}

	// Update active list
	var cmd tea.Cmd
	switch m.activePanel {
	case CategoryPanel:
		m.categories, cmd = m.categories.Update(msg)
		m.updateToolsForCategory()
	case ToolPanel:
		m.tools, cmd = m.tools.Update(msg)
	case ActionPanel:
		m.actions, cmd = m.actions.Update(msg)
	}

	return m, cmd
}

func (m *Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.activePanel {
	case CategoryPanel:
		m.activePanel = ToolPanel
	case ToolPanel:
		if item, ok := m.tools.SelectedItem().(Tool); ok {
			m.selectedTool = &item
			m.activePanel = ActionPanel
		}
	case ActionPanel:
		if m.selectedTool != nil {
			if action, ok := m.actions.SelectedItem().(Action); ok {
				m.running = true
				m.runningText = fmt.Sprintf("%s → %s", m.selectedTool.Name, action.Name)
				m.popupTitle = m.runningText
				return m, tea.Batch(m.executeAction(action), m.spinner.Tick)
			}
		}
	}
	return m, nil
}

func (m *Model) executeAction(action Action) tea.Cmd {
	toolName := m.selectedTool.Name
	command := action.Command
	showOutput := action.ShowOutput

	return func() tea.Msg {
		// Execute tool-manager.sh with tool name and action
		cmd := exec.Command("bash", m.toolManagerScript, toolName, command)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		cmd.Stdin = nil

		// Run in new session to prevent /dev/tty access
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

		err := cmd.Run()

		output := out.String()
		if err != nil && output == "" {
			output = fmt.Sprintf("Command failed: %v", err)
		}

		return commandDoneMsg{output: output, err: err, showOutput: showOutput}
	}
}

func (m *Model) updateToolsForCategory() {
	if item, ok := m.categories.SelectedItem().(Category); ok {
		var filtered []list.Item
		for _, t := range m.allTools {
			if item.Name == "All" || t.Category == item.Name {
				filtered = append(filtered, t)
			}
		}
		m.tools.SetItems(filtered)
	}
}

func (m *Model) updateListSizes() {
	panelWidth := (m.width - 10) / 3
	panelHeight := m.height - 10

	m.categories.SetSize(panelWidth-2, panelHeight-2)
	m.tools.SetSize(panelWidth-2, panelHeight-2)
	m.actions.SetSize(panelWidth-2, panelHeight-2)
}

func (m *Model) updatePopupSize() {
	popupWidth := m.width / 3
	popupHeight := m.height / 3
	if popupWidth < 40 {
		popupWidth = 40
	}
	if popupHeight < 10 {
		popupHeight = 10
	}
	m.popupContent.Width = popupWidth - 6
	m.popupContent.Height = popupHeight - 6
}

func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	// Calculate consistent dimensions
	panelWidth := (m.width - 10) / 3
	panelHeight := m.height - 10

	// Title
	title := titleStyle.Render("🔧 Butler Tools Manager")

	// Panel styles based on active panel
	catStyle := panelStyle.Copy()
	toolStyle := panelStyle.Copy()
	actStyle := panelStyle.Copy()

	switch m.activePanel {
	case CategoryPanel:
		catStyle = activePanelStyle.Copy()
	case ToolPanel:
		toolStyle = activePanelStyle.Copy()
	case ActionPanel:
		actStyle = activePanelStyle.Copy()
	}

	// Render panels with consistent sizes
	catPanel := catStyle.
		Width(panelWidth).
		Height(panelHeight).
		Render(m.categories.View())

	toolPanel := toolStyle.
		Width(panelWidth).
		Height(panelHeight).
		Render(m.tools.View())

	var actPanel string
	if m.selectedTool != nil {
		actPanel = actStyle.
			Width(panelWidth).
			Height(panelHeight).
			Render(m.actions.View())
	} else {
		placeholder := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Padding(2).
			Render("Select a tool\nto see actions")
		actPanel = panelStyle.Copy().
			Width(panelWidth).
			Height(panelHeight).
			Render(placeholder)
	}

	// Join panels horizontally
	panels := lipgloss.JoinHorizontal(lipgloss.Top, catPanel, toolPanel, actPanel)

	// Status bar - show spinner if running
	var statusText string
	if m.running {
		statusText = fmt.Sprintf("%s %s", m.spinner.View(), m.runningText)
	} else {
		statusText = "tab: switch panel • enter: select • esc: back • q: quit"
	}
	status := statusStyle.Width(m.width - 4).Render(statusText)

	// Main view
	mainView := appStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, title, "", panels, "", status),
	)

	// If popup is showing, overlay it
	if m.showPopup {
		popupWidth := m.width / 3
		popupHeight := m.height / 3
		if popupWidth < 40 {
			popupWidth = 40
		}
		if popupHeight < 10 {
			popupHeight = 10
		}

		popupHeader := popupTitleStyle.Render(m.popupTitle)
		popupFooter := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Render("\n[esc/enter • ↑↓ scroll]")

		popupBody := popupStyle.
			Width(popupWidth).
			Height(popupHeight).
			Render(lipgloss.JoinVertical(lipgloss.Left, popupHeader, "", m.popupContent.View(), popupFooter))

		// Center the popup
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			popupBody,
		)
	}

	return mainView
}

// loadToolsConfig loads tools from YAML config file
func loadToolsConfig(configPath string) ([]Tool, map[string]CategoryConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, nil, err
	}

	var config ToolConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, nil, err
	}

	var tools []Tool
	for catName, catConfig := range config.Categories {
		for _, toolEntry := range catConfig.Tools {
			tools = append(tools, Tool{
				Name:     toolEntry.Name,
				Desc:     toolEntry.Description,
				Category: catName,
			})
		}
	}

	return tools, config.Categories, nil
}

func newModel(configPath, toolManagerScript string) Model {
	// Load tools from YAML config
	tools, categoryConfigs, err := loadToolsConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Build categories
	var categories []list.Item
	totalTools := 0

	for catName, catConfig := range categoryConfigs {
		count := len(catConfig.Tools)
		totalTools += count
		categories = append(categories, Category{
			Name:  catName,
			Icon:  catConfig.Icon,
			Count: count,
		})
	}

	// Add "All" category
	categories = append([]list.Item{Category{
		Name:  "All",
		Icon:  "📋",
		Count: totalTools,
	}}, categories...)

	// Create list models with compact delegate
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(2)

	categoryList := list.New(categories, delegate, 0, 0)
	categoryList.Title = "Categories"
	categoryList.SetShowStatusBar(false)
	categoryList.SetFilteringEnabled(false)
	categoryList.SetShowHelp(false)

	// Convert tools to list items
	var toolItems []list.Item
	for _, t := range tools {
		toolItems = append(toolItems, t)
	}

	toolList := list.New(toolItems, delegate, 0, 0)
	toolList.Title = "Tools"
	toolList.SetShowStatusBar(false)
	toolList.SetFilteringEnabled(true)
	toolList.SetShowHelp(false)

	actionList := list.New(defaultActions, delegate, 0, 0)
	actionList.Title = "Actions"
	actionList.SetShowStatusBar(false)
	actionList.SetFilteringEnabled(false)
	actionList.SetShowHelp(false)

	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	// Create viewport for popup
	vp := viewport.New(80, 20)

	return Model{
		categories:        categoryList,
		tools:             toolList,
		actions:           actionList,
		allTools:          tools,
		activePanel:       CategoryPanel,
		spinner:           s,
		popupContent:      vp,
		toolManagerScript: toolManagerScript,
	}
}

func main() {
	// Default paths
	configPath := "/usr/share/apparatus/tools.yaml"
	toolManagerScript := "/opt/bin/tool-manager.sh"

	// Allow overrides via args
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}
	if len(os.Args) > 2 {
		toolManagerScript = os.Args[2]
	}

	p := tea.NewProgram(newModel(configPath, toolManagerScript), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
