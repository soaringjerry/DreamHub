package service

import (
	"context"

	"github.com/pgvector/pgvector-go"
	"github.com/soaringjerry/dreamhub/internal/entity"
	"github.com/soaringjerry/dreamhub/internal/repository"
	"github.com/soaringjerry/dreamhub/pkg/apperr"
	"github.com/soaringjerry/dreamhub/pkg/logger"
)

// ragServiceImpl 是 RAGService 接口的实现。
type ragServiceImpl struct {
	vectorRepo        repository.VectorRepository // 向量仓库
	embeddingProvider EmbeddingProvider           // 嵌入服务提供者
}

// NewRAGService 创建一个新的 ragServiceImpl 实例。
func NewRAGService(vectorRepo repository.VectorRepository, embeddingProvider EmbeddingProvider) RAGService {
	return &ragServiceImpl{
		vectorRepo:        vectorRepo,
		embeddingProvider: embeddingProvider,
	}
}

// RetrieveRelevantChunks 实现 RAGService 接口。
// 它将查询文本转换为向量，然后使用向量仓库搜索相似的文档块。
func (s *ragServiceImpl) RetrieveRelevantChunks(ctx context.Context, userID string, query string, limit int) ([]*entity.DocumentChunk, error) {
	logger.InfoContext(ctx, "开始检索相关文档块", "userID", userID, "query", query, "limit", limit)

	// 1. 将查询文本转换为嵌入向量
	queryEmbeddings, err := s.embeddingProvider.CreateEmbeddings(ctx, []string{query})
	if err != nil {
		logger.ErrorContext(ctx, "创建查询嵌入失败", "error", err)
		// 不直接返回 embeddingProvider 的错误，而是包装它
		return nil, apperr.Wrap(err, apperr.CodeInternal, "无法为查询生成嵌入向量")
	}

	if len(queryEmbeddings) == 0 || len(queryEmbeddings[0]) == 0 {
		logger.ErrorContext(ctx, "嵌入服务返回了空的查询向量", "query", query)
		return nil, apperr.New(apperr.CodeInternal, "嵌入服务未能生成有效的查询向量")
	}

	// 将 []float32 转换为 pgvector.Vector
	queryVector := pgvector.NewVector(queryEmbeddings[0])

	// 2. 使用向量仓库搜索相似块
	// 2. 使用向量仓库搜索相似块
	// userID is now passed as a parameter.
	// 如果需要额外的元数据过滤，可以在这里传递 filter map
	searchResults, err := s.vectorRepo.SearchSimilarChunks(ctx, userID, queryVector, limit, nil) // 使用传入的 userID
	if err != nil {
		logger.ErrorContext(ctx, "向量搜索失败", "error", err, "userID", userID)
		// vectorRepo 应该已经包装了错误，这里可以不再包装，或者根据需要再次包装
		return nil, err // 直接返回 vectorRepo 的错误
	}

	// 3. 从搜索结果中提取 DocumentChunk
	chunks := make([]*entity.DocumentChunk, 0, len(searchResults))
	for _, result := range searchResults {
		// 可以在这里添加距离阈值过滤，例如：
		// if result.Distance > 0.5 { continue }
		chunks = append(chunks, result.Chunk)
		logger.DebugContext(ctx, "找到相关块", "chunk_id", result.Chunk.ID, "document_id", result.Chunk.DocumentID, "distance", result.Distance)
	}

	logger.InfoContext(ctx, "成功检索到相关文档块", "count", len(chunks))
	return chunks, nil
}
