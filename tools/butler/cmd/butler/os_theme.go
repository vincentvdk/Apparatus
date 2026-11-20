package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)
type stagedSwap struct {
	staged string // the ".new" path holding the new content
	target string // the real path it will replace
}

// stageSymlink prepares target's replacement as a symlink to source at a
// ".new" sibling path, without touching target itself. Returns the staged
// swap to commit later, or an error if the source is missing or the
// symlink couldn't be created.
func stageSymlink(source, target string) (stagedSwap, error) {
	if _, err := os.Stat(source); err != nil {
		return stagedSwap{}, fmt.Errorf("%s: %w", filepath.Base(target), err)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return stagedSwap{}, fmt.Errorf("%s: %w", filepath.Base(target), err)
	}
	staged := target + ".new"
	os.Remove(staged) // clean up any leftover from a prior failed apply
	if err := os.Symlink(source, staged); err != nil {
		return stagedSwap{}, fmt.Errorf("%s: %w", filepath.Base(target), err)
	}
	return stagedSwap{staged: staged, target: target}, nil
}

// stageFile is like stageSymlink but copies source's content instead of
// symlinking to it (used for nvim.lua, which is written directly rather
// than symlinked so it survives being synced into a distrobox).
func stageFile(source, target string) (stagedSwap, error) {
	data, err := os.ReadFile(source)
	if err != nil {
		return stagedSwap{}, fmt.Errorf("%s: %w", filepath.Base(target), err)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return stagedSwap{}, fmt.Errorf("%s: %w", filepath.Base(target), err)
	}
	staged := target + ".new"
	if err := os.WriteFile(staged, data, 0644); err != nil {
		return stagedSwap{}, fmt.Errorf("%s: %w", filepath.Base(target), err)
	}
	return stagedSwap{staged: staged, target: target}, nil
}

// applyTheme applies a theme across all config files. Every app's new
// config is staged as a sibling ".new" file first; only once every app has
// staged successfully do we commit by renaming each one over its live
// target (an atomic operation on the same filesystem). If staging fails
// partway through, everything staged so far is rolled back and the old
// theme's config is left completely untouched -- so a failure never leaves
// some apps on the new theme and others broken or stuck on the old one.
func applyTheme(themeName string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	themesDir := getThemesDir()
	if _, err := os.Stat(themesDir); err != nil {
		return fmt.Errorf("themes directory not available: %s (inside a distrobox this should be the host's /run/host mount)", themesDir)
	}
	themeDir := filepath.Join(themesDir, themeName)

	var pending []stagedSwap
	rollback := func() {
		for _, s := range pending {
			os.Remove(s.staged)
		}
	}

	stage := func(source, target string) error {
		s, err := stageSymlink(source, target)
		if err != nil {
			rollback()
			return err
		}
		pending = append(pending, s)
		return nil
	}

	if err := stage(filepath.Join(themeDir, "kitty.conf"), filepath.Join(homeDir, ".config", "kitty", "theme.conf")); err != nil {
		return err
	}
	if err := stage(filepath.Join(themeDir, "waybar.css"), filepath.Join(homeDir, ".config", "waybar", "theme.css")); err != nil {
		return err
	}
	if err := stage(filepath.Join(themeDir, "mako.conf"), filepath.Join(homeDir, ".config", "mako", "config")); err != nil {
		return err
	}
	if err := stage(filepath.Join(themeDir, "hyprland.conf"), filepath.Join(homeDir, ".config", "hypr", "theme.conf")); err != nil {
		return err
	}
	if err := stage(filepath.Join(themeDir, "satty", "overrides.css"), filepath.Join(homeDir, ".config", "satty", "overrides.css")); err != nil {
		return err
	}

	// nvim.lua is optional -- not every theme ships one
	nvimSource := filepath.Join(themeDir, "nvim.lua")
	if _, err := os.Stat(nvimSource); err == nil {
		nvimTarget := filepath.Join(homeDir, ".config", "nvim", "lua", "config", "theme.lua")
		s, err := stageFile(nvimSource, nvimTarget)
		if err != nil {
			rollback()
			return err
		}
		pending = append(pending, s)
	}

	// Commit: atomically rename every staged file over its live target.
	for _, s := range pending {
		if err := os.Rename(s.staged, s.target); err != nil {
			return fmt.Errorf("committing %s: %w", filepath.Base(s.target), err)
		}
	}

	// Apply GTK theme (best-effort, outside the atomic config swap)
	isDark := themeIsDark(themeName)
	if isDark {
		exec.Command("gsettings", "set", "org.gnome.desktop.interface", "color-scheme", "prefer-dark").Run()
		exec.Command("gsettings", "set", "org.gnome.desktop.interface", "gtk-theme", "Adwaita-dark").Run()
	} else {
		exec.Command("gsettings", "set", "org.gnome.desktop.interface", "color-scheme", "prefer-light").Run()
		exec.Command("gsettings", "set", "org.gnome.desktop.interface", "gtk-theme", "Adwaita").Run()
	}

	// Save current theme marker -- only after every app's config has been
	// committed, so it can never claim a theme that isn't actually applied.
	appConfigDir := filepath.Join(homeDir, ".config", "apparatus")
	os.MkdirAll(appConfigDir, 0755)
	currentThemeFile := filepath.Join(appConfigDir, "current-theme")
	tmpThemeFile := currentThemeFile + ".new"
	if err := os.WriteFile(tmpThemeFile, []byte(themeName), 0644); err == nil {
		os.Rename(tmpThemeFile, currentThemeFile)
	}

	// Reload services
	exec.Command("hyprctl", "reload").Run()
	exec.Command("pkill", "-SIGUSR1", "kitty").Run()
	exec.Command("makoctl", "reload").Run()

	// Reload waybar using SIGUSR2 signal
	exec.Command("killall", "-SIGUSR2", "waybar").Run()

	// Reload satty (must be restarted as it doesn't support signal-based reload)
	exec.Command("pkill", "satty").Run()

	return nil
}

// applyFont applies a font across all config files
func applyFont(fontFamily, fontName string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		return
	}

	// Update kitty font
	kittyConf := filepath.Join(homeDir, ".config", "kitty", "kitty.conf")
	if data, err := os.ReadFile(kittyConf); err == nil {
		content := string(data)
		content = strings.ReplaceAll(content, "font_family      "+"IoskeleyMono Nerd Font", "font_family      "+fontFamily)
		content = strings.ReplaceAll(content, "font_family      "+"JetBrainsMono Nerd Font", "font_family      "+fontFamily)
		content = strings.ReplaceAll(content, "font_family      "+"Hack Nerd Font", "font_family      "+fontFamily)
		os.WriteFile(kittyConf, []byte(content), 0644)
	}

	// Update waybar font
	waybarCSS := filepath.Join(homeDir, ".config", "waybar", "style.css")
	if data, err := os.ReadFile(waybarCSS); err == nil {
		content := string(data)
		content = strings.ReplaceAll(content, "\"Hack Nerd Font\"", "\""+fontFamily+"\"")
		content = strings.ReplaceAll(content, "\"JetBrainsMono Nerd Font\"", "\""+fontFamily+"\"")
		content = strings.ReplaceAll(content, "\"IoskeleyMono Nerd Font\"", "\""+fontFamily+"\"")
		os.WriteFile(waybarCSS, []byte(content), 0644)
	}

	// Update mako font
	makoConf := filepath.Join(homeDir, ".config", "mako", "config")
	if data, err := os.ReadFile(makoConf); err == nil {
		content := string(data)
		content = strings.ReplaceAll(content, "Hack Nerd Font", fontFamily)
		content = strings.ReplaceAll(content, "JetBrainsMono Nerd Font", fontFamily)
		content = strings.ReplaceAll(content, "IoskeleyMono Nerd Font", fontFamily)
		os.WriteFile(makoConf, []byte(content), 0644)
	}

	// Save current font
	appConfigDir := filepath.Join(homeDir, ".config", "apparatus")
	os.MkdirAll(appConfigDir, 0755)
	os.WriteFile(filepath.Join(appConfigDir, "current-font"), []byte(fontName), 0644)

	// Reload services
	exec.Command("pkill", "-SIGUSR1", "kitty").Run()
	exec.Command("hyprctl", "reload").Run()
	exec.Command("makoctl", "reload").Run()

	// Reload waybar using SIGUSR2 signal
	exec.Command("killall", "-SIGUSR2", "waybar").Run()
}
