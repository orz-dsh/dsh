package dsh_utils

import "testing"

func TestLogger(t *testing.T) {
	logger := NewLogger(LogLevelAll)
	logger.Debug("debug %s", "format")
	logger.Info("info %s", "format")
	logger.Warn("warn %s", "format")
	logger.Error("error %s", "format")
	logger.Fatal("critical error%+v", NewError("crash error", map[string]any{"key": "value"}))
	logger.Panic("critical error%+v", NewError("crash error", map[string]any{"key": "value"}))
}
