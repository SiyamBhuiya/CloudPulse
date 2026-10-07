package logger

import (
	"log/slog"
	"os"
)

// Setup makes slog print readable lines to the terminal.
func Setup() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
}