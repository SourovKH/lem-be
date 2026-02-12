package utils

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

// GlobalLogger is the raw slog instance
var GlobalLogger *slog.Logger

// Logger wraps slog.Logger and provides formatted logging methods
type Logger struct {
	inner *slog.Logger
	ctx   context.Context
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
		ctx:   context.Background(),
	}
}

// WithContext returns a new logger with the provided context
func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{
		inner: l.inner,
		ctx:   ctx,
	}
}

// recordEvent adds the log message as an event to the span in the context
func (l *Logger) recordEvent(level slog.Level, msg string) {
	if l.ctx == nil {
		return
	}
	span := trace.SpanFromContext(l.ctx)
	if span.IsRecording() {
		span.AddEvent(fmt.Sprintf("[%s] %s", level.String(), msg))
	}
}

// --- Standard Logging Methods ---

func (l *Logger) Info(msg string, args ...any) {
	l.inner.Info(msg, args...)
	l.recordEvent(slog.LevelInfo, msg)
}

func (l *Logger) Error(msg string, args ...any) {
	l.inner.Error(msg, args...)
	l.recordEvent(slog.LevelError, msg)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.inner.Warn(msg, args...)
	l.recordEvent(slog.LevelWarn, msg)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.inner.Debug(msg, args...)
	l.recordEvent(slog.LevelDebug, msg)
}

// --- Formatted Logging Methods ---

func (l *Logger) Infof(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.inner.Info(msg)
	l.recordEvent(slog.LevelInfo, msg)
}

func (l *Logger) Errorf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.inner.Error(msg)
	l.recordEvent(slog.LevelError, msg)
}

func (l *Logger) Warnf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.inner.Warn(msg)
	l.recordEvent(slog.LevelWarn, msg)
}

func (l *Logger) Debugf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.inner.Debug(msg)
	l.recordEvent(slog.LevelDebug, msg)
}
