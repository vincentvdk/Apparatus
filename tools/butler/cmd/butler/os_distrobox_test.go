package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAssembleIniPath verifies the path is under ~/.config/distrobox/.
func TestAssembleIniPath(t *testing.T) {
	got := assembleIniPath()
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".config", "distrobox", "distrobox.ini")
	if got != want {
		t.Errorf("assembleIniPath() = %q, want %q", got, want)
	}
}

// TestAssembleCreateBoxWritesIni verifies the ini section format without
// actually running distrobox-assemble (we mock by checking file contents
// before the runCommand call fails).
func TestAssembleCreateBoxWritesIni(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// assembleCreateBox will fail at runCommand (no distrobox-assemble),
	// but the ini file should already be written.
	_, err := assembleCreateBox("testbox", "")
	// Error expected (distrobox-assemble not available), but file should exist
	ini := filepath.Join(tmp, ".config", "distrobox", "distrobox.ini")
	data, readErr := os.ReadFile(ini)
	if readErr != nil {
		t.Fatalf("ini file not written: %v (assemble error: %v)", readErr, err)
	}
	content := string(data)
	if !strings.Contains(content, "[testbox]") {
		t.Errorf("ini missing [testbox] section:\n%s", content)
	}
	if !strings.Contains(content, "image = ghcr.io/vincentvdk/apparatus-box:latest") {
		t.Errorf("ini missing image line:\n%s", content)
	}
	if !strings.Contains(content, "init_hooks = distrobox-host-exec butler sync testbox") {
		t.Errorf("ini missing init_hooks line:\n%s", content)
	}
}

// TestAssembleCreateBoxWithHome verifies home dir is included.
func TestAssembleCreateBoxWithHome(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	_, _ = assembleCreateBox("homebox", "/custom/home")
	ini := filepath.Join(tmp, ".config", "distrobox", "distrobox.ini")
	data, _ := os.ReadFile(ini)
	content := string(data)
	if !strings.Contains(content, "home = /custom/home") {
		t.Errorf("ini missing home line:\n%s", content)
	}
}

// TestAssembleCreateBoxIdempotent verifies duplicate sections aren't added.
func TestAssembleCreateBoxIdempotent(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	_, _ = assembleCreateBox("dupbox", "")
	_, _ = assembleCreateBox("dupbox", "")

	ini := filepath.Join(tmp, ".config", "distrobox", "distrobox.ini")
	data, _ := os.ReadFile(ini)
	count := strings.Count(string(data), "[dupbox]")
	if count != 1 {
		t.Errorf("expected 1 [dupbox] section, got %d", count)
	}
}
