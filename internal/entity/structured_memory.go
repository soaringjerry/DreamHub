package entity

import (
	"time"

	"github.com/google/uuid"
)

// StructuredMemory represents a single piece of structured information
// stored for a specific user.
type StructuredMemory struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"` // Foreign key to users table
	Key       string    `json:"key" gorm:"type:varchar(255);not null;index"`
	Value     string    `json:"value" gorm:"type:text;not null"` // Use TEXT for flexibility
	CreatedAt time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`

	// Define unique constraint for UserID and Key
	// This is typically defined in the migration, but can be hinted here for ORM awareness
	// gorm:"uniqueIndex:idx_user_key"`
}

// TableName specifies the table name for the StructuredMemory entity.
func (StructuredMemory) TableName() string {
	return "structured_memories"
}
