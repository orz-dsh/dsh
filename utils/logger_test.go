package utils

import "testing"

func TestLogger1(t *testing.T) {
	logger := NewLogger(LogLevelAll)
	logger.Debug("debug %s", "format")
	logger.Info("info %s", "format")
	logger.Warn("warn %s", "format")
	logger.Error("error %s", "format")
	logger.Fatal("critical error%+v", ErrN("crash error", KV("key", "value")))
	logger.Panic("critical error%+v", ErrN("crash error", KV("key", "value")))
}

func TestLogger2(t *testing.T) {
	logger := NewLogger(LogLevelAll)
	logger.DebugDesc("debug desc", KV("key", "value"))
	logger.InfoDesc("info desc", KV("key", "value"))
	logger.WarnDesc("warn desc", KV("key", "value"))
	logger.ErrorDesc("error desc", KV("key", "value"))
	logger.FatalDesc("critical error", KV("key", "value"))
	logger.PanicDesc("critical error", KV("key", "value"))
}
