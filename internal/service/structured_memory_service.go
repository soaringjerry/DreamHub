package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/soaringjerry/dreamhub/internal/entity"
)

// StructuredMemoryService defines the business logic for managing structured memories.
type StructuredMemoryService interface {
	// CreateMemory creates a new structured memory entry for the given user.
	CreateMemory(ctx context.Context, userID uuid.UUID, key, value string) (*entity.StructuredMemory, error)

	// GetMemoryByKey retrieves a specific memory entry for the user by key.
	GetMemoryByKey(ctx context.Context, userID uuid.UUID, key string) (*entity.StructuredMemory, error)

	// GetUserMemories retrieves all memory entries for the given user.
	GetUserMemories(ctx context.Context, userID uuid.UUID) ([]*entity.StructuredMemory, error)

	// UpdateMemory updates the value of an existing memory entry for the user.
	UpdateMemory(ctx context.Context, userID uuid.UUID, key, value string) (*entity.StructuredMemory, error)

	// DeleteMemory deletes a memory entry for the user by key.
	DeleteMemory(ctx context.Context, userID uuid.UUID, key string) error
}
