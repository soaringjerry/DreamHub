package ctxutil

import (
	"context"
)

// CtxKey 定义了用于 context.Context 的键类型，以避免冲突。
type CtxKey string

const (
	// UserIDKey 是存储用户 ID 的 context key。
	UserIDKey CtxKey = "user_id"
	// TraceIDKey 是存储追踪 ID 的 context key。
	TraceIDKey CtxKey = "trace_id"
	// TenantIDKey 是存储租户 ID 的 context key (如果需要多租户)。
	// TenantIDKey CtxKey = "tenant_id"
)

// WithUserID 将用户 ID 添加到 context 中。
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetUserID 从 context 中获取用户 ID。如果不存在则返回空字符串。
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

// WithTraceID 将追踪 ID 添加到 context 中。
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// GetTraceID 从 context 中获取追踪 ID。如果不存在则返回空字符串。
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

/*
// 如果需要多租户支持，可以取消以下注释

// WithTenantID 将租户 ID 添加到 context 中。
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// GetTenantID 从 context 中获取租户 ID。如果不存在则返回空字符串。
func GetTenantID(ctx context.Context) string {
	if tenantID, ok := ctx.Value(TenantIDKey).(string); ok {
		return tenantID
	}
	return ""
}
*/
