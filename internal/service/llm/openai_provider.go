package llm

import (
	"context"
	"errors" // Import standard errors package

	"github.com/soaringjerry/dreamhub/internal/entity"
	"github.com/soaringjerry/dreamhub/internal/service"
	"github.com/soaringjerry/dreamhub/pkg/apperr"
	"github.com/soaringjerry/dreamhub/pkg/config"
	"github.com/soaringjerry/dreamhub/pkg/logger"
	"github.com/tmc/langchaingo/llms" // Ensure llms is imported
	"github.com/tmc/langchaingo/llms/openai"
	// Ensure schema is imported
)

// openAIProvider 是 LLMProvider 接口的 OpenAI 实现。
type openAIProvider struct {
	client *openai.LLM
	// 可以添加其他配置，如模型名称、温度等
	modelName string
}

// NewOpenAIProvider 创建一个新的 openAIProvider 实例。
func NewOpenAIProvider(cfg *config.Config) (service.LLMProvider, error) {
	if cfg.OpenAIAPIKey == "" {
		// 使用 pkg/apperr 中定义的错误类型
		return nil, apperr.New(apperr.CodeInvalidArgument, "OpenAI API Key 未配置")
	}
	// TODO: 从配置中读取模型名称等参数
	modelName := "gpt-3.5-turbo" // 或者 "gpt-4", "gpt-4o" 等
	// modelName := cfg.OpenAIModelName

	// 使用 functional options 创建 OpenAI LLM 客户端
	// 更多选项见: https://pkg.go.dev/github.com/tmc/langchaingo/llms/openai#New
	opts := []openai.Option{
		openai.WithModel(modelName),
		openai.WithToken(cfg.OpenAIAPIKey),
		// openai.WithBaseURL(cfg.OpenAIBaseURL), // 如果使用代理或 Azure OpenAI
		// openai.WithOrganization(cfg.OpenAIOrgID),
		// openai.WithStreamingFunc(func(ctx context.Context, chunk []byte) error { ... }), // 全局流式处理函数 (通常在调用时指定)
	}

	llm, err := openai.New(opts...)
	if err != nil {
		logger.Error("创建 OpenAI LLM 客户端失败", "error", err)
		return nil, apperr.Wrap(err, apperr.CodeInternal, "无法创建 OpenAI 客户端")
	}

	logger.Info("OpenAI LLM Provider 初始化成功。", "model", modelName)
	return &openAIProvider{
		client:    llm,
		modelName: modelName,
	}, nil
}

// GenerateContent 使用 OpenAI 生成回复。
func (p *openAIProvider) GenerateContent(ctx context.Context, messages []*entity.Message) (string, error) {
	// 将 entity.Message 转换为 langchaingo 的 schema.ChatMessage
	chatMessages := convertToLangchainMessages(messages)

	// 调用 langchaingo 的 GenerateContent 方法处理聊天消息
	// TODO: 添加重试、限流等逻辑 (参考 DETAILED_PLAN.md 2.3)
	// 移除无效的 llms.WithStreaming 选项
	resp, err := p.client.GenerateContent(ctx, chatMessages)
	// 可以添加其他选项: llms.WithTemperature(0.7), llms.WithModel(p.modelName) 等

	if err != nil {
		logger.ErrorContext(ctx, "调用 OpenAI API 失败", "error", err)
		// 尝试解析 OpenAI 的具体错误类型 (如果 langchaingo 支持)
		// 使用 pkg/apperr 中定义的错误类型
		return "", apperr.Wrap(err, apperr.CodeUnavailable, "调用 LLM 服务失败") // Use CodeUnavailable for external service issues
	}

	// GenerateContent 返回 []*llms.Generation，我们需要提取第一个选择的内容
	if len(resp.Choices) == 0 {
		logger.ErrorContext(ctx, "OpenAI API 返回了空的 Choices", "response", resp)
		return "", apperr.New(apperr.CodeInternal, "LLM 服务返回空结果")
	}
	// 假设我们总是取第一个 Choice 的 Message Content
	return resp.Choices[0].Content, nil
}

// GenerateContentStream 使用 OpenAI 流式生成回复。
func (p *openAIProvider) GenerateContentStream(ctx context.Context, messages []*entity.Message, streamFn func(chunk string)) error {
	chatMessages := convertToLangchainMessages(messages)

	// 调用 langchaingo 的 GenerateContent 方法
	// TODO: 正确实现流式处理。当前 langchaingo 的 openai.LLM 可能需要不同的方法
	// 或选项来实现流式。暂时调用非流式方法并记录警告。
	logger.WarnContext(ctx, "GenerateContentStream: 流式处理尚未完全实现，将使用非流式调用。")

	// 移除无效的流式选项
	resp, err := p.client.GenerateContent(ctx, chatMessages)
	// 可以添加其他选项: llms.WithTemperature(0.7), llms.WithModel(p.modelName) 等

	if err != nil && !errors.Is(err, context.Canceled) { // 忽略由 context 取消引起的错误
		logger.ErrorContext(ctx, "调用 OpenAI (非流式) API 失败", "error", err)
		// 使用 pkg/apperr 中定义的错误类型
		return apperr.Wrap(err, apperr.CodeUnavailable, "调用 LLM 服务失败") // Use CodeUnavailable
	}

	// 如果非流式调用成功，将完整结果作为单个块传递
	if err == nil && len(resp.Choices) > 0 {
		streamFn(resp.Choices[0].Content)
	} else if err == nil {
		logger.WarnContext(ctx, "GenerateContentStream: 非流式调用成功但未返回任何内容")
	}

	// 返回 nil 因为错误已处理，或者模拟流结束（虽然只有一个块）
	return nil
}

// convertToLangchainMessages 将内部的 Message 实体转换为 langchaingo 的 llms.MessageContent。
func convertToLangchainMessages(messages []*entity.Message) []llms.MessageContent {
	lcMessages := make([]llms.MessageContent, 0, len(messages))
	for _, msg := range messages {
		var role llms.ChatMessageType
		switch msg.SenderRole {
		case entity.SenderRoleUser:
			role = llms.ChatMessageTypeHuman
		case entity.SenderRoleAI:
			role = llms.ChatMessageTypeAI
		case entity.SenderRoleSystem:
			role = llms.ChatMessageTypeSystem
		default:
			logger.Warn("转换消息时遇到未知发送者角色，作为 System 消息处理", "role", msg.SenderRole)
			role = llms.ChatMessageTypeSystem // Default to System
		}
		// Create MessageContent with a single TextPart
		content := llms.MessageContent{
			Role: role,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: msg.Content},
			},
		}
		lcMessages = append(lcMessages, content)
	}
	// Add system message if needed
	// systemMsg := llms.MessageContent{
	// 	Role: llms.ChatMessageTypeSystem,
	// 	Parts: []llms.ContentPart{
	// 		llms.TextContent{Text: "You are a helpful assistant."},
	// 	},
	// }
	// lcMessages = append([]llms.MessageContent{systemMsg}, lcMessages...)
	return lcMessages
}
