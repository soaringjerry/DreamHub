package repository

import (
	"context"
	"errors" // Import errors package for defining errors

	"github.com/soaringjerry/dreamhub/internal/entity" // Use correct module path
	// "github.com/google/uuid" // Removed unused import
)

// Common repository errors
var (
	ErrNotFound     = errors.New("entity not found")
	ErrDuplicateKey = errors.New("duplicate key violates unique constraint")
)

// StructuredMemoryRepository defines the interface for interacting with
// structured memory data storage.
type StructuredMemoryRepository interface {
	// Create adds a new structured memory entry to the storage.
	Create(ctx context.Context, memory *entity.StructuredMemory) error

	// GetByKey retrieves a specific structured memory entry for a user by its key.
	// Returns ErrNotFound if the entry does not exist.
	// Changed userID type from uuid.UUID to string (UUID)
	GetByKey(ctx context.Context, userID string, key string) (*entity.StructuredMemory, error)

	// GetByUserID retrieves all structured memory entries for a specific user.
	// Changed userID type from uuid.UUID to string (UUID)
	GetByUserID(ctx context.Context, userID string) ([]*entity.StructuredMemory, error)

	// Update modifies an existing structured memory entry.
	// It accepts UserConfig with UserID as string (UUID).
	// It should typically update the 'Value' and 'UpdatedAt' fields based on UserID and Key.
	Update(ctx context.Context, memory *entity.StructuredMemory) error

	// Delete removes a structured memory entry from the storage based on UserID and Key.
	// Changed userID type from uuid.UUID to string (UUID)
	Delete(ctx context.Context, userID string, key string) error
}
