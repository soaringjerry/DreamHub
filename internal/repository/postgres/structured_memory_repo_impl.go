package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/soaringjerry/dreamhub/internal/entity"
	"github.com/soaringjerry/dreamhub/internal/repository"
)

// ErrDuplicateKey is returned when trying to create an entry with a key that already exists for the user.
var ErrDuplicateKey = errors.New("duplicate key for user")

// ErrNotFound is returned when a requested resource is not found.
// Consider defining this in a shared errors package if used across multiple repositories.
var ErrNotFound = pgx.ErrNoRows // Map pgx.ErrNoRows to a repository-level error

type structuredMemoryRepoImpl struct {
	db *pgxpool.Pool
}

// NewStructuredMemoryRepository creates a new instance of StructuredMemoryRepository.
func NewStructuredMemoryRepository(db *pgxpool.Pool) repository.StructuredMemoryRepository {
	return &structuredMemoryRepoImpl{db: db}
}

// Create adds a new structured memory entry.
func (r *structuredMemoryRepoImpl) Create(ctx context.Context, memory *entity.StructuredMemory) error {
	query := `
		INSERT INTO structured_memories (user_id, key, value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	now := time.Now()
	err := r.db.QueryRow(ctx, query,
		memory.UserID,
		memory.Key,
		memory.Value,
		now, // Set CreatedAt server-side
		now, // Set UpdatedAt server-side
	).Scan(&memory.ID, &memory.CreatedAt, &memory.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		// Check for unique constraint violation (code 23505)
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrDuplicateKey
		}
		return err // Return other errors directly
	}
	return nil
}

// GetByKey retrieves a specific structured memory entry by user ID and key.
func (r *structuredMemoryRepoImpl) GetByKey(ctx context.Context, userID uuid.UUID, key string) (*entity.StructuredMemory, error) {
	query := `
		SELECT id, user_id, key, value, created_at, updated_at
		FROM structured_memories
		WHERE user_id = $1 AND key = $2`

	memory := &entity.StructuredMemory{}
	err := r.db.QueryRow(ctx, query, userID, key).Scan(
		&memory.ID,
		&memory.UserID,
		&memory.Key,
		&memory.Value,
		&memory.CreatedAt,
		&memory.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return memory, nil
}

// GetByUserID retrieves all structured memory entries for a user.
func (r *structuredMemoryRepoImpl) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.StructuredMemory, error) {
	query := `
		SELECT id, user_id, key, value, created_at, updated_at
		FROM structured_memories
		WHERE user_id = $1
		ORDER BY created_at DESC` // Or order by key, depending on desired default sorting

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	memories := []*entity.StructuredMemory{}
	for rows.Next() {
		memory := &entity.StructuredMemory{}
		err := rows.Scan(
			&memory.ID,
			&memory.UserID,
			&memory.Key,
			&memory.Value,
			&memory.CreatedAt,
			&memory.UpdatedAt,
		)
		if err != nil {
			return nil, err // Return error encountered during row scanning
		}
		memories = append(memories, memory)
	}

	if err = rows.Err(); err != nil {
		return nil, err // Check for errors during iteration
	}

	return memories, nil
}

// Update modifies an existing structured memory entry's value.
func (r *structuredMemoryRepoImpl) Update(ctx context.Context, memory *entity.StructuredMemory) error {
	query := `
		UPDATE structured_memories
		SET value = $1, updated_at = $2
		WHERE user_id = $3 AND key = $4
		RETURNING updated_at` // Return updated_at to confirm update and potentially update the entity

	now := time.Now()
	err := r.db.QueryRow(ctx, query,
		memory.Value,
		now, // Set UpdatedAt server-side
		memory.UserID,
		memory.Key,
	).Scan(&memory.UpdatedAt) // Scan the updated timestamp back into the entity

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// If the row to update wasn't found, return ErrNotFound
			return ErrNotFound
		}
		return err
	}
	return nil
}

// Delete removes a structured memory entry by user ID and key.
func (r *structuredMemoryRepoImpl) Delete(ctx context.Context, userID uuid.UUID, key string) error {
	query := `DELETE FROM structured_memories WHERE user_id = $1 AND key = $2`

	result, err := r.db.Exec(ctx, query, userID, key)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		// If no rows were deleted, it means the entry didn't exist
		return ErrNotFound
	}

	return nil
}
