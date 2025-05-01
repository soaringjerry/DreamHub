package service

import (
	"context"

	"github.com/soaringjerry/dreamhub/internal/entity"
)

// LoginCredentials holds the username and password for login attempts.
type LoginCredentials struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterPayload holds the necessary information for user registration.
type RegisterPayload struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8"` // Consider adding password complexity rules
}

// LoginResponse contains the JWT and user information upon successful login.
type LoginResponse struct {
	Token string               `json:"token"`
	User  entity.SanitizedUser `json:"user"`
}

// AuthService defines the interface for authentication operations.
type AuthService interface {
	// Register handles new user registration.
	// It validates input, hashes the password, and creates the user via UserRepository.
	Register(ctx context.Context, payload RegisterPayload) (*entity.SanitizedUser, error)

	// Login authenticates a user based on credentials.
	// It verifies the password and generates a JWT upon success.
	Login(ctx context.Context, creds LoginCredentials) (*LoginResponse, error)

	// ValidateToken parses and validates a JWT string.
	// It returns the user ID contained within the token if valid, otherwise returns an error.
	ValidateToken(ctx context.Context, tokenString string) (string, error) // Returns UserID (UUID as string)
}
