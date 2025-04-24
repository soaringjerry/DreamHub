package repository

import (
	"context"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

// DocumentRepository defines the interface for interacting with document vectors.
// Implementations should handle user/tenant isolation internally based on context.
type DocumentRepository interface {
	// AddDocuments adds documents to the vector store for a specific user.
	// The user information should be extracted from the context.
	AddDocuments(ctx context.Context, docs []schema.Document) error

	// SimilaritySearch performs a similarity search within a specific user's documents.
	// The user information should be extracted from the context to apply filtering.
	// The 'options' parameter can be used for additional search parameters like score threshold,
	// but filtering based on user/tenant should be handled internally.
	SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...vectorstores.Option) ([]schema.Document, error)

	// TODO: Add other necessary methods like DeleteDocuments, etc.
}

// // Example context key for user ID (can be defined in a central place like pkg/ctxutil)
// type ctxKey string
// const UserIDKey ctxKey = "user_id"
