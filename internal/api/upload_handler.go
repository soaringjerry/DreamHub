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
		errResp := apperr.ErrValidation("Missing user_id field") // Use ErrValidation helper
		c.JSON(errResp.HTTPStatus, gin.H{"error": errResp.Message, "code": errResp.Code})
		return
	}

	// 2. Get file from form data
	fileHeader, err := c.FormFile("file")
	if err != nil {
		logger.Warn("UploadHandler: Failed to get file from form", "userID", userID, "error", err)
		// Use CodeInvalidArgument and correct Wrap signature: Wrap(err, code, message)
		errResp := apperr.Wrap(err, apperr.CodeInvalidArgument, "Failed to get file from request")
		c.JSON(errResp.HTTPStatus, gin.H{"error": errResp.Message, "code": errResp.Code})
		return
	}

	// 3. Create context with UserID
	ctx := ctxutil.WithUserID(c.Request.Context(), userID)

	// 4. Open the uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		logger.Error("UploadHandler: Failed to open uploaded file", "userID", userID, "filename", fileHeader.Filename, "error", err)
		errResp := apperr.Wrap(err, apperr.CodeInternal, "Failed to open uploaded file")
		c.JSON(errResp.HTTPStatus, gin.H{"error": errResp.Message, "code": errResp.Code})
		return
	}
	defer file.Close() // Ensure file is closed

	// 5. Call FileService to handle upload and enqueue task, passing the new context and correct parameters
	// UploadFile expects: ctx, userID, filename, fileSize, contentType, fileData io.Reader
	// It returns: *entity.Document, taskID string, error
	// Add the missing userID argument
	_, taskID, err := h.fileService.UploadFile(ctx, userID, fileHeader.Filename, fileHeader.Size, fileHeader.Header.Get("Content-Type"), file)
	if err != nil {
		logger.Error("UploadHandler: FileService failed", "userID", userID, "filename", fileHeader.Filename, "error", err)

		// Handle specific AppError types
		var appErr *apperr.AppError
		if errors.As(err, &appErr) {
			// Use HTTPStatus from AppError directly
			// Use correct ErrorCode constants
			// Assuming FileUploadError and FileSaveError map to CodeInternal for now
			// switch appErr.Code {
			// case apperr.CodeValidation: // Already handled above? Maybe for content validation?
			// case apperr.CodeInternal: // For FileUploadError, FileSaveError
			// case apperr.CodeNotFound:
			// }
			c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message, "code": appErr.Code})
		} else {
			// Handle unexpected errors
			unexpectedErr := apperr.Wrap(err, apperr.CodeUnknown, "An unexpected error occurred during file upload")
			c.JSON(unexpectedErr.HTTPStatus, gin.H{"error": unexpectedErr.Message, "code": unexpectedErr.Code})
		}
		return
	}

	// 6. Return success response (202 Accepted) using the returned taskID
	logger.Info("UploadHandler: File upload accepted for processing", "userID", userID, "filename", fileHeader.Filename, "taskID", taskID)
	c.JSON(http.StatusAccepted, gin.H{
		"message":  "File upload accepted, processing in background.",
		"filename": fileHeader.Filename,
		"task_id":  taskID, // Use the returned taskID string
	})
}
