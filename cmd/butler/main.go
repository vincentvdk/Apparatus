package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Mode represents which environment butler is running in
type Mode int

const (
	ModeUnknown Mode = iota
	ModeOS
	ModeBox
)

// Styles
var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#6C50FF")).
			Padding(0, 2).
			Bold(true)

	modeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Italic(true).
			Bold(true)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6C50FF")).
			Padding(0, 1)

	activePanelStyle = panelStyle.Copy().
				BorderForeground(lipgloss.Color("#04B575"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#333333")).
			Padding(0, 1)

	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))

	popupStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#04B575")).
			Padding(1, 2)

	popupTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#04B575")).
				Bold(true)
)

// Command completion message
type commandDoneMsg struct {
	output     string
	err        error
	showOutput bool
}

// Popup model
type popupModel struct {
	title    string
	content  string
	quitting bool
	width    int
	height   int
}

func (m popupModel) Init() tea.Cmd {
	return nil
}

func (m popupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "enter":
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m popupModel) View() string {
	if m.quitting {
		return ""
	}

	popupWidth := m.width / 3
	popupHeight := m.height / 3
	if popupWidth < 40 {
		popupWidth = 40
	}
	if popupHeight < 10 {
		popupHeight = 10
	}

	popupHeader := popupTitleStyle.Render(m.title)
	popupFooter := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Render("\n[esc/enter • q to close]")

	popupBody := popupStyle.
		Width(popupWidth).
		Height(popupHeight).
		Render(lipgloss.JoinVertical(lipgloss.Left, popupHeader, "\n", m.content, popupFooter))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, popupBody)
}

// runCommand executes a shell command and returns combined output
func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// runPrivilegedCommand executes a command requiring root via pkexec
func runPrivilegedCommand(description string, command string) (string, error) {
	cmd := exec.Command("pkexec", "bash", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("%v: %s", err, string(output))
	}
	return string(output), nil
}

// detectMode auto-detects the environment
func detectMode() Mode {
	// Check explicit env override
	if mode := os.Getenv("APPARATUS_MODE"); mode != "" {
		switch strings.ToLower(mode) {
		case "os":
			return ModeOS
		case "box":
			return ModeBox
		}
	}

	// Auto-detect: check for environment markers
	_, err := exec.LookPath("hyprctl")
	if err == nil {
		// hyprctl exists but might also exist in box via distrobox-host-exec
		// Check for mise to distinguish
		_, miseErr := exec.LookPath("mise")
		if miseErr == nil {
			// Both exist - check if we're in a distrobox
			if _, err := os.Stat("/run/.containerenv"); err == nil {
				return ModeBox
			}
			return ModeOS
		}
		return ModeOS
	}

	_, err = exec.LookPath("mise")
	if err == nil {
		return ModeBox
	}

	// Check if we're in a distrobox
	if _, err := os.Stat("/run/.containerenv"); err == nil {
		return ModeBox
	}

	return ModeUnknown
}

// applyThemeCLI applies a theme from the command line
func applyThemeCLI(themeName string) {
	themesDir := "/usr/share/apparatus/themes"
	themePath := fmt.Sprintf("%s/%s", themesDir, themeName)

	if _, err := os.Stat(themePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Theme '%s' not found in %s\n", themeName, themesDir)
		os.Exit(1)
	}

	applyTheme(themeName)
	fmt.Printf("Theme applied: %s\n", themeName)
}

// applyFontCLI applies a font from the command line
func applyFontCLI(fontName string) {
	fontMap := map[string]string{
		"ioskeley-mono":    "IoskeleyMono Nerd Font",
		"jetbrains-mono":   "JetBrainsMono Nerd Font",
		"hack-nerd-font":   "Hack Nerd Font",
	}

	fontFamily, ok := fontMap[fontName]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: Unknown font '%s'. Available: ioskeley-mono, jetbrains-mono, hack-nerd-font\n", fontName)
		os.Exit(1)
	}

	applyFont(fontFamily, fontName)
	fmt.Printf("Font applied: %s\n", fontName)
}

// modeString returns a display name for the mode
func modeString(m Mode) string {
	switch m {
	case ModeOS:
		return "🖥  Host OS"
	case ModeBox:
		return "📦  Distrobox"
	default:
		return "❓  Unknown"
	}
}

func printUsage() {
	fmt.Println("Butler - Apparatus OS Configuration Tool")
	fmt.Println()
	fmt.Println("Usage: butler [mode]")
	fmt.Println()
	fmt.Println("Modes:")
	fmt.Println("  os    Run in Host OS mode (system config, distrobox management)")
	fmt.Println("  box   Run in Distrobox mode (dev tool management)")
	fmt.Println()
	fmt.Println("If no mode is specified, butler auto-detects the environment.")
	fmt.Println("Override with APPARATUS_MODE=os or APPARATUS_MODE=box.")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  butler          # Auto-detect")
	fmt.Println("  butler os       # Force OS mode")
	fmt.Println("  butler box      # Force box mode")
	fmt.Println("  APPARATUS_MODE=box butler  # Override via env var")
}

func main() {
	mode := ModeUnknown

	// Check for explicit mode from CLI args
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "os":
			mode = ModeOS
		case "box":
			mode = ModeBox
		case "theme":
			if len(os.Args) > 2 {
				applyThemeCLI(os.Args[2])
				os.Exit(0)
			}
			fmt.Fprintln(os.Stderr, "Usage: butler theme <name>")
			fmt.Fprintln(os.Stderr, "Available themes: catppuccin-mocha, catppuccin-latte")
			os.Exit(1)
		case "font":
			if len(os.Args) > 2 {
				applyFontCLI(os.Args[2])
				os.Exit(0)
			}
			fmt.Fprintln(os.Stderr, "Usage: butler font <name>")
			fmt.Fprintln(os.Stderr, "Available fonts: ioskeley-mono, jetbrains-mono, hack-nerd-font")
			os.Exit(1)
		case "help", "--help", "-h":
			printUsage()
			os.Exit(0)
		case "version", "--version", "-v":
			fmt.Println("butler v0.1.0")
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown mode: %s\n", os.Args[1])
			printUsage()
			os.Exit(1)
		}
	} else {
		mode = detectMode()
	}

	if mode == ModeUnknown {
		fmt.Fprintln(os.Stderr, "Error: Could not detect Apparatus environment.")
		fmt.Fprintln(os.Stderr, "Usage: butler [os|box]  (or set APPARATUS_MODE=os|box)")
		os.Exit(1)
	}

	fmt.Printf("Butler running in %s mode\n\n", modeString(mode))

	var p *tea.Program
	switch mode {
	case ModeOS:
		model := newOSModel()
		p = tea.NewProgram(model, tea.WithAltScreen())
	case ModeBox:
		p = tea.NewProgram(newBoxModel(), tea.WithAltScreen())
	}

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
