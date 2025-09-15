package slogger

import (
	"context"
	"log/slog"
	"os"
	"runtime/debug"
)

var (
	LevelTrace slog.Level = -8
	LevelFatal slog.Level = 12

	levelNames = map[slog.Level]string{
		LevelTrace: "TRACE",
		LevelFatal: "FATAL",
	}
)

// Trace calls Logger.LogAttrs on the trace logger
func Trace(ctx context.Context, msg string, attrs ...slog.Attr) {
	slog.LogAttrs(ctx, LevelTrace, msg, attrs...)
}

// Debug calls Logger.LogAttrs on the debug logger
func Debug(ctx context.Context, msg string, attrs ...slog.Attr) {
	slog.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
}

// Info calls Logger.LogAttrs on the info logger
func Info(ctx context.Context, msg string, attrs ...slog.Attr) {
	slog.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
}

// Warn calls Logger.LogAttrs on the warn logger
func Warn(ctx context.Context, msg string, attrs ...slog.Attr) {
	slog.LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
}

// Error calls Logger.LogAttrs on the error logger
func Error(ctx context.Context, msg string, attrs ...slog.Attr) {
	slog.LogAttrs(ctx, slog.LevelError, msg, attrs...)
}

// Fatal calls Logger.LogAttrs on the fatal logger with stack trace, then it
// will exit the application with 1 status
func Fatal(ctx context.Context, msg string, attrs ...slog.Attr) {
	attrs = append(attrs, slog.String("stack_trace", string(debug.Stack())))
	slog.LogAttrs(ctx, LevelFatal, msg, attrs...)
	os.Exit(1)
}
