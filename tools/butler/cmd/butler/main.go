package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// syncConfigsToDistroboxCLI syncs configs to a specific distrobox
func syncConfigsToDistroboxCLI(name string) (string, error) {
	// Get the distrobox info
	homeDir, err := getDistroboxHomeDirCLI(name)
	if err != nil {
		return "", fmt.Errorf("failed to get home dir for %s: %v", name, err)
	}
	
	src := "/usr/share/apparatus"
	dst := homeDir
	if dst == "" {
		dst = filepath.Join("/home", name, ".config")
	} else {
		dst = filepath.Join(dst, ".config")
	}
	
	// Ensure destination exists
	if err := os.MkdirAll(dst, 0755); err != nil {
		return "", fmt.Errorf("failed to create config dir: %v", err)
	}
	
	// Sync configs
	configs := []string{"hypr", "kitty", "waybar", "mako", "walker", "uwsm", "satty", "atuin", "nvim", "zsh"}
	var output strings.Builder
	
	for _, config := range configs {
		srcPath := filepath.Join(src, config)
		dstPath := filepath.Join(dst, config)
		
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue
		}
		
		os.RemoveAll(dstPath)
		if err := copyDirCLI(srcPath, dstPath); err != nil {
			return output.String(), fmt.Errorf("failed to copy %s: %v", config, err)
		}
		output.WriteString(fmt.Sprintf("Synced %s\n", config))
	}
	
	// Sync themes
	themesSrc := filepath.Join(src, "themes")
	themesDst := filepath.Join(dst, "apparatus", "themes")
	os.RemoveAll(themesDst)
	if err := copyDirCLI(themesSrc, themesDst); err != nil {
		return output.String(), fmt.Errorf("failed to copy themes: %v", err)
	}
	output.WriteString("Synced themes\n")
	
	return output.String(), nil
}

// getDistroboxHomeDirCLI returns the home directory for a distrobox
func getDistroboxHomeDirCLI(name string) (string, error) {
	cmd := exec.Command("podman", "inspect", name, "--format", "{{json .Args}}")
	output, err := cmd.Output()
	if err != nil {
		return "", nil
	}
	
	var args []string
	if err := json.Unmarshal(output, &args); err == nil {
		for i, arg := range args {
			if arg == "--home" && i+1 < len(args) {
				return args[i+1], nil
			}
		}
	}
	return "", nil
}

// copyDirCLI copies a directory tree from src to dst
func copyDirCLI(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		
		if entry.IsDir() {
			if err := copyDirCLI(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return err
			}
		}
	}
	
	return nil
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
	fmt.Println("Usage: butler [mode|command]")
	fmt.Println()
	fmt.Println("Modes:")
	fmt.Println("  os    Run in Host OS mode (system config, distrobox management)")
	fmt.Println("  box   Run in Distrobox mode (dev tool management)")
	fmt.Println()
	fmt.Println("Commands (OS mode only):")
	fmt.Println("  theme <name>   Apply theme (catppuccin-mocha, catppuccin-latte)")
	fmt.Println("  font <name>   Apply font (ioskeley-mono, jetbrains-mono, hack-nerd-font)")
	fmt.Println("  sync [name]    Sync configs to all distroboxes (or specific one)")
	fmt.Println()
	fmt.Println("If no mode is specified, butler auto-detects the environment.")
	fmt.Println("Override with APPARATUS_MODE=os or APPARATUS_MODE=box.")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  butler          # Auto-detect, launch TUI")
	fmt.Println("  butler os       # Force OS mode, launch TUI")
	fmt.Println("  butler box      # Force box mode, launch TUI")
	fmt.Println("  butler sync     # Sync configs to all distroboxes")
	fmt.Println("  butler sync mybox  # Sync configs to specific distrobox")
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
		case "sync":
			if len(os.Args) > 2 {
				// Sync specific distrobox
				output, err := syncConfigsToDistroboxCLI(os.Args[2])
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n%s\n", err, output)
					os.Exit(1)
				}
				fmt.Println(output)
			} else {
				// Sync all distroboxes
				output, err := syncConfigsToAllDistroboxes()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n%s\n", err, output)
					os.Exit(1)
				}
				fmt.Println(output)
			}
			os.Exit(0)
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
