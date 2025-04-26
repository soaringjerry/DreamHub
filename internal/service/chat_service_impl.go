package service

import (
	"context" // Import fmt for error formatting
	"strings" // Import strings for builder

	"github.com/google/uuid"
	"github.com/soaringjerry/dreamhub/internal/entity"
	"github.com/soaringjerry/dreamhub/internal/repository"
	"github.com/soaringjerry/dreamhub/internal/repository/postgres" // 需要访问 postgres.GetUserIDFromCtx
	"github.com/soaringjerry/dreamhub/pkg/logger"
)

// chatServiceImpl 是 ChatService 接口的实现。
type chatServiceImpl struct {
	chatRepo    repository.ChatRepository // 对话历史仓库
	llmProvider LLMProvider               // LLM 服务提供者
	// ragService    RAGService                // RAG 服务 (未来添加)
	// memoryService MemoryService             // 对话记忆服务 (未来添加)
	// maxHistory int // 最大历史消息数 (可以从配置读取)
}

// NewChatService 创建一个新的 chatServiceImpl 实例。
func NewChatService(
	chatRepo repository.ChatRepository,
	llm LLMProvider,
	// rag RAGService,
	// mem MemoryService,
) ChatService {
	return &chatServiceImpl{
		chatRepo:    chatRepo,
		llmProvider: llm,
		// ragService:    rag,
		// memoryService: mem,
		// maxHistory: 10, // Example default
	}
}

// HandleChatMessage 处理传入的聊天消息。
func (s *chatServiceImpl) HandleChatMessage(ctx context.Context, conversationID uuid.UUID, message string) (reply string, newConversationID uuid.UUID, err error) {
	userID, err := postgres.GetUserIDFromCtx(ctx) // 强制获取 UserID
	if err != nil {
		return "", uuid.Nil, err
	}

	isNewConversation := (conversationID == uuid.Nil)
	if isNewConversation {
		conversationID = uuid.New() // 为新对话生成 ID
		logger.InfoContext(ctx, "开始新对话", "user_id", userID, "conversation_id", conversationID)
	} else {
		logger.InfoContext(ctx, "继续现有对话", "user_id", userID, "conversation_id", conversationID)
	}
	newConversationID = conversationID // 返回当前的（可能是新的）对话 ID

	// 1. 保存用户消息
	userMessage := entity.NewMessage(conversationID, userID, entity.SenderRoleUser, message)
	if err := s.chatRepo.SaveMessage(ctx, userMessage); err != nil {
		// 保存失败，直接返回错误
		return "", newConversationID, err // SaveMessage 内部已包装错误
	}

	// 2. 获取对话历史 (和 RAG 上下文 - 未来)
	// TODO: 结合 MemoryService 获取可能包含摘要的历史记录
	// TODO: 结合 RAGService 获取相关文档块
	const historyLimit = 10 // 获取最近 10 条消息作为上下文 (应可配置)
	historyMessages, err := s.chatRepo.GetConversationHistory(ctx, conversationID, historyLimit)
	if err != nil {
		// 获取历史失败，可以尝试无历史记录调用 LLM，或返回错误
		logger.WarnContext(ctx, "获取对话历史失败，将无历史记录调用 LLM", "error", err, "conversation_id", conversationID)
		// return "", newConversationID, err // 或者选择返回错误
		historyMessages = []*entity.Message{} // 使用空历史
	}

	// 准备 LLM 输入 (当前用户消息 + 历史消息)
	llmInputMessages := append(historyMessages, userMessage) // 注意顺序

	// 3. 调用 LLM 生成回复
	logger.InfoContext(ctx, "准备调用 LLM", "conversation_id", conversationID, "message_count", len(llmInputMessages))
	aiReplyContent, err := s.llmProvider.GenerateContent(ctx, llmInputMessages)
	if err != nil {
		// LLM 调用失败
		return "", newConversationID, err // GenerateContent 内部已包装错误
	}
	logger.InfoContext(ctx, "LLM 生成回复成功", "conversation_id", conversationID)

	// 4. 保存 AI 回复
	aiMessage := entity.NewMessage(conversationID, userID, entity.SenderRoleAI, aiReplyContent)
	if err := s.chatRepo.SaveMessage(ctx, aiMessage); err != nil {
		// 保存 AI 回复失败，这是一个问题，但用户已经收到了回复
		// 记录严重错误，但仍然返回成功获取的回复
		logger.ErrorContext(ctx, "保存 AI 回复失败", "error", err, "conversation_id", conversationID)
		// 可以考虑返回一个特殊的错误或标记，指示保存失败
	}

	// 5. (未来) 更新对话摘要
	// err = s.memoryService.SummarizeAndSave(ctx, conversationID, append(llmInputMessages, aiMessage))
	// if err != nil { logger.ErrorContext(ctx, "更新对话摘要失败", ...) }

	return aiReplyContent, newConversationID, nil
}

// HandleStreamChatMessage 处理流式聊天消息。
func (s *chatServiceImpl) HandleStreamChatMessage(ctx context.Context, conversationID uuid.UUID, message string, streamCh chan<- string) (newConversationID uuid.UUID, err error) {
	defer close(streamCh) // 确保 channel 在函数退出时关闭

	userID, err := postgres.GetUserIDFromCtx(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	isNewConversation := (conversationID == uuid.Nil)
	if isNewConversation {
		conversationID = uuid.New()
		logger.InfoContext(ctx, "开始新对话 (流式)", "user_id", userID, "conversation_id", conversationID)
	} else {
		logger.InfoContext(ctx, "继续现有对话 (流式)", "user_id", userID, "conversation_id", conversationID)
	}
	newConversationID = conversationID

	// 1. 保存用户消息
	userMessage := entity.NewMessage(conversationID, userID, entity.SenderRoleUser, message)
	if err := s.chatRepo.SaveMessage(ctx, userMessage); err != nil {
		return newConversationID, err
	}

	// 2. 获取对话历史 (和 RAG 上下文 - 未来)
	const historyLimit = 10
	historyMessages, err := s.chatRepo.GetConversationHistory(ctx, conversationID, historyLimit)
	if err != nil {
		logger.WarnContext(ctx, "获取对话历史失败 (流式)，将无历史记录调用 LLM", "error", err, "conversation_id", conversationID)
		historyMessages = []*entity.Message{}
	}
	llmInputMessages := append(historyMessages, userMessage)

	// 3. 调用 LLM 流式生成回复
	logger.InfoContext(ctx, "准备调用 LLM (流式)", "conversation_id", conversationID, "message_count", len(llmInputMessages))

	var fullReply strings.Builder // 用于拼接完整回复以保存
	streamErr := s.llmProvider.GenerateContentStream(ctx, llmInputMessages, func(chunk string) {
		// 将块发送到 channel
		select {
		case streamCh <- chunk:
			fullReply.WriteString(chunk) // 拼接回复
		case <-ctx.Done():
			// 如果上下文被取消（例如客户端断开连接），停止发送
			logger.InfoContext(ctx, "上下文取消，停止发送流式块", "conversation_id", conversationID)
			return // 虽然不能直接从回调中返回错误，但这会停止处理后续块
		}
	})

	if streamErr != nil {
		logger.ErrorContext(ctx, "LLM 流式调用失败", "error", streamErr, "conversation_id", conversationID)
		return newConversationID, streamErr // GenerateContentStream 内部已包装错误
	}
	logger.InfoContext(ctx, "LLM 流式回复完成", "conversation_id", conversationID)

	// 4. 保存完整的 AI 回复
	aiReplyContent := fullReply.String()
	if aiReplyContent != "" { // 确保有内容才保存
		aiMessage := entity.NewMessage(conversationID, userID, entity.SenderRoleAI, aiReplyContent)
		if err := s.chatRepo.SaveMessage(ctx, aiMessage); err != nil {
			logger.ErrorContext(ctx, "保存 AI 回复失败 (流式)", "error", err, "conversation_id", conversationID)
			// 记录错误，但不影响流式发送
		}
	} else {
		logger.WarnContext(ctx, "LLM 流式调用未生成任何内容", "conversation_id", conversationID)
	}

	// 5. (未来) 更新对话摘要
	// ...

	return newConversationID, nil
}

// GetConversationMessages 获取对话消息列表。
func (s *chatServiceImpl) GetConversationMessages(ctx context.Context, conversationID uuid.UUID, limit int, offset int) ([]*entity.Message, error) {
	// GetMessagesByConversationID 内部会根据 ctx 中的 user_id 过滤
	messages, err := s.chatRepo.GetMessagesByConversationID(ctx, conversationID, limit, offset)
	if err != nil {
		// 内部已记录日志和包装错误
		return nil, err
	}
	return messages, nil
}
