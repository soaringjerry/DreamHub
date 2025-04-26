package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/soaringjerry/dreamhub/internal/entity"
)

// ChatService 定义了处理聊天交互的业务逻辑接口。
type ChatService interface {
	// HandleChatMessage 处理传入的聊天消息。
	// 如果 conversationID 为零值 UUID，则表示开始新对话。
	// 它会检索上下文（历史记录，未来可能包括 RAG），调用 LLM，
	// 保存用户消息和 AI 回复，并返回 AI 的回复和对话 ID。
	// 需要从 ctx 中获取 user_id。
	HandleChatMessage(ctx context.Context, conversationID uuid.UUID, message string) (reply string, newConversationID uuid.UUID, err error)

	// HandleStreamChatMessage 处理流式聊天消息 (用于 WebSocket)。
	// 实现逻辑与 HandleChatMessage 类似，但通过 channel 流式返回 AI 回复块。
	// 需要从 ctx 中获取 user_id。
	HandleStreamChatMessage(ctx context.Context, conversationID uuid.UUID, message string, streamCh chan<- string) (newConversationID uuid.UUID, err error)

	// GetConversationMessages 获取指定对话的消息列表（带分页）。
	// 需要从 ctx 中获取 user_id。
	GetConversationMessages(ctx context.Context, conversationID uuid.UUID, limit int, offset int) ([]*entity.Message, error)

	// TODO: 可能需要添加其他方法，例如：
	// ListConversations(ctx context.Context, limit int, offset int) ([]*entity.ConversationSummary, error)
	// DeleteConversation(ctx context.Context, conversationID uuid.UUID) error
}

// LLMProvider 定义了与大型语言模型交互的接口。
// 这允许我们将具体的 LLM 实现（如 OpenAI, Anthropic 等）解耦。
type LLMProvider interface {
	// GenerateContent 根据提供的消息历史生成回复。
	GenerateContent(ctx context.Context, messages []*entity.Message) (string, error)
	// GenerateContentStream 根据提供的消息历史流式生成回复。
	GenerateContentStream(ctx context.Context, messages []*entity.Message, streamFn func(chunk string)) error
}

// EmbeddingProvider 定义了生成文本嵌入向量的接口。
type EmbeddingProvider interface {
	// CreateEmbeddings 为一批文本生成嵌入向量。
	CreateEmbeddings(ctx context.Context, texts []string) ([][]float32, error)
	// GetEmbeddingDimension 返回嵌入向量的维度。
	GetEmbeddingDimension() int
}

// RAGService 定义了执行 RAG 检索的接口 (暂时定义，后续实现)。
type RAGService interface {
	RetrieveRelevantChunks(ctx context.Context, query string, limit int) ([]*entity.DocumentChunk, error)
}

// MemoryService 定义了管理对话记忆和摘要的接口 (暂时定义，后续实现)。
type MemoryService interface {
	GetSummarizedHistory(ctx context.Context, conversationID uuid.UUID, currentMessages []*entity.Message) ([]*entity.Message, error)
	SummarizeAndSave(ctx context.Context, conversationID uuid.UUID, messages []*entity.Message) error
}
