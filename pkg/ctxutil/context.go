package ctxutil

import (
	"context"
	"errors"
)

// ctxKey is an unexported type for context keys defined in this package.
// This prevents collisions with keys defined in other packages.
type ctxKey string

// userIDKey is the context key for the user ID.
const userIDKey ctxKey = "user_id"

// ErrUserIDNotFound indicates that the user ID was not found in the context.
var ErrUserIDNotFound = errors.New("user ID not found in context")

// WithUserID returns a new context with the provided user ID added.
func WithUserID(ctx context.Context, userID string) context.Context {
	if userID == "" {
		// Avoid storing empty user IDs, though validation should happen earlier.
		return ctx
	}
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext retrieves the user ID from the context.
// It returns the user ID and nil error if found, otherwise returns an empty string
// and ErrUserIDNotFound.
func UserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok || userID == "" {
		return "", ErrUserIDNotFound
	}
	return userID, nil
}

// TODO: Add similar functions for TraceID or other context values if needed.
