package worker

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid" // Add uuid import
	"github.com/hibiken/asynq"
	"github.com/pgvector/pgvector-go" // Add pgvector import
	"github.com/soaringjerry/dreamhub/internal/entity"
	"github.com/soaringjerry/dreamhub/internal/repository"
	"github.com/soaringjerry/dreamhub/pkg/ctxutil" // Import ctxutil
	"github.com/soaringjerry/dreamhub/pkg/logger"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

// EmbeddingTaskHandler handles tasks of type entity.TaskTypeEmbedding.
type EmbeddingTaskHandler struct {
	docRepo    repository.DocumentRepository
	vectorRepo repository.VectorRepository
	embedder   embeddings.Embedder
	// TODO: Add config for chunk size/overlap if needed
}

// NewEmbeddingTaskHandler creates a new handler for embedding tasks.
func NewEmbeddingTaskHandler(docRepo repository.DocumentRepository, vectorRepo repository.VectorRepository, emb embeddings.Embedder) *EmbeddingTaskHandler {
	if docRepo == nil || vectorRepo == nil || emb == nil {
		panic("DocumentRepository, VectorRepository and Embedder cannot be nil for EmbeddingTaskHandler")
	}
	return &EmbeddingTaskHandler{
		docRepo:    docRepo,
		vectorRepo: vectorRepo,
		embedder:   emb,
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
	chunksText, err := splitter.SplitText(content) // Renamed variable to avoid conflict
	if err != nil {
		logger.Error("Worker: Failed to split text", "taskID", t.ResultWriter().TaskID(), "filename", payload.OriginalFilename, "error", err)
		// Splitting is deterministic, retrying won't help. Consider SkipRetry.
		return fmt.Errorf("failed to split text for %s: %w", payload.OriginalFilename, err)
	}
	logger.Info("Worker: Text split into chunks", "taskID", t.ResultWriter().TaskID(), "filename", payload.OriginalFilename, "chunks", len(chunksText))

	// 3. Create documents with metadata (using schema.Document for embedding)
	docs := make([]schema.Document, len(chunksText))
	for i, chunk := range chunksText {
		docs[i] = schema.Document{
			PageContent: chunk,
			Metadata: map[string]any{
				"source":   payload.OriginalFilename,
				"chunk_id": i,
				"user_id":  payload.UserID, // UserID comes from the payload
			},
		}
	}

	// 4. Generate embeddings for chunks
	// Extract text content from documents
	texts := make([]string, len(docs))
	for i, doc := range docs {
		texts[i] = doc.PageContent
	}

	// Generate embeddings
	embeddingsResult, err := h.embedder.EmbedDocuments(ctx, texts) // Renamed variable
	if err != nil {
		logger.Error("Worker: Failed to generate embeddings", "taskID", t.ResultWriter().TaskID(), "userID", payload.UserID, "filename", payload.OriginalFilename, "error", err)
		return fmt.Errorf("failed to generate embeddings for %s: %w", payload.OriginalFilename, err)
	}

	// 5. Convert to DocumentChunks
	// Create a new context containing the user ID from the payload
	taskCtx := ctxutil.WithUserID(ctx, payload.UserID)

	// TODO: Ideally, fetch or create the main Document record here using h.docRepo
	// and get its ID. For now, generate a new ID for the chunks.
	documentID := uuid.New()
	logger.Warn("Generated new DocumentID for chunks, consider linking to a Document record", "generatedID", documentID, "taskID", t.ResultWriter().TaskID())

	// Convert schema.Document to entity.DocumentChunk
	entityChunks := make([]*entity.DocumentChunk, len(docs)) // Correct variable name and type
	for i, doc := range docs {
		// Convert metadata
		metadata := make(map[string]any)
		if doc.Metadata != nil { // Add nil check for safety
			for k, v := range doc.Metadata {
				metadata[k] = v
			}
		}

		// Create pgvector.Vector from []float32
		vector := pgvector.NewVector(embeddingsResult[i]) // Use renamed variable

		// Create DocumentChunk
		entityChunks[i] = entity.NewDocumentChunk( // Assign to the correct slice
			documentID.String(), // Convert UUID to string
			payload.UserID,
			i, // chunk index
			doc.PageContent,
			vector,
			metadata,
		)
	}

	// 6. Add chunks to vector store via repository
	err = h.vectorRepo.AddChunks(taskCtx, entityChunks) // Pass the correct slice
	if err != nil {
		logger.Error("Worker: Failed to add chunks to vector store", "taskID", t.ResultWriter().TaskID(), "userID", payload.UserID, "filename", payload.OriginalFilename, "error", err)
		// Retry might help if it's a DB connection issue.
		return fmt.Errorf("failed to add chunks for %s: %w", payload.OriginalFilename, err)
	}
	// TODO: Update the main Document record status to Completed using h.docRepo.UpdateDocumentStatus

	logger.Info("Embedding task processed successfully", "taskID", t.ResultWriter().TaskID(), "userID", payload.UserID, "filename", payload.OriginalFilename, "chunks", len(docs))
	return nil // Return nil to indicate success
}
