package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestStageSymlink verifies staged symlink creation.
func TestStageSymlink(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "source.conf")
	target := filepath.Join(tmp, "target.conf")

	if err := os.WriteFile(src, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	sw, err := stageSymlink(src, target)
	if err != nil {
		t.Fatalf("stageSymlink() error = %v", err)
	}
	if sw.staged != target+".new" {
		t.Errorf("staged = %q, want %q", sw.staged, target+".new")
	}
	if sw.target != target {
		t.Errorf("target = %q, want %q", sw.target, target)
	}
	// Verify symlink exists at staged path
	link, err := os.Readlink(sw.staged)
	if err != nil {
		t.Fatalf("reading staged symlink: %v", err)
	}
	if link != src {
		t.Errorf("symlink target = %q, want %q", link, src)
	}
	// Cleanup
	os.Remove(sw.staged)
}

// TestStageFile verifies staged file copy.
func TestStageFile(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "source.lua")
	target := filepath.Join(tmp, "target.lua")

	content := []byte("-- lua content")
	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatal(err)
	}

	sw, err := stageFile(src, target)
	if err != nil {
		t.Fatalf("stageFile() error = %v", err)
	}
	data, err := os.ReadFile(sw.staged)
	if err != nil {
		t.Fatalf("reading staged file: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("staged content = %q, want %q", data, content)
	}
	os.Remove(sw.staged)
}

// TestStageSymlinkMissingSource verifies error when source doesn't exist.
func TestStageSymlinkMissingSource(t *testing.T) {
	tmp := t.TempDir()
	_, err := stageSymlink(filepath.Join(tmp, "nonexistent"), filepath.Join(tmp, "target"))
	if err == nil {
		t.Error("expected error for missing source, got nil")
	}
}
