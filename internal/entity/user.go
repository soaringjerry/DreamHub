package entity

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system.
type User struct {
	ID           uuid.UUID `json:"id" db:"id"`                 // Unique identifier for the user
	Username     string    `json:"username" db:"username"`     // User's chosen username (must be unique)
	PasswordHash string    `json:"-" db:"password_hash"`       // Hashed password (never expose this in JSON)
	CreatedAt    time.Time `json:"created_at" db:"created_at"` // Timestamp when the user was created
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"` // Timestamp when the user was last updated
}

// SanitizedUser represents a user structure safe for API responses, excluding sensitive data.
type SanitizedUser struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Sanitize removes sensitive information (like password hash) before sending user data to the client.
func (u *User) Sanitize() SanitizedUser {
	return SanitizedUser{
		ID:        u.ID,
		Username:  u.Username,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
