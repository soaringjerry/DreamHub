package repository

import (
	"context"
	"time"

	"dreamhub/internal/entity"
)

// SyncRepository defines the interface for sync-related database operations
type SyncRepository interface {
	// Sync status management
	GetSyncStatus(ctx context.Context, userID, deviceID string) (*entity.SyncStatus, error)
	CreateOrUpdateSyncStatus(ctx context.Context, status *entity.SyncStatus) error
	UpdateSyncVersion(ctx context.Context, userID string, version int64) error

	// Get changes since last sync
	GetConversationsSince(ctx context.Context, userID string, since time.Time) ([]entity.Conversation, error)
	GetMessagesSince(ctx context.Context, userID string, since time.Time) ([]entity.Message, error)
	GetStructuredMemoriesSince(ctx context.Context, userID string, since time.Time) ([]entity.StructuredMemory, error)
	GetUserConfigSince(ctx context.Context, userID string, since time.Time) (*entity.UserConfig, error)

	// Get deleted items
	GetDeletedConversations(ctx context.Context, userID string, since time.Time) ([]string, error)
	GetDeletedMessages(ctx context.Context, userID string, since time.Time) ([]string, error)
	GetDeletedStructuredMemories(ctx context.Context, userID string, since time.Time) ([]string, error)

	// Conflict detection
	DetectConflicts(ctx context.Context, userID string, changes entity.SyncChanges) ([]entity.Conflict, error)

	// Apply sync changes
	ApplySyncChanges(ctx context.Context, userID string, changes entity.SyncChanges) error
}