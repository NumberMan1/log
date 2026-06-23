package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	// Test with existing file
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "exists.txt")
	if err := os.WriteFile(existingFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	if !FileExists(existingFile) {
		t.Error("FileExists should return true for existing file")
	}

	// Test with non-existing file
	if FileExists(filepath.Join(tmpDir, "missing.txt")) {
		t.Error("FileExists should return false for missing file")
	}
}

func TestFileExists_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	if !FileExists(tmpDir) {
		t.Error("FileExists should return true for directories")
	}
}
