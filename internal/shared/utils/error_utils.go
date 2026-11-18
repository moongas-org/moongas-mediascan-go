package utils

import (
	"log/slog"
	"os"
)

func Check(e error, msg string) {
	if e != nil {
		slog.Error("Operation failed", "message", msg, "error", e)
		os.Exit(1)
	}
}
