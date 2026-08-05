package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

// Box mode structures
type toolConfig struct {
	Categories map[string]categoryConfig `yaml:"categories"`
}

type categoryConfig struct {
	Icon  string     `yaml:"icon"`
	Tools []toolEntry `yaml:"tools"`
}

type toolEntry struct {
	Name        string     `yaml:"name"`
	Description string     `yaml:"description"`
	LSP         *lspConfig `yaml:"lsp,omitempty"`
}

type lspConfig struct {
	Name       string `yaml:"name"`
	InstallCmd string `yaml:"install_cmd"`
}

// Box mode items
type boxTool struct {
	Name     string
	Desc     string
	Category string
	LSP      *lspConfig
}

func (t boxTool) Title() string       { return t.Name }
func (t boxTool) Description() string { return t.Desc }
func (t boxTool) FilterValue() string { return t.Name }
func (t boxTool) HasLSP() bool        { return t.LSP != nil }

type boxCategory struct {
	Name  string
	Icon  string
	Count int
}

func (c boxCategory) Title() string       { return fmt.Sprintf("%s %s", c.Icon, c.Name) }
func (c boxCategory) Description() string { return fmt.Sprintf("%d tools", c.Count) }
func (c boxCategory) FilterValue() string { return c.Name }

// Box mode panel types
type boxPanel int

const (
	boxCategoryPanel boxPanel = iota
	boxToolPanel
	boxActionPanel
)

// Box mode model
type boxModel struct {
	width, height int
	quitting      bool
	activePanel   boxPanel
	selectedTool  *boxTool
	allTools      []boxTool
	categories    list.Model
	tools         list.Model
	actions       list.Model
	running       bool
	spinner       spinner.Model
	runningText   string
	showPopup     bool
	popupTitle    string
	popupContent  viewport.Model

	// Paths
	configPath        string
	toolManagerScript string
	lspTemplatesDir   string

	// Pre-computed category items (includes "All")
	categoriesItems []list.Item
}

func newBoxModel() *boxModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	vp := viewport.New(80, 20)

	// Default paths (override via env vars)
	configPath := os.Getenv("BUTLER_CONFIG_PATH")
	if configPath == "" {
		configPath = "/usr/share/apparatus/tools.yaml"
	}
	toolManagerScript := os.Getenv("BUTLER_TOOL_MANAGER")
	if toolManagerScript == "" {
		toolManagerScript = "/opt/bin/tool-manager.sh"
	}
	lspTemplatesDir := os.Getenv("BUTLER_LSP_TEMPLATES")
	if lspTemplatesDir == "" {
		lspTemplatesDir = "/usr/share/apparatus/lsp"
	}

	m := &boxModel{
		configPath:        configPath,
		toolManagerScript: toolManagerScript,
		lspTemplatesDir:   lspTemplatesDir,
		spinner:           s,
		popupContent:      vp,
	}

	m.loadTools()
	m.buildCategories()
	m.initLists()

	return m
}

func (m *boxModel) loadTools() {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not load config %s: %v\n", m.configPath, err)
		return
	}

	var config toolConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not parse config: %v\n", err)
		return
	}

	for catName, catConfig := range config.Categories {
		for _, toolEntry := range catConfig.Tools {
			m.allTools = append(m.allTools, boxTool{
				Name:     toolEntry.Name,
				Desc:     toolEntry.Description,
				Category: catName,
				LSP:      toolEntry.LSP,
			})
		}
	}
}

func (m *boxModel) buildCategories() {
	var categories []list.Item
	totalTools := 0

	for catName, catConfig := range m.categoriesFromConfig() {
		count := len(catConfig.Tools)
		totalTools += count
		categories = append(categories, boxCategory{
			Name:  catName,
			Icon:  catConfig.Icon,
			Count: count,
		})
	}

	// Add "All" category
	categories = append([]list.Item{boxCategory{
		Name:  "All",
		Icon:  "📋",
		Count: totalTools,
	}}, categories...)

	m.categoriesItems = categories
}

func (m *boxModel) categoriesFromConfig() map[string]categoryConfig {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return map[string]categoryConfig{}
	}
	var config toolConfig
	yaml.Unmarshal(data, &config)
	return config.Categories
}

func (m *boxModel) initLists() {
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(1)

	// Category list
	catList := list.New(m.categoriesItems, delegate, 0, 0)
	catList.Title = "Categories"
	catList.SetShowStatusBar(false)
	catList.SetFilteringEnabled(false)
	catList.SetShowHelp(false)
	m.categories = catList

	// Tool list (show all initially)
	var toolItems []list.Item
	for _, t := range m.allTools {
		toolItems = append(toolItems, t)
	}
	toolList := list.New(toolItems, delegate, 0, 0)
	toolList.Title = "Tools"
	toolList.SetShowStatusBar(false)
	toolList.SetFilteringEnabled(true)
	toolList.SetShowHelp(false)
	m.tools = toolList

	// Action list
	actionItems := []list.Item{
		boxTool{Name: "latest", Desc: "Install latest version", Category: "action"},
		boxTool{Name: "show current", Desc: "Show current version", Category: "action"},
		boxTool{Name: "show available", Desc: "Show available versions", Category: "action"},
		boxTool{Name: "set version", Desc: "Set active version", Category: "action"},
	}
	actionList := list.New(actionItems, delegate, 0, 0)
	actionList.Title = "Actions"
	actionList.SetShowStatusBar(false)
	actionList.SetFilteringEnabled(false)
	actionList.SetShowHelp(false)
	m.actions = actionList
}

func (m *boxModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *boxModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle popup mode
	if m.showPopup {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc", "q", "enter":
				m.showPopup = false
				return m, nil
			}
			var cmd tea.Cmd
			m.popupContent, cmd = m.popupContent.Update(msg)
			return m, cmd
		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height
			m.updatePanelSizes()
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
				m.popupTitle = "Error"
				m.popupContent.SetContent(fmt.Sprintf("Error: %v\n\n%s", msg.err, msg.output))
			} else {
				m.popupTitle = "Output"
				m.popupContent.SetContent(msg.output)
			}
			m.updatePopupSize()
		}
		return m, nil
	}

	// If running a command
	if m.running {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if win, ok := msg.(tea.WindowSizeMsg); ok {
			m.width = win.Width
			m.height = win.Height
			m.updatePanelSizes()
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
			m.advancePanel(1)
			return m, nil
		case "shift+tab":
			m.advancePanel(-1)
			return m, nil
		case "esc":
			m.retreatPanel()
			return m, nil
		case "enter":
			return m.handleEnter()
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updatePanelSizes()
		return m, nil
	}

	// Update active list
	var cmd tea.Cmd
	switch m.activePanel {
	case boxCategoryPanel:
		m.categories, cmd = m.categories.Update(msg)
		m.updateToolsForCategory()
	case boxToolPanel:
		m.tools, cmd = m.tools.Update(msg)
	case boxActionPanel:
		m.actions, cmd = m.actions.Update(msg)
	}

	return m, cmd
}

func (m *boxModel) advancePanel(dir int) {
	next := int(m.activePanel) + dir
	if next < 0 {
		next = 2
	}
	if next > 2 {
		next = 0
	}
	m.activePanel = boxPanel(next)
}

func (m *boxModel) retreatPanel() {
	switch m.activePanel {
	case boxActionPanel:
		m.activePanel = boxToolPanel
		m.selectedTool = nil
	case boxToolPanel:
		m.activePanel = boxCategoryPanel
	}
}

func (m *boxModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.activePanel {
	case boxCategoryPanel:
		m.activePanel = boxToolPanel
	case boxToolPanel:
		if item, ok := m.tools.SelectedItem().(boxTool); ok {
			m.selectedTool = &item
			m.activePanel = boxActionPanel
		}
	case boxActionPanel:
		if m.selectedTool != nil {
			if action, ok := m.actions.SelectedItem().(boxTool); ok {
				m.running = true
				m.runningText = fmt.Sprintf("%s → %s", m.selectedTool.Name, action.Name)
				m.popupTitle = m.runningText
				return m, tea.Batch(m.executeToolAction(action.Name), m.spinner.Tick)
			}
		}
	}
	return m, nil
}

func (m *boxModel) updateToolsForCategory() {
	if item, ok := m.categories.SelectedItem().(boxCategory); ok {
		var filtered []list.Item
		for _, t := range m.allTools {
			if item.Name == "All" || t.Category == item.Name {
				filtered = append(filtered, t)
			}
		}
		m.tools.SetItems(filtered)
	}
}

func (m *boxModel) updatePanelSizes() {
	panelWidth := (m.width - 10) / 3
	panelHeight := m.height - 8

	m.categories.SetSize(panelWidth-2, panelHeight-2)
	m.tools.SetSize(panelWidth-2, panelHeight-2)
	m.actions.SetSize(panelWidth-2, panelHeight-2)
}

func (m *boxModel) updatePopupSize() {
	popupWidth := m.width / 2
	popupHeight := m.height / 2
	if popupWidth < 40 {
		popupWidth = 40
	}
	if popupHeight < 10 {
		popupHeight = 10
	}
	m.popupContent.Width = popupWidth - 6
	m.popupContent.Height = popupHeight - 6
}

func (m *boxModel) executeToolAction(action string) tea.Cmd {
	toolName := m.selectedTool.Name
	lsp := m.selectedTool.LSP

	return func() tea.Msg {
		var out bytes.Buffer

		// Execute tool-manager.sh
		cmd := exec.Command("bash", m.toolManagerScript, toolName, action)
		cmd.Stdout = &out
		cmd.Stderr = &out
		cmd.Stdin = nil
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

		err := cmd.Run()

		// If installing latest and tool has LSP config, set up LSP
		if action == "latest" && lsp != nil && err == nil {
			out.WriteString("\n--- LSP Setup ---\n")

			// Install LSP
			out.WriteString(fmt.Sprintf("Installing LSP (%s)...\n", lsp.Name))
			lspCmd := exec.Command("bash", "-c", lsp.InstallCmd)
			lspCmd.Stdout = &out
			lspCmd.Stderr = &out
			lspCmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
			if lspErr := lspCmd.Run(); lspErr != nil {
				out.WriteString(fmt.Sprintf("LSP install warning: %v\n", lspErr))
			}

			// Copy LSP config template if it exists
			homeDir, _ := os.UserHomeDir()
			nvimLspDir := filepath.Join(homeDir, ".config", "nvim", "lsp")
			lspConfigSrc := filepath.Join(m.lspTemplatesDir, lsp.Name+".lua")
			lspConfigDst := filepath.Join(nvimLspDir, lsp.Name+".lua")

			os.MkdirAll(nvimLspDir, 0755)

			if srcData, readErr := os.ReadFile(lspConfigSrc); readErr == nil {
				if writeErr := os.WriteFile(lspConfigDst, srcData, 0644); writeErr == nil {
					out.WriteString(fmt.Sprintf("LSP config copied to %s\n", lspConfigDst))
				} else {
					out.WriteString(fmt.Sprintf("Failed to write LSP config: %v\n", writeErr))
				}
			} else {
				out.WriteString(fmt.Sprintf("LSP config template not found: %s\n", lspConfigSrc))
			}
		}

		output := out.String()
		if err != nil && output == "" {
			output = fmt.Sprintf("Command failed: %v", err)
		}

		return commandDoneMsg{output: output, err: err, showOutput: true}
	}
}

func (m boxModel) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	panelWidth := (m.width - 10) / 3
	panelHeight := m.height - 8

	// Title
	title := titleStyle.Render("🔧 Butler") + "  " + modeStyle.Render(modeString(ModeBox))

	// Panel styles
	catStyle := panelStyle.Copy()
	toolStyle := panelStyle.Copy()
	actStyle := panelStyle.Copy()

	switch m.activePanel {
	case boxCategoryPanel:
		catStyle = activePanelStyle.Copy()
	case boxToolPanel:
		toolStyle = activePanelStyle.Copy()
	case boxActionPanel:
		actStyle = activePanelStyle.Copy()
	}

	// Render panels
	catPanel := catStyle.Width(panelWidth).Height(panelHeight).Render(m.categories.View())
	toolPanel := toolStyle.Width(panelWidth).Height(panelHeight).Render(m.tools.View())

	var actPanel string
	if m.selectedTool != nil {
		actPanel = actStyle.Width(panelWidth).Height(panelHeight).Render(m.actions.View())
	} else {
		placeholder := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Padding(2).
			Render("Select a tool\nto see actions")
		actPanel = panelStyle.Copy().Width(panelWidth).Height(panelHeight).Render(placeholder)
	}

	panels := lipgloss.JoinHorizontal(lipgloss.Top, catPanel, toolPanel, actPanel)

	// Status bar
	var statusText string
	if m.running {
		statusText = fmt.Sprintf("%s %s", m.spinner.View(), m.runningText)
	} else {
		statusText = "tab: switch panel • enter: select • esc: back • q: quit"
	}
	status := statusStyle.Width(m.width - 4).Render(statusText)

	// Main view
	mainView := appStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, title, panels, status),
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
		popupFooter := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Render("\n[esc/enter • ↑↓ scroll • q to close]")

		popupBody := popupStyle.
			Width(popupWidth).
			Height(popupHeight).
			Render(lipgloss.JoinVertical(lipgloss.Left, popupHeader, m.popupContent.View(), popupFooter))

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, popupBody)
	}

	return mainView
}
