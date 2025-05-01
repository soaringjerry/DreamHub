package api

import (
	// Import fmt for error formatting
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid" // Import uuid
	"github.com/soaringjerry/dreamhub/internal/service"
	"github.com/soaringjerry/dreamhub/pkg/apperr"
	"github.com/soaringjerry/dreamhub/pkg/logger"
)

// FileHandler 负责处理与文件和任务相关的 API 请求。
type FileHandler struct {
	fileService service.FileService
}

// NewFileHandler 创建一个新的 FileHandler 实例。
func NewFileHandler(fs service.FileService) *FileHandler {
	return &FileHandler{
		fileService: fs,
	}
}

// RegisterRoutes 将文件和任务相关的路由注册到 Gin 引擎。
func (h *FileHandler) RegisterRoutes(router *gin.RouterGroup) {
	// 文件上传路由
	router.POST("/upload", h.handleUploadFile)

	// 任务状态查询路由
	tasksGroup := router.Group("/tasks")
	{
		tasksGroup.GET("/:task_id/status", h.handleGetTaskStatus)
	}

	// 文档列表和详情路由 (可以放在 /documents 下)
	docsGroup := router.Group("/documents")
	{
		// TODO: Add authentication middleware here
		docsGroup.GET("", h.handleListDocuments)             // GET /api/v1/documents?user_id=...
		docsGroup.GET("/:doc_id", h.handleGetDocument)       // GET /api/v1/documents/{doc_id}
		docsGroup.DELETE("/:doc_id", h.handleDeleteDocument) // DELETE /api/v1/documents/{doc_id}
	}
}

// handleUploadFile 处理文件上传请求。
// 使用 multipart/form-data。
func (h *FileHandler) handleUploadFile(c *gin.Context) {
	// Get UserID from the context set by the auth middleware
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		logger.ErrorContext(c.Request.Context(), "无法从上下文中获取用户 ID (UploadFile)")
		appErr := apperr.New(apperr.CodeInternal, "无法处理请求，缺少用户信息")
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}
	// Log the userID
	logger.DebugContext(c.Request.Context(), "处理文件上传", "user_id", userID)
	ctx := c.Request.Context() // Use original context

	// 从表单获取文件
	fileHeader, err := c.FormFile("file")
	if err != nil {
		logger.WarnContext(ctx, "无法获取上传的文件", "error", err)
		// Use apperr.Wrap function instead of chaining from helper
		appErr := apperr.Wrap(err, apperr.CodeInvalidArgument, "缺少文件或表单字段名错误 ('file')")
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}

	// 打开文件
	fileData, err := fileHeader.Open()
	if err != nil {
		logger.ErrorContext(ctx, "打开上传的文件失败", "error", err, "filename", fileHeader.Filename)
		appErr := apperr.Wrap(err, apperr.CodeInternal, "无法处理上传的文件")
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}
	defer fileData.Close() // 确保文件被关闭

	// 调用 FileService 处理上传
	// Correctly call UploadFile with required arguments
	doc, taskID, err := h.fileService.UploadFile(ctx, fileHeader.Filename, fileHeader.Size, fileHeader.Header.Get("Content-Type"), fileData)
	if err != nil {
		// UploadFile 内部应该已经记录日志并包装错误
		appErr, ok := err.(*apperr.AppError)
		if !ok {
			// Use a generic message for unknown errors during upload
			// Correct Wrap usage: Wrap(originalError, code, message)
			appErr = apperr.Wrap(err, apperr.CodeInternal, "处理文件上传时发生未知错误")
		}
		// Example of checking specific error codes if needed:
		// switch {
		// case apperr.Is(err, apperr.CodeValidation): // Check for validation error from service
		// 	// Use the existing appErr as it came from the service
		// case apperr.Is(err, apperr.CodeNotFound): // Check for not found error from service
		// 	// Use the existing appErr
		// default:
		// 	// If not AppError or specific known code, wrap as internal
		// 	if !ok {
		// 		appErr = apperr.Wrap(err, apperr.CodeInternal, "处理文件上传时发生未知错误")
		// 	}
		// }

		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}

	// 返回成功响应 (HTTP 202 Accepted 表示后台处理中)
	c.JSON(http.StatusAccepted, gin.H{
		"message":  "文件上传成功，正在后台处理中...",
		"filename": doc.OriginalFilename,
		"doc_id":   doc.ID.String(),
		"task_id":  taskID,
	})
}

// handleGetTaskStatus 处理获取任务状态的请求。
func (h *FileHandler) handleGetTaskStatus(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		appErr := apperr.New(apperr.CodeInvalidArgument, "缺少 task_id 路径参数")
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}

	// TODO: 从认证中间件获取 user_id 并传递给 ctx (如果需要按用户隔离任务视图)
	ctx := c.Request.Context()

	task, err := h.fileService.GetTaskStatus(ctx, taskID)
	if err != nil {
		appErr, ok := err.(*apperr.AppError)
		if !ok {
			appErr = apperr.Wrap(err, apperr.CodeInternal, "获取任务状态时发生未知错误")
		}
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}

	// 返回任务状态信息 (entity.Task 结构体已经有 json tag)
	c.JSON(http.StatusOK, task)
}

// handleListDocuments 处理列出用户文档的请求。
func (h *FileHandler) handleListDocuments(c *gin.Context) {
	// Get UserID from the context set by the auth middleware
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		logger.ErrorContext(c.Request.Context(), "无法从上下文中获取用户 ID (ListDocuments)")
		appErr := apperr.New(apperr.CodeInternal, "无法处理请求，缺少用户信息")
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}
	logger.DebugContext(c.Request.Context(), "列出文档", "user_id", userID)
	ctx := c.Request.Context() // Use original context

	// 获取分页参数
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	limit, errL := strconv.Atoi(limitStr)
	offset, errO := strconv.Atoi(offsetStr)
	if errL != nil || limit <= 0 {
		limit = 20
	}
	if errO != nil || offset < 0 {
		offset = 0
	}

	docs, err := h.fileService.ListUserDocuments(ctx, userID, limit, offset)
	if err != nil {
		appErr, ok := err.(*apperr.AppError)
		if !ok {
			appErr = apperr.Wrap(err, apperr.CodeInternal, "获取文档列表时发生未知错误")
		}
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}

	c.JSON(http.StatusOK, docs)
}

// handleGetDocument 处理获取单个文档详情的请求。
func (h *FileHandler) handleGetDocument(c *gin.Context) {
	docIDStr := c.Param("doc_id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		appErr := apperr.New(apperr.CodeInvalidArgument, "无效的文档 ID 格式")
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}

	// Get UserID from the context set by the auth middleware
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		logger.ErrorContext(c.Request.Context(), "无法从上下文中获取用户 ID (GetDocument)")
		appErr := apperr.New(apperr.CodeInternal, "无法处理请求，缺少用户信息")
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}
	logger.DebugContext(c.Request.Context(), "获取文档", "user_id", userID, "doc_id", docIDStr)
	ctx := c.Request.Context()
	// Assume GetDocument service method performs authorization based on context/userID implicitly or needs update
	doc, err := h.fileService.GetDocument(ctx, docID) // Pass userID if needed: h.fileService.GetDocument(ctx, userID, docID)
	if err != nil {
		appErr, ok := err.(*apperr.AppError)
		if !ok {
			appErr = apperr.Wrap(err, apperr.CodeInternal, "获取文档详情时发生未知错误")
		}
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}

	c.JSON(http.StatusOK, doc)
}

// handleDeleteDocument 处理删除文档的请求。
func (h *FileHandler) handleDeleteDocument(c *gin.Context) {
	docIDStr := c.Param("doc_id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		appErr := apperr.New(apperr.CodeInvalidArgument, "无效的文档 ID 格式")
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}

	// Get UserID from the context set by the auth middleware
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		logger.ErrorContext(c.Request.Context(), "无法从上下文中获取用户 ID (DeleteDocument)")
		appErr := apperr.New(apperr.CodeInternal, "无法处理请求，缺少用户信息")
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}
	logger.DebugContext(c.Request.Context(), "删除文档", "user_id", userID, "doc_id", docIDStr)
	ctx := c.Request.Context()
	// Assume DeleteDocument service method performs authorization based on context/userID implicitly or needs update
	err = h.fileService.DeleteDocument(ctx, docID) // Pass userID if needed: h.fileService.DeleteDocument(ctx, userID, docID)
	if err != nil {
		appErr, ok := err.(*apperr.AppError)
		if !ok {
			appErr = apperr.Wrap(err, apperr.CodeInternal, "删除文档时发生未知错误")
		}
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文档已成功删除"})
}
