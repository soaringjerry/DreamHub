package repository

import (
	"context"

	"github.com/soaringjerry/dreamhub/internal/entity"
)

// UserRepository defines the interface for interacting with user data storage.
type UserRepository interface {
	// CreateUser creates a new user record in the storage.
	CreateUser(ctx context.Context, user *entity.User) error
	// GetUserByUsername retrieves a user by their username.
	// It should return an error (e.g., ErrNotFound from a custom error package or pgx.ErrNoRows)
	// if the user with the given username does not exist.
	GetUserByUsername(ctx context.Context, username string) (*entity.User, error)
	// GetUserByID retrieves a user by their ID.
	// It should return an error if the user with the given ID does not exist.
	GetUserByID(ctx context.Context, id string) (*entity.User, error) // Assuming ID is string (UUID)
}
