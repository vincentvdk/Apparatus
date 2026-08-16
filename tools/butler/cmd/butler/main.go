package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
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

// Butler renders itself using the terminal's own ANSI 16-color palette
// (indices 0-15) rather than literal hex values. lipgloss/termenv pass an
// ANSI index straight through as a raw SGR code; the terminal resolves it
// against whatever palette it currently has loaded. There is no querying
// and nothing to cache, so this can never go stale after a theme switch --
// the very next repaint uses whatever the terminal has live at that moment.
//
// Two approaches were tried and discarded before this one:
//  1. A hardcoded dark/light palette guessed from the theme name -- wrong
//     whenever a theme didn't fit the guess (e.g. tokyonight-moon is dark
//     but matched the "light" bucket).
//  2. Parsing the real hex values out of the active theme's kitty.conf and
//     caching them, plus forcing lipgloss.SetHasDarkBackground() to match.
//     This still went stale in-session because it duplicated state that
//     already lives in the terminal, and fighting lipgloss's own cached
//     termenv.HasDarkBackground() (queried once via OSC 11 and never
//     re-checked) caused other elements to lose their color entirely.
//
// color4 (blue/accent) and color2 (green/success) are valid hex in every
// theme file this project ships. color8 (bright black, the conventional
// "muted" slot) is NOT -- it's an unresolved template placeholder in all
// four Tokyo Night variants' kitty.conf (a separate bug in the theme
// generator, worth fixing there). So muted text uses Faint(true) instead
// of a color: it dims whatever the terminal's real foreground is, which
// works correctly regardless of that bug.
func getAccentColor() lipgloss.Color {
	return lipgloss.Color("4")
}

func getSuccessColor() lipgloss.Color {
	return lipgloss.Color("2")
}

func getErrorColor() lipgloss.Color {
	return lipgloss.Color("1")
}

// currentThemeName reads the theme name butler last wrote via applyTheme(),
// for display only (e.g. an orientation/status line) -- not used for any
// color decision, see the comment above getAccentColor().
func currentThemeName() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(homeDir, ".config", "apparatus", "current-theme"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// currentFontName reads the font name butler last wrote via applyFont(), for
// display only.
func currentFontName() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(homeDir, ".config", "apparatus", "current-font"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// Styles
//
// Foreground is deliberately left unset almost everywhere below: an unset
// Foreground/Background means "use the terminal's real default text/background
// color," which is exactly the theme's `foreground`/`background` value from
// kitty.conf, applied by the terminal itself. There is no ANSI index for
// "the terminal's actual default fg/bg" (that's a distinct OSC 10/11 concept
// from the 16-color table), so the only correct way to get it is to not set
// a color at all and let it inherit through.
var appStyle = lipgloss.NewStyle().Padding(1, 2)

func getTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(getAccentColor()).
		Padding(0, 2).
		Bold(true)
}

func getModeStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(getSuccessColor()).
		Italic(true).
		Bold(true)
}

func getPanelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(getAccentColor()).
		Padding(0, 1)
}

func getActivePanelStyle() lipgloss.Style {
	return getPanelStyle().Copy().
		BorderForeground(getSuccessColor())
}

func getStatusStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Faint(true).
		Padding(0, 1)
}

func getSpinnerStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(getSuccessColor())
}

// getPopupStyle returns the popup border style. isError switches the border
// to the error color (ANSI red) instead of success (ANSI green) -- popups
// showing a failed command should not look identical to ones showing normal
// output.
func getPopupStyle(isError bool) lipgloss.Style {
	color := getSuccessColor()
	if isError {
		color = getErrorColor()
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(1, 2)
}

func getPopupTitleStyle(isError bool) lipgloss.Style {
	color := getSuccessColor()
	if isError {
		color = getErrorColor()
	}
	return lipgloss.NewStyle().
		Foreground(color).
		Bold(true)
}

// getListDelegate creates a theme-aware list delegate
func getListDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	// Deliberately NOT calling delegate.SetSpacing(0) here: it changes the
	// list's rendered line count between the first (pre-WindowSizeMsg,
	// zero-sized) frame and the real sized frame, which corrupts
	// bubbletea's incremental line-diffing renderer and drops the app
	// title bar above the list entirely. Confirmed by bisecting against
	// this session's changes -- removing this one line was the fix.
	delegate.Styles.NormalTitle = lipgloss.NewStyle()
	delegate.Styles.NormalDesc = lipgloss.NewStyle().Faint(true)
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(getAccentColor()).Bold(true)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(getAccentColor())
	delegate.Styles.DimmedTitle = lipgloss.NewStyle().Faint(true)
	delegate.Styles.DimmedDesc = lipgloss.NewStyle().Faint(true)
	delegate.Styles.FilterMatch = lipgloss.NewStyle().Foreground(getSuccessColor())
	return delegate
}

// getListStyles builds the list chrome (title badge, filter prompt/cursor,
// status bar, pagination dots, help/no-items text) from the same ANSI
// colors as the rest of butler, instead of list.DefaultStyles()'s baked-in
// lipgloss.AdaptiveColor values (see the comment above getAccentColor()).
func getListStyles() list.Styles {
	s := list.DefaultStyles()
	s.Title = lipgloss.NewStyle().
		Background(getAccentColor()).
		Padding(0, 1)
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

// newList creates a list.Model with butler's theme-aware chrome applied.
// Use this instead of list.New() everywhere so every list -- main menu and
// every popup -- stays in sync with the active theme.
func newList(items []list.Item, delegate list.ItemDelegate, width, height int) list.Model {
	l := list.New(items, delegate, width, height)
	l.Styles = getListStyles()

	// list.Model embeds its own help.Model for the keybinding line
	// ("↑/k up • ↓/j down • ..."), which has a *separate* Styles struct
	// that list.Styles above does not touch. help.New() defaults it to
	// lipgloss.AdaptiveColor grays too, so left alone it renders the same
	// unthemed, hardcoded color regardless of anything above.
	l.Help.Styles.ShortKey = lipgloss.NewStyle().Faint(true)
	l.Help.Styles.ShortDesc = lipgloss.NewStyle().Faint(true)
	l.Help.Styles.ShortSeparator = lipgloss.NewStyle().Faint(true)
	l.Help.Styles.FullKey = lipgloss.NewStyle().Faint(true)
	l.Help.Styles.FullDesc = lipgloss.NewStyle().Faint(true)
	l.Help.Styles.FullSeparator = lipgloss.NewStyle().Faint(true)
	l.Help.Styles.Ellipsis = lipgloss.NewStyle().Faint(true)

	return l
}

// Command completion message
type commandDoneMsg struct {
	output     string
	err        error
	showOutput bool
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

	if err := applyTheme(themeName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to apply theme '%s': %v\n", themeName, err)
		os.Exit(1)
	}
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
	configs := syncConfigDirs()
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
	fmt.Println("  theme <name>   Apply theme (catppuccin-mocha, catppuccin-latte,")
	fmt.Println("                  tokyonight-night, tokyonight-storm, tokyonight-moon, tokyonight-day)")
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
			fmt.Fprintln(os.Stderr, "Available themes: catppuccin-mocha, catppuccin-latte,")
			fmt.Fprintln(os.Stderr, "                  tokyonight-night, tokyonight-storm, tokyonight-moon, tokyonight-day")
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
