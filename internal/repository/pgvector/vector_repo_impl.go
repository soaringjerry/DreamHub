package pgvector

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pgvector/pgvector-go"
	"github.com/soaringjerry/dreamhub/internal/entity"
	"github.com/soaringjerry/dreamhub/internal/repository"
	"github.com/soaringjerry/dreamhub/internal/repository/postgres" // 需要访问 postgres.DB 和 GetUserIDFromCtx
	"github.com/soaringjerry/dreamhub/pkg/apperr"
	"github.com/soaringjerry/dreamhub/pkg/logger"
)

const (
	// tableName 是存储向量嵌入的表名。
	// 注意：这基于 README.md 中提到的 langchain_pg_embedding。如果实际表名不同，需要修改。
	tableName = "langchain_pg_embedding"
	// metadataUserIDKey 是 cmetadata JSONB 字段中存储用户 ID 的键。
	metadataUserIDKey = "user_id"
	// metadataDocumentIDKey 是 cmetadata JSONB 字段中存储文档 ID 的键。
	metadataDocumentIDKey = "document_id"
	// metadataChunkIDKey 是 cmetadata JSONB 字段中存储块 ID 的键 (可选，但推荐)。
	metadataChunkIDKey = "chunk_id"
	// metadataContentKey 是 cmetadata JSONB 字段中存储块内容的键 (可选，如果 document 列不存在或不适用)。
	// metadataContentKey = "content"
	// metadataChunkIndexKey 是 cmetadata JSONB 字段中存储块索引的键 (可选)。
	// metadataChunkIndexKey = "chunk_index"
)

// pgVectorRepository 是 VectorRepository 接口的 PGVector 实现。
type pgVectorRepository struct {
	db *postgres.DB // 嵌入 DB 连接池
}

// NewPGVectorRepository 创建一个新的 pgVectorRepository 实例。
func NewPGVectorRepository(db *postgres.DB) repository.VectorRepository {
	return &pgVectorRepository{db: db}
}

// AddChunks 使用 pgx.CopyFrom 批量添加文档块及其向量。
// user_id 和 document_id 被存储在 cmetadata 字段中。
func (r *pgVectorRepository) AddChunks(ctx context.Context, chunks []*entity.DocumentChunk) error {
	if len(chunks) == 0 {
		return nil // 没有需要添加的块
	}

	// 准备 CopyFrom 需要的数据源
	// 每行包含: collection_id (可选, 如果 langchaingo 需要), embedding, document (可选), cmetadata
	// 我们假设 collection_id 不是必需的，或者可以为 nil/default。
	// 我们将 user_id, document_id, chunk_id 等关键信息存入 cmetadata。
	// 我们假设 'document' 列存储块内容，如果不是，则需要将内容也存入 cmetadata。
	rows := make([][]interface{}, len(chunks))
	for i, chunk := range chunks {
		// 确保元数据包含必要信息
		if chunk.Metadata == nil {
			chunk.Metadata = make(map[string]any)
		}
		chunk.Metadata[metadataUserIDKey] = chunk.UserID
		chunk.Metadata[metadataDocumentIDKey] = chunk.DocumentID.String() // 存储为字符串
		chunk.Metadata[metadataChunkIDKey] = chunk.ID.String()            // 存储块 ID
		// chunk.Metadata[metadataContentKey] = chunk.Content // 如果 document 列不存在
		// chunk.Metadata[metadataChunkIndexKey] = chunk.ChunkIndex

		metadataBytes, err := json.Marshal(chunk.Metadata)
		if err != nil {
			logger.ErrorContext(ctx, "序列化块元数据失败", "error", err, "chunk_id", chunk.ID)
			return apperr.Wrap(err, apperr.CodeInternal, fmt.Sprintf("无法序列化块 %s 的元数据", chunk.ID))
		}

		// 假设表结构为 (embedding vector, document text, cmetadata jsonb)
		// 如果有 collection_id 列，需要调整
		rows[i] = []interface{}{chunk.Embedding, chunk.Content, metadataBytes}
	}

	// 定义要复制的列名
	// 调整这里的列名以匹配你的 langchain_pg_embedding 表结构
	columnNames := []string{"embedding", "document", "cmetadata"}
	// columnNames := []string{"collection_id", "embedding", "document", "cmetadata"} // 如果有 collection_id

	// 执行批量插入
	copyCount, err := r.db.Pool.CopyFrom(
		ctx,
		pgx.Identifier{tableName},
		columnNames,
		pgx.CopyFromRows(rows),
	)

	if err != nil {
		logger.ErrorContext(ctx, "批量插入向量块失败", "error", err)
		// 考虑错误类型，例如唯一约束冲突可能表示 CodeAlreadyExists
		return apperr.Wrap(err, apperr.CodeInternal, "无法批量添加向量块")
	}

	if int(copyCount) != len(chunks) {
		logger.WarnContext(ctx, "批量插入向量块数量不匹配", "expected", len(chunks), "inserted", copyCount)
		// 这可能表示部分插入成功，需要更复杂的错误处理或回滚逻辑
		// 暂时返回一个通用错误
		return apperr.New(apperr.CodeInternal, fmt.Sprintf("预期插入 %d 个块，但实际插入 %d 个", len(chunks), copyCount))
	}

	logger.InfoContext(ctx, "成功批量插入向量块", "count", copyCount)
	return nil
}

// SearchSimilarChunks 搜索与查询向量相似的文档块。
// 强制使用 ctx 中的 user_id 进行过滤。
// 支持基于 filter map 的额外元数据过滤。
func (r *pgVectorRepository) SearchSimilarChunks(ctx context.Context, queryVector pgvector.Vector, limit int, filter map[string]any) ([]repository.SearchResult, error) {
	userID, err := postgres.GetUserIDFromCtx(ctx) // 强制获取 UserID
	if err != nil {
		return nil, err
	}

	// -- 构建 SQL 查询 --
	// 选择列：块 ID (从元数据), 文档 ID (从元数据), 内容 (从 document 列), 元数据, 距离
	// 假设我们从 cmetadata 中恢复 chunk_id 和 document_id
	selectClause := fmt.Sprintf(`
		SELECT
			cmetadata->>'%s' AS chunk_id,
			cmetadata->>'%s' AS document_id,
			document AS content,
			cmetadata,
			embedding <-> $1 AS distance
		FROM %s
	`, metadataChunkIDKey, metadataDocumentIDKey, tableName) // 使用 cosine distance '<->'

	// -- 构建 WHERE 子句 --
	whereClauses := []string{
		// 强制用户 ID 过滤
		fmt.Sprintf("cmetadata @> '{\"%s\": \"%s\"}'::jsonb", metadataUserIDKey, userID),
	}
	args := []interface{}{queryVector} // $1 是查询向量
	argCounter := 2                    // 从 $2 开始用于其他过滤器

	// 添加来自 filter map 的额外过滤条件
	for key, value := range filter {
		// 对 key 进行基本清理，防止注入 (虽然这里 key 来自内部，但仍是好习惯)
		// 更健壮的方法是使用白名单验证 key
		safeKey := key // 假设 key 是安全的元数据字段名
		whereClauses = append(whereClauses, fmt.Sprintf("cmetadata->>'%s' = $%d", safeKey, argCounter))
		args = append(args, value)
		argCounter++
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// -- 构建 ORDER BY 和 LIMIT 子句 --
	orderByClause := "ORDER BY distance ASC" // 按距离升序排序
	limitClause := fmt.Sprintf("LIMIT $%d", argCounter)
	args = append(args, limit)

	// -- 组合最终 SQL --
	sql := fmt.Sprintf("%s %s %s %s", selectClause, whereClause, orderByClause, limitClause)

	// -- 执行查询 --
	logger.DebugContext(ctx, "执行向量搜索查询", "sql", sql, "args_count", len(args)) // 不记录 args 的值以防敏感信息
	rows, err := r.db.Pool.Query(ctx, sql, args...)
	if err != nil {
		logger.ErrorContext(ctx, "向量搜索查询失败", "error", err, "user_id", userID)
		return nil, apperr.Wrap(err, apperr.CodeInternal, "向量搜索失败")
	}
	defer rows.Close()

	// -- 处理结果 --
	results := make([]repository.SearchResult, 0)
	for rows.Next() {
		var chunkIDStr, docIDStr, content string
		var metadataBytes []byte
		var distance float32
		var metadataMap map[string]any // 用于解析 JSONB

		// 扫描结果行
		if err := rows.Scan(&chunkIDStr, &docIDStr, &content, &metadataBytes, &distance); err != nil {
			logger.ErrorContext(ctx, "扫描向量搜索结果行失败", "error", err)
			return nil, apperr.Wrap(err, apperr.CodeInternal, "处理向量搜索结果时出错")
		}

		// 解析 UUID
		chunkID, err := uuid.Parse(chunkIDStr)
		if err != nil {
			logger.WarnContext(ctx, "无法解析搜索结果中的 chunk_id", "error", err, "chunk_id_str", chunkIDStr)
			continue // 跳过无效的行
		}
		docID, err := uuid.Parse(docIDStr)
		if err != nil {
			logger.WarnContext(ctx, "无法解析搜索结果中的 document_id", "error", err, "doc_id_str", docIDStr)
			continue // 跳过无效的行
		}

		// 解析元数据 JSONB
		if err := json.Unmarshal(metadataBytes, &metadataMap); err != nil {
			logger.WarnContext(ctx, "无法解析搜索结果中的 cmetadata", "error", err, "chunk_id", chunkID)
			metadataMap = make(map[string]any) // 使用空 map
		}

		// 创建 DocumentChunk 实体 (不包含 Embedding)
		chunk := &entity.DocumentChunk{
			ID:         chunkID,
			DocumentID: docID,
			UserID:     userID, // 从上下文中获取，或从元数据中验证
			Content:    content,
			Metadata:   metadataMap,
			// ChunkIndex, CreatedAt 等字段无法直接从这个查询中获取，除非它们也存储在元数据中
		}

		results = append(results, repository.SearchResult{
			Chunk:    chunk,
			Distance: distance,
		})
	}

	if err := rows.Err(); err != nil {
		logger.ErrorContext(ctx, "处理向量搜索结果集时出错", "error", err)
		return nil, apperr.Wrap(err, apperr.CodeInternal, "处理向量搜索结果时出错")
	}

	return results, nil
}

// DeleteChunksByDocumentID 删除指定文档的所有相关向量块。
// 强制使用 ctx 中的 user_id 和传入的 documentID 进行过滤。
func (r *pgVectorRepository) DeleteChunksByDocumentID(ctx context.Context, documentID uuid.UUID) error {
	userID, err := postgres.GetUserIDFromCtx(ctx) // 强制获取 UserID
	if err != nil {
		return err
	}

	// 构建 WHERE 子句以匹配 user_id 和 document_id
	// 使用 JSONB 操作符 @>
	whereClause := fmt.Sprintf(`cmetadata @> '{"%s": "%s", "%s": "%s"}'::jsonb`,
		metadataUserIDKey, userID,
		metadataDocumentIDKey, documentID.String(),
	)

	sql := fmt.Sprintf("DELETE FROM %s WHERE %s", tableName, whereClause)

	logger.DebugContext(ctx, "执行向量块删除操作", "sql", sql) // SQL 不包含敏感信息
	cmdTag, err := r.db.Pool.Exec(ctx, sql)
	if err != nil {
		logger.ErrorContext(ctx, "删除向量块失败", "error", err, "document_id", documentID, "user_id", userID)
		return apperr.Wrap(err, apperr.CodeInternal, "无法删除向量块")
	}

	logger.InfoContext(ctx, "向量块删除操作完成", "document_id", documentID, "user_id", userID, "rows_affected", cmdTag.RowsAffected())
	// 注意：即使 RowsAffected 为 0 也可能不是错误（可能该文档没有块，或已被删除）
	return nil
}
