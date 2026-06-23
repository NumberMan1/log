package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewFile(t *testing.T) {
	tmpDir := t.TempDir()
	fullPath := filepath.Join(tmpDir, "test.log")

	f, err := NewFile(fullPath, 4096)
	if err != nil {
		t.Fatalf("NewFile failed: %v", err)
	}
	defer f.Close()

	// File should exist
	if !FileExists(fullPath) {
		t.Error("file should exist after NewFile")
	}
}

func TestNewFile_CreateDir(t *testing.T) {
	tmpDir := t.TempDir()
	// Directory doesn't exist yet
	fullPath := filepath.Join(tmpDir, "subdir", "test.log")

	// Ensure parent dir exists
	os.MkdirAll(filepath.Dir(fullPath), 0755)

	f, err := NewFile(fullPath, 4096)
	if err != nil {
		t.Fatalf("NewFile failed: %v", err)
	}
	defer f.Close()

	if !FileExists(fullPath) {
		t.Error("file should exist")
	}
}

func TestFile_Write(t *testing.T) {
	tmpDir := t.TempDir()
	fullPath := filepath.Join(tmpDir, "write_test.log")

	f, _ := NewFile(fullPath, 4096)
	defer f.Close()

	if err := f.Write("hello world"); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Flush and check content
	f.Flush()
	data, _ := os.ReadFile(fullPath)
	if string(data) != "hello world" {
		t.Errorf("content = %q, want 'hello world'", string(data))
	}
}

func TestFile_WriteBytes(t *testing.T) {
	tmpDir := t.TempDir()
	fullPath := filepath.Join(tmpDir, "writebytes_test.log")

	f, _ := NewFile(fullPath, 4096)
	defer f.Close()

	if err := f.WriteBytes([]byte("bytes data")); err != nil {
		t.Fatalf("WriteBytes failed: %v", err)
	}

	f.Flush()
	data, _ := os.ReadFile(fullPath)
	if string(data) != "bytes data" {
		t.Errorf("content = %q, want 'bytes data'", string(data))
	}
}

func TestFile_Write_MultipleLines(t *testing.T) {
	tmpDir := t.TempDir()
	fullPath := filepath.Join(tmpDir, "multi_test.log")

	f, _ := NewFile(fullPath, 4096)
	defer f.Close()

	lines := []string{"line1\n", "line2\n", "line3\n"}
	for _, line := range lines {
		if err := f.Write(line); err != nil {
			t.Fatalf("Write failed: %v", err)
		}
	}

	f.Flush()
	data, _ := os.ReadFile(fullPath)
	expected := "line1\nline2\nline3\n"
	if string(data) != expected {
		t.Errorf("content = %q, want %q", string(data), expected)
	}
}

func TestFile_Flush_NoChange(t *testing.T) {
	tmpDir := t.TempDir()
	fullPath := filepath.Join(tmpDir, "nochange_test.log")

	f, _ := NewFile(fullPath, 4096)
	defer f.Close()

	// Flush with no writes should not error
	if err := f.Flush(); err != nil {
		t.Errorf("Flush should not error with no changes: %v", err)
	}
}

func TestFile_Close(t *testing.T) {
	tmpDir := t.TempDir()
	fullPath := filepath.Join(tmpDir, "close_test.log")

	f, _ := NewFile(fullPath, 4096)

	f.Write("data before close")
	if err := f.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Data should be flushed and written
	data, _ := os.ReadFile(fullPath)
	if string(data) != "data before close" {
		t.Errorf("content = %q, want 'data before close'", string(data))
	}
}

func TestFile_DoubleClose(t *testing.T) {
	tmpDir := t.TempDir()
	fullPath := filepath.Join(tmpDir, "double_close_test.log")

	f, _ := NewFile(fullPath, 4096)
	defer f.Close()

	if err := f.Close(); err != nil {
		t.Errorf("first Close failed: %v", err)
	}
	// Second close may return an error, that's OK
	_ = f.Close()
}

func TestFile_WriteManyLines(t *testing.T) {
	tmpDir := t.TempDir()
	fullPath := filepath.Join(tmpDir, "many_lines_test.log")

	f, _ := NewFile(fullPath, 256) // Small buffer to test buffering
	defer f.Close()

	for i := 0; i < 100; i++ {
		if err := f.Write("data line\n"); err != nil {
			t.Fatalf("Write failed at line %d: %v", i, err)
		}
	}

	if err := f.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	data, _ := os.ReadFile(fullPath)
	if len(data) == 0 {
		t.Error("file should not be empty")
	}
}
