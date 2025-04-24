package worker

import (
	"context"
	"fmt"
	"os"

	"github.com/hibiken/asynq"
	"github.com/soaringjerry/dreamhub/internal/entity"
	"github.com/soaringjerry/dreamhub/internal/repository"
	"github.com/soaringjerry/dreamhub/pkg/logger"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

// EmbeddingTaskHandler handles tasks of type entity.TaskTypeEmbedding.
type EmbeddingTaskHandler struct {
	docRepo  repository.DocumentRepository
	embedder embeddings.Embedder
	// TODO: Add config for chunk size/overlap if needed
}

// NewEmbeddingTaskHandler creates a new handler for embedding tasks.
func NewEmbeddingTaskHandler(repo repository.DocumentRepository, emb embeddings.Embedder) *EmbeddingTaskHandler {
	if repo == nil || emb == nil {
		panic("DocumentRepository and Embedder cannot be nil for EmbeddingTaskHandler")
	}
	return &EmbeddingTaskHandler{
		docRepo:  repo,
		embedder: emb,
	}
}

// ProcessTask implements the asynq.Handler interface.
func (h *EmbeddingTaskHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	if t.Type() != entity.TaskTypeEmbedding {
		// Should not happen if routing is correct, but good practice to check.
		return fmt.Errorf("unexpected task type: %s", t.Type())
	}

	payload, err := entity.UnmarshalEmbeddingPayload(t.Payload())
	if err != nil {
		// Error already includes context, log it here.
		logger.Error("Failed to unmarshal embedding payload", "error", err, "taskID", t.ResultWriter().TaskID())
		// Returning error will cause Asynq to retry based on config.
		// If unmarshaling fails, retrying likely won't help, consider asynq.SkipRetry.
		return fmt.Errorf("cannot process task: %w", err)
	}

	logger.Info("Processing embedding task", "taskID", t.ResultWriter().TaskID(), "userID", payload.UserID, "filename", payload.OriginalFilename, "filePath", payload.FilePath)

	// 1. Read file content
	contentBytes, err := os.ReadFile(payload.FilePath)
	if err != nil {
		logger.Error("Worker: Failed to read file", "taskID", t.ResultWriter().TaskID(), "filePath", payload.FilePath, "error", err)
		// Retry might help if it's a temporary filesystem issue.
		return fmt.Errorf("failed to read file %s: %w", payload.FilePath, err)
	}
	content := string(contentBytes)
	logger.Debug("Worker: File read successfully", "taskID", t.ResultWriter().TaskID(), "filePath", payload.FilePath, "size", len(contentBytes))

	// 2. Split text into chunks
	// TODO: Make chunk size/overlap configurable
	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(1000),
		textsplitter.WithChunkOverlap(200),
	)
	chunks, err := splitter.SplitText(content)
	if err != nil {
		logger.Error("Worker: Failed to split text", "taskID", t.ResultWriter().TaskID(), "filename", payload.OriginalFilename, "error", err)
		// Splitting is deterministic, retrying won't help. Consider SkipRetry.
		return fmt.Errorf("failed to split text for %s: %w", payload.OriginalFilename, err)
	}
	logger.Info("Worker: Text split into chunks", "taskID", t.ResultWriter().TaskID(), "filename", payload.OriginalFilename, "chunks", len(chunks))

	// 3. Create documents with metadata
	docs := make([]schema.Document, len(chunks))
	for i, chunk := range chunks {
		docs[i] = schema.Document{
			PageContent: chunk,
			Metadata: map[string]any{
				"source":   payload.OriginalFilename,
				"chunk_id": i,
				"user_id":  payload.UserID, // UserID comes from the payload
			},
		}
	}

	// 4. Add documents to vector store via repository
	// The repository implementation handles adding user_id from context/payload correctly.
	// We need to pass a context that potentially holds the user_id for the repository.
	// For now, we use the task context `ctx`, assuming user_id might be added later.
	// Alternatively, the repository could be modified to accept user_id directly if needed.
	// TODO: Refine context propagation for user_id to repository.
	err = h.docRepo.AddDocuments(ctx, docs)
	if err != nil {
		logger.Error("Worker: Failed to add documents to vector store", "taskID", t.ResultWriter().TaskID(), "userID", payload.UserID, "filename", payload.OriginalFilename, "error", err)
		// Retry might help if it's a DB connection issue.
		return fmt.Errorf("failed to add documents for %s: %w", payload.OriginalFilename, err)
	}

	logger.Info("Embedding task processed successfully", "taskID", t.ResultWriter().TaskID(), "userID", payload.UserID, "filename", payload.OriginalFilename, "chunks", len(docs))
	return nil // Return nil to indicate success
}
