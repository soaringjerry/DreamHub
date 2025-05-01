package api

import (
	"fmt" // Actually add fmt import
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/soaringjerry/dreamhub/internal/dto"
	"github.com/soaringjerry/dreamhub/internal/service"
	"github.com/soaringjerry/dreamhub/pkg/apperr" // Import apperr
	"github.com/soaringjerry/dreamhub/pkg/logger"
)

// ConfigHandler handles API requests related to user configuration.
type ConfigHandler struct {
	configService service.ConfigService
}

// NewConfigHandler creates a new instance of ConfigHandler.
func NewConfigHandler(configService service.ConfigService) *ConfigHandler {
	return &ConfigHandler{configService: configService}
}

// RegisterRoutes registers the configuration API routes with the Gin router group.
// It assumes the group already has the authentication middleware applied.
func (h *ConfigHandler) RegisterRoutes(rg *gin.RouterGroup) {
	configGroup := rg.Group("/users/me/config") // Base path for user's own config
	{
		configGroup.GET("", h.GetUserConfig)
		configGroup.PUT("", h.UpdateUserConfig)
	}
}

// GetUserConfig godoc
// @Summary Get current user's configuration
// @Description Retrieves the configuration settings for the currently authenticated user, merging user-specific settings with global defaults.
// @Tags Config
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserConfigDTO "Successfully retrieved configuration"
// @Failure 401 {object} apperr.ErrorResponse "Unauthorized"
// @Failure 500 {object} apperr.ErrorResponse "Internal Server Error"
// @Router /api/v1/users/me/config [get]
func (h *ConfigHandler) GetUserConfig(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		logger.ErrorContext(c, "无法从上下文中获取 user_id", "handler", "GetUserConfig")
		err := apperr.New(apperr.CodeUnauthenticated, "无效的认证信息") // Use CodeUnauthenticated
		c.AbortWithStatusJSON(apperr.GetHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	// Ensure userID is of the correct type (uint in our service)
	userIDUint, ok := userID.(uint)
	if !ok {
		logger.ErrorContext(c, "上下文中的 user_id 类型不正确", "handler", "GetUserConfig", "type", fmt.Sprintf("%T", userID)) // Use fmt.Sprintf
		err := apperr.New(apperr.CodeInternal, "服务器内部错误")
		c.AbortWithStatusJSON(apperr.GetHTTPStatus(err), gin.H{"error": err.Error()}) // Replace HandleError
		return
	}

	configDTO, err := h.configService.GetUserConfig(c.Request.Context(), userIDUint)
	if err != nil {
		// Service layer should return wrapped apperr errors
		logger.ErrorContext(c, "GetUserConfig service error", "error", err)
		c.AbortWithStatusJSON(apperr.GetHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, configDTO)
}

// UpdateUserConfig godoc
// @Summary Update current user's configuration
// @Description Updates or creates the configuration settings for the currently authenticated user. Allows partial updates. Send empty string "" for api_key to clear it.
// @Tags Config
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param config body dto.UpdateUserConfigDTO true "Configuration settings to update"
// @Success 200 {object} map[string]string "Successfully updated configuration"
// @Failure 400 {object} apperr.ErrorResponse "Bad Request (e.g., invalid JSON)"
// @Failure 401 {object} apperr.ErrorResponse "Unauthorized"
// @Failure 500 {object} apperr.ErrorResponse "Internal Server Error"
// @Router /api/v1/users/me/config [put]
func (h *ConfigHandler) UpdateUserConfig(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		logger.ErrorContext(c, "无法从上下文中获取 user_id", "handler", "UpdateUserConfig")
		err := apperr.New(apperr.CodeUnauthenticated, "无效的认证信息") // Use CodeUnauthenticated
		c.AbortWithStatusJSON(apperr.GetHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	userIDUint, ok := userID.(uint)
	if !ok {
		logger.ErrorContext(c, "上下文中的 user_id 类型不正确", "handler", "UpdateUserConfig", "type", fmt.Sprintf("%T", userID)) // Use fmt.Sprintf
		err := apperr.New(apperr.CodeInternal, "服务器内部错误")
		c.AbortWithStatusJSON(apperr.GetHTTPStatus(err), gin.H{"error": err.Error()}) // Replace HandleError
		return
	}

	var updateDTO dto.UpdateUserConfigDTO
	if err := c.ShouldBindJSON(&updateDTO); err != nil {
		logger.WarnContext(c, "无效的请求体", "handler", "UpdateUserConfig", "error", err)
		bindErr := apperr.Wrap(err, apperr.CodeInvalidArgument, "请求体格式错误") // Use CodeInvalidArgument for binding errors
		c.AbortWithStatusJSON(apperr.GetHTTPStatus(bindErr), gin.H{"error": bindErr.Error()})
		return
	}

	err := h.configService.UpdateUserConfig(c.Request.Context(), userIDUint, &updateDTO)
	if err != nil {
		// Service layer should return wrapped apperr errors
		logger.ErrorContext(c, "UpdateUserConfig service error", "error", err)
		c.AbortWithStatusJSON(apperr.GetHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "配置更新成功"})
}

// End of file
