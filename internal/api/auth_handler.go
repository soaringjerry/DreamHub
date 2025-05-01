package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/soaringjerry/dreamhub/internal/service"
	"github.com/soaringjerry/dreamhub/pkg/apperr" // Import apperr
	"github.com/soaringjerry/dreamhub/pkg/logger"
)

// AuthHandler handles authentication related API requests.
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new instance of AuthHandler.
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRoutes registers the authentication routes with the Gin router group.
func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	authGroup := rg.Group("/auth") // Group auth routes under /api/v1/auth
	{
		authGroup.POST("/register", h.register)
		authGroup.POST("/login", h.login)
		// Maybe add a /validate endpoint later if needed for frontend checks
	}
}

// register handles the POST /api/v1/auth/register request.
func (h *AuthHandler) register(c *gin.Context) {
	var payload service.RegisterPayload
	// Bind JSON payload to the struct, handling potential binding errors
	if err := c.ShouldBindJSON(&payload); err != nil {
		logger.WarnContext(c.Request.Context(), "注册请求绑定失败", "error", err)
		// Use apperr.Wrap for validation errors
		appErr := apperr.Wrap(err, apperr.CodeValidation, "无效的注册信息")
		c.Error(appErr) // Pass error to the error handling middleware
		return
	}

	// Call the auth service to register the user
	user, err := h.authService.Register(c.Request.Context(), payload)
	if err != nil {
		// Service layer should return appropriate AppError
		logger.ErrorContext(c.Request.Context(), "注册服务失败", "username", payload.Username, "error", err)
		c.Error(err) // Pass error to the error handling middleware
		return
	}

	// Return the sanitized user information upon successful registration
	c.JSON(http.StatusCreated, user) // 201 Created
}

// login handles the POST /api/v1/auth/login request.
func (h *AuthHandler) login(c *gin.Context) {
	var creds service.LoginCredentials
	// Bind JSON payload
	if err := c.ShouldBindJSON(&creds); err != nil {
		logger.WarnContext(c.Request.Context(), "登录请求绑定失败", "error", err)
		// Use apperr.Wrap for validation errors
		appErr := apperr.Wrap(err, apperr.CodeValidation, "无效的登录凭证")
		c.Error(appErr)
		return
	}

	// Call the auth service to log in the user
	loginResponse, err := h.authService.Login(c.Request.Context(), creds)
	if err != nil {
		// Service layer should return appropriate AppError (e.g., Unauthenticated, Internal)
		logger.WarnContext(c.Request.Context(), "登录服务失败", "username", creds.Username, "error", err)
		c.Error(err)
		return
	}

	// Return the JWT and user info upon successful login
	c.JSON(http.StatusOK, loginResponse) // 200 OK
}
