package api

import (
	"errors"
	"fmt" // Add fmt for error messages
	"net/http"

	"github.com/gin-gonic/gin"
	// "github.com/google/uuid" // No longer asserting to UUID here
	"github.com/soaringjerry/dreamhub/internal/entity"
	"github.com/soaringjerry/dreamhub/internal/repository" // Use base repository for errors like ErrDuplicateKey, ErrNotFound if defined there
	"github.com/soaringjerry/dreamhub/internal/service"
	"github.com/soaringjerry/dreamhub/pkg/apperr" // Import apperr
	"github.com/soaringjerry/dreamhub/pkg/logger" // Import logger
)

// MemoryHandler handles API requests related to structured memories.
type MemoryHandler struct {
	service service.StructuredMemoryService
}

// NewMemoryHandler creates a new MemoryHandler.
func NewMemoryHandler(s service.StructuredMemoryService) *MemoryHandler {
	return &MemoryHandler{service: s}
}

// CreateMemoryRequest defines the structure for the create memory request body.
type CreateMemoryRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

// UpdateMemoryRequest defines the structure for the update memory request body.
type UpdateMemoryRequest struct {
	Value string `json:"value" binding:"required"`
}

// CreateMemory godoc
// @Summary Create a structured memory entry
// @Description Adds a new key-value pair to the user's structured memory.
// @Tags Memory
// @Accept json
// @Produce json
// @Param memory body CreateMemoryRequest true "Memory Key and Value"
// @Success 201 {object} entity.StructuredMemory
// @Failure 400 {object} gin.H{"error": "string"} "Invalid input"
// @Failure 401 {object} gin.H{"error": "string"} "Unauthorized"
// @Failure 409 {object} gin.H{"error": "string"} "Duplicate key"
// @Failure 500 {object} gin.H{"error": "string"} "Internal server error"
// @Security BearerAuth
// @Router /api/v1/memory/structured [post]
func (h *MemoryHandler) CreateMemory(c *gin.Context) {
	// Get userID string from context using the correct key
	userIDVal, exists := c.Get(authorizationPayloadKey)
	if !exists {
		logger.ErrorContext(c, "无法从上下文中获取 user_id", "handler", "CreateMemory")
		err := apperr.New(apperr.CodeUnauthenticated, "无效的认证信息")
		c.AbortWithStatusJSON(apperr.GetHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok {
		logger.ErrorContext(c, "上下文中的 user_id 类型不正确", "handler", "CreateMemory", "expected", "string", "actual", fmt.Sprintf("%T", userIDVal))
		err := apperr.New(apperr.CodeInternal, "服务器内部错误 (用户标识类型错误)")
		c.AbortWithStatusJSON(apperr.GetHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	var req CreateMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WarnContext(c, "无效的请求体", "handler", "CreateMemory", "error", err)
		bindErr := apperr.Wrap(err, apperr.CodeInvalidArgument, "请求体格式错误")
		c.AbortWithStatusJSON(apperr.GetHTTPStatus(bindErr), gin.H{"error": bindErr.Error()})
		return
	}

	// Pass string userID to the service (assuming service accepts string)
	memory, err := h.service.CreateMemory(c.Request.Context(), userIDStr, req.Key, req.Value)
	if err != nil {
		logger.ErrorContext(c, "CreateMemory service error", "user_id", userIDStr, "key", req.Key, "error", err)
		// Handle specific errors (assuming repository defines ErrDuplicateKey)
		if errors.Is(err, repository.ErrDuplicateKey) { // Use repository error
			dupErr := apperr.Wrap(err, apperr.CodeConflict, "内存键已存在")
			c.AbortWithStatusJSON(apperr.GetHTTPStatus(dupErr), gin.H{"error": dupErr.Error()})
		} else { // Handle other potential AppErrors from service or generic internal error
			c.AbortWithStatusJSON(apperr.GetHTTPStatus(err), gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, memory)
}

// GetUserMemories godoc
// @Summary Get all structured memory entries for the current user
// @Description Retrieves all key-value pairs stored for the logged-in user.
// @Tags Memory
// @Produce json
// @Success 200 {array} entity.StructuredMemory
// @Failure 401 {object} gin.H{"error": "string"} "Unauthorized"
// @Failure 500 {object} gin.H{"error": "string"} "Internal server error"
// @Security BearerAuth
// @Router /api/v1/memory/structured [get]
func (h *MemoryHandler) GetUserMemories(c *gin.Context) {
	// Get userID string from context using the correct key
	userIDVal, exists := c.Get(authorizationPayloadKey)
	if !exists {
		logger.ErrorContext(c, "无法从上下文中获取 user_id", "handler", "GetUserMemories")
		err := apperr.New(apperr.CodeUnauthenticated, "无效的认证信息")
		c.AbortWithStatusJSON(apperr.GetHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok {
		logger.ErrorContext(c, "上下文中的 user_id 类型不正确", "handler", "GetUserMemories", "expected", "string", "actual", fmt.Sprintf("%T", userIDVal))
		err := apperr.New(apperr.CodeInternal, "服务器内部错误 (用户标识类型错误)")
		c.AbortWithStatusJSON(apperr.GetHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	// Pass string userID to the service (assuming service accepts string)
	memories, err := h.service.GetUserMemories(c.Request.Context(), userIDStr)
	if err != nil {
		logger.ErrorContext(c, "GetUserMemories service error", "user_id", userIDStr, "error", err)
		c.AbortWithStatusJSON(apperr.GetHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	// Return empty array instead of null if no memories found
	if memories == nil {
		memories = []*entity.StructuredMemory{} // Use pointer slice type
	}

	c.JSON(http.StatusOK, memories)
}

// GetMemoryByKey godoc
// @Summary Get a specific structured memory entry by key
// @Description Retrieves the value associated with a specific key for the logged-in user.
// @Tags Memory
// @Produce json
// @Param key path string true "Memory Key"
// @Success 200 {object} entity.StructuredMemory
// @Failure 400 {object} gin.H{"error": "string"} "Invalid key"
// @Failure 401 {object} gin.H{"error": "string"} "Unauthorized"
// @Failure 404 {object} gin.H{"error": "string"} "Memory not found"
// @Failure 500 {object} gin.H{"error": "string"} "Internal server error"
// @Security BearerAuth
// @Router /api/v1/memory/structured/{key} [get]
func (h *MemoryHandler) GetMemoryByKey(c *gin.Context) {
	// Get userID string from context using the correct key
	userIDVal, exists := c.Get(authorizationPayloadKey)
	if !exists {
		logger.ErrorContext(c, "无法从上下文中获取 user_id", "handler", "GetMemoryByKey")
		err := apperr.New(apperr.CodeUnauthenticated, "无效的认证信息")
		c.AbortWithStatusJSON(apperr.GetHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok {
		logger.ErrorContext(c, "上下文中的 user_id 类型不正确", "handler", "GetMemoryByKey", "expected", "string", "actual", fmt.Sprintf("%T", userIDVal))
		err := apperr.New(apperr.CodeInternal, "服务器内部错误 (用户标识类型错误)")
		c.AbortWithStatusJSON(apperr.GetHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	key := c.Param("key")
	if key == "" {
		err := apperr.New(apperr.CodeInvalidArgument, "内存键参数是必需的")
		c.AbortWithStatusJSON(apperr.GetHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	// Pass string userID to the service (assuming service accepts string)
	memory, err := h.service.GetMemoryByKey(c.Request.Context(), userIDStr, key)
	if err != nil {
		logger.ErrorContext(c, "GetMemoryByKey service error", "user_id", userIDStr, "key", key, "error", err)
		// Handle specific errors (assuming repository defines ErrNotFound)
		if errors.Is(err, repository.ErrNotFound) { // Use repository error
			nfErr := apperr.Wrap(err, apperr.CodeNotFound, "找不到指定键的内存")
			c.AbortWithStatusJSON(apperr.GetHTTPStatus(nfErr), gin.H{"error": nfErr.Error()})
		} else { // Handle other potential AppErrors from service or generic internal error
			c.AbortWithStatusJSON(apperr.GetHTTPStatus(err), gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, memory)
}

// UpdateMemory godoc
// @Summary Update a structured memory entry by key
// @Description Modifies the value associated with a specific key for the logged-in user.
// @Tags Memory
// @Accept json
// @Produce json
// @Param key path string true "Memory Key"
// @Param memory body UpdateMemoryRequest true "New Memory Value"
// @Success 200 {object} entity.StructuredMemory
// @Failure 400 {object} gin.H{"error": "string"} "Invalid input"
// @Failure 401 {object} gin.H{"error": "string"} "Unauthorized"
// @Failure 404 {object} gin.H{"error": "string"} "Memory not found"
// @Failure 500 {object} gin.H{"error": "string"} "Internal server error"
// @Security BearerAuth
// @Router /api/v1/memory/structured/{key} [put]
func (h *MemoryHandler) UpdateMemory(c *gin.Context) {
	// Get userID string from context using the correct key
	userIDVal, exists := c.Get(authorizationPayloadKey)
	if !exists {
		logger.ErrorContext(c, "无法从上下文中获取 user_id", "handler", "UpdateMemory")
		err := apperr.New(apperr.CodeUnauthenticated, "无效的认证信息")
		c.AbortWithStatusJSON(apperr.GetHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok {
		logger.ErrorContext(c, "上下文中的 user_id 类型不正确", "handler", "UpdateMemory", "expected", "string", "actual", fmt.Sprintf("%T", userIDVal))
		err := apperr.New(apperr.CodeInternal, "服务器内部错误 (用户标识类型错误)")
		c.AbortWithStatusJSON(apperr.GetHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	key := c.Param("key")
	if key == "" {
		err := apperr.New(apperr.CodeInvalidArgument, "内存键参数是必需的")
		c.AbortWithStatusJSON(apperr.GetHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	var req UpdateMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WarnContext(c, "无效的请求体", "handler", "UpdateMemory", "error", err)
		bindErr := apperr.Wrap(err, apperr.CodeInvalidArgument, "请求体格式错误")
		c.AbortWithStatusJSON(apperr.GetHTTPStatus(bindErr), gin.H{"error": bindErr.Error()})
		return
	}

	// Pass string userID to the service (assuming service accepts string)
	memory, err := h.service.UpdateMemory(c.Request.Context(), userIDStr, key, req.Value)
	if err != nil {
		logger.ErrorContext(c, "UpdateMemory service error", "user_id", userIDStr, "key", key, "error", err)
		// Handle specific errors (assuming repository defines ErrNotFound)
		if errors.Is(err, repository.ErrNotFound) { // Use repository error
			nfErr := apperr.Wrap(err, apperr.CodeNotFound, "找不到指定键的内存")
			c.AbortWithStatusJSON(apperr.GetHTTPStatus(nfErr), gin.H{"error": nfErr.Error()})
		} else { // Handle other potential AppErrors from service or generic internal error
			c.AbortWithStatusJSON(apperr.GetHTTPStatus(err), gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, memory)
}

// DeleteMemory godoc
// @Summary Delete a structured memory entry by key
// @Description Removes a key-value pair from the user's structured memory.
// @Tags Memory
// @Param key path string true "Memory Key"
// @Success 204 "Successfully deleted"
// @Failure 400 {object} gin.H{"error": "string"} "Invalid key"
// @Failure 401 {object} gin.H{"error": "string"} "Unauthorized"
// @Failure 404 {object} gin.H{"error": "string"} "Memory not found"
// @Failure 500 {object} gin.H{"error": "string"} "Internal server error"
// @Security BearerAuth
// @Router /api/v1/memory/structured/{key} [delete]
func (h *MemoryHandler) DeleteMemory(c *gin.Context) {
	// Get userID string from context using the correct key
	userIDVal, exists := c.Get(authorizationPayloadKey)
	if !exists {
		logger.ErrorContext(c, "无法从上下文中获取 user_id", "handler", "DeleteMemory")
		err := apperr.New(apperr.CodeUnauthenticated, "无效的认证信息")
		c.AbortWithStatusJSON(apperr.GetHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok {
		logger.ErrorContext(c, "上下文中的 user_id 类型不正确", "handler", "DeleteMemory", "expected", "string", "actual", fmt.Sprintf("%T", userIDVal))
		err := apperr.New(apperr.CodeInternal, "服务器内部错误 (用户标识类型错误)")
		c.AbortWithStatusJSON(apperr.GetHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	key := c.Param("key")
	if key == "" {
		err := apperr.New(apperr.CodeInvalidArgument, "内存键参数是必需的")
		c.AbortWithStatusJSON(apperr.GetHTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	// Pass string userID to the service (assuming service accepts string)
	err := h.service.DeleteMemory(c.Request.Context(), userIDStr, key)
	if err != nil {
		logger.ErrorContext(c, "DeleteMemory service error", "user_id", userIDStr, "key", key, "error", err)
		// Handle specific errors (assuming repository defines ErrNotFound)
		if errors.Is(err, repository.ErrNotFound) { // Use repository error
			nfErr := apperr.Wrap(err, apperr.CodeNotFound, "找不到指定键的内存")
			c.AbortWithStatusJSON(apperr.GetHTTPStatus(nfErr), gin.H{"error": nfErr.Error()})
		} else { // Handle other potential AppErrors from service or generic internal error
			c.AbortWithStatusJSON(apperr.GetHTTPStatus(err), gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
