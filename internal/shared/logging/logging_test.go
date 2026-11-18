package logging

import (
	"context"
	"log/slog"
	"testing"
)

func TestInitSetsInfoAsDefaultLevel(t *testing.T) {
	previous := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previous) })

	Init()

	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		t.Fatal("debug logging should be disabled by default")
	}
	if !slog.Default().Enabled(context.Background(), slog.LevelInfo) {
		t.Fatal("info logging should be enabled by default")
	}
}
