package repository

import (
	"context"

	"github.com/soaringjerry/dreamhub/internal/entity"
)

// ConfigRepository defines the interface for interacting with user configuration data.
type ConfigRepository interface {
	// GetByUserID retrieves the user configuration for a given user ID.
	// It returns entity.ErrNotFound if no configuration exists for the user.
	// Changed userID type from uint to string (UUID)
	GetByUserID(ctx context.Context, userID string) (*entity.UserConfig, error)

	// Upsert creates or updates the user configuration.
	// It accepts UserConfig with UserID as string (UUID).
	// If a configuration for the user ID already exists, it updates it; otherwise, it creates a new one.
	Upsert(ctx context.Context, config *entity.UserConfig) error
}
