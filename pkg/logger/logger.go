package logger

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/soaringjerry/dreamhub/pkg/ctxutil" // 引入我们刚创建的 ctxutil 包
)

var (
	defaultLogger *slog.Logger
	once          sync.Once
)

// InitDefaultLogger 初始化默认的全局 logger。
// 使用 JSON handler，日志级别为 Info。
func InitDefaultLogger() {
	once.Do(func() {
		logLevel := slog.LevelInfo // 默认级别
		// 可以考虑从环境变量读取日志级别
		// if levelStr := os.Getenv("LOG_LEVEL"); levelStr != "" { ... }

		opts := &slog.HandlerOptions{
			Level: logLevel,
			// 可以添加 AddSource: true 来包含源文件和行号，但会影响性能
		}
		handler := slog.NewJSONHandler(os.Stdout, opts)
		defaultLogger = slog.New(handler)
		slog.SetDefault(defaultLogger) // 设置为标准库 slog 的默认 logger
	})
}

// GetLogger 返回默认的全局 logger 实例。
// 如果尚未初始化，会先进行初始化。
func GetLogger() *slog.Logger {
	InitDefaultLogger() // 确保已初始化
	return defaultLogger
}

// Ctx retrieves the logger from the context or returns the default logger.
// This allows associating logs with specific requests via trace IDs.
// 注意：目前我们还没有将 logger 注入 context 的中间件，
// 所以这个函数暂时总是返回 defaultLogger。
// 后续在实现 API 中间件时，可以将携带 trace_id 的 logger 注入 context。
func Ctx(ctx context.Context) *slog.Logger {
	// TODO: Implement middleware to inject logger with trace_id into context.
	// Example (pseudo-code in middleware):
	// traceID := ctxutil.GetTraceID(ctx)
	// reqLogger := GetLogger().With(slog.String("trace_id", traceID))
	// ctx = context.WithValue(ctx, loggerKey, reqLogger)

	// For now, return the default logger.
	l := GetLogger()

	// 如果 context 中有 trace_id，则添加到日志字段中
	if traceID := ctxutil.GetTraceID(ctx); traceID != "" {
		l = l.With(slog.String(string(ctxutil.TraceIDKey), traceID))
	}
	// 如果 context 中有 user_id，也可以考虑添加
	// if userID := ctxutil.GetUserID(ctx); userID != "" {
	//  l = l.With(slog.String(string(ctxutil.UserIDKey), userID))
	// }

	return l
}

// Info logs a message at the Info level using the default logger.
func Info(msg string, args ...any) {
	GetLogger().Info(msg, args...)
}

// Warn logs a message at the Warn level using the default logger.
func Warn(msg string, args ...any) {
	GetLogger().Warn(msg, args...)
}

// Error logs a message at the Error level using the default logger.
func Error(msg string, args ...any) {
	GetLogger().Error(msg, args...)
}

// Debug logs a message at the Debug level using the default logger.
// Note: Default level is Info, so Debug logs might not appear unless level is changed.
func Debug(msg string, args ...any) {
	GetLogger().Debug(msg, args...)
}

// InfoContext logs a message at the Info level using the logger derived from the context.
func InfoContext(ctx context.Context, msg string, args ...any) {
	Ctx(ctx).Info(msg, args...)
}

// WarnContext logs a message at the Warn level using the logger derived from the context.
func WarnContext(ctx context.Context, msg string, args ...any) {
	Ctx(ctx).Warn(msg, args...)
}

// ErrorContext logs a message at the Error level using the logger derived from the context.
func ErrorContext(ctx context.Context, msg string, args ...any) {
	Ctx(ctx).Error(msg, args...)
}

// DebugContext logs a message at the Debug level using the logger derived from the context.
func DebugContext(ctx context.Context, msg string, args ...any) {
	Ctx(ctx).Debug(msg, args...)
}
