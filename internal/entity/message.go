package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// SenderRole 定义了消息发送者的角色。
type SenderRole string

const (
	SenderRoleUser   SenderRole = "user"
	SenderRoleAI     SenderRole = "ai"
	SenderRoleSystem SenderRole = "system" // Add System role
)

// Message 代表一条对话消息。
// 这对应于数据库中的 conversation_history 表。
type Message struct {
	ID             uuid.UUID       `json:"id"`              // 消息的唯一 ID
	ConversationID uuid.UUID       `json:"conversation_id"` // 所属对话的 ID
	UserID         string          `json:"user_id"`         // 发送消息的用户 ID (用于数据隔离)
	SenderRole     SenderRole      `json:"sender_role"`     // 发送者角色 ('user' 或 'ai')
	Content        string          `json:"content"`         // 消息内容
	Timestamp      time.Time       `json:"timestamp"`       // 消息时间戳
	Metadata       json.RawMessage `json:"metadata"`        // 可选的元数据 (JSONB)
	// 可以考虑添加其他字段，例如：
	// Tokens         int             `json:"tokens"`          // 消息消耗的 token 数 (如果需要统计)
	// ModelName      string          `json:"model_name"`      // 生成 AI 回复的模型名称
}

// NewMessage 创建一个新的 Message 实例。
func NewMessage(conversationID uuid.UUID, userID string, role SenderRole, content string) *Message {
	return &Message{
		ID:             uuid.New(), // 自动生成新的 UUID
		ConversationID: conversationID,
		UserID:         userID,
		SenderRole:     role,
		Content:        content,
		Timestamp:      time.Now(), // 设置当前时间
		Metadata:       nil,        // 默认为空
	}
}

// SetMetadata 设置消息的元数据。
func (m *Message) SetMetadata(data map[string]interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err // 或者返回一个 apperr.Wrap 错误
	}
	m.Metadata = jsonData
	return nil
}

// GetMetadata 将元数据解析到提供的 map 中。
func (m *Message) GetMetadata(target map[string]interface{}) error {
	if m.Metadata == nil {
		return nil // 没有元数据
	}
	return json.Unmarshal(m.Metadata, &target)
}
