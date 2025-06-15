package api

import (
	"dreamhub/internal/entity"
	"dreamhub/internal/service"
	"dreamhub/pkg/apperr"
	"dreamhub/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SyncHandler handles sync-related HTTP requests
type SyncHandler struct {
	syncService service.SyncService
	log         *logger.Logger
}

// NewSyncHandler creates a new sync handler
func NewSyncHandler(syncService service.SyncService, log *logger.Logger) *SyncHandler {
	return &SyncHandler{
		syncService: syncService,
		log:         log,
	}
}

// GetSyncStatus handles GET /api/v1/sync/status
func (h *SyncHandler) GetSyncStatus(c *gin.Context) {
	userID, exists := GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	deviceID := c.Query("device_id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_id is required"})
		return
	}

	status, err := h.syncService.GetSyncStatus(c.Request.Context(), userID, deviceID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, status)
}

// PullChanges handles POST /api/v1/sync/pull
func (h *SyncHandler) PullChanges(c *gin.Context) {
	userID, exists := GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req entity.SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	response, err := h.syncService.PullChanges(c.Request.Context(), userID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// PushChanges handles POST /api/v1/sync/push
func (h *SyncHandler) PushChanges(c *gin.Context) {
	userID, exists := GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req entity.SyncPushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	response, err := h.syncService.PushChanges(c.Request.Context(), userID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ResolveConflicts handles POST /api/v1/sync/conflicts/resolve
func (h *SyncHandler) ResolveConflicts(c *gin.Context) {
	userID, exists := GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var conflicts []entity.Conflict
	if err := c.ShouldBindJSON(&conflicts); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err := h.syncService.ResolveConflicts(c.Request.Context(), userID, conflicts)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "conflicts resolved successfully"})
}

func (h *SyncHandler) handleError(c *gin.Context, err error) {
	if appErr, ok := err.(*apperr.AppError); ok {
		switch appErr.Code {
		case apperr.CodeNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": appErr.Message})
		case apperr.CodeInvalidArgument:
			c.JSON(http.StatusBadRequest, gin.H{"error": appErr.Message})
		case apperr.CodeConflict:
			c.JSON(http.StatusConflict, gin.H{"error": appErr.Message})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}