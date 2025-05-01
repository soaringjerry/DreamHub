package service

import (
	"context"
	"errors"
	"strings"

	// "github.com/google/uuid" // Removed unused import
	"github.com/soaringjerry/dreamhub/internal/entity"
	"github.com/soaringjerry/dreamhub/internal/repository"
)

// structuredMemoryServiceImpl implements the StructuredMemoryService interface.
type structuredMemoryServiceImpl struct {
	repo repository.StructuredMemoryRepository
}

// NewStructuredMemoryService creates a new instance of StructuredMemoryService.
func NewStructuredMemoryService(repo repository.StructuredMemoryRepository) StructuredMemoryService {
	return &structuredMemoryServiceImpl{repo: repo}
}

// CreateMemory creates a new structured memory entry.
// Changed userID type from uuid.UUID to string (UUID)
func (s *structuredMemoryServiceImpl) CreateMemory(ctx context.Context, userID string, key, value string) (*entity.StructuredMemory, error) {
	// Basic validation
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, errors.New("memory key cannot be empty")
	}
	// Value validation might depend on requirements (e.g., max length, format if JSON)
	if strings.TrimSpace(value) == "" {
		return nil, errors.New("memory value cannot be empty")
	}

	memory := &entity.StructuredMemory{
		UserID: userID,
		Key:    key,
		Value:  value,
		// ID, CreatedAt, UpdatedAt will be set by the repository/database
	}

	err := s.repo.Create(ctx, memory)
	if err != nil {
		// The repository already maps pgx errors to specific repo errors like ErrDuplicateKey
		return nil, err
	}
	return memory, nil
}

// GetMemoryByKey retrieves a specific memory entry by key.
// Changed userID type from uuid.UUID to string (UUID)
func (s *structuredMemoryServiceImpl) GetMemoryByKey(ctx context.Context, userID string, key string) (*entity.StructuredMemory, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, errors.New("memory key cannot be empty") // TODO: Use apperr
	}
	// The repository handles ErrNotFound mapping
	return s.repo.GetByKey(ctx, userID, key)
}

// GetUserMemories retrieves all memory entries for a user.
// Changed userID type from uuid.UUID to string (UUID)
func (s *structuredMemoryServiceImpl) GetUserMemories(ctx context.Context, userID string) ([]*entity.StructuredMemory, error) {
	// Pass string userID to repository
	return s.repo.GetByUserID(ctx, userID)
}

// UpdateMemory updates an existing memory entry.
// Changed userID type from uuid.UUID to string (UUID)
func (s *structuredMemoryServiceImpl) UpdateMemory(ctx context.Context, userID string, key, value string) (*entity.StructuredMemory, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, errors.New("memory key cannot be empty") // TODO: Use apperr
	}
	if strings.TrimSpace(value) == "" {
		return nil, errors.New("memory value cannot be empty") // TODO: Use apperr
	}

	// First, check if the memory exists (optional, repo update also checks)
	// _, err := s.repo.GetByKey(ctx, userID, key)
	// if err != nil {
	// 	return nil, err // Return ErrNotFound or other errors
	// }

	memory := &entity.StructuredMemory{
		UserID: userID,
		Key:    key,
		Value:  value,
		// UpdatedAt will be set by the repository/database
	}

	err := s.repo.Update(ctx, memory)
	if err != nil {
		// Repository handles ErrNotFound if the key doesn't exist for the user
		return nil, err
	}
	// We need to return the full updated entity, including the new UpdatedAt.
	// The repo.Update currently scans UpdatedAt back, but not the full entity.
	// Let's fetch it again to return the complete, updated state.
	// Alternatively, the repo.Update could be modified to return the full entity.
	updatedMemory, fetchErr := s.repo.GetByKey(ctx, userID, key)
	if fetchErr != nil {
		// This shouldn't ideally happen if the update succeeded, but handle defensively.
		return nil, fetchErr
	}

	return updatedMemory, nil
}

// DeleteMemory deletes a memory entry by key.
// Changed userID type from uuid.UUID to string (UUID)
func (s *structuredMemoryServiceImpl) DeleteMemory(ctx context.Context, userID string, key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return errors.New("memory key cannot be empty") // TODO: Use apperr
	}
	// Repository handles ErrNotFound if the key doesn't exist
	// Pass string userID to repository
	return s.repo.Delete(ctx, userID, key)
}
