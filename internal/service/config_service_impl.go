package service

import (
	"context"

	"github.com/soaringjerry/dreamhub/internal/dto"
	"github.com/soaringjerry/dreamhub/internal/entity"
	"github.com/soaringjerry/dreamhub/internal/repository"
	"github.com/soaringjerry/dreamhub/internal/util" // For encryption/decryption
	"github.com/soaringjerry/dreamhub/pkg/apperr"
	"github.com/soaringjerry/dreamhub/pkg/config" // For defaults and secret
	"github.com/soaringjerry/dreamhub/pkg/logger"
)

// Ensure configServiceImpl implements ConfigService interface.
var _ ConfigService = (*configServiceImpl)(nil)

type configServiceImpl struct {
	configRepo       repository.ConfigRepository
	globalConfig     *config.Config // Store global config for defaults
	encryptionSecret string         // Store encryption secret
}

// NewConfigService creates a new instance of ConfigService.
func NewConfigService(configRepo repository.ConfigRepository) ConfigService {
	cfg := config.Get()
	// Secret presence is checked in repo constructor, but good to have it here too if needed directly.
	// if cfg.UserAPIKeyEncryptionSecret == "" {
	// 	log.Fatal("Config Service: USER_API_KEY_ENCRYPTION_SECRET is not set")
	// }
	return &configServiceImpl{
		configRepo:       configRepo,
		globalConfig:     cfg,
		encryptionSecret: cfg.UserAPIKeyEncryptionSecret,
	}
}

// GetUserConfig retrieves the configuration for a specific user, merging with defaults.
// Changed userID type from uint to string (UUID)
func (s *configServiceImpl) GetUserConfig(ctx context.Context, userID string) (*dto.UserConfigDTO, error) {
	// Pass string userID to the repository
	userConfig, err := s.configRepo.GetByUserID(ctx, userID)

	// Prepare default DTO using global config
	defaultDTO := &dto.UserConfigDTO{
		ApiEndpoint: &s.globalConfig.OpenAIAPIKey, // Default API endpoint (Note: Using OpenAI key field for now, might need dedicated default field) - Let's assume global config has defaults
		ModelName:   &s.globalConfig.OpenAIModel,  // Default Model Name
		ApiKeyIsSet: false,                        // Default to false
	}
	// Adjust if global defaults are empty strings, make them nil pointers? Or handle in frontend?
	// Let's assume empty string defaults are valid and frontend handles display.
	// If global config values are empty, the pointers will point to empty strings.

	if err != nil {
		if appErr, ok := err.(*apperr.AppError); ok && appErr.Code == apperr.CodeNotFound {
			// User has no specific config, return defaults
			logger.InfoContext(ctx, "用户配置未找到，返回默认值", "user_id", userID) // Log string userID
			return defaultDTO, nil
		}
		// Other error fetching config
		logger.ErrorContext(ctx, "获取用户配置时出错", "user_id", userID, "error", err) // Log string userID
		return nil, err                                                        // Return the original error (already wrapped by repo)
	}

	// User config found, merge with defaults
	resultDTO := &dto.UserConfigDTO{
		ApiEndpoint: userConfig.ApiEndpoint,
		ModelName:   userConfig.ModelName,
		ApiKeyIsSet: userConfig.ApiKey != nil && len(*userConfig.ApiKey) > 0, // Check if encrypted key exists
	}

	// Fill missing fields with defaults
	if resultDTO.ApiEndpoint == nil || *resultDTO.ApiEndpoint == "" { // Check for nil or empty string
		resultDTO.ApiEndpoint = defaultDTO.ApiEndpoint
	}
	if resultDTO.ModelName == nil || *resultDTO.ModelName == "" { // Check for nil or empty string
		resultDTO.ModelName = defaultDTO.ModelName
	}

	// Note: We don't decrypt the API key here, just indicate if it's set.
	logger.InfoContext(ctx, "成功获取用户配置（含默认值）", "user_id", userID) // Log string userID
	return resultDTO, nil
}

// UpdateUserConfig updates or creates the configuration for a specific user.
// Changed userID type from uint to string (UUID)
func (s *configServiceImpl) UpdateUserConfig(ctx context.Context, userID string, updateDTO *dto.UpdateUserConfigDTO) error {
	// 1. Get existing config (or determine if it's new)
	// Pass string userID to the repository
	existingConfig, err := s.configRepo.GetByUserID(ctx, userID)
	isNew := false
	if err != nil {
		if appErr, ok := err.(*apperr.AppError); ok && appErr.Code == apperr.CodeNotFound {
			// Config doesn't exist, we'll create a new one
			isNew = true
			// Create a base entity with the string UserID
			existingConfig = &entity.UserConfig{UserID: userID}
			logger.InfoContext(ctx, "用户配置不存在，将创建新配置", "user_id", userID) // Log string userID
		} else {
			// Other error fetching config
			logger.ErrorContext(ctx, "更新前获取用户配置失败", "user_id", userID, "error", err) // Log string userID
			return err                                                               // Return the original error
		}
	}

	// 2. Prepare the entity to be saved by applying updates
	configToSave := *existingConfig // Make a copy to modify

	// Apply updates from DTO
	if updateDTO.ApiEndpoint != nil {
		configToSave.ApiEndpoint = updateDTO.ApiEndpoint // Update pointer directly
	}
	if updateDTO.ModelName != nil {
		configToSave.ModelName = updateDTO.ModelName // Update pointer directly
	}

	// Handle API Key update (encryption/clearing)
	if updateDTO.ApiKey != nil {
		plaintextKey := *updateDTO.ApiKey
		if plaintextKey == "" {
			// User wants to clear the key
			configToSave.ApiKey = nil
			logger.InfoContext(ctx, "用户请求清除 API Key", "user_id", userID) // Log string userID
		} else {
			// User provided a new key, encrypt it
			encryptedKeyBytes, err := util.EncryptString(plaintextKey, s.encryptionSecret)
			if err != nil {
				logger.ErrorContext(ctx, "加密用户 API Key 失败", "user_id", userID, "error", err) // Log string userID
				// Return an internal error, don't expose encryption details
				return apperr.New(apperr.CodeInternal, "无法处理 API Key")
			}
			configToSave.ApiKey = &encryptedKeyBytes                          // Store the encrypted key
			logger.InfoContext(ctx, "用户提供了新的 API Key，已加密", "user_id", userID) // Log string userID
		}
	}
	// If updateDTO.ApiKey is nil, configToSave.ApiKey retains its existing value (or nil if new)

	// 3. Upsert the configuration
	// Upsert should now accept UserConfig with string UserID
	err = s.configRepo.Upsert(ctx, &configToSave)
	if err != nil {
		logger.ErrorContext(ctx, "Upsert 用户配置失败", "user_id", userID, "error", err) // Log string userID
		return err                                                                 // Return the error from repository (already wrapped)
	}

	logAction := "更新"
	if isNew {
		logAction = "创建"
	}
	logger.InfoContext(ctx, "成功"+logAction+"用户配置", "user_id", userID) // Log string userID
	return nil
}
