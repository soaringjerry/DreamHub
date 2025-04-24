package service

import (
	"context"
	"encoding/json"
	"io" // 移到这里
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/hibiken/asynq"
	"github.com/soaringjerry/dreamhub/internal/entity"
	"github.com/soaringjerry/dreamhub/pkg/apperr" // Import apperr
	"github.com/soaringjerry/dreamhub/pkg/logger"
)

const uploadDir = "uploads" // Consider making this configurable

// FileService defines the interface for file operations.
type FileService interface {
	UploadFile(ctx context.Context, userID string, fileHeader *multipart.FileHeader) (*asynq.TaskInfo, error)
}

// fileServiceImpl implements the FileService interface.
type fileServiceImpl struct {
	asynqClient *asynq.Client
	// Add config if uploadDir becomes configurable
}

// NewFileService creates a new FileService implementation.
func NewFileService(client *asynq.Client) FileService {
	if client == nil {
		panic("Asynq client cannot be nil for FileService")
	}
	return &fileServiceImpl{
		asynqClient: client,
	}
}

// UploadFile saves the uploaded file and enqueues an embedding task.
func (s *fileServiceImpl) UploadFile(ctx context.Context, userID string, fileHeader *multipart.FileHeader) (*asynq.TaskInfo, error) {
	filename := filepath.Base(fileHeader.Filename)

	// 1. Create user-specific directory
	userUploadDir := filepath.Join(uploadDir, userID)
	if err := os.MkdirAll(userUploadDir, os.ModePerm); err != nil {
		logger.Error("FileService: Failed to create user upload directory", "path", userUploadDir, "userID", userID, "error", err)
		// Wrap the error using apperr
		return nil, apperr.WrapInternalError("Failed to create upload directory", err)
	}

	// 2. Save the file
	dst := filepath.Join(userUploadDir, filename)
	// Need to open the file from the header to save it
	srcFile, err := fileHeader.Open()
	if err != nil {
		logger.Error("FileService: Failed to open uploaded file", "filename", filename, "userID", userID, "error", err)
		return nil, apperr.Wrap(apperr.FileUploadError, "Failed to open uploaded file", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		logger.Error("FileService: Failed to create destination file", "path", dst, "userID", userID, "error", err)
		return nil, apperr.WrapInternalError("Failed to create destination file", err)
	}
	defer dstFile.Close()

	// Copy the file content - Consider using io.CopyBuffer for large files
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		logger.Error("FileService: Failed to save file content", "path", dst, "userID", userID, "error", err)
		// Attempt to remove partially written file
		os.Remove(dst)
		return nil, apperr.Wrap(apperr.FileSaveError, "Failed to save file content", err)
	}
	logger.Info("FileService: File saved successfully", "filename", filename, "path", dst, "userID", userID)

	// 3. Enqueue embedding task
	payload := entity.EmbeddingPayload{
		UserID:           userID,
		FilePath:         dst,
		OriginalFilename: filename,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Error("FileService: Failed to marshal embedding payload", "userID", userID, "filename", filename, "error", err)
		// Don't necessarily delete the uploaded file here, maybe allow manual retry?
		return nil, apperr.WrapInternalError("Failed to create processing task payload", err)
	}

	task := asynq.NewTask(entity.TaskTypeEmbedding, payloadBytes)
	taskInfo, err := s.asynqClient.EnqueueContext(ctx, task) // Use context from handler
	if err != nil {
		logger.Error("FileService: Failed to enqueue embedding task", "userID", userID, "filename", filename, "error", err)
		return nil, apperr.Wrap(apperr.InternalError, "Failed to schedule file processing", err) // Use a generic internal error code
	}

	logger.Info("FileService: Embedding task enqueued", "userID", userID, "filename", filename, "taskID", taskInfo.ID, "queue", taskInfo.Queue)

	return taskInfo, nil
}

// Need to import "io" for io.Copy
// import "io" // Duplicate removed, already imported at the top
