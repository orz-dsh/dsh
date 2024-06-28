package utils

import "testing"

func TestLogger1(t *testing.T) {
	logger := NewLogger(LogLevelAll)
	logger.Debug("debug %s", "format")
	logger.Info("info %s", "format")
	logger.Warn("warn %s", "format")
	logger.Error("error %s", "format")
	logger.Fatal("critical error%+v", errN("crash error", kv("key", "value")))
	logger.Panic("critical error%+v", errN("crash error", kv("key", "value")))
}

func TestLogger2(t *testing.T) {
	logger := NewLogger(LogLevelAll)
	logger.DebugDesc("debug desc", kv("key", "value"))
	logger.InfoDesc("info desc", kv("key", "value"))
	logger.WarnDesc("warn desc", kv("key", "value"))
	logger.ErrorDesc("error desc", kv("key", "value"))
	logger.FatalDesc("critical error", kv("key", "value"))
	logger.PanicDesc("critical error", kv("key", "value"))
}
