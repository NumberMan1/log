package fileutil

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDaySwitch_DirAndFileName(t *testing.T) {
	ds := daySwitch{}
	dir, file := ds.DirAndFileName("/base/log")

	now := time.Now()
	expectedDir := "/base/log/" + now.Format("2006/01/02/")
	expectedFile := expectedDir + now.Format("2006-01-02")

	if dir != expectedDir {
		t.Errorf("dir = %q, want %q", dir, expectedDir)
	}
	if file != expectedFile {
		t.Errorf("file = %q, want %q", file, expectedFile)
	}
}

func TestDaySwitch_FirstSwitchTime(t *testing.T) {
	ds := daySwitch{}
	d := ds.FirstSwitchTime()
	if d <= 0 {
		t.Errorf("FirstSwitchTime should be positive, got %v", d)
	}
	// Should be less than or equal to 24 hours
	if d > 24*time.Hour {
		t.Errorf("FirstSwitchTime should be <= 24h, got %v", d)
	}
}

func TestDaySwitch_NextSwitchTime(t *testing.T) {
	ds := daySwitch{}
	d := ds.NextSwitchTime()
	if d != 24*time.Hour {
		t.Errorf("NextSwitchTime = %v, want 24h", d)
	}
}

func TestHoursSwitch_DirAndFileName(t *testing.T) {
	hs := hoursSwitch{}
	dir, file := hs.DirAndFileName("/base/log")

	now := time.Now()
	expectedDir := "/base/log/" + now.Format("2006/01/02/")
	expectedFile := expectedDir + now.Format("2006-01-02_15")

	if dir != expectedDir {
		t.Errorf("dir = %q, want %q", dir, expectedDir)
	}
	if file != expectedFile {
		t.Errorf("file = %q, want %q", file, expectedFile)
	}
}

func TestHoursSwitch_FirstSwitchTime(t *testing.T) {
	hs := hoursSwitch{}
	d := hs.FirstSwitchTime()
	if d <= 0 {
		t.Errorf("FirstSwitchTime should be positive, got %v", d)
	}
	// Should be less than or equal to 1 hour
	if d > time.Hour {
		t.Errorf("FirstSwitchTime should be <= 1h, got %v", d)
	}
}

func TestHoursSwitch_NextSwitchTime(t *testing.T) {
	hs := hoursSwitch{}
	d := hs.NextSwitchTime()
	if d != time.Hour {
		t.Errorf("NextSwitchTime = %v, want 1h", d)
	}
}

func TestSwitcher_Interfaces(t *testing.T) {
	// Verify DAY_SWITCHER and HOURS_SWITCHER implement Switcher
	var _ Switcher = DAY_SWITCHER
	var _ Switcher = HOURS_SWITCHER
}

func TestHoursSwitch_FileNameFormat(t *testing.T) {
	hs := hoursSwitch{}
	_, file := hs.DirAndFileName("/tmp/logs")
	base := filepath.Base(file)

	// Filename should match YYYY-MM-DD_HH format
	// Example: 2026-06-23_14
	if len(base) != 13 { // "YYYY-MM-DD_HH" = 13 chars
		t.Errorf("expected 13-char filename, got %q (%d chars)", base, len(base))
	}
	parts := strings.Split(base, "_")
	if len(parts) != 2 {
		t.Errorf("expected filename format YYYY-MM-DD_HH, got %q", base)
	}
}

func TestDaySwitch_FileNameFormat(t *testing.T) {
	ds := daySwitch{}
	_, file := ds.DirAndFileName("/tmp/logs")
	base := filepath.Base(file)

	// Filename should match YYYY-MM-DD format
	// Example: 2026-06-23
	if len(base) != 10 { // "YYYY-MM-DD" = 10 chars
		t.Errorf("expected 10-char filename, got %q (%d chars)", base, len(base))
	}
}
