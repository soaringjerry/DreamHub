package service

import (
	"context"

	"github.com/soaringjerry/dreamhub/internal/entity"
)

// StructuredMemoryService defines the business logic for managing structured memories.
type StructuredMemoryService interface {
	// CreateMemory creates a new structured memory entry for the given user.
	// Changed userID type from uuid.UUID to string (UUID)
	CreateMemory(ctx context.Context, userID string, key, value string) (*entity.StructuredMemory, error)

	// GetMemoryByKey retrieves a specific memory entry for the user by key.
	// Changed userID type from uuid.UUID to string (UUID)
	GetMemoryByKey(ctx context.Context, userID string, key string) (*entity.StructuredMemory, error)

	// GetUserMemories retrieves all memory entries for the given user.
	// Changed userID type from uuid.UUID to string (UUID)
	GetUserMemories(ctx context.Context, userID string) ([]*entity.StructuredMemory, error)

	// UpdateMemory updates the value of an existing memory entry for the user.
	// Changed userID type from uuid.UUID to string (UUID)
	UpdateMemory(ctx context.Context, userID string, key, value string) (*entity.StructuredMemory, error)

	// DeleteMemory deletes a memory entry for the user by key.
	// Changed userID type from uuid.UUID to string (UUID)
	DeleteMemory(ctx context.Context, userID string, key string) error
}
