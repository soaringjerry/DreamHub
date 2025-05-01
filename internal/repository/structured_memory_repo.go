package repository

import (
	"context"

	"github.com/soaringjerry/dreamhub/internal/entity" // Use correct module path

	"github.com/google/uuid"
)

// StructuredMemoryRepository defines the interface for interacting with
// structured memory data storage.
type StructuredMemoryRepository interface {
	// Create adds a new structured memory entry to the storage.
	Create(ctx context.Context, memory *entity.StructuredMemory) error

	// GetByKey retrieves a specific structured memory entry for a user by its key.
	// Returns ErrNotFound if the entry does not exist.
	GetByKey(ctx context.Context, userID uuid.UUID, key string) (*entity.StructuredMemory, error)

	// GetByUserID retrieves all structured memory entries for a specific user.
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.StructuredMemory, error)

	// Update modifies an existing structured memory entry.
	// It should typically update the 'Value' and 'UpdatedAt' fields based on UserID and Key.
	Update(ctx context.Context, memory *entity.StructuredMemory) error

	// Delete removes a structured memory entry from the storage based on UserID and Key.
	Delete(ctx context.Context, userID uuid.UUID, key string) error
}
