package postgres

import (
	"context"
	"database/sql"
	"github.com/soaringjerry/dreamhub/internal/entity"
	"github.com/soaringjerry/dreamhub/internal/repository"
	"github.com/soaringjerry/dreamhub/pkg/apperr"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type syncRepositoryImpl struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

// NewSyncRepository creates a new PostgreSQL-based sync repository
func NewSyncRepository(db *pgxpool.Pool, log *slog.Logger) repository.SyncRepository {
	return &syncRepositoryImpl{db: db, log: log}
}

func (r *syncRepositoryImpl) GetSyncStatus(ctx context.Context, userID, deviceID string) (*entity.SyncStatus, error) {
	query := `
		SELECT id, user_id, device_id, last_sync_at, sync_version, created_at, updated_at
		FROM sync_status
		WHERE user_id = $1 AND device_id = $2
	`

	var status entity.SyncStatus
	err := r.db.QueryRow(ctx, query, userID, deviceID).Scan(
		&status.ID,
		&status.UserID,
		&status.DeviceID,
		&status.LastSyncAt,
		&status.SyncVersion,
		&status.CreatedAt,
		&status.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperr.New(apperr.CodeNotFound, "sync status not found")
		}
		return nil, apperr.Wrap(err, apperr.CodeInternal, "failed to get sync status")
	}

	return &status, nil
}

func (r *syncRepositoryImpl) CreateOrUpdateSyncStatus(ctx context.Context, status *entity.SyncStatus) error {
	query := `
		INSERT INTO sync_status (id, user_id, device_id, last_sync_at, sync_version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, device_id) 
		DO UPDATE SET 
			last_sync_at = EXCLUDED.last_sync_at,
			sync_version = EXCLUDED.sync_version,
			updated_at = EXCLUDED.updated_at
	`

	if status.ID == uuid.Nil {
		status.ID = uuid.New()
	}
	
	now := time.Now()
	status.UpdatedAt = now
	if status.CreatedAt.IsZero() {
		status.CreatedAt = now
	}

	_, err := r.db.Exec(ctx, query,
		status.ID,
		status.UserID,
		status.DeviceID,
		status.LastSyncAt,
		status.SyncVersion,
		status.CreatedAt,
		status.UpdatedAt,
	)

	if err != nil {
		return apperr.Wrap(err, apperr.CodeInternal, "failed to create/update sync status")
	}

	return nil
}

func (r *syncRepositoryImpl) UpdateSyncVersion(ctx context.Context, userID string, version int64) error {
	query := `UPDATE sync_status SET sync_version = $1, updated_at = $2 WHERE user_id = $3`
	_, err := r.db.Exec(ctx, query, version, time.Now(), userID)
	if err != nil {
		return apperr.Wrap(err, apperr.CodeInternal, "failed to update sync version")
	}
	return nil
}

func (r *syncRepositoryImpl) GetConversationsSince(ctx context.Context, userID string, since time.Time) ([]entity.Conversation, error) {
	query := `
		SELECT id, user_id, title, created_at, last_updated_at, version, sync_at
		FROM conversations
		WHERE user_id = $1 AND sync_at > $2 AND deleted_at IS NULL
		ORDER BY sync_at ASC
	`

	rows, err := r.db.Query(ctx, query, userID, since)
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeInternal, "failed to query conversations")
	}
	defer rows.Close()

	var conversations []entity.Conversation
	for rows.Next() {
		var conv entity.Conversation
		var version sql.NullInt64
		var syncAt sql.NullTime
		
		err := rows.Scan(
			&conv.ID,
			&conv.UserID,
			&conv.Title,
			&conv.CreatedAt,
			&conv.LastUpdatedAt,
			&version,
			&syncAt,
		)
		if err != nil {
			return nil, apperr.Wrap(err, apperr.CodeInternal, "failed to scan conversation")
		}
		conversations = append(conversations, conv)
	}

	return conversations, nil
}

func (r *syncRepositoryImpl) GetMessagesSince(ctx context.Context, userID string, since time.Time) ([]entity.Message, error) {
	query := `
		SELECT id, conversation_id, user_id, sender_role, content, timestamp, metadata, version, sync_at
		FROM conversation_history
		WHERE user_id = $1 AND sync_at > $2 AND deleted_at IS NULL
		ORDER BY sync_at ASC
	`

	rows, err := r.db.Query(ctx, query, userID, since)
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeInternal, "failed to query messages")
	}
	defer rows.Close()

	var messages []entity.Message
	for rows.Next() {
		var msg entity.Message
		var version sql.NullInt64
		var syncAt sql.NullTime
		
		err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.UserID,
			&msg.SenderRole,
			&msg.Content,
			&msg.Timestamp,
			&msg.Metadata,
			&version,
			&syncAt,
		)
		if err != nil {
			return nil, apperr.Wrap(err, apperr.CodeInternal, "failed to scan message")
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (r *syncRepositoryImpl) GetStructuredMemoriesSince(ctx context.Context, userID string, since time.Time) ([]entity.StructuredMemory, error) {
	query := `
		SELECT id, user_id, key, value, created_at, updated_at, version, sync_at
		FROM structured_memories
		WHERE user_id = $1 AND sync_at > $2 AND deleted_at IS NULL
		ORDER BY sync_at ASC
	`

	rows, err := r.db.Query(ctx, query, userID, since)
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeInternal, "failed to query structured memories")
	}
	defer rows.Close()

	var memories []entity.StructuredMemory
	for rows.Next() {
		var mem entity.StructuredMemory
		var version sql.NullInt64
		var syncAt sql.NullTime
		
		err := rows.Scan(
			&mem.ID,
			&mem.UserID,
			&mem.Key,
			&mem.Value,
			&mem.CreatedAt,
			&mem.UpdatedAt,
			&version,
			&syncAt,
		)
		if err != nil {
			return nil, apperr.Wrap(err, apperr.CodeInternal, "failed to scan structured memory")
		}
		memories = append(memories, mem)
	}

	return memories, nil
}

func (r *syncRepositoryImpl) GetUserConfigSince(ctx context.Context, userID string, since time.Time) (*entity.UserConfig, error) {
	query := `
		SELECT id, user_id, api_endpoint, model_name, encrypted_api_key, created_at, updated_at, version, sync_at
		FROM user_configs
		WHERE user_id = $1 AND sync_at > $2
		LIMIT 1
	`

	var config entity.UserConfig
	var version sql.NullInt64
	var syncAt sql.NullTime
	
	err := r.db.QueryRow(ctx, query, userID, since).Scan(
		&config.ID,
		&config.UserID,
		&config.APIEndpoint,
		&config.ModelName,
		&config.EncryptedAPIKey,
		&config.CreatedAt,
		&config.UpdatedAt,
		&version,
		&syncAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, apperr.Wrap(err, apperr.CodeInternal, "failed to get user config")
	}

	return &config, nil
}

func (r *syncRepositoryImpl) GetDeletedConversations(ctx context.Context, userID string, since time.Time) ([]string, error) {
	query := `
		SELECT id
		FROM conversations
		WHERE user_id = $1 AND deleted_at IS NOT NULL AND deleted_at > $2
		ORDER BY deleted_at ASC
	`

	rows, err := r.db.Query(ctx, query, userID, since)
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeInternal, "failed to query deleted conversations")
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, apperr.Wrap(err, apperr.CodeInternal, "failed to scan conversation id")
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (r *syncRepositoryImpl) GetDeletedMessages(ctx context.Context, userID string, since time.Time) ([]string, error) {
	query := `
		SELECT id
		FROM conversation_history
		WHERE user_id = $1 AND deleted_at IS NOT NULL AND deleted_at > $2
		ORDER BY deleted_at ASC
	`

	rows, err := r.db.Query(ctx, query, userID, since)
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeInternal, "failed to query deleted messages")
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, apperr.Wrap(err, apperr.CodeInternal, "failed to scan message id")
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (r *syncRepositoryImpl) GetDeletedStructuredMemories(ctx context.Context, userID string, since time.Time) ([]string, error) {
	query := `
		SELECT key
		FROM structured_memories
		WHERE user_id = $1 AND deleted_at IS NOT NULL AND deleted_at > $2
		ORDER BY deleted_at ASC
	`

	rows, err := r.db.Query(ctx, query, userID, since)
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeInternal, "failed to query deleted structured memories")
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, apperr.Wrap(err, apperr.CodeInternal, "failed to scan memory key")
		}
		keys = append(keys, key)
	}

	return keys, nil
}

func (r *syncRepositoryImpl) DetectConflicts(ctx context.Context, userID string, changes entity.SyncChanges) ([]entity.Conflict, error) {
	// TODO: Implement conflict detection logic
	// For now, we'll use a simple Last-Write-Wins strategy
	return []entity.Conflict{}, nil
}

func (r *syncRepositoryImpl) ApplySyncChanges(ctx context.Context, userID string, changes entity.SyncChanges) error {
	// Start a transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return apperr.Wrap(err, apperr.CodeInternal, "failed to start transaction")
	}
	defer tx.Rollback(ctx)

	// Apply conversation changes
	for _, conv := range changes.Conversations {
		query := `
			INSERT INTO conversations (id, user_id, title, created_at, last_updated_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO UPDATE SET
				title = EXCLUDED.title,
				last_updated_at = EXCLUDED.last_updated_at
		`
		_, err := tx.Exec(ctx, query, conv.ID, userID, conv.Title, conv.CreatedAt, conv.LastUpdatedAt)
		if err != nil {
			return apperr.Wrap(err, apperr.CodeInternal, "failed to apply conversation change")
		}
	}

	// Apply message changes
	for _, msg := range changes.Messages {
		query := `
			INSERT INTO conversation_history (id, conversation_id, user_id, sender_role, content, timestamp)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO UPDATE SET
				content = EXCLUDED.content,
				timestamp = EXCLUDED.timestamp
		`
		_, err := tx.Exec(ctx, query, msg.ID, msg.ConversationID, userID, msg.SenderRole, msg.Content, msg.Timestamp)
		if err != nil {
			return apperr.Wrap(err, apperr.CodeInternal, "failed to apply message change")
		}
	}

	// Apply structured memory changes
	for _, mem := range changes.StructuredMemory {
		query := `
			INSERT INTO structured_memories (id, user_id, key, value, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (user_id, key) DO UPDATE SET
				value = EXCLUDED.value,
				updated_at = EXCLUDED.updated_at
		`
		_, err := tx.Exec(ctx, query, mem.ID, userID, mem.Key, mem.Value, mem.CreatedAt, mem.UpdatedAt)
		if err != nil {
			return apperr.Wrap(err, apperr.CodeInternal, "failed to apply structured memory change")
		}
	}

	// Apply user config changes
	if changes.UserConfig != nil {
		query := `
			INSERT INTO user_configs (id, user_id, api_endpoint, model_name, encrypted_api_key, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (user_id) DO UPDATE SET
				api_endpoint = EXCLUDED.api_endpoint,
				model_name = EXCLUDED.model_name,
				encrypted_api_key = EXCLUDED.encrypted_api_key,
				updated_at = EXCLUDED.updated_at
		`
		_, err := tx.Exec(ctx, query,
			changes.UserConfig.ID,
			userID,
			changes.UserConfig.APIEndpoint,
			changes.UserConfig.ModelName,
			changes.UserConfig.EncryptedAPIKey,
			changes.UserConfig.CreatedAt,
			changes.UserConfig.UpdatedAt,
		)
		if err != nil {
			return apperr.Wrap(err, apperr.CodeInternal, "failed to apply user config change")
		}
	}

	// Apply deletions
	if len(changes.DeletedItems.ConversationIDs) > 0 {
		query := `UPDATE conversations SET deleted_at = CURRENT_TIMESTAMP WHERE id = ANY($1) AND user_id = $2`
		_, err := tx.Exec(ctx, query, changes.DeletedItems.ConversationIDs, userID)
		if err != nil {
			return apperr.Wrap(err, apperr.CodeInternal, "failed to apply conversation deletions")
		}
	}

	if len(changes.DeletedItems.MessageIDs) > 0 {
		query := `UPDATE conversation_history SET deleted_at = CURRENT_TIMESTAMP WHERE id = ANY($1) AND user_id = $2`
		_, err := tx.Exec(ctx, query, changes.DeletedItems.MessageIDs, userID)
		if err != nil {
			return apperr.Wrap(err, apperr.CodeInternal, "failed to apply message deletions")
		}
	}

	if len(changes.DeletedItems.StructuredMemoryKeys) > 0 {
		query := `UPDATE structured_memories SET deleted_at = CURRENT_TIMESTAMP WHERE key = ANY($1) AND user_id = $2`
		_, err := tx.Exec(ctx, query, changes.DeletedItems.StructuredMemoryKeys, userID)
		if err != nil {
			return apperr.Wrap(err, apperr.CodeInternal, "failed to apply structured memory deletions")
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return apperr.Wrap(err, apperr.CodeInternal, "failed to commit transaction")
	}

	return nil
}