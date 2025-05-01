package repository

import (
	"context"

	"github.com/soaringjerry/dreamhub/internal/entity"
)

// ConfigRepository defines the interface for interacting with user configuration data.
type ConfigRepository interface {
	// GetByUserID retrieves the user configuration for a given user ID.
	// It returns entity.ErrNotFound if no configuration exists for the user.
	GetByUserID(ctx context.Context, userID uint) (*entity.UserConfig, error)

	// Upsert creates or updates the user configuration.
	// If a configuration for the user ID already exists, it updates it; otherwise, it creates a new one.
	Upsert(ctx context.Context, config *entity.UserConfig) error
}
