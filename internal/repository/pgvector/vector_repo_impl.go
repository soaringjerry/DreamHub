package pgvector

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
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

// AddChunks 使用循环和单条 INSERT 语句添加文档块及其向量。
// user_id 和 document_id 被存储在 cmetadata 字段中。
func (r *pgVectorRepository) AddChunks(ctx context.Context, chunks []*entity.DocumentChunk) error {
	logger.InfoContext(ctx, ">>> Executing AddChunks with single INSERT logic <<<") // 添加日志标记
	logger.InfoContext(ctx, ">>> Executing AddChunks with single INSERT logic <<<") // 添加日志标记
	if len(chunks) == 0 {
		return nil // 没有需要添加的块
	}

	// 使用事务确保原子性
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "开始事务失败 (AddChunks)", "error", err)
		return apperr.Wrap(err, apperr.CodeInternal, "无法开始数据库事务")
	}
	defer func() {
		// 使用新的 context 以防原始 context 被取消
		rollbackCtx := context.Background()
		if p := recover(); p != nil {
			_ = tx.Rollback(rollbackCtx)
			panic(p)
		} else if err != nil {
			// 如果循环中出现错误，回滚事务
			if rollbackErr := tx.Rollback(rollbackCtx); rollbackErr != nil {
				logger.ErrorContext(rollbackCtx, "事务回滚失败 (AddChunks)", "rollback_error", rollbackErr, "original_error", err)
			}
		} else {
			// 如果循环成功，提交事务
			err = tx.Commit(ctx)
			if err != nil {
				logger.ErrorContext(ctx, "提交事务失败 (AddChunks)", "error", err)
				err = apperr.Wrap(err, apperr.CodeInternal, "无法提交数据库事务") // 包装错误以便上层处理
			}
		}
	}()

	// 准备 INSERT 语句
	// 假设表结构为 (embedding vector, document text, cmetadata jsonb)
	sql := fmt.Sprintf("INSERT INTO %s (embedding, document, cmetadata) VALUES ($1, $2, $3)", tableName)
	stmt, err := tx.Prepare(ctx, "insert_chunk", sql)
	if err != nil {
		logger.ErrorContext(ctx, "准备 INSERT 语句失败", "error", err)
		return apperr.Wrap(err, apperr.CodeInternal, "无法准备插入语句")
	}

	insertedCount := 0
	for _, chunk := range chunks {
		// 确保元数据包含必要信息
		if chunk.Metadata == nil {
			chunk.Metadata = make(map[string]any)
		}
		chunk.Metadata[metadataUserIDKey] = chunk.UserID
		chunk.Metadata[metadataDocumentIDKey] = chunk.DocumentID.String() // 存储为字符串
		chunk.Metadata[metadataChunkIDKey] = chunk.ID.String()            // 存储块 ID

		metadataBytes, errJson := json.Marshal(chunk.Metadata)
		if errJson != nil {
			logger.ErrorContext(ctx, "序列化块元数据失败", "error", errJson, "chunk_id", chunk.ID)
			err = apperr.Wrap(errJson, apperr.CodeInternal, fmt.Sprintf("无法序列化块 %s 的元数据", chunk.ID))
			return err // 返回错误，触发 defer 中的 Rollback
		}

		// 更彻底地清理文本内容，移除所有非法的 UTF8 字符和 C0 控制字符 (除了换行和制表符)
		cleanedContent := strings.Map(func(r rune) rune {
			// 移除 NULL 字节和 C0 控制字符 (U+0000 to U+001F) 但保留 Tab(U+0009) 和 LF(U+000A)
			if r == '\x00' || (r < ' ' && r != '\t' && r != '\n') {
				return -1 // -1 表示移除该 rune
			}
			// 检查是否为有效的 UTF8 rune (虽然 Go 字符串通常是有效的，但以防万一)
			// if !utf8.ValidRune(r) && r != utf8.RuneError {
			// 	return -1
			// }
			return r // 保留其他字符
		}, chunk.Content)

		// 直接传递 pgvector.Vector 类型和清理后的文本
		_, errExec := tx.Exec(ctx, stmt.Name, chunk.Embedding, cleanedContent, metadataBytes)
		if errExec != nil {
			logger.ErrorContext(ctx, "执行 INSERT 语句失败", "error", errExec, "chunk_id", chunk.ID)
			// Log the problematic content for debugging (be careful with large content)
			// logger.DebugContext(ctx, "Problematic chunk content (first 100 chars)", "content_prefix", cleanedContent[:min(100, len(cleanedContent))])
			err = apperr.Wrap(errExec, apperr.CodeInternal, fmt.Sprintf("无法插入块 %s", chunk.ID))
			return err // 返回错误，触发 defer 中的 Rollback
		}
		insertedCount++
	}

	// 如果没有错误，defer 中的 Commit 会被执行
	if err == nil {
		logger.InfoContext(ctx, "成功插入向量块 (逐条)", "count", insertedCount)
	}
	return err // 返回 nil 或 Commit 错误
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
