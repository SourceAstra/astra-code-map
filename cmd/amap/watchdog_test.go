package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestHangWatchdogFires reproduces a stalled index phase in a child process and
// asserts the watchdog dumps goroutines and exits instead of hanging forever.
func TestHangWatchdogFires(t *testing.T) {
	dumpDir := t.TempDir()

	if os.Getenv("AMAP_WATCHDOG_CHILD") == "1" {
		stop := startHangWatchdog(os.Getenv("AMAP_WATCHDOG_DIR"), 100*time.Millisecond)
		defer stop()
		select {} // stall: the watchdog must terminate this process
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestHangWatchdogFires", "-test.v")
	cmd.Env = append(os.Environ(), "AMAP_WATCHDOG_CHILD=1", "AMAP_WATCHDOG_DIR="+dumpDir)
	out, err := cmd.CombinedOutput()
	if ee, ok := err.(*exec.ExitError); !ok || ee.ExitCode() != 1 {
		t.Fatalf("watchdog did not exit(1): %v\noutput: %s", err, out)
	}

	dump, err := os.ReadFile(filepath.Join(dumpDir, ".astramap", "hang-dump.txt"))
	if err != nil || len(dump) == 0 {
		t.Fatalf("expected non-empty hang-dump.txt: %v\nchild output: %s", err, out)
	}
}
