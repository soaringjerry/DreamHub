package apperr

import "fmt"

// ErrorCode defines the type for application-specific error codes.
type ErrorCode string

// Predefined error codes
const (
	// General Errors (00xx)
	UnknownError      ErrorCode = "ERR-0000" // 未知错误
	InternalError     ErrorCode = "ERR-0001" // 内部服务错误
	ValidationError   ErrorCode = "ERR-0002" // 输入验证错误
	NotFoundError     ErrorCode = "ERR-0003" // 资源未找到
	UnauthorizedError ErrorCode = "ERR-0004" // 未授权
	ForbiddenError    ErrorCode = "ERR-0005" // 禁止访问

	// Database Errors (01xx)
	DBConnectionError ErrorCode = "ERR-0100" // 数据库连接错误
	DBQueryError      ErrorCode = "ERR-0101" // 数据库查询错误
	DBInsertError     ErrorCode = "ERR-0102" // 数据库插入错误
	DBUpdateError     ErrorCode = "ERR-0103" // 数据库更新错误
	DBDeleteError     ErrorCode = "ERR-0104" // 数据库删除错误

	// Vector Store Errors (02xx)
	VectorStoreAddError    ErrorCode = "ERR-0200" // 向量存储添加错误
	VectorStoreSearchError ErrorCode = "ERR-0201" // 向量存储搜索错误
	VectorStoreInitError   ErrorCode = "ERR-0202" // 向量存储初始化错误

	// File Handling Errors (03xx)
	FileUploadError ErrorCode = "ERR-0300" // 文件上传错误
	FileReadError   ErrorCode = "ERR-0301" // 文件读取错误
	FileSaveError   ErrorCode = "ERR-0302" // 文件保存错误
	FileSplitError  ErrorCode = "ERR-0303" // 文件分割错误

	// External Service Errors (04xx)
	LLMAPIError    ErrorCode = "ERR-0400" // LLM API 调用错误
	EmbeddingError ErrorCode = "ERR-0401" // Embedding 生成错误

	// Add more specific codes as needed...
)

// AppError represents a custom application error.
type AppError struct {
	Code    ErrorCode // Application-specific error code
	Message string    // User-friendly error message
	Err     error     // Original underlying error (optional)
}

// Error returns the string representation of the AppError.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for error chaining (errors.Is, errors.As).
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError without an underlying error.
func New(code ErrorCode, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Wrap creates a new AppError with an underlying error.
func Wrap(code ErrorCode, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// --- Helper functions for common errors ---

func NewValidationError(message string) *AppError {
	return New(ValidationError, message)
}

func WrapInternalError(message string, err error) *AppError {
	if message == "" {
		message = "Internal server error"
	}
	return Wrap(InternalError, message, err)
}

func NewNotFoundError(message string) *AppError {
	return New(NotFoundError, message)
}

// Add more helpers as needed...
