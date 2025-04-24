package logger

import (
	"log/slog"
	"os"
)

// Logger is a wrapper around slog.Logger for application-wide logging.
var Logger *slog.Logger

func init() {
	// Initialize the default logger
	// TODO: Make log level and format configurable (e.g., via pkg/config)
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug, // Default to Debug for now
		// AddSource: true, // Uncomment to include source file and line number
	}

	// Use JSON handler for structured logging
	handler := slog.NewJSONHandler(os.Stdout, opts)
	Logger = slog.New(handler)

	slog.SetDefault(Logger) // Optionally set as the default slog logger
	Logger.Info("Logger initialized")
}

// Info logs an informational message.
func Info(msg string, args ...any) {
	Logger.Info(msg, args...)
}

// Warn logs a warning message.
func Warn(msg string, args ...any) {
	Logger.Warn(msg, args...)
}

// Error logs an error message.
func Error(msg string, args ...any) {
	Logger.Error(msg, args...)
}

// Debug logs a debug message.
func Debug(msg string, args ...any) {
	Logger.Debug(msg, args...)
}
