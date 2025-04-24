package entity

import (
	"encoding/json"
	"fmt"
)

// TaskTypeEmbedding defines the type name for the embedding task.
const TaskTypeEmbedding = "embedding:process_file"

// EmbeddingPayload defines the data needed for the file embedding task.
type EmbeddingPayload struct {
	UserID           string `json:"user_id"`
	FilePath         string `json:"file_path"`
	OriginalFilename string `json:"original_filename"`
}

// NewEmbeddingTask creates a new Asynq task for file embedding.
// func NewEmbeddingTask(userID, filePath, originalFilename string) (*asynq.Task, error) {
// 	payload, err := json.Marshal(EmbeddingPayload{
// 		UserID:           userID,
// 		FilePath:         filePath,
// 		OriginalFilename: originalFilename,
// 	})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to marshal embedding payload: %w", err)
// 	}
// 	// Task type is TaskTypeEmbedding
// 	return asynq.NewTask(TaskTypeEmbedding, payload), nil
// }

// Note: The NewEmbeddingTask helper function depends on asynq package.
// It might be better placed in a package that handles task enqueuing (e.g., internal/service or a dedicated task client package)
// to avoid circular dependencies or entity package depending on infrastructure.
// For now, it's commented out here for reference.

// Helper to unmarshal payload in the worker
func UnmarshalEmbeddingPayload(data []byte) (*EmbeddingPayload, error) {
	var p EmbeddingPayload
	err := json.Unmarshal(data, &p)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal embedding payload: %w", err)
	}
	return &p, nil
}
