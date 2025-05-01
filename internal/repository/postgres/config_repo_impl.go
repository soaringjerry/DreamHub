package postgres

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/soaringjerry/dreamhub/internal/entity"
	"github.com/soaringjerry/dreamhub/internal/repository"

	// Removed unused: "github.com/soaringjerry/dreamhub/internal/util"
	"github.com/soaringjerry/dreamhub/pkg/apperr"
	"github.com/soaringjerry/dreamhub/pkg/config" // Import config to get secret
	"github.com/soaringjerry/dreamhub/pkg/logger"
)

// Ensure postgresConfigRepository implements ConfigRepository interface.
var _ repository.ConfigRepository = (*postgresConfigRepository)(nil)

type postgresConfigRepository struct {
	db               *pgxpool.Pool
	encryptionSecret string // Store the encryption secret
}

// NewPostgresConfigRepository creates a new instance of postgresConfigRepository.
func NewPostgresConfigRepository(db *pgxpool.Pool) repository.ConfigRepository {
	cfg := config.Get() // Get loaded config
	if cfg.UserAPIKeyEncryptionSecret == "" {
		// This should have been caught during LoadConfig, but double-check
		log.Fatal("Config Repository: USER_API_KEY_ENCRYPTION_SECRET is not set") // Use log.Fatal
	}
	return &postgresConfigRepository{
		db:               db,
		encryptionSecret: cfg.UserAPIKeyEncryptionSecret,
	}
}

// GetByUserID retrieves the user configuration for a given user ID.
// It returns the configuration with the API key still encrypted.
func (r *postgresConfigRepository) GetByUserID(ctx context.Context, userID uint) (*entity.UserConfig, error) {
	query := `
		SELECT id, user_id, api_endpoint, model_name, api_key, created_at, updated_at
		FROM user_configs
		WHERE user_id = $1
	`
	var config entity.UserConfig
	// Scan directly into the struct fields, including the *[]byte for api_key
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&config.ID,
		&config.UserID,
		&config.ApiEndpoint,
		&config.ModelName,
		&config.ApiKey, // Scan directly into *[]byte? Check pgx docs. Yes, *[]byte works for BYTEA.
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.InfoContext(ctx, "用户配置未找到", "user_id", userID)
			// Return specific not found error
			return nil, apperr.New(apperr.CodeNotFound, "用户配置未找到")
		}
		logger.ErrorContext(ctx, "按用户 ID 获取配置时数据库出错", "user_id", userID, "error", err)
		return nil, apperr.Wrap(err, apperr.CodeInternal, "获取用户配置失败")
	}

	// No decryption here. Return the entity as is from the DB.
	return &config, nil
}

// Upsert creates or updates the user configuration.
// It expects the ApiKey in the input config to be already encrypted (*[]byte).
func (r *postgresConfigRepository) Upsert(ctx context.Context, config *entity.UserConfig) error {
	// Assumes config.ApiKey (*[]byte) is already encrypted by the service layer.
	query := `
		INSERT INTO user_configs (user_id, api_endpoint, model_name, api_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			api_endpoint = EXCLUDED.api_endpoint,
			model_name = EXCLUDED.model_name,
			api_key = EXCLUDED.api_key,
			updated_at = NOW()
	`

	_, err := r.db.Exec(ctx, query,
		config.UserID,
		config.ApiEndpoint,
		config.ModelName,
		config.ApiKey, // Pass the *[]byte directly
	)

	if err != nil {
		logger.ErrorContext(ctx, "Upsert 用户配置时数据库出错", "user_id", config.UserID, "error", err)
		return apperr.Wrap(err, apperr.CodeInternal, "更新用户配置失败")
	}

	logger.InfoContext(ctx, "用户配置 Upsert 成功", "user_id", config.UserID)
	return nil
}
