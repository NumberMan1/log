package serverlog

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func readLogLines(t *testing.T, dir string) []string {
	t.Helper()
	var lines []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		return scanner.Err()
	})
	if err != nil {
		t.Fatal(err)
	}
	return lines
}

func TestFilePrinterConcurrentWritesDrainOnClose(t *testing.T) {
	dir := t.TempDir()
	p, err := NewFilePrinter(FilePrinterConfig{Dir: dir, LogFormat: NewOnlyMessageFormat(), LogChanSize: 3})
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				p.Log(InfoLevel, fmt.Sprintf("%d-%d", i, j), nil)
			}
		}(i)
	}
	wg.Wait()
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); p.Close() }()
	}
	wg.Wait()
	p.Log(InfoLevel, "late", nil)
	lines := readLogLines(t, dir)
	if len(lines) != 200 {
		t.Fatalf("got %d records, want 200", len(lines))
	}
	seen := map[string]bool{}
	for _, line := range lines {
		if seen[line] {
			t.Fatalf("duplicate %s", line)
		}
		seen[line] = true
	}
	if p.Err() != nil {
		t.Fatal(p.Err())
	}
}

func TestFilePrinterReportsWriteFailureWithoutPanic(t *testing.T) {
	failures := make(chan error, 10)
	p, err := NewFilePrinter(FilePrinterConfig{Dir: t.TempDir(), OnError: func(err error) { failures <- err }})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.file.Close(); err != nil {
		t.Fatal(err)
	}
	p.Log(ErrorLevel, "failed write", M{"k": "v"})
	select {
	case err := <-failures:
		if !errors.Is(err, os.ErrClosed) {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("missing write failure")
	}
	p.Close()
	if p.Err() == nil {
		t.Fatal("failure not queryable")
	}
}

func TestFilePrinterDropPolicyReportsOverflow(t *testing.T) {
	// Pause consumption to make queue saturation deterministic.
	p := &FilePrinter{config: FilePrinterConfig{OverflowPolicy: DropNewest, OnError: func(error) {}}, logChan: make(chan string, 1), logFormat: NewOnlyMessageFormat()}
	p.Log(InfoLevel, "accepted", nil)
	p.Log(InfoLevel, "dropped", nil)
	if p.Dropped() != 1 || !errors.Is(p.Err(), ErrLogQueueFull) {
		t.Fatal("overflow not reported")
	}
	if record := <-p.logChan; record != "accepted" {
		t.Fatal("oldest record overwritten")
	}
}

type testSwitcher struct {
	fail      atomic.Bool
	rotations atomic.Int32
}

func (*testSwitcher) FirstSwitchTime() time.Duration { return time.Millisecond }
func (*testSwitcher) NextSwitchTime() time.Duration  { return time.Hour }
func (s *testSwitcher) DirAndFileName(base string) (string, string) {
	index := s.rotations.Add(1)
	if s.fail.Load() {
		return filepath.Join(base, "blocked"), filepath.Join(base, "blocked", "log")
	}
	return base, filepath.Join(base, fmt.Sprintf("file%d", index))
}

func TestFileRotationFailureKeepsOldFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "blocked"), []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}
	switcher := &testSwitcher{}
	p := &FilePrinter{config: FilePrinterConfig{Dir: dir, Switcher: switcher, FileType: ".log", WriteBufferSize: 128}}
	if err := p.switchFile(); err != nil {
		t.Fatal(err)
	}
	first := p.file
	switcher.fail.Store(true)
	if err := p.switchFile(); err == nil {
		t.Fatal("rotation should fail")
	}
	if p.file != first {
		t.Fatal("old file replaced on failed rotation")
	}
	if err := p.file.Write("still writable\n"); err != nil {
		t.Fatal(err)
	}
	if err := p.file.Close(); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(dir, "file1.01.log"))
	if err != nil || !strings.Contains(string(content), "still writable") {
		t.Fatalf("old file lost: %s %v", content, err)
	}
}

func TestFilePrinterRotationAndPeriodicFlush(t *testing.T) {
	dir := t.TempDir()
	switcher := &testSwitcher{}
	p, err := NewFilePrinter(FilePrinterConfig{Dir: dir, Switcher: switcher, FlushTick: time.Millisecond, LogFormat: NewOnlyMessageFormat()})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	p.Log(InfoLevel, "before", nil)
	deadline := time.After(3 * time.Second)
	for switcher.rotations.Load() < 2 {
		select {
		case <-deadline:
			t.Fatal("rotation did not happen")
		case <-time.After(time.Millisecond):
		}
	}
	p.Log(InfoLevel, "after", nil)
	// Observe the periodic flush before Close, rather than relying on shutdown flushing.
	for {
		if strings.Contains(strings.Join(readLogLines(t, dir), "\n"), "after") {
			break
		}
		select {
		case <-deadline:
			t.Fatal("periodic flush did not happen")
		case <-time.After(time.Millisecond):
		}
	}
	p.Close()
	if len(readLogLines(t, dir)) != 2 {
		t.Fatal("rotation lost records")
	}
}

func TestFilePrintersReserveDistinctFiles(t *testing.T) {
	dir := t.TempDir()
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			p, err := NewFilePrinter(FilePrinterConfig{Dir: dir, LogFormat: NewOnlyMessageFormat()})
			if err != nil {
				t.Error(err)
				return
			}
			p.Log(InfoLevel, fmt.Sprint(i), nil)
			p.Close()
		}(i)
	}
	wg.Wait()
	if len(readLogLines(t, dir)) != 12 {
		t.Fatal("concurrent printers overwrote each other")
	}
}
