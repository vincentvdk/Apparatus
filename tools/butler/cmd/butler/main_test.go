package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestThemeIsDarkFallback verifies the name-based fallback used when
// colors.json is not readable (e.g. in CI without the themes dir installed).
func TestThemeIsDarkFallback(t *testing.T) {
	tests := []struct {
		name     string
		theme    string
		expected bool
	}{
		{"catppuccin-mocha is dark", "catppuccin-mocha", true},
		{"tokyonight-night is dark", "tokyonight-night", true},
		{"tokyonight-storm is dark", "tokyonight-storm", true},
		{"tokyonight-moon is dark", "tokyonight-moon", true},
		{"catppuccin-latte is light", "catppuccin-latte", false},
		{"tokyonight-day is light", "tokyonight-day", false},
		{"unknown theme defaults to light", "some-random-theme", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := themeIsDarkFallback(tt.theme)
			if got != tt.expected {
				t.Errorf("themeIsDark(%q) = %v, want %v", tt.theme, got, tt.expected)
			}
		})
	}
}

// TestCurrentThemeNameReadsFile verifies reading the marker file from a temp HOME.
func TestCurrentThemeNameReadsFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// No file -> empty
	if got := currentThemeName(); got != "" {
		t.Errorf("currentThemeName() = %q, want empty", got)
	}

	// Write marker -> reads value
	dir := filepath.Join(tmp, ".config", "apparatus")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "current-theme"), []byte("catppuccin-mocha\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := currentThemeName(); got != "catppuccin-mocha" {
		t.Errorf("currentThemeName() = %q, want catppuccin-mocha", got)
	}
}

// TestGetThemesDir verifies the path selection logic.
func TestGetThemesDir(t *testing.T) {
	got := getThemesDir()
	if _, err := os.Stat("/run/.containerenv"); err == nil {
		if got != "/run/host/usr/share/apparatus/themes" {
			t.Errorf("in container: getThemesDir() = %q, want /run/host/...", got)
		}
	} else {
		if got != "/usr/share/apparatus/themes" {
			t.Errorf("on host: getThemesDir() = %q, want /usr/share/...", got)
		}
	}
}
