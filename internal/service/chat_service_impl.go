package service

import (
	"context" // Import fmt for error formatting
	"fmt"     // Import fmt for string formatting
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
	ragService  RAGService                // RAG 服务
	// memoryService MemoryService             // 对话记忆服务 (未来添加)
	// maxHistory int // 最大历史消息数 (可以从配置读取)
}

// NewChatService 创建一个新的 chatServiceImpl 实例。
func NewChatService(
	chatRepo repository.ChatRepository,
	llm LLMProvider,
	rag RAGService, // 添加 RAG 服务依赖
	// mem MemoryService,
) ChatService {
	return &chatServiceImpl{
		chatRepo:    chatRepo,
		llmProvider: llm,
		ragService:  rag, // 初始化 RAG 服务
		// memoryService: mem,
		// maxHistory: 10, // Example default
	}
}

// HandleChatMessage 处理传入的聊天消息。
func (s *chatServiceImpl) HandleChatMessage(ctx context.Context, conversationID uuid.UUID, message string, modelName string) (reply string, newConversationID uuid.UUID, err error) {
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

	// 2. 获取对话历史
	const historyLimit = 10 // 获取最近 10 条消息作为上下文 (应可配置)
	historyMessages, err := s.chatRepo.GetConversationHistory(ctx, conversationID, historyLimit)
	if err != nil {
		logger.WarnContext(ctx, "获取对话历史失败，将继续执行", "error", err, "conversation_id", conversationID)
		historyMessages = []*entity.Message{} // 使用空历史
	}

	// 2.5 (RAG) 检索相关文档块
	const ragLimit = 3 // 获取最多 3 个相关块 (应可配置)
	var ragContextMessage *entity.Message
	relevantChunks, ragErr := s.ragService.RetrieveRelevantChunks(ctx, message, ragLimit)
	if ragErr != nil {
		// RAG 检索失败，记录警告但继续
		logger.WarnContext(ctx, "RAG 检索相关文档块失败", "error", ragErr, "conversation_id", conversationID)
	} else if len(relevantChunks) > 0 {
		// 构建 RAG 上下文消息
		var contextBuilder strings.Builder
		contextBuilder.WriteString("Relevant context:\n")
		for i, chunk := range relevantChunks {
			contextBuilder.WriteString(fmt.Sprintf("--- Context %d (Doc: %s, Chunk: %s) ---\n", i+1, chunk.DocumentID.String(), chunk.ID.String()))
			// 可以在这里添加清理或截断 chunk.Content 的逻辑
			contextBuilder.WriteString(chunk.Content)
			contextBuilder.WriteString("\n")
		}
		ragContextMessage = entity.NewMessage(conversationID, userID, entity.SenderRoleSystem, contextBuilder.String())
		logger.InfoContext(ctx, "成功检索 RAG 上下文", "chunk_count", len(relevantChunks), "conversation_id", conversationID)
	}

	// 准备 LLM 输入 (历史消息 + RAG 上下文 (如果存在) + 当前用户消息)
	llmInputMessages := make([]*entity.Message, 0, len(historyMessages)+2)
	llmInputMessages = append(llmInputMessages, historyMessages...)
	if ragContextMessage != nil {
		llmInputMessages = append(llmInputMessages, ragContextMessage) // 在用户消息前插入 RAG 上下文
	}
	llmInputMessages = append(llmInputMessages, userMessage)

	// 3. 调用 LLM 生成回复
	logger.InfoContext(ctx, "准备调用 LLM", "conversation_id", conversationID, "message_count", len(llmInputMessages), "model_name", modelName, "with_rag", ragContextMessage != nil) // Log model name and RAG status
	// 将 modelName 传递给 LLMProvider
	aiReplyContent, err := s.llmProvider.GenerateContent(ctx, llmInputMessages, modelName)
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
func (s *chatServiceImpl) HandleStreamChatMessage(ctx context.Context, conversationID uuid.UUID, message string, modelName string, streamCh chan<- string) (newConversationID uuid.UUID, err error) {
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

	// 2. 获取对话历史
	const historyLimit = 10 // (应可配置)
	historyMessages, err := s.chatRepo.GetConversationHistory(ctx, conversationID, historyLimit)
	if err != nil {
		logger.WarnContext(ctx, "获取对话历史失败 (流式)，将继续执行", "error", err, "conversation_id", conversationID)
		historyMessages = []*entity.Message{}
	}

	// 2.5 (RAG) 检索相关文档块
	const ragLimitStream = 3 // (应可配置)
	var ragContextMessageStream *entity.Message
	relevantChunksStream, ragErrStream := s.ragService.RetrieveRelevantChunks(ctx, message, ragLimitStream)
	if ragErrStream != nil {
		logger.WarnContext(ctx, "RAG 检索相关文档块失败 (流式)", "error", ragErrStream, "conversation_id", conversationID)
	} else if len(relevantChunksStream) > 0 {
		var contextBuilder strings.Builder
		contextBuilder.WriteString("Relevant context:\n")
		for i, chunk := range relevantChunksStream {
			contextBuilder.WriteString(fmt.Sprintf("--- Context %d (Doc: %s, Chunk: %s) ---\n", i+1, chunk.DocumentID.String(), chunk.ID.String()))
			contextBuilder.WriteString(chunk.Content)
			contextBuilder.WriteString("\n")
		}
		ragContextMessageStream = entity.NewMessage(conversationID, userID, entity.SenderRoleSystem, contextBuilder.String())
		logger.InfoContext(ctx, "成功检索 RAG 上下文 (流式)", "chunk_count", len(relevantChunksStream), "conversation_id", conversationID)
	}

	// 准备 LLM 输入 (历史消息 + RAG 上下文 (如果存在) + 当前用户消息)
	llmInputMessages := make([]*entity.Message, 0, len(historyMessages)+2)
	llmInputMessages = append(llmInputMessages, historyMessages...)
	if ragContextMessageStream != nil {
		llmInputMessages = append(llmInputMessages, ragContextMessageStream)
	}
	llmInputMessages = append(llmInputMessages, userMessage)

	// 3. 调用 LLM 流式生成回复
	logger.InfoContext(ctx, "准备调用 LLM (流式)", "conversation_id", conversationID, "message_count", len(llmInputMessages), "model_name", modelName, "with_rag", ragContextMessageStream != nil) // Log model name and RAG status

	var fullReply strings.Builder // 用于拼接完整回复以保存
	// 将 modelName 传递给 LLMProvider
	streamErr := s.llmProvider.GenerateContentStream(ctx, llmInputMessages, modelName, func(chunk string) {
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

// GetUserConversations 获取指定用户的所有对话基本信息。
func (s *chatServiceImpl) GetUserConversations(ctx context.Context, userID string) ([]*entity.Conversation, error) {
	// 直接调用 repository 层的方法
	conversations, err := s.chatRepo.GetUserConversations(ctx, userID)
	if err != nil {
		// 错误已在 repository 层记录和包装
		return nil, err
	}
	return conversations, nil
}
