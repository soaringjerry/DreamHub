package entity

import (
	"time"

	"github.com/google/uuid"
)

// SyncStatus represents the synchronization status for a user
type SyncStatus struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	UserID         string     `json:"user_id" db:"user_id"`
	LastSyncAt     time.Time  `json:"last_sync_at" db:"last_sync_at"`
	DeviceID       string     `json:"device_id" db:"device_id"`
	SyncVersion    int64      `json:"sync_version" db:"sync_version"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// SyncRequest represents a sync request from client
type SyncRequest struct {
	DeviceID    string    `json:"device_id" binding:"required"`
	LastSyncAt  time.Time `json:"last_sync_at"`
	SyncVersion int64     `json:"sync_version"`
}

// SyncResponse represents the sync response to client
type SyncResponse struct {
	SyncVersion int64       `json:"sync_version"`
	Changes     SyncChanges `json:"changes"`
	Conflicts   []Conflict  `json:"conflicts,omitempty"`
	ServerTime  time.Time   `json:"server_time"`
}

// SyncChanges contains all changed entities since last sync
type SyncChanges struct {
	Conversations     []Conversation     `json:"conversations,omitempty"`
	Messages          []Message          `json:"messages,omitempty"`
	StructuredMemory  []StructuredMemory `json:"structured_memory,omitempty"`
	UserConfig        *UserConfig        `json:"user_config,omitempty"`
	DeletedItems      DeletedItems       `json:"deleted_items,omitempty"`
}

// DeletedItems tracks deleted entities
type DeletedItems struct {
	ConversationIDs      []string `json:"conversation_ids,omitempty"`
	MessageIDs           []string `json:"message_ids,omitempty"`
	StructuredMemoryKeys []string `json:"structured_memory_keys,omitempty"`
}

// Conflict represents a sync conflict
type Conflict struct {
	EntityType   string      `json:"entity_type"`
	EntityID     string      `json:"entity_id"`
	LocalValue   interface{} `json:"local_value"`
	RemoteValue  interface{} `json:"remote_value"`
	Resolution   string      `json:"resolution"` // "local", "remote", "merge"
}

// SyncPushRequest represents changes pushed from client
type SyncPushRequest struct {
	DeviceID    string      `json:"device_id" binding:"required"`
	SyncVersion int64       `json:"sync_version"`
	Changes     SyncChanges `json:"changes"`
}

// SyncPushResponse represents the response to a push request
type SyncPushResponse struct {
	Success     bool       `json:"success"`
	SyncVersion int64      `json:"sync_version"`
	Conflicts   []Conflict `json:"conflicts,omitempty"`
	ServerTime  time.Time  `json:"server_time"`
}