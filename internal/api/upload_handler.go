package api

import (
	"errors" // Import errors package
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/soaringjerry/dreamhub/internal/service"
	"github.com/soaringjerry/dreamhub/pkg/apperr"  // Import apperr
	"github.com/soaringjerry/dreamhub/pkg/ctxutil" // Import ctxutil
	"github.com/soaringjerry/dreamhub/pkg/logger"
)

// UploadHandler handles file upload requests.
type UploadHandler struct {
	fileService service.FileService
}

// NewUploadHandler creates a new UploadHandler.
func NewUploadHandler(fs service.FileService) *UploadHandler {
	if fs == nil {
		panic("FileService cannot be nil for UploadHandler")
	}
	return &UploadHandler{
		fileService: fs,
	}
}

// HandleUpload is the Gin handler function for POST /upload.
func (h *UploadHandler) HandleUpload(c *gin.Context) {
	// 1. Get user ID from form data
	// TODO: Get UserID from authenticated context instead of form data later.
	userID := c.PostForm("user_id")
	if userID == "" {
		// Use apperr for validation error
		errResp := apperr.NewValidationError("Missing user_id field")
		c.JSON(http.StatusBadRequest, gin.H{"error": errResp.Message, "code": errResp.Code})
		return
	}

	// 2. Get file from form data
	fileHeader, err := c.FormFile("file")
	if err != nil {
		logger.Warn("UploadHandler: Failed to get file from form", "userID", userID, "error", err)
		errResp := apperr.Wrap(apperr.ValidationError, "Failed to get file from request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": errResp.Message, "code": errResp.Code})
		return
	}

	// 3. Create context with UserID
	ctx := ctxutil.WithUserID(c.Request.Context(), userID)

	// 4. Call FileService to handle upload and enqueue task, passing the new context
	taskInfo, err := h.fileService.UploadFile(ctx, userID, fileHeader)
	if err != nil {
		logger.Error("UploadHandler: FileService failed", "userID", userID, "filename", fileHeader.Filename, "error", err)

		// Handle specific AppError types
		var appErr *apperr.AppError
		if errors.As(err, &appErr) {
			statusCode := http.StatusInternalServerError // Default to 500
			switch appErr.Code {
			case apperr.ValidationError, apperr.FileUploadError, apperr.FileSaveError:
				statusCode = http.StatusBadRequest
			case apperr.NotFoundError:
				statusCode = http.StatusNotFound
				// Add other specific cases as needed
			}
			c.JSON(statusCode, gin.H{"error": appErr.Message, "code": appErr.Code})
		} else {
			// Handle unexpected errors
			c.JSON(http.StatusInternalServerError, gin.H{"error": "An unexpected error occurred during file upload", "code": apperr.UnknownError})
		}
		return
	}

	// 4. Return success response (202 Accepted)
	logger.Info("UploadHandler: File upload accepted for processing", "userID", userID, "filename", fileHeader.Filename, "taskID", taskInfo.ID)
	c.JSON(http.StatusAccepted, gin.H{
		"message":  "File upload accepted, processing in background.",
		"filename": fileHeader.Filename,
		"task_id":  taskInfo.ID,
	})
}
