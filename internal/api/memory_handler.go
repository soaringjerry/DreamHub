package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/soaringjerry/dreamhub/internal/entity"
	postgres "github.com/soaringjerry/dreamhub/internal/repository/postgres" // Add alias and remove unused import
	"github.com/soaringjerry/dreamhub/internal/service"
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
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}
	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID format in context"})
		return
	}

	var req CreateMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	memory, err := h.service.CreateMemory(c.Request.Context(), userIDUUID, req.Key, req.Value)
	if err != nil {
		// Check against the specific error defined in the postgres implementation
		if errors.Is(err, postgres.ErrDuplicateKey) {
			c.JSON(http.StatusConflict, gin.H{"error": "Memory key already exists for this user"})
		} else if err.Error() == "memory key cannot be empty" || err.Error() == "memory value cannot be empty" { // Service level validation errors
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create memory: " + err.Error()})
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
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}
	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID format in context"})
		return
	}

	memories, err := h.service.GetUserMemories(c.Request.Context(), userIDUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve memories: " + err.Error()})
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
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}
	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID format in context"})
		return
	}

	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Memory key parameter is required"})
		return
	}

	memory, err := h.service.GetMemoryByKey(c.Request.Context(), userIDUUID, key)
	if err != nil {
		// Check against the specific error defined in the postgres implementation
		if errors.Is(err, postgres.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Memory not found for the given key"})
		} else if err.Error() == "memory key cannot be empty" { // Service level validation error
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve memory: " + err.Error()})
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
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}
	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID format in context"})
		return
	}

	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Memory key parameter is required"})
		return
	}

	var req UpdateMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	memory, err := h.service.UpdateMemory(c.Request.Context(), userIDUUID, key, req.Value)
	if err != nil {
		// Check against the specific error defined in the postgres implementation
		if errors.Is(err, postgres.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Memory not found for the given key"})
		} else if err.Error() == "memory key cannot be empty" || err.Error() == "memory value cannot be empty" { // Service level validation errors
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update memory: " + err.Error()})
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
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}
	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID format in context"})
		return
	}

	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Memory key parameter is required"})
		return
	}

	err := h.service.DeleteMemory(c.Request.Context(), userIDUUID, key)
	if err != nil {
		// Check against the specific error defined in the postgres implementation
		if errors.Is(err, postgres.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Memory not found for the given key"})
		} else if err.Error() == "memory key cannot be empty" { // Service level validation error
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete memory: " + err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
