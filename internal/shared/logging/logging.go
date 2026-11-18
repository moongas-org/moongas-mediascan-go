// Package logging configures the project's default structured logger.
package logging

import (
	"log/slog"
	"os"
)

// Init installs the default text logger at INFO level.
func Init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))
}
