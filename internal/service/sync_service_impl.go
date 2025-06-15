package service

import (
	"context"
	"dreamhub/internal/entity"
	"dreamhub/internal/repository"
	"dreamhub/pkg/apperr"
	"dreamhub/pkg/logger"
	"time"

	"github.com/google/uuid"
)

type syncServiceImpl struct {
	syncRepo    repository.SyncRepository
	chatRepo    repository.ChatRepository
	memoryRepo  repository.StructuredMemoryRepository
	configRepo  repository.ConfigRepository
	log         *logger.Logger
}

// NewSyncService creates a new sync service instance
func NewSyncService(
	syncRepo repository.SyncRepository,
	chatRepo repository.ChatRepository,
	memoryRepo repository.StructuredMemoryRepository,
	configRepo repository.ConfigRepository,
	log *logger.Logger,
) SyncService {
	return &syncServiceImpl{
		syncRepo:   syncRepo,
		chatRepo:   chatRepo,
		memoryRepo: memoryRepo,
		configRepo: configRepo,
		log:        log,
	}
}

func (s *syncServiceImpl) GetSyncStatus(ctx context.Context, userID, deviceID string) (*entity.SyncStatus, error) {
	status, err := s.syncRepo.GetSyncStatus(ctx, userID, deviceID)
	if err != nil {
		if apperr.IsNotFound(err) {
			// Create new sync status for first-time sync
			status = &entity.SyncStatus{
				ID:          uuid.New(),
				UserID:      userID,
				DeviceID:    deviceID,
				LastSyncAt:  time.Now().Add(-24 * 365 * time.Hour), // Set to 1 year ago for first sync
				SyncVersion: 0,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			if err := s.syncRepo.CreateOrUpdateSyncStatus(ctx, status); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return status, nil
}

func (s *syncServiceImpl) PullChanges(ctx context.Context, userID string, req entity.SyncRequest) (*entity.SyncResponse, error) {
	// Get or create sync status
	status, err := s.GetSyncStatus(ctx, userID, req.DeviceID)
	if err != nil {
		return nil, err
	}

	// Use the client's last sync time if provided, otherwise use server's record
	lastSyncTime := status.LastSyncAt
	if !req.LastSyncAt.IsZero() && req.LastSyncAt.Before(lastSyncTime) {
		lastSyncTime = req.LastSyncAt
	}

	// Fetch all changes since last sync
	changes := entity.SyncChanges{}

	// Get conversations
	conversations, err := s.syncRepo.GetConversationsSince(ctx, userID, lastSyncTime)
	if err != nil {
		return nil, err
	}
	changes.Conversations = conversations

	// Get messages
	messages, err := s.syncRepo.GetMessagesSince(ctx, userID, lastSyncTime)
	if err != nil {
		return nil, err
	}
	changes.Messages = messages

	// Get structured memories
	memories, err := s.syncRepo.GetStructuredMemoriesSince(ctx, userID, lastSyncTime)
	if err != nil {
		return nil, err
	}
	changes.StructuredMemory = memories

	// Get user config
	config, err := s.syncRepo.GetUserConfigSince(ctx, userID, lastSyncTime)
	if err != nil {
		return nil, err
	}
	changes.UserConfig = config

	// Get deleted items
	deletedConvs, err := s.syncRepo.GetDeletedConversations(ctx, userID, lastSyncTime)
	if err != nil {
		return nil, err
	}
	deletedMsgs, err := s.syncRepo.GetDeletedMessages(ctx, userID, lastSyncTime)
	if err != nil {
		return nil, err
	}
	deletedMemories, err := s.syncRepo.GetDeletedStructuredMemories(ctx, userID, lastSyncTime)
	if err != nil {
		return nil, err
	}

	changes.DeletedItems = entity.DeletedItems{
		ConversationIDs:      deletedConvs,
		MessageIDs:           deletedMsgs,
		StructuredMemoryKeys: deletedMemories,
	}

	// Update sync status
	now := time.Now()
	status.LastSyncAt = now
	status.SyncVersion++
	if err := s.syncRepo.CreateOrUpdateSyncStatus(ctx, status); err != nil {
		return nil, err
	}

	return &entity.SyncResponse{
		SyncVersion: status.SyncVersion,
		Changes:     changes,
		Conflicts:   []entity.Conflict{},
		ServerTime:  now,
	}, nil
}

func (s *syncServiceImpl) PushChanges(ctx context.Context, userID string, req entity.SyncPushRequest) (*entity.SyncPushResponse, error) {
	// Get current sync status
	status, err := s.GetSyncStatus(ctx, userID, req.DeviceID)
	if err != nil {
		return nil, err
	}

	// Detect conflicts
	conflicts, err := s.syncRepo.DetectConflicts(ctx, userID, req.Changes)
	if err != nil {
		return nil, err
	}

	// If no conflicts, apply changes
	if len(conflicts) == 0 {
		if err := s.syncRepo.ApplySyncChanges(ctx, userID, req.Changes); err != nil {
			return nil, err
		}

		// Update sync version
		status.SyncVersion++
		status.LastSyncAt = time.Now()
		if err := s.syncRepo.CreateOrUpdateSyncStatus(ctx, status); err != nil {
			return nil, err
		}

		return &entity.SyncPushResponse{
			Success:     true,
			SyncVersion: status.SyncVersion,
			Conflicts:   nil,
			ServerTime:  time.Now(),
		}, nil
	}

	// Return conflicts for client to resolve
	return &entity.SyncPushResponse{
		Success:     false,
		SyncVersion: status.SyncVersion,
		Conflicts:   conflicts,
		ServerTime:  time.Now(),
	}, nil
}

func (s *syncServiceImpl) ResolveConflicts(ctx context.Context, userID string, conflicts []entity.Conflict) error {
	// TODO: Implement conflict resolution logic
	// For now, we'll use Last-Write-Wins strategy
	return nil
}