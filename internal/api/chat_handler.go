package api

import (
	"fmt" // Import fmt for Sscan
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/soaringjerry/dreamhub/internal/service"
	"github.com/soaringjerry/dreamhub/pkg/apperr" // 需要 apperr 来处理错误
	"github.com/soaringjerry/dreamhub/pkg/logger"
	// "github.com/gorilla/websocket" // 后续实现 WebSocket 时需要
)

// ChatHandler 负责处理与聊天相关的 API 请求。
type ChatHandler struct {
	chatService service.ChatService
	// upgrader websocket.Upgrader // 用于 WebSocket
}

// NewChatHandler 创建一个新的 ChatHandler 实例。
func NewChatHandler(cs service.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: cs,
		// upgrader: websocket.Upgrader{ // 初始化 WebSocket 升级器
		// 	ReadBufferSize:  1024,
		// 	WriteBufferSize: 1024,
		// 	CheckOrigin: func(r *http.Request) bool {
		// 		// TODO: 实现更严格的来源检查
		// 		return true
		// 	},
		// },
	}
}

// RegisterRoutes 将聊天相关的路由注册到 Gin 引擎。
func (h *ChatHandler) RegisterRoutes(router *gin.RouterGroup) {
	chatGroup := router.Group("/chat")
	{
		chatGroup.POST("", h.handlePostChat) // 处理 POST /api/v1/chat
		// chatGroup.GET("/ws", h.handleChatWebSocket) // 处理 GET /api/v1/chat/ws (未来)
		chatGroup.GET("/:conversation_id/messages", h.handleGetMessages) // 处理 GET /api/v1/chat/{conversation_id}/messages
	}
}

// ChatRequest 定义了聊天请求的 JSON 结构体。
type ChatRequest struct {
	// UserID is now obtained from the authentication context
	ConversationID string `json:"conversation_id,omitempty"` // 可选，如果为空则开始新对话
	Message        string `json:"message" binding:"required"`
	ModelName      string `json:"model_name,omitempty"` // 新增：可选的模型名称
}

// ChatResponse 定义了聊天响应的 JSON 结构体。
type ChatResponse struct {
	ConversationID string `json:"conversation_id"`
	Reply          string `json:"reply"`
}

// handlePostChat 处理非流式的聊天请求。
func (h *ChatHandler) handlePostChat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WarnContext(c.Request.Context(), "无效的聊天请求体", "error", err)
		// 使用 apperr 包装错误并返回标准响应
		// Use apperr.Wrap function instead of chaining
		appErr := apperr.Wrap(err, apperr.CodeInvalidArgument, "请求体无效")
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}

	// Get UserID from the context set by the auth middleware
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		// This should ideally not happen if middleware is applied correctly
		logger.ErrorContext(c.Request.Context(), "无法从上下文中获取用户 ID (ChatHandler)")
		appErr := apperr.New(apperr.CodeInternal, "无法处理请求，缺少用户信息")
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}
	// Log the userID to mark it as used (or pass it to service if needed)
	logger.DebugContext(c.Request.Context(), "处理聊天请求", "user_id", userID)
	// Use the original context, user ID is implicitly available via logger/service calls if needed
	ctx := c.Request.Context()

	// conversationID is now string, no need to parse or use uuid.Nil
	conversationID := req.ConversationID // Use the string directly, empty string means new conversation

	// 调用 ChatService 处理消息，传入 ModelName
	// Pass string conversationID directly
	reply, newConvID, err := h.chatService.HandleChatMessage(ctx, conversationID, req.Message, req.ModelName)
	if err != nil {
		// HandleChatMessage 内部应该已经记录了日志并包装了错误
		// 直接使用返回的 apperr
		appErr, ok := err.(*apperr.AppError)
		if !ok {
			// 如果不是 AppError，包装为内部错误
			appErr = apperr.Wrap(err, apperr.CodeInternal, "处理聊天消息时发生未知错误")
		}
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}

	// 返回成功响应
	// newConvID is now string, no need for .String()
	c.JSON(http.StatusOK, ChatResponse{
		ConversationID: newConvID,
		Reply:          reply,
	})
}

// handleGetMessages 处理获取对话消息列表的请求。
func (h *ChatHandler) handleGetMessages(c *gin.Context) {
	conversationIDStr := c.Param("conversation_id")
	// conversationID is now string, no need to parse
	if conversationIDStr == "" { // Basic validation
		logger.WarnContext(c.Request.Context(), "缺少 conversation_id 路径参数")
		appErr := apperr.New(apperr.CodeInvalidArgument, "缺少 conversation_id")
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}

	// 获取分页参数 (可选)
	limit := c.DefaultQuery("limit", "50")  // 默认获取 50 条
	offset := c.DefaultQuery("offset", "0") // 默认从头开始

	var limitInt, offsetInt int
	// TODO: 更健壮的参数转换和验证
	_, errLimit := fmt.Sscan(limit, &limitInt)
	_, errOffset := fmt.Sscan(offset, &offsetInt)
	if errLimit != nil || limitInt <= 0 {
		limitInt = 50 // Use default if conversion fails or value is invalid
	}
	if errOffset != nil || offsetInt < 0 {
		offsetInt = 0 // Use default if conversion fails or value is invalid
	}

	// Get UserID from the context set by the auth middleware
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		logger.ErrorContext(c.Request.Context(), "无法从上下文中获取用户 ID (GetMessages)")
		appErr := apperr.New(apperr.CodeInternal, "无法处理请求，缺少用户信息")
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}
	// Log the userID to mark it as used
	logger.DebugContext(c.Request.Context(), "获取消息列表", "user_id", userID, "conversation_id", conversationIDStr)
	// Use the original context
	ctx := c.Request.Context()
	// Note: We might need to pass userID explicitly to GetConversationMessages if it needs it for authorization/filtering
	// For now, assume the service layer handles context implicitly or doesn't need explicit userID for this call.

	// Pass string conversationIDStr directly
	messages, err := h.chatService.GetConversationMessages(ctx, conversationIDStr, limitInt, offsetInt) // Pass userID if needed: h.chatService.GetConversationMessages(ctx, userID, conversationIDStr, limitInt, offsetInt)
	if err != nil {
		appErr, ok := err.(*apperr.AppError)
		if !ok {
			appErr = apperr.Wrap(err, apperr.CodeInternal, "获取对话消息时发生未知错误")
		}
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}

	// 返回消息列表 (entity.Message 结构体已经有 json tag)
	c.JSON(http.StatusOK, messages)
}

// GetUserConversationsHandler 处理获取用户所有对话列表的请求。
func (h *ChatHandler) GetUserConversationsHandler(c *gin.Context) {
	// 从认证中间件设置的上下文中获取 UserID
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		logger.ErrorContext(c.Request.Context(), "无法从上下文中获取用户 ID (GetUserConversations)")
		appErr := apperr.New(apperr.CodeInternal, "无法处理请求，缺少用户信息")
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}
	logger.DebugContext(c.Request.Context(), "获取用户对话列表", "user_id", userID)
	ctx := c.Request.Context()

	// 调用 ChatService 获取对话列表
	conversations, err := h.chatService.GetUserConversations(ctx, userID)
	if err != nil {
		// GetUserConversations 内部应该已经记录了日志并包装了错误
		appErr, ok := err.(*apperr.AppError)
		if !ok {
			// 如果不是 AppError，包装为内部错误
			appErr = apperr.Wrap(err, apperr.CodeInternal, "获取用户对话列表时发生未知错误")
		}
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}

	// 返回对话列表 (entity.Conversation 结构体已经有 json tag)
	// 如果列表为空，也会返回一个空的 JSON 数组 `[]`
	c.JSON(http.StatusOK, conversations)
}

// handleChatWebSocket 处理 WebSocket 连接请求 (未来实现)。
// func (h *ChatHandler) handleChatWebSocket(c *gin.Context) {
//  // TODO: 实现 WebSocket 逻辑
//  // 1. 升级 HTTP 连接到 WebSocket
//  // 2. 从认证信息获取 user_id
//  // 3. 启动 goroutine 读取客户端消息
//  // 4. 调用 HandleStreamChatMessage 处理消息
//  // 5. 将流式回复写回 WebSocket 连接
//  // 6. 处理连接关闭和错误
// 	c.JSON(http.StatusNotImplemented, gin.H{"message": "WebSocket not implemented yet"})
// }
