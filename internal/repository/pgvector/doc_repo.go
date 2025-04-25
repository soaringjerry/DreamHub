package pgvector

import (
	"context"
	"fmt" // Import fmt for error formatting

	"github.com/soaringjerry/dreamhub/internal/repository" // Import the repository interface package
	"github.com/soaringjerry/dreamhub/pkg/apperr"          // Import apperr for returning errors
	"github.com/soaringjerry/dreamhub/pkg/ctxutil"         // Import ctxutil for UserID extraction
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/pgvector"
)

// Compile-time check to ensure PGVectorDocumentRepository implements DocumentRepository
var _ repository.DocumentRepository = (*PGVectorDocumentRepository)(nil)

// PGVectorDocumentRepository implements the DocumentRepository interface using pgvector.
type PGVectorDocumentRepository struct {
	store pgvector.Store
	// We might need a helper to extract user ID from context later
	// userIDExtractor func(ctx context.Context) (string, error)
}

// New creates a new PGVectorDocumentRepository.
// It assumes the provided store is a valid, non-nil instance,
// as error handling should occur where the store is initially created.
func New(store pgvector.Store) *PGVectorDocumentRepository {
	return &PGVectorDocumentRepository{
		store: store,
	}
}

// AddDocuments adds documents, ensuring user_id metadata is present.
// It extracts user_id from context.
func (r *PGVectorDocumentRepository) AddDocuments(ctx context.Context, docs []schema.Document) error {
	// Extract user ID from context using ctxutil
	userID, err := ctxutil.UserIDFromContext(ctx)
	if err != nil {
		// Return an application error if user ID is missing
		return apperr.Wrap(apperr.UnauthorizedError, "User ID not found in context for AddDocuments", err)
	}

	// Ensure all documents have the correct user_id metadata
	processedDocs := make([]schema.Document, len(docs))
	for i, doc := range docs {
		// Create a mutable copy of metadata or initialize if nil
		newMetadata := make(map[string]any)
		if doc.Metadata != nil {
			for k, v := range doc.Metadata {
				newMetadata[k] = v
			}
		}
		// Set/overwrite the user_id from context
		newMetadata["user_id"] = userID
		processedDocs[i] = schema.Document{
			PageContent: doc.PageContent,
			Metadata:    newMetadata,
			// Score is usually set during retrieval, keep it default here
		}
	}

	// Add documents using the underlying pgvector store
	_, err = r.store.AddDocuments(ctx, processedDocs) // Use = instead of := because err is already declared
	if err != nil {
		// Wrap error using apperr
		return apperr.Wrap(apperr.VectorStoreAddError, "Failed to add documents to pgvector", err)
	}
	return nil
}

// SimilaritySearch performs similarity search, enforcing user_id filtering.
// It extracts user_id from context.
func (r *PGVectorDocumentRepository) SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...vectorstores.Option) ([]schema.Document, error) {
	// Extract user ID from context using ctxutil
	userID, err := ctxutil.UserIDFromContext(ctx)
	if err != nil {
		// Return an application error if user ID is missing
		return nil, apperr.Wrap(apperr.UnauthorizedError, "User ID not found in context for SimilaritySearch", err)
	}

	// Create the mandatory user_id filter
	mandatoryFilter := map[string]any{"user_id": userID}

	// Prepare the final options, ensuring our filter is included.
	// We create a new slice to avoid modifying the input `options`.
	finalOptions := make([]vectorstores.Option, 0, len(options)+1)

	// Add the mandatory filter first. If WithFilters is called multiple times,
	// the behavior might depend on the specific vectorstores implementation.
	// Assuming later calls might override or merge, adding ours first seems reasonable,
	// but ideally, we'd handle potential conflicts if the input options also contain WithFilters.
	// For simplicity now, we just add ours. A more robust approach might inspect
	// options and merge/replace the filter map.
	finalOptions = append(finalOptions, vectorstores.WithFilters(mandatoryFilter))

	// Append the original options provided by the caller
	finalOptions = append(finalOptions, options...)

	// Perform the search using the underlying pgvector store with the enforced filter
	results, err := r.store.SimilaritySearch(ctx, query, numDocuments, finalOptions...)
	if err != nil {
		return nil, fmt.Errorf("pgvector SimilaritySearch failed: %w", err)
	}

	return results, nil
}
