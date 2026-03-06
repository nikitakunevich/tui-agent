package logging

import (
	"io"
	"log/slog"
	"os"
)

// Setup initializes slog with a file handler and optional stderr output.
// Returns a cleanup function to close the log file.
func Setup(logPath string, debug bool) (func(), error) {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	var w io.Writer = f
	if debug {
		w = io.MultiWriter(f, os.Stderr)
	}

	handler := slog.NewTextHandler(w, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))

	cleanup := func() { f.Close() }
	return cleanup, nil
}
