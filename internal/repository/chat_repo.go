package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/soaringjerry/dreamhub/internal/entity"
)

// ChatRepository 定义了与对话历史数据存储交互的方法。
type ChatRepository interface {
	// SaveMessage 保存一条新的对话消息。
	// 需要确保实现中处理了 user_id 以进行数据隔离。
	SaveMessage(ctx context.Context, message *entity.Message) error

	// GetMessagesByConversationID 获取指定对话的所有消息。
	// 需要确保实现中根据 ctx 中的 user_id 进行了过滤。
	// 可以添加分页、排序等参数。
	GetMessagesByConversationID(ctx context.Context, conversationID uuid.UUID, limit int, offset int) ([]*entity.Message, error)

	// GetConversationHistory 获取指定对话的最近 N 条消息 (用于 RAG 或 LLM 上下文)。
	// 需要确保实现中根据 ctx 中的 user_id 进行了过滤。
	GetConversationHistory(ctx context.Context, conversationID uuid.UUID, lastN int) ([]*entity.Message, error)

	// TODO: 可能需要添加其他方法，例如：
	// DeleteMessagesByConversationID(ctx context.Context, conversationID uuid.UUID) error
	// GetConversationsByUser(ctx context.Context, userID string, limit int, offset int) ([]*entity.ConversationSummary, error) // 需要 ConversationSummary 实体
}
