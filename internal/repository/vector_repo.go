package repository

import (
	"context"

	// "github.com/google/uuid" // Removed unused import (assuming AddChunks implementation doesn't need it directly)
	"github.com/pgvector/pgvector-go" // 引入 pgvector 类型
	"github.com/soaringjerry/dreamhub/internal/entity"
)

// SearchResult 代表向量搜索的结果项。
type SearchResult struct {
	Chunk    *entity.DocumentChunk // 匹配的文档块
	Distance float32               // 与查询向量的距离 (相似度得分的某种度量)
}

// VectorRepository 定义了与向量数据存储交互的方法。
// 这通常对应于向量数据库 (如 PGVector, Milvus, Qdrant 等)。
type VectorRepository interface {
	// AddChunks 批量添加文档块及其向量。
	// 需要确保实现中根据 ctx 中的 user_id 进行了隔离 (例如通过元数据或命名空间)。
	// 实现应考虑幂等性 (例如基于 chunk_hash)。
	AddChunks(ctx context.Context, chunks []*entity.DocumentChunk) error

	// SearchSimilarChunks 搜索与查询向量相似的文档块。
	// 需要确保实现中根据 ctx 中的 user_id 进行了过滤。
	// limit 参数指定返回结果的数量。
	// filter 参数允许根据元数据进行额外过滤 (可选)。
	// Added userID string parameter for filtering
	SearchSimilarChunks(ctx context.Context, userID string, queryVector pgvector.Vector, limit int, filter map[string]any) ([]SearchResult, error)

	// DeleteChunksByDocumentID 删除指定文档的所有相关向量块。
	// Added userID string parameter for filtering
	// Changed documentID type from uuid.UUID to string
	DeleteChunksByDocumentID(ctx context.Context, userID string, documentID string) error

	// TODO: 可能需要添加其他方法，例如：
	// GetChunkByID(ctx context.Context, userID string, chunkID string) (*entity.DocumentChunk, error) // Changed chunkID to string, added userID
	// DeleteChunkByID(ctx context.Context, chunkID uuid.UUID) error
}
