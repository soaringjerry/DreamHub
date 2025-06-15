package service

import (
	"context"
	"github.com/soaringjerry/dreamhub/internal/entity"
	"time"
)

// SyncService defines the interface for synchronization logic
type SyncService interface {
	// GetSyncStatus retrieves the sync status for a user's device
	GetSyncStatus(ctx context.Context, userID, deviceID string) (*entity.SyncStatus, error)

	// PullChanges retrieves all changes since the last sync
	PullChanges(ctx context.Context, userID string, req entity.SyncRequest) (*entity.SyncResponse, error)

	// PushChanges applies changes from the client and returns conflicts if any
	PushChanges(ctx context.Context, userID string, req entity.SyncPushRequest) (*entity.SyncPushResponse, error)

	// ResolveConflicts resolves sync conflicts based on the chosen strategy
	ResolveConflicts(ctx context.Context, userID string, conflicts []entity.Conflict) error
}