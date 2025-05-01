package api

import (
	"errors" // Import errors package
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/soaringjerry/dreamhub/internal/service"
	"github.com/soaringjerry/dreamhub/pkg/apperr" // Import apperr
	"github.com/soaringjerry/dreamhub/pkg/logger"
)

const (
	authorizationHeaderKey  = "Authorization"
	authorizationTypeBearer = "Bearer"
	authorizationPayloadKey = "authorization_payload_user_id" // Key to store user ID in context
)

// AuthMiddleware provides Gin middleware for authentication.
type AuthMiddleware struct {
	authService service.AuthService
}

// NewAuthMiddleware creates a new instance of AuthMiddleware.
func NewAuthMiddleware(authService service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// Authenticate is the Gin middleware function to enforce authentication.
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(authorizationHeaderKey)
		if len(authHeader) == 0 {
			err := apperr.New(apperr.CodeUnauthenticated, "未提供认证头")
			logger.WarnContext(c.Request.Context(), "认证失败: "+err.Message)
			c.AbortWithStatusJSON(err.HTTPStatus, gin.H{"error": err})
			return
		}

		fields := strings.Fields(authHeader)
		if len(fields) < 2 {
			err := apperr.New(apperr.CodeUnauthenticated, "认证头格式无效")
			logger.WarnContext(c.Request.Context(), "认证失败: "+err.Message)
			c.AbortWithStatusJSON(err.HTTPStatus, gin.H{"error": err})
			return
		}

		authType := strings.ToLower(fields[0])
		if authType != strings.ToLower(authorizationTypeBearer) {
			err := apperr.New(apperr.CodeUnauthenticated, "不支持的认证类型: "+fields[0])
			logger.WarnContext(c.Request.Context(), "认证失败: "+err.Message)
			c.AbortWithStatusJSON(err.HTTPStatus, gin.H{"error": err})
			return
		}

		accessToken := fields[1]
		userID, err := m.authService.ValidateToken(c.Request.Context(), accessToken)
		if err != nil {
			// ValidateToken should return an appropriate AppError
			logger.WarnContext(c.Request.Context(), "令牌验证失败", "error", err)
			// Try to extract AppError details for response
			var appErr *apperr.AppError
			if errors.As(err, &appErr) {
				c.AbortWithStatusJSON(appErr.HTTPStatus, gin.H{"error": appErr})
			} else {
				// Fallback for unexpected errors from ValidateToken
				genericErr := apperr.New(apperr.CodeUnauthenticated, "无效或过期的认证令牌")
				c.AbortWithStatusJSON(genericErr.HTTPStatus, gin.H{"error": genericErr})
			}
			return
		}

		// Set the user ID in the context for downstream handlers
		c.Set(authorizationPayloadKey, userID)
		logger.DebugContext(c.Request.Context(), "认证成功", "user_id", userID)

		// Proceed to the next handler
		c.Next()
	}
}

// GetUserIDFromContext retrieves the authenticated user ID from the Gin context.
// It should only be called in handlers protected by the Authenticate middleware.
// Returns the user ID string and true if found, otherwise empty string and false.
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	userIDVal, exists := c.Get(authorizationPayloadKey)
	if !exists {
		return "", false
	}
	userID, ok := userIDVal.(string)
	return userID, ok
}
