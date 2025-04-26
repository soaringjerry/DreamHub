package embedding

import (
	"context"

	"github.com/soaringjerry/dreamhub/internal/service"
	"github.com/soaringjerry/dreamhub/pkg/apperr"
	"github.com/soaringjerry/dreamhub/pkg/config"
	"github.com/soaringjerry/dreamhub/pkg/logger"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai" // 回到原始的 openai 包
)

// openAIEmbeddingProvider 是 EmbeddingProvider 接口的 OpenAI 实现。
type openAIEmbeddingProvider struct {
	embedder  embeddings.Embedder // langchaingo embeddings client
	dimension int                 // 存储向量维度
}

// NewOpenAIEmbeddingProvider 创建一个新的 openAIEmbeddingProvider 实例。
func NewOpenAIEmbeddingProvider(cfg *config.Config) (service.EmbeddingProvider, error) {
	if cfg.OpenAIAPIKey == "" {
		return nil, apperr.New(apperr.CodeInvalidArgument, "OpenAI API Key 未配置")
	}

	// TODO: 从配置中读取 Embedding 模型名称
	// 常用的 OpenAI Embedding 模型: "text-embedding-ada-002", "text-embedding-3-small", "text-embedding-3-large"
	embeddingModel := "text-embedding-3-small" // Example model
	// embeddingModel := cfg.OpenAIEmbeddingModel

	// 创建标准的 OpenAI LLM 客户端
	llmClient, err := openai.New(
		openai.WithModel(embeddingModel),
		openai.WithToken(cfg.OpenAIAPIKey),
		// openai.WithBaseURL(cfg.OpenAIBaseURL), // 如果需要
	)
	if err != nil {
		logger.Error("创建 OpenAI 客户端失败", "error", err, "model", embeddingModel)
		return nil, apperr.Wrap(err, apperr.CodeInternal, "无法创建 OpenAI 客户端")
	}

	// 使用 embeddings.NewEmbedder 创建嵌入器
	embedder, err := embeddings.NewEmbedder(llmClient)
	if err != nil {
		logger.Error("创建 Embedder 失败", "error", err, "model", embeddingModel)
		return nil, apperr.Wrap(err, apperr.CodeInternal, "无法创建 Embedder")
	}

	// 获取并存储维度 (保持硬编码逻辑)
	var dimension int
	switch embeddingModel {
	case "text-embedding-ada-002":
		dimension = 1536
	case "text-embedding-3-small":
		dimension = 1536 // 根据 OpenAI 文档确认
	case "text-embedding-3-large":
		dimension = 3072 // 根据 OpenAI 文档确认
	default:
		dimension = 1536 // 默认值，可能不准确
		logger.Warn("未知的 OpenAI Embedding 模型，使用默认维度", "model", embeddingModel, "dimension", dimension)
	}

	logger.Info("OpenAI Embedding Provider 初始化成功。", "model", embeddingModel, "dimension", dimension)
	return &openAIEmbeddingProvider{
		embedder:  embedder,
		dimension: dimension,
	}, nil
}

// CreateEmbeddings 为一批文本生成嵌入向量。
func (p *openAIEmbeddingProvider) CreateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	// TODO: 添加重试、限流逻辑
	embeddings, err := p.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		logger.ErrorContext(ctx, "调用 OpenAI Embedding API 失败", "error", err, "text_count", len(texts))
		return nil, apperr.Wrap(err, apperr.CodeUnavailable, "调用 Embedding 服务失败")
	}

	if len(embeddings) != len(texts) {
		logger.ErrorContext(ctx, "Embedding API 返回的向量数量与输入文本数量不匹配", "expected", len(texts), "received", len(embeddings))
		return nil, apperr.New(apperr.CodeInternal, "Embedding 服务返回结果数量不一致")
	}

	return embeddings, nil
}

// GetEmbeddingDimension 返回嵌入向量的维度。
func (p *openAIEmbeddingProvider) GetEmbeddingDimension() int {
	return p.dimension
}

// 不再需要自定义适配器，因为 openai.LLM 已经实现了 embeddings.EmbedderClient 接口
// 可以直接传递给 embeddings.NewEmbedder 函数
