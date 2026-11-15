// soundctl - A TUI for managing audio devices on Linux
// Works with both PulseAudio and PipeWire (via pactl)
// Part of Apparatus OS
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Device types
const (
	DeviceTypeOutput = "output"
	DeviceTypeInput  = "input"
)

// Panel types
type panel int

const (
	panelDevices panel = iota
	panelOutputs
	panelInputs
	panelSettings
)

// SoundDevice represents an audio device
type SoundDevice struct {
	Index   uint32
	Name    string
	Desc    string
	Type    string
	Volume  float64
	Muted   bool
	Default bool
}

func (d SoundDevice) Title() string {
	if d.Desc != "" {
		return d.Desc
	}
	return d.Name
}
func (d SoundDevice) Description() string {
	if d.Desc != "" && d.Desc != d.Name {
		return d.Name
	}
	return ""
}
func (d SoundDevice) FilterValue() string { return d.Name + " " + d.Desc }

// ========== Styles ==========

func getAccentColor() lipgloss.Color  { return lipgloss.Color("4") }
func getSuccessColor() lipgloss.Color { return lipgloss.Color("2") }
func getErrorColor() lipgloss.Color   { return lipgloss.Color("1") }

var appStyle = lipgloss.NewStyle().Padding(1, 2)

func getTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(getAccentColor()).
		Padding(0, 2).
		Bold(true)
}

func getPanelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(getAccentColor()).
		Padding(0, 1)
}

func getActivePanelStyle() lipgloss.Style {
	return getPanelStyle().Copy().BorderForeground(getSuccessColor())
}

func getStatusStyle() lipgloss.Style {
	return lipgloss.NewStyle().Faint(true).Padding(0, 1)
}

func getVolumeBar(current, max, width int) string {
	if max == 0 {
		max = 1
	}
	filled := (current * width) / max
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("█", filled)
	empty := strings.Repeat("░", width-filled)
	return lipgloss.NewStyle().Foreground(getSuccessColor()).Render(bar) + 
		lipgloss.NewStyle().Faint(true).Render(empty)
}

func getListDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.NormalTitle = lipgloss.NewStyle()
	delegate.Styles.NormalDesc = lipgloss.NewStyle().Faint(true)
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(getAccentColor()).Bold(true)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(getAccentColor())
	delegate.Styles.DimmedTitle = lipgloss.NewStyle().Faint(true)
	delegate.Styles.DimmedDesc = lipgloss.NewStyle().Faint(true)
	delegate.Styles.FilterMatch = lipgloss.NewStyle().Foreground(getSuccessColor())
	return delegate
}

func getListStyles() list.Styles {
	s := list.DefaultStyles()
	s.Title = lipgloss.NewStyle().Background(getAccentColor()).Padding(0, 1)
	s.Spinner = lipgloss.NewStyle().Faint(true)
	s.FilterPrompt = lipgloss.NewStyle().Foreground(getSuccessColor())
	s.FilterCursor = lipgloss.NewStyle().Foreground(getAccentColor())
	s.StatusBar = lipgloss.NewStyle().Faint(true).Padding(0, 0, 1, 2)
	s.StatusEmpty = lipgloss.NewStyle().Faint(true)
	s.StatusBarActiveFilter = lipgloss.NewStyle().Bold(true)
	s.StatusBarFilterCount = lipgloss.NewStyle().Faint(true)
	s.NoItems = lipgloss.NewStyle().Faint(true)
	s.ArabicPagination = lipgloss.NewStyle().Faint(true)
	s.ActivePaginationDot = lipgloss.NewStyle().Foreground(getAccentColor()).SetString("•")
	s.InactivePaginationDot = lipgloss.NewStyle().Faint(true).SetString("•")
	s.DividerDot = lipgloss.NewStyle().Faint(true).SetString(" • ")
	return s
}

func newStyledList(items []list.Item, delegate list.ItemDelegate, width, height int) list.Model {
	l := list.New(items, delegate, width, height)
	l.Styles = getListStyles()
	l.Help.Styles.ShortKey = lipgloss.NewStyle().Faint(true)
	l.Help.Styles.ShortDesc = lipgloss.NewStyle().Faint(true)
	l.Help.Styles.ShortSeparator = lipgloss.NewStyle().Faint(true)
	l.Help.Styles.FullKey = lipgloss.NewStyle().Faint(true)
	l.Help.Styles.FullDesc = lipgloss.NewStyle().Faint(true)
	l.Help.Styles.FullSeparator = lipgloss.NewStyle().Faint(true)
	l.Help.Styles.Ellipsis = lipgloss.NewStyle().Faint(true)
	return l
}

// ========== Messages ==========

type refreshMsg struct{}
type errorMsg struct {
	text string
}

// ========== Model ==========

type model struct {
	width, height int
	quitting      bool
	activePanel   panel
	
	outputDevices list.Model
	inputDevices  list.Model
	mainMenu     list.Model
	
	outputs []SoundDevice
	inputs  []SoundDevice
	
	volumeValue float64
	volumeTarget *SoundDevice
	
	showPopup    bool
	popupTitle   string
	popupContent string
	popupViewport viewport.Model
	
	lastError string
}

func newModel() *model {
	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().Padding(1)
	
	m := &model{
		activePanel:   panelDevices,
		popupViewport: vp,
	}
	
	m.loadDevices()
	m.initLists()
	
	return m
}

func (m *model) loadDevices() {
	m.outputs = getSinks()
	m.inputs = getSources()
}

func (m *model) initLists() {
	delegate := getListDelegate()
	delegate.SetHeight(1)
	
	mainItems := []list.Item{
		&SoundDevice{Name: "🔊 Output Devices", Desc: "Speakers, headphones", Type: DeviceTypeOutput},
		&SoundDevice{Name: "🎤 Input Devices", Desc: "Microphones, line-in", Type: DeviceTypeInput},
		&SoundDevice{Name: "⚙️  Settings", Desc: "Audio configuration", Type: "settings"},
		&SoundDevice{Name: "🔄 Refresh", Desc: "Reload device list", Type: "refresh"},
	}
	m.mainMenu = newStyledList(mainItems, delegate, 0, 0)
	m.mainMenu.Title = "Sound Control"
	m.mainMenu.SetShowStatusBar(false)
	m.mainMenu.SetFilteringEnabled(false)
	m.mainMenu.SetShowHelp(false)
	
	var outputItems []list.Item
	for i := range m.outputs {
		outputItems = append(outputItems, &m.outputs[i])
	}
	m.outputDevices = newStyledList(outputItems, delegate, 0, 0)
	m.outputDevices.Title = "Output Devices"
	m.outputDevices.SetShowStatusBar(false)
	m.outputDevices.SetFilteringEnabled(true)
	m.outputDevices.SetShowHelp(false)
	
	var inputItems []list.Item
	for i := range m.inputs {
		inputItems = append(inputItems, &m.inputs[i])
	}
	m.inputDevices = newStyledList(inputItems, delegate, 0, 0)
	m.inputDevices.Title = "Input Devices"
	m.inputDevices.SetShowStatusBar(false)
	m.inputDevices.SetFilteringEnabled(true)
	m.inputDevices.SetShowHelp(false)
}

func (m *model) refreshDevices() tea.Cmd {
	return func() tea.Msg {
		m.loadDevices()
		
		delegate := getListDelegate()
		delegate.SetHeight(1)
		
		var outputItems []list.Item
		for i := range m.outputs {
			outputItems = append(outputItems, &m.outputs[i])
		}
		m.outputDevices.SetItems(outputItems)
		
		var inputItems []list.Item
		for i := range m.inputs {
			inputItems = append(inputItems, &m.inputs[i])
		}
		m.inputDevices.SetItems(inputItems)
		
		m.lastError = ""
		return refreshMsg{}
	}
}

// Audio helper functions using pactl (works with both PulseAudio and PipeWire)

func runPactl(args ...string) (string, error) {
	cmd := exec.Command("pactl", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func getSinks() []SoundDevice {
	var devices []SoundDevice
	
	// Get list of sinks
	output, err := runPactl("list", "sinks", "short")
	if err != nil {
		return devices
	}
	
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		
		index, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			continue
		}
		name := parts[1]
		
		// Get detailed info for this sink
		info, err := runPactl("list", "sinks")
		if err != nil {
			continue
		}
		
		// Parse volume, mute, and description from detailed info
		volume, muted, desc := parseSinkInfo(info, uint32(index))
		
		// Check if default
		defaultSink, _ := runPactl("get-default-sink")
		isDefault := defaultSink == name
		
		devices = append(devices, SoundDevice{
			Index:   uint32(index),
			Name:    name,
			Desc:    desc,
			Type:    DeviceTypeOutput,
			Volume:  volume,
			Muted:   muted,
			Default: isDefault,
		})
	}
	
	return devices
}

func getSources() []SoundDevice {
	var devices []SoundDevice
	
	output, err := runPactl("list", "sources", "short")
	if err != nil {
		return devices
	}
	
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		
		index, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			continue
		}
		name := parts[1]
		
		info, err := runPactl("list", "sources")
		if err != nil {
			continue
		}
		
		volume, muted, desc := parseSourceInfo(info, uint32(index))
		
		defaultSource, _ := runPactl("get-default-source")
		isDefault := defaultSource == name
		
		devices = append(devices, SoundDevice{
			Index:   uint32(index),
			Name:    name,
			Desc:    desc,
			Type:    DeviceTypeInput,
			Volume:  volume,
			Muted:   muted,
			Default: isDefault,
		})
	}
	
	return devices
}



func parseSinkInfo(info string, index uint32) (float64, bool, string) {
	lines := strings.Split(info, "\n")
	var inSink bool
	var volume float64 = 50.0
	var muted bool = false
	var desc string = ""
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, fmt.Sprintf("Sink #%d", index)) {
			inSink = true
			continue
		}
		
		if inSink && strings.HasPrefix(line, "Sink #") && !strings.HasPrefix(line, fmt.Sprintf("Sink #%d", index)) {
			break
		}
		
		if inSink {
			if strings.Contains(line, "Description:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) >= 2 {
					desc = strings.TrimSpace(parts[1])
				}
			}
			if strings.Contains(line, "Volume:") {
				// Parse: "Volume: front-left: 32768 / 50% / -16.00 dB"
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					volPart := strings.TrimSpace(parts[1])
					// Split by "/" to get percentage
					percentParts := strings.Split(volPart, "/")
					if len(percentParts) >= 2 {
						percentStr := strings.TrimSpace(percentParts[1])
						percentStr = strings.TrimSuffix(percentStr, "%")
						if pct, err := strconv.ParseFloat(strings.TrimSpace(percentStr), 64); err == nil {
							volume = pct
						}
					}
				}
			}
			if strings.Contains(line, "Mute:") {
				muted = strings.Contains(line, "yes")
			}
		}
	}
	
	return volume, muted, desc
}

func parseSourceInfo(info string, index uint32) (float64, bool, string) {
	lines := strings.Split(info, "\n")
	var inSource bool
	var volume float64 = 50.0
	var muted bool = false
	var desc string = ""
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, fmt.Sprintf("Source #%d", index)) {
			inSource = true
			continue
		}
		
		if inSource && strings.HasPrefix(line, "Source #") && !strings.HasPrefix(line, fmt.Sprintf("Source #%d", index)) {
			break
		}
		
		if inSource {
			if strings.Contains(line, "Description:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) >= 2 {
					desc = strings.TrimSpace(parts[1])
				}
			}
			if strings.Contains(line, "Volume:") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					volPart := strings.TrimSpace(parts[1])
					percentParts := strings.Split(volPart, "/")
					if len(percentParts) >= 2 {
						percentStr := strings.TrimSpace(percentParts[1])
						percentStr = strings.TrimSuffix(percentStr, "%")
						if pct, err := strconv.ParseFloat(strings.TrimSpace(percentStr), 64); err == nil {
							volume = pct
						}
					}
				}
			}
			if strings.Contains(line, "Mute:") {
				muted = strings.Contains(line, "yes")
			}
		}
	}
	
	return volume, muted, desc
}

func (m *model) setDefaultSink(index uint32) error {
	cmd := exec.Command("pactl", "set-default-sink", fmt.Sprintf("%d", index))
	return cmd.Run()
}

func (m *model) setDefaultSource(index uint32) error {
	cmd := exec.Command("pactl", "set-default-source", fmt.Sprintf("%d", index))
	return cmd.Run()
}

func (m *model) setSinkVolume(index uint32, volume float64) error {
	if volume < 0 {
		volume = 0
	}
	if volume > 100 {
		volume = 100
	}
	percentStr := fmt.Sprintf("%d%%", int(volume))
	cmd := exec.Command("pactl", "set-sink-volume", fmt.Sprintf("%d", index), percentStr)
	return cmd.Run()
}

func (m *model) setSourceVolume(index uint32, volume float64) error {
	if volume < 0 {
		volume = 0
	}
	if volume > 100 {
		volume = 100
	}
	percentStr := fmt.Sprintf("%d%%", int(volume))
	cmd := exec.Command("pactl", "set-source-volume", fmt.Sprintf("%d", index), percentStr)
	return cmd.Run()
}

func (m *model) toggleSinkMute(index uint32) error {
	cmd := exec.Command("pactl", "set-sink-mute", fmt.Sprintf("%d", index), "toggle")
	return cmd.Run()
}

func (m *model) toggleSourceMute(index uint32) error {
	cmd := exec.Command("pactl", "set-source-mute", fmt.Sprintf("%d", index), "toggle")
	return cmd.Run()
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.showPopup {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc", "q", "enter":
				m.showPopup = false
				return m, nil
			}
			var cmd tea.Cmd
			m.popupViewport, cmd = m.popupViewport.Update(msg)
			return m, cmd
		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height
			m.popupViewport.Width = m.width / 2
			m.popupViewport.Height = m.height / 2
			return m, nil
		}
		return m, nil
	}
	
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "tab":
			switch m.activePanel {
			case panelDevices:
				m.activePanel = panelOutputs
			case panelOutputs:
				m.activePanel = panelInputs
			case panelInputs:
				m.activePanel = panelSettings
			case panelSettings:
				m.activePanel = panelOutputs
			}
			return m, nil
		case "shift+tab":
			switch m.activePanel {
			case panelDevices:
				m.activePanel = panelSettings
			case panelOutputs:
				m.activePanel = panelSettings
			case panelInputs:
				m.activePanel = panelOutputs
			case panelSettings:
				m.activePanel = panelInputs
			}
			return m, nil
		case "esc":
			m.activePanel = panelDevices
			return m, nil
		case "enter":
			m.handleEnter()
			return m, nil
		case "+", "=":
			m.handleVolumeChange(5)
			return m, nil
		case "-":
			m.handleVolumeChange(-5)
			return m, nil
		case "m":
			m.handleMuteToggle()
			return m, nil
		case "d":
			m.handleSetDefault()
			return m, nil
		case "r":
			return m, m.refreshDevices()
		case "left", "h":
			if m.volumeTarget != nil {
				m.volumeValue -= 5
				if m.volumeValue < 0 {
					m.volumeValue = 0
				}
				var err error
				if m.volumeTarget.Type == DeviceTypeOutput {
					err = m.setSinkVolume(m.volumeTarget.Index, m.volumeValue)
				} else {
					err = m.setSourceVolume(m.volumeTarget.Index, m.volumeValue)
				}
				if err != nil {
					m.lastError = err.Error()
				}
				m.volumeTarget.Volume = m.volumeValue
				_ = m.refreshDevices()
			}
			return m, nil
		case "right", "l":
			if m.volumeTarget != nil {
				m.volumeValue += 5
				if m.volumeValue > 100 {
					m.volumeValue = 100
				}
				var err error
				if m.volumeTarget.Type == DeviceTypeOutput {
					err = m.setSinkVolume(m.volumeTarget.Index, m.volumeValue)
				} else {
					err = m.setSourceVolume(m.volumeTarget.Index, m.volumeValue)
				}
				if err != nil {
					m.lastError = err.Error()
				}
				m.volumeTarget.Volume = m.volumeValue
				_ = m.refreshDevices()
			}
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updatePanelSizes()
		return m, nil
	case refreshMsg:
		return m, nil
	case errorMsg:
		m.lastError = msg.text
		return m, nil
	}
	
	var cmd tea.Cmd
	switch m.activePanel {
	case panelDevices:
		m.mainMenu, cmd = m.mainMenu.Update(msg)
	case panelOutputs:
		m.outputDevices, cmd = m.outputDevices.Update(msg)
	case panelInputs:
		m.inputDevices, cmd = m.inputDevices.Update(msg)
	}
	
	return m, cmd
}

func (m *model) handleEnter() {
	switch m.activePanel {
	case panelDevices:
		if item, ok := m.mainMenu.SelectedItem().(*SoundDevice); ok {
			switch item.Type {
			case DeviceTypeOutput:
				m.activePanel = panelOutputs
			case DeviceTypeInput:
				m.activePanel = panelInputs
			case "settings":
				m.activePanel = panelSettings
			case "refresh":
				_ = m.refreshDevices()
			}
		}
	case panelOutputs:
		if item, ok := m.outputDevices.SelectedItem().(*SoundDevice); ok {
			m.volumeTarget = item
			m.volumeValue = item.Volume
		}
	case panelInputs:
		if item, ok := m.inputDevices.SelectedItem().(*SoundDevice); ok {
			m.volumeTarget = item
			m.volumeValue = item.Volume
		}
	case panelSettings:
		m.showPopup = true
		m.popupTitle = "Settings & Help"
		m.popupContent = `Keyboard Shortcuts

Navigation:
  tab / shift+tab  Switch between panels
  up/dn              Navigate device list
  enter            Select / adjust volume
  esc              Return to main menu

Device Control:
  + / =            Increase volume by 5%
  -               Decrease volume by 5%
  m               Toggle mute
  d               Set as default device
  r               Refresh device list

General:
  q               Quit application

About:
  - Uses pactl for audio control (PulseAudio/PipeWire)
  - Works with both PulseAudio and PipeWire
  - Changes take effect immediately
  - Default device changes affect new applications

Requirements:
  - pactl command must be available
  - For PipeWire: pipewire-pulse must be installed`
		m.popupViewport.SetContent(m.popupContent)
	}
}

func (m *model) handleVolumeChange(delta int) {
	var device *SoundDevice
	
	switch m.activePanel {
	case panelOutputs:
		if item, ok := m.outputDevices.SelectedItem().(*SoundDevice); ok {
			device = item
		}
	case panelInputs:
		if item, ok := m.inputDevices.SelectedItem().(*SoundDevice); ok {
			device = item
		}
	default:
		return
	}
	
	if device == nil {
		return
	}
	
	newVolume := device.Volume + float64(delta)
	if newVolume < 0 {
		newVolume = 0
	}
	if newVolume > 100 {
		newVolume = 100
	}
	
	var err error
	if device.Type == DeviceTypeOutput {
		err = m.setSinkVolume(device.Index, newVolume)
	} else {
		err = m.setSourceVolume(device.Index, newVolume)
	}
	
	if err != nil {
		m.lastError = err.Error()
		m.showErrorPopup("Volume Error", err.Error())
	} else {
		device.Volume = newVolume
		_ = m.refreshDevices()
	}
}

func (m *model) handleMuteToggle() {
	var device *SoundDevice
	
	switch m.activePanel {
	case panelOutputs:
		if item, ok := m.outputDevices.SelectedItem().(*SoundDevice); ok {
			device = item
		}
	case panelInputs:
		if item, ok := m.inputDevices.SelectedItem().(*SoundDevice); ok {
			device = item
		}
	default:
		return
	}
	
	if device == nil {
		return
	}
	
	var err error
	if device.Type == DeviceTypeOutput {
		err = m.toggleSinkMute(device.Index)
	} else {
		err = m.toggleSourceMute(device.Index)
	}
	
	if err != nil {
		m.lastError = err.Error()
		m.showErrorPopup("Mute Error", err.Error())
	} else {
		device.Muted = !device.Muted
		_ = m.refreshDevices()
	}
}

func (m *model) handleSetDefault() {
	var device *SoundDevice
	
	switch m.activePanel {
	case panelOutputs:
		if item, ok := m.outputDevices.SelectedItem().(*SoundDevice); ok {
			device = item
		}
	case panelInputs:
		if item, ok := m.inputDevices.SelectedItem().(*SoundDevice); ok {
			device = item
		}
	default:
		return
	}
	
	if device == nil {
		return
	}
	
	var err error
	if device.Type == DeviceTypeOutput {
		err = m.setDefaultSink(device.Index)
	} else {
		err = m.setDefaultSource(device.Index)
	}
	
	if err != nil {
		m.lastError = err.Error()
		m.showErrorPopup("Default Error", err.Error())
	} else {
		for i := range m.outputs {
			m.outputs[i].Default = (m.outputs[i].Index == device.Index)
		}
		for i := range m.inputs {
			m.inputs[i].Default = (m.inputs[i].Index == device.Index)
		}
		_ = m.refreshDevices()
	}
}

func (m *model) showErrorPopup(title, content string) {
	m.showPopup = true
	m.popupTitle = title
	m.popupContent = content
	m.popupViewport.SetContent(content)
}

func (m *model) updatePanelSizes() {
	panelWidth := (m.width - 10) / 2
	panelHeight := m.height - 8
	m.mainMenu.SetSize(panelWidth-2, panelHeight-2)
	m.outputDevices.SetSize(panelWidth-2, panelHeight-2)
	m.inputDevices.SetSize(panelWidth-2, panelHeight-2)
}

func (m model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}
	
	if m.showPopup {
		popupWidth := m.width / 2
		popupHeight := m.height / 2
		if popupWidth < 40 {
			popupWidth = 40
		}
		if popupHeight < 10 {
			popupHeight = 10
		}
		
		popupHeader := lipgloss.NewStyle().
			Foreground(getSuccessColor()).
			Bold(true).
			Render(m.popupTitle)
		
		popupFooter := lipgloss.NewStyle().
			Faint(true).
			Render("\n[esc/enter or q to close]")
		
		popupBody := lipgloss.JoinVertical(lipgloss.Left, popupHeader, m.popupViewport.View(), popupFooter)
		
		popupStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(getSuccessColor()).
			Padding(1, 2).
			Width(popupWidth).
			Height(popupHeight)
		
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, popupStyle.Render(popupBody))
	}
	
	
	panelWidth := (m.width - 10) / 2
	panelHeight := m.height - 8
	
	title := getTitleStyle().Render("🔊 Sound Control")
	
	var leftPanel, rightPanel string
	
	switch m.activePanel {
	case panelDevices:
		leftPanel = m.mainMenu.View()
		rightPanel = m.renderQuickInfo(panelWidth, panelHeight)
	case panelOutputs:
		leftPanel = m.mainMenu.View()
		rightPanel = m.renderOutputDevices(panelWidth, panelHeight)
	case panelInputs:
		leftPanel = m.mainMenu.View()
		rightPanel = m.renderInputDevices(panelWidth, panelHeight)
	case panelSettings:
		leftPanel = m.mainMenu.View()
		rightPanel = m.renderSettings(panelWidth, panelHeight)
	default:
		leftPanel = m.mainMenu.View()
		rightPanel = m.renderQuickInfo(panelWidth, panelHeight)
	}
	
	leftStyle := getPanelStyle().Copy()
	rightStyle := getPanelStyle().Copy()
	
	switch m.activePanel {
	case panelOutputs:
		rightStyle = getActivePanelStyle().Copy()
	case panelInputs:
		rightStyle = getActivePanelStyle().Copy()
	case panelSettings:
		rightStyle = getActivePanelStyle().Copy()
	default:
		leftStyle = getActivePanelStyle().Copy()
	}
	
	leftPanel = leftStyle.Width(panelWidth).Height(panelHeight).Render(leftPanel)
	rightPanel = rightStyle.Width(panelWidth).Height(panelHeight).Render(rightPanel)
	
	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	
	statusText := "tab: switch • enter: select • +/-: vol • m: mute • d: default • r: refresh • q: quit"
	if m.lastError != "" {
		statusText += "   " + lipgloss.NewStyle().Foreground(getErrorColor()).Render("Error: "+m.lastError)
	}
	status := getStatusStyle().Width(m.width - 4).Render(statusText)
	
	return appStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, title, panels, status),
	)
}

func (m *model) renderQuickInfo(width, height int) string {
	info := lipgloss.NewStyle().Padding(1)
	
	var sb strings.Builder
	sb.WriteString("Audio Information\n\n")
	sb.WriteString(fmt.Sprintf("Outputs: %d\n", len(m.outputs)))
	sb.WriteString(fmt.Sprintf("Inputs:  %d\n", len(m.inputs)))
	
	for _, dev := range m.outputs {
		if dev.Default {
			sb.WriteString(fmt.Sprintf("\nDefault Output: %s", dev.Name))
		}
	}
	for _, dev := range m.inputs {
		if dev.Default {
			sb.WriteString(fmt.Sprintf("\nDefault Input:  %s", dev.Name))
		}
	}
	
	return info.Render(sb.String())
}

func (m *model) renderOutputDevices(width, height int) string {
	if len(m.outputs) == 0 {
		return lipgloss.NewStyle().Faint(true).Padding(2).Render("No output devices found")
	}
	
	var sb strings.Builder
	sb.WriteString(m.outputDevices.View())
	
	if item, ok := m.outputDevices.SelectedItem().(*SoundDevice); ok {
		sb.WriteString("\n")
		volIcon := "🔊"
		if item.Muted {
			volIcon = "🔇"
		}
		defaultMarker := ""
		if item.Default {
			defaultMarker = " [DEFAULT]"
		}
		sb.WriteString(fmt.Sprintf("%s Volume: %.0f%%%s", volIcon, item.Volume, defaultMarker))
		
		// Show slider if this is the volume target
		if m.volumeTarget != nil && m.volumeTarget.Name == item.Name {
			volumePercent := int(m.volumeValue)
			barWidth := 30
			volumeBar := getVolumeBar(volumePercent, 100, barWidth)
			sb.WriteString("\n")
			sb.WriteString(fmt.Sprintf("[%s] %d%%", volumeBar, volumePercent))
			sb.WriteString("\n[←/→ or h/l to adjust]")
		}
	}
	
	return sb.String()
}

func (m *model) renderInputDevices(width, height int) string {
	if len(m.inputs) == 0 {
		return lipgloss.NewStyle().Faint(true).Padding(2).Render("No input devices found")
	}
	
	var sb strings.Builder
	sb.WriteString(m.inputDevices.View())
	
	if item, ok := m.inputDevices.SelectedItem().(*SoundDevice); ok {
		sb.WriteString("\n")
		volIcon := "🎤"
		if item.Muted {
			volIcon = "🔇"
		}
		defaultMarker := ""
		if item.Default {
			defaultMarker = " [DEFAULT]"
		}
		sb.WriteString(fmt.Sprintf("%s Volume: %.0f%%%s", volIcon, item.Volume, defaultMarker))
		
		// Show slider if this is the volume target
		if m.volumeTarget != nil && m.volumeTarget.Name == item.Name {
			volumePercent := int(m.volumeValue)
			barWidth := 30
			volumeBar := getVolumeBar(volumePercent, 100, barWidth)
			sb.WriteString("\n")
			sb.WriteString(fmt.Sprintf("[%s] %d%%", volumeBar, volumePercent))
			sb.WriteString("\n[←/→ or h/l to adjust]")
		}
	}
	
	return sb.String()
}

func (m *model) renderSettings(width, height int) string {
	info := lipgloss.NewStyle().Padding(1)
	return info.Render(`Keyboard Shortcuts

Navigation:
  tab / shift+tab  Switch panels
  up/dn            Navigate list
  enter          Select / adjust
  esc            Back to menu

Control:
  + / =          Volume +5%
  -             Volume -5%
  m             Toggle mute
  d             Set default
  r             Refresh

Quit:
  q             Quit

Notes:
  Uses pactl (PulseAudio/PipeWire)
  Works with both audio systems`)
}

func main() {
	// Check if pactl is available
	if _, err := exec.LookPath("pactl"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: pactl not found.\n\n")
		fmt.Fprintf(os.Stderr, "Please install PulseAudio or PipeWire:\n")
		fmt.Fprintf(os.Stderr, "  For PulseAudio: pulseaudio\n")
		fmt.Fprintf(os.Stderr, "  For PipeWire: pipewire pipewire-pulse\n")
		os.Exit(1)
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	p := tea.NewProgram(newModel(), tea.WithAltScreen(), tea.WithContext(ctx))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
