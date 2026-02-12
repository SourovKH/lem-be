package utils

import (
	"fmt"
	"log/slog"
	"os"
)

// GlobalLogger is the raw slog instance
var GlobalLogger *slog.Logger

// Logger wraps slog.Logger and provides formatted logging methods
type Logger struct {
	inner *slog.Logger
}

// InitLogger initializes the global structured logger.
func InitLogger(service, method string) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	GlobalLogger = slog.New(handler).With(slog.String("service", service), slog.String("method", method))
	slog.SetDefault(GlobalLogger)
}

// NewLogger creates a new scoped Logger instance
func NewLogger(service, method string) *Logger {
	if GlobalLogger == nil {
		InitLogger("App", "Init")
	}
	return &Logger{
		inner: GlobalLogger.With(slog.String("service", service), slog.String("method", method)),
	}
}

// --- Standard Logging Methods ---

func (l *Logger) Info(msg string, args ...any) {
	l.inner.Info(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.inner.Error(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.inner.Warn(msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.inner.Debug(msg, args...)
}

// --- Formatted Logging Methods ---

func (l *Logger) Infof(format string, args ...any) {
	l.inner.Info(fmt.Sprintf(format, args...))
}

func (l *Logger) Errorf(format string, args ...any) {
	l.inner.Error(fmt.Sprintf(format, args...))
}

func (l *Logger) Warnf(format string, args ...any) {
	l.inner.Warn(fmt.Sprintf(format, args...))
}

func (l *Logger) Debugf(format string, args ...any) {
	l.inner.Debug(fmt.Sprintf(format, args...))
}

// --- Backward Compatibility Helpers (Optional, can be removed once refactor is complete) ---

func LogInfo(msg string, args ...any) {
	if GlobalLogger == nil { InitLogger("App", "Init") }
	slog.Info(msg, args...)
}

func LogError(msg string, err error, args ...any) {
	if GlobalLogger == nil { InitLogger("App", "Init") }
	slog.Error(msg, append([]any{slog.Any("error", err)}, args...)...)
}
