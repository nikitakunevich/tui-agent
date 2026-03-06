package logging

import (
	"log/slog"
	"os"
	"testing"
)

func TestSetup(t *testing.T) {
	tmp := t.TempDir()
	logPath := tmp + "/test.log"

	cleanup, err := Setup(logPath, false)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	defer cleanup()

	slog.Info("test message", "key", "value")

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("reading log file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("log file is empty")
	}
	if got := string(data); !contains(got, "test message") {
		t.Errorf("log file does not contain 'test message': %s", got)
	}
}

func TestSetupDebug(t *testing.T) {
	tmp := t.TempDir()
	logPath := tmp + "/test-debug.log"

	cleanup, err := Setup(logPath, true)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	defer cleanup()

	slog.Debug("debug message")

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("reading log file: %v", err)
	}
	if got := string(data); !contains(got, "debug message") {
		t.Errorf("log file does not contain 'debug message': %s", got)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
