package main

import (
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	tea "github.com/charmbracelet/bubbletea"
)
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
var defaultSyncConfigs = []string{"kitty", "waybar", "mako", "walker", "uwsm", "satty", "atuin", "nvim", "zsh"}

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
