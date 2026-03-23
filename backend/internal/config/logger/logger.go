package logger

import "log/slog"

// Logger wraps slog for structured logging
type Logger struct {
	*slog.Logger
}

// New creates a new logger with the given name
func New(name string) *Logger {
	return &Logger{
		Logger: slog.With("component", name),
	}
}
