package entity

// ChatMessage 定义用户发送的聊天消息结构体
type ChatMessage struct {
	ConversationID string `json:"conversation_id"`            // 对话 ID (可选，为空则新建)
	UserID         string `json:"user_id" binding:"required"` // 用户 ID (临时添加, 设为必填)
	Message        string `json:"message" binding:"required"` // 用户消息
}

// ChatResponse 定义返回给客户端的结构体
type ChatResponse struct {
	ConversationID string `json:"conversation_id"` // 返回当前或新的对话 ID
	Reply          string `json:"reply"`           // AI 回复
}
