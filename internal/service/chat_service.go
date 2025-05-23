package service

import (
	"context"

	"github.com/soaringjerry/dreamhub/internal/entity"
)

// ChatService 定义了处理聊天交互的业务逻辑接口。
type ChatService interface {
	// HandleChatMessage 处理传入的聊天消息。
	// 如果 conversationID 为零值 UUID，则表示开始新对话。
	// 它会检索上下文（历史记录，未来可能包括 RAG），调用 LLM，
	// 保存用户消息和 AI 回复，并返回 AI 的回复和对话 ID。
	// modelName 参数用于指定要使用的 LLM 模型，如果为空则使用默认模型。
	// conversationID is now string
	HandleChatMessage(ctx context.Context, userID string, conversationID string, message string, modelName string) (reply string, newConversationID string, err error)

	// HandleStreamChatMessage 处理流式聊天消息 (用于 WebSocket)。
	// 实现逻辑与 HandleChatMessage 类似，但通过 channel 流式返回 AI 回复块。
	// modelName 参数用于指定要使用的 LLM 模型，如果为空则使用默认模型。
	// conversationID is now string
	HandleStreamChatMessage(ctx context.Context, userID string, conversationID string, message string, modelName string, streamCh chan<- string) (newConversationID string, err error)

	// GetConversationMessages 获取指定对话的消息列表（带分页）。
	// conversationID is now string
	GetConversationMessages(ctx context.Context, userID string, conversationID string, limit int, offset int) ([]*entity.Message, error)

	// GetUserConversations 获取指定用户的所有对话基本信息。
	GetUserConversations(ctx context.Context, userID string) ([]*entity.Conversation, error)

	// TODO: 可能需要添加其他方法，例如：
	// DeleteConversation(ctx context.Context, conversationID uuid.UUID) error
}

// LLMProvider 定义了与大型语言模型交互的接口。
// 这允许我们将具体的 LLM 实现（如 OpenAI, Anthropic 等）解耦。
type LLMProvider interface {
	// GenerateContent 根据提供的消息历史生成回复。
	// modelName 参数用于指定要使用的 LLM 模型，如果为空则使用默认模型。
	GenerateContent(ctx context.Context, messages []*entity.Message, modelName string) (string, error)
	// GenerateContentStream 根据提供的消息历史流式生成回复。
	// modelName 参数用于指定要使用的 LLM 模型，如果为空则使用默认模型。
	GenerateContentStream(ctx context.Context, messages []*entity.Message, modelName string, streamFn func(chunk string)) error
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
	RetrieveRelevantChunks(ctx context.Context, userID string, query string, limit int) ([]*entity.DocumentChunk, error)
}

// MemoryService 定义了管理对话记忆和摘要的接口 (暂时定义，后续实现)。
type MemoryService interface {
	// conversationID is now string
	GetSummarizedHistory(ctx context.Context, conversationID string, currentMessages []*entity.Message) ([]*entity.Message, error)
	// conversationID is now string
	SummarizeAndSave(ctx context.Context, conversationID string, messages []*entity.Message) error
}
