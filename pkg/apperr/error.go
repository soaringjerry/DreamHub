package apperr

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode 定义了应用程序错误的唯一代码。
type ErrorCode string

// 预定义的错误代码
const (
	CodeUnknown          ErrorCode = "UNKNOWN"           // 未知错误
	CodeInvalidArgument  ErrorCode = "INVALID_ARGUMENT"  // 无效参数
	CodeNotFound         ErrorCode = "NOT_FOUND"         // 资源未找到
	CodeAlreadyExists    ErrorCode = "ALREADY_EXISTS"    // 资源已存在
	CodePermissionDenied ErrorCode = "PERMISSION_DENIED" // 权限不足
	CodeUnauthenticated  ErrorCode = "UNAUTHENTICATED"   // 未认证
	CodeInternal         ErrorCode = "INTERNAL_ERROR"    // 内部服务器错误
	CodeUnimplemented    ErrorCode = "UNIMPLEMENTED"     // 未实现
	CodeUnavailable      ErrorCode = "UNAVAILABLE"       // 服务不可用
	CodeConflict         ErrorCode = "CONFLICT"          // 状态冲突
	CodeRateLimited      ErrorCode = "RATE_LIMITED"      // 请求频率限制
	CodeValidation       ErrorCode = "VALIDATION_ERROR"  // 数据验证失败
)

// AppError 是应用程序自定义错误类型。
type AppError struct {
	Code       ErrorCode // 错误代码
	Message    string    // 用户友好的错误消息
	Details    []string  // 可选的错误详情
	Err        error     // 原始错误 (可选)
	HTTPStatus int       // 对应的 HTTP 状态码
}

// Error 实现 error 接口。
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 返回原始错误，支持 errors.Is 和 errors.As。
func (e *AppError) Unwrap() error {
	return e.Err
}

// New 创建一个新的 AppError。
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: errorCodeToHTTPStatus(code), // 根据 Code 设置默认 HTTP 状态码
	}
}

// Wrap 包装一个现有错误为 AppError。
func Wrap(err error, code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Err:        err,
		HTTPStatus: errorCodeToHTTPStatus(code),
	}
}

// WithDetails 添加错误详情。
func (e *AppError) WithDetails(details ...string) *AppError {
	e.Details = append(e.Details, details...)
	return e
}

// WithHTTPStatus 设置自定义的 HTTP 状态码。
func (e *AppError) WithHTTPStatus(status int) *AppError {
	e.HTTPStatus = status
	return e
}

// errorCodeToHTTPStatus 将内部错误代码映射到推荐的 HTTP 状态码。
func errorCodeToHTTPStatus(code ErrorCode) int {
	switch code {
	case CodeInvalidArgument, CodeValidation:
		return http.StatusBadRequest // 400
	case CodeNotFound:
		return http.StatusNotFound // 404
	case CodeAlreadyExists, CodeConflict:
		return http.StatusConflict // 409
	case CodePermissionDenied:
		return http.StatusForbidden // 403
	case CodeUnauthenticated:
		return http.StatusUnauthorized // 401
	case CodeInternal:
		return http.StatusInternalServerError // 500
	case CodeUnimplemented:
		return http.StatusNotImplemented // 501
	case CodeUnavailable:
		return http.StatusServiceUnavailable // 503
	case CodeRateLimited:
		return http.StatusTooManyRequests // 429
	default:
		return http.StatusInternalServerError // 500
	}
}

// GetCode 从错误中提取 ErrorCode。如果错误不是 AppError 或为 nil，则返回 CodeUnknown。
func GetCode(err error) ErrorCode {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	if err != nil {
		// 如果不是 AppError 但也不是 nil，认为是内部错误
		return CodeInternal
	}
	return CodeUnknown // 或者可以返回一个表示“无错误”的值
}

// GetHTTPStatus 从错误中提取 HTTP 状态码。如果错误不是 AppError 或为 nil，则返回 500。
func GetHTTPStatus(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPStatus
	}
	if err != nil {
		return http.StatusInternalServerError
	}
	// 对于 nil 错误，通常不应该调用此函数，但可以返回 200 OK
	return http.StatusOK
}

// Is checks if the target error is an AppError with the specified code.
func Is(err error, code ErrorCode) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

// Helper functions for common error types
func ErrNotFound(message string) *AppError {
	return New(CodeNotFound, message)
}

func ErrInvalidArgument(message string) *AppError {
	return New(CodeInvalidArgument, message)
}

func ErrInternal(message string, originalErr error) *AppError {
	if originalErr != nil {
		return Wrap(originalErr, CodeInternal, message)
	}
	return New(CodeInternal, message)
}

func ErrPermissionDenied(message string) *AppError {
	return New(CodePermissionDenied, message)
}

func ErrUnauthenticated(message string) *AppError {
	return New(CodeUnauthenticated, message)
}

func ErrAlreadyExists(message string) *AppError {
	return New(CodeAlreadyExists, message)
}

func ErrValidation(message string) *AppError {
	return New(CodeValidation, message)
}

// ... 可以根据需要添加更多辅助函数 ...
