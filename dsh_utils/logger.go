package dsh_utils

import (
	"io"
	"log"
	"os"
)

type LogLevel int

const (
	LogLevelAll   LogLevel = 0
	LogLevelDebug LogLevel = 1
	LogLevelInfo  LogLevel = 2
	LogLevelWarn  LogLevel = 3
	LogLevelError LogLevel = 4
	LogLevelNone  LogLevel = 5
)

type Logger struct {
	Level        LogLevel
	normalLogger *log.Logger
	errorLogger  *log.Logger
}

func NewLogger(level LogLevel) *Logger {
	return &Logger{
		Level:        level,
		normalLogger: log.New(os.Stdout, "", 0),
		errorLogger:  log.New(os.Stderr, "", 0),
	}
}

func (l *Logger) IsDebugEnabled() bool {
	return l.Level <= LogLevelDebug
}

func (l *Logger) GetDebugWriter() io.Writer {
	return l.normalLogger.Writer()
}

func (l *Logger) Debug(format string, v ...any) {
	if l.IsDebugEnabled() {
		l.normalLogger.Printf("[DEBUG] "+format, v...)
	}
}

func (l *Logger) DebugDesc(title string, kvs ...DescKeyValue) {
	if l.IsDebugEnabled() {
		l.normalLogger.Printf("[DEBUG] %+v", NewDesc(title, kvs).ToString("", "\t\t"))
	}
}

func (l *Logger) IsInfoEnabled() bool {
	return l.Level <= LogLevelInfo
}

func (l *Logger) GetInfoWriter() io.Writer {
	return l.normalLogger.Writer()
}

func (l *Logger) Info(format string, v ...any) {
	if l.IsInfoEnabled() {
		l.normalLogger.Printf("[INFO ] "+format, v...)
	}
}

func (l *Logger) InfoDesc(title string, kvs ...DescKeyValue) {
	if l.IsInfoEnabled() {
		l.normalLogger.Printf("[INFO ] %+v", NewDesc(title, kvs).ToString("", "\t\t"))
	}
}

func (l *Logger) IsWarnEnabled() bool {
	return l.Level <= LogLevelWarn
}

func (l *Logger) GetWarnWriter() io.Writer {
	return l.normalLogger.Writer()
}

func (l *Logger) Warn(format string, v ...any) {
	if l.IsWarnEnabled() {
		l.normalLogger.Printf("[WARN ] "+format, v...)
	}
}

func (l *Logger) WarnDesc(title string, kvs ...DescKeyValue) {
	if l.IsWarnEnabled() {
		l.normalLogger.Printf("[WARN ] %+v", NewDesc(title, kvs).ToString("", "\t\t"))
	}
}

func (l *Logger) IsErrorEnabled() bool {
	return l.Level <= LogLevelError
}

func (l *Logger) GetErrorWriter() io.Writer {
	return l.errorLogger.Writer()
}

func (l *Logger) Error(format string, v ...any) {
	if l.IsErrorEnabled() {
		l.errorLogger.Printf("[ERROR] "+format, v...)
	}
}

func (l *Logger) ErrorDesc(title string, kvs ...DescKeyValue) {
	if l.IsErrorEnabled() {
		l.errorLogger.Printf("[ERROR] %+v", NewDesc(title, kvs).ToString("", "\t\t"))
	}
}

func (l *Logger) GetFatalWriter() io.Writer {
	return l.errorLogger.Writer()
}

func (l *Logger) Fatal(format string, v ...any) {
	l.errorLogger.Fatalf("[FATAL] "+format, v...)
}

func (l *Logger) FatalDesc(title string, kvs ...DescKeyValue) {
	l.errorLogger.Fatalf("[FATAL] %+v", NewDesc(title, kvs).ToString("", "\t\t"))
}

func (l *Logger) GetPanicWriter() io.Writer {
	return l.errorLogger.Writer()
}

func (l *Logger) Panic(format string, v ...any) {
	l.errorLogger.Panicf("[PANIC] "+format, v...)
}

func (l *Logger) PanicDesc(title string, kvs ...DescKeyValue) {
	l.errorLogger.Panicf("[PANIC] %+v", NewDesc(title, kvs).ToString("", "\t\t"))
}
