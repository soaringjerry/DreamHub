package service

import (
	"context"

	"github.com/soaringjerry/dreamhub/internal/dto"
)

// ConfigService defines the interface for managing user configurations.
type ConfigService interface {
	// GetUserConfig retrieves the configuration for a specific user.
	// It merges user-specific settings with global defaults and returns a DTO.
	// The DTO indicates whether a user-specific API key is set, but does not return the key itself.
	GetUserConfig(ctx context.Context, userID uint) (*dto.UserConfigDTO, error)

	// UpdateUserConfig updates or creates the configuration for a specific user.
	// It takes an UpdateUserConfigDTO containing the plaintext API key (if provided).
	// The service is responsible for encrypting the API key before storing it.
	// Passing an empty string "" for ApiKey in the DTO clears the stored key.
	// Passing nil for ApiKey leaves the stored key unchanged.
	UpdateUserConfig(ctx context.Context, userID uint, updateDTO *dto.UpdateUserConfigDTO) error
}
