package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/soaringjerry/dreamhub/internal/entity"
	"github.com/soaringjerry/dreamhub/internal/service"
	"github.com/soaringjerry/dreamhub/pkg/apperr"
	"github.com/soaringjerry/dreamhub/pkg/logger"
)

// ChatHandler handles chat requests.
type ChatHandler struct {
	chatService service.ChatService
}

// NewChatHandler creates a new ChatHandler.
func NewChatHandler(cs service.ChatService) *ChatHandler {
	if cs == nil {
		panic("ChatService cannot be nil for ChatHandler")
	}
	return &ChatHandler{
		chatService: cs,
	}
}

// HandleChat is the Gin handler function for POST /chat.
func (h *ChatHandler) HandleChat(c *gin.Context) {
	var req entity.ChatMessage
	// 1. Bind JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("ChatHandler: Invalid request format", "error", err)
		errResp := apperr.Wrap(apperr.ValidationError, "Invalid request format", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": errResp.Message, "code": errResp.Code})
		return
	}

	// Basic validation (UserID is required by ChatMessage struct tag)
	// if req.UserID == "" { ... } // Already handled by ShouldBindJSON + binding tag

	// TODO: Get UserID from authenticated context instead of request body later.
	userID := req.UserID
	logger.Info("ChatHandler: Received chat message", "userID", userID, "conversationID", req.ConversationID)

	// 2. Call ChatService
	// Pass Gin context for potential tracing, cancellation, etc.
	chatResponse, err := h.chatService.HandleChatMessage(c.Request.Context(), userID, req.ConversationID, req.Message)
	if err != nil {
		logger.Error("ChatHandler: ChatService failed", "userID", userID, "conversationID", req.ConversationID, "error", err)

		// Handle specific AppError types
		var appErr *apperr.AppError
		if errors.As(err, &appErr) {
			statusCode := http.StatusInternalServerError // Default to 500
			switch appErr.Code {
			case apperr.ValidationError:
				statusCode = http.StatusBadRequest
			case apperr.NotFoundError:
				statusCode = http.StatusNotFound
			case apperr.LLMAPIError:
				// Could be 500 or maybe 503 Service Unavailable?
				statusCode = http.StatusInternalServerError
				// Add other specific cases as needed
			}
			c.JSON(statusCode, gin.H{"error": appErr.Message, "code": appErr.Code})
		} else {
			// Handle unexpected errors
			c.JSON(http.StatusInternalServerError, gin.H{"error": "An unexpected error occurred while processing your message", "code": apperr.UnknownError})
		}
		return
	}

	// 3. Return success response
	logger.Info("ChatHandler: Sending chat reply", "userID", userID, "conversationID", chatResponse.ConversationID)
	c.JSON(http.StatusOK, chatResponse)
}
