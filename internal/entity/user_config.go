package entity

import "time"

// UserConfig represents user-specific configuration settings.
type UserConfig struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      string    `gorm:"type:uuid;uniqueIndex;not null"`  // Foreign key to users table (assuming users.id is UUID)
	User        User      `gorm:"foreignKey:UserID;references:ID"` // Belongs to User, references User.ID (UUID)
	ApiEndpoint *string   // Nullable string for API endpoint
	ModelName   *string   // Nullable string for model name
	ApiKey      *[]byte   // Nullable byte slice for encrypted API key
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}
