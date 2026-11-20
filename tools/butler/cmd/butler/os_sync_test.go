package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDefaultSyncConfigs verifies hypr is excluded (host-only) and rio is gone.
func TestDefaultSyncConfigs(t *testing.T) {
	for _, dir := range defaultSyncConfigs {
		if dir == "hypr" {
			t.Error("hypr should not be in defaultSyncConfigs (host-only)")
		}
		if dir == "rio" {
			t.Error("rio should not be in defaultSyncConfigs (removed)")
		}
	}
	// Verify expected dirs are present
	expected := []string{"kitty", "waybar", "mako", "zsh", "nvim"}
	for _, e := range expected {
		found := false
		for _, d := range defaultSyncConfigs {
			if d == e {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %q in defaultSyncConfigs", e)
		}
	}
}

// TestLoadSkipListEmpty verifies empty map when file doesn't exist.
func TestLoadSkipListEmpty(t *testing.T) {
	skip, err := loadSkipList()
	if err != nil {
		t.Fatalf("loadSkipList() error = %v", err)
	}
	if len(skip) != 0 {
		t.Errorf("expected empty skip list, got %d entries", len(skip))
	}
}

// TestSyncConfigDirsDefault verifies fallback to defaults when no override file.
func TestSyncConfigDirsDefault(t *testing.T) {
	dirs := syncConfigDirs()
	if len(dirs) == 0 {
		t.Error("syncConfigDirs() returned empty list")
	}
	// Should match defaults (no override file in test env)
	if len(dirs) != len(defaultSyncConfigs) {
		t.Errorf("syncConfigDirs() len = %d, want %d (defaults)", len(dirs), len(defaultSyncConfigs))
	}
}

// TestPushExcludes verifies rio is not in push excludes.
func TestPushExcludes(t *testing.T) {
	if _, ok := pushExcludes["rio"]; ok {
		t.Error("rio should not be in pushExcludes (removed)")
	}
	if _, ok := pushExcludes["hypr"]; !ok {
		t.Error("hypr should be in pushExcludes")
	}
}

// TestSaveAndLoadSkipList round-trips through a temp HOME.
func TestSaveAndLoadSkipList(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	want := map[string]bool{"box1": true, "box2": true}
	if err := saveSkipList(want); err != nil {
		t.Fatalf("saveSkipList() error = %v", err)
	}

	// Verify file contents
	skipFile := filepath.Join(tmp, ".config", "apparatus", "skip-distroboxes")
	data, err := os.ReadFile(skipFile)
	if err != nil {
		t.Fatalf("reading skip file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "box1") || !strings.Contains(content, "box2") {
		t.Errorf("skip file content = %q, want box1 and box2", content)
	}

	got, err := loadSkipList()
	if err != nil {
		t.Fatalf("loadSkipList() error = %v", err)
	}
	if len(got) != 2 || !got["box1"] || !got["box2"] {
		t.Errorf("loadSkipList() = %v, want {box1:true, box2:true}", got)
	}
}
