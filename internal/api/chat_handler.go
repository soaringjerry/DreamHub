package api

import (
	"fmt" // Import fmt for Sscan
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/soaringjerry/dreamhub/internal/service"
	"github.com/soaringjerry/dreamhub/pkg/apperr"  // 需要 apperr 来处理错误
	"github.com/soaringjerry/dreamhub/pkg/ctxutil" // 需要 ctxutil 来设置 user_id
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
	UserID         string `json:"user_id" binding:"required"` // 临时从请求体获取 user_id
	ConversationID string `json:"conversation_id"`            // 可选，如果为空则开始新对话
	Message        string `json:"message" binding:"required"`
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
		appErr := apperr.New(apperr.CodeInvalidArgument, "请求体无效").WithDetails(err.Error())
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}

	// TODO: 从认证中间件获取 user_id，而不是从请求体
	ctx := ctxutil.WithUserID(c.Request.Context(), req.UserID)

	var conversationID uuid.UUID
	var err error
	if req.ConversationID != "" {
		conversationID, err = uuid.Parse(req.ConversationID)
		if err != nil {
			logger.WarnContext(ctx, "无效的 conversation_id 格式", "error", err, "conversation_id", req.ConversationID)
			appErr := apperr.New(apperr.CodeInvalidArgument, "无效的 conversation_id 格式")
			c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
			return
		}
	} else {
		conversationID = uuid.Nil // 表示新对话
	}

	// 调用 ChatService 处理消息
	reply, newConvID, err := h.chatService.HandleChatMessage(ctx, conversationID, req.Message)
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
	c.JSON(http.StatusOK, ChatResponse{
		ConversationID: newConvID.String(),
		Reply:          reply,
	})
}

// handleGetMessages 处理获取对话消息列表的请求。
func (h *ChatHandler) handleGetMessages(c *gin.Context) {
	conversationIDStr := c.Param("conversation_id")
	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		logger.WarnContext(c.Request.Context(), "无效的 conversation_id 格式 (URL)", "error", err, "conversation_id", conversationIDStr)
		appErr := apperr.New(apperr.CodeInvalidArgument, "无效的 conversation_id 格式")
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

	// TODO: 从认证中间件获取 user_id
	// 暂时硬编码一个 user_id 用于测试，或者需要修改 GetConversationMessages 不强制 user_id
	// ctx := ctxutil.WithUserID(c.Request.Context(), "temp_user_for_get_messages")
	ctx := c.Request.Context() // 假设 user_id 已在中间件中设置

	messages, err := h.chatService.GetConversationMessages(ctx, conversationID, limitInt, offsetInt)
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
