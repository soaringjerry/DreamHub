package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/soaringjerry/dreamhub/internal/entity"
	"github.com/soaringjerry/dreamhub/internal/repository"
	"github.com/soaringjerry/dreamhub/pkg/apperr"
	"github.com/soaringjerry/dreamhub/pkg/logger"
)

// postgresChatRepository 是 ChatRepository 接口的 PostgreSQL 实现。
type postgresChatRepository struct {
	db *DB // 嵌入 DB 连接池
}

// NewPostgresChatRepository 创建一个新的 postgresChatRepository 实例。
func NewPostgresChatRepository(db *DB) repository.ChatRepository {
	return &postgresChatRepository{db: db}
}

// SaveMessage 保存一条新的对话消息到 conversation_history 表。
func (r *postgresChatRepository) SaveMessage(ctx context.Context, message *entity.Message) error {
	const sql = `
		INSERT INTO conversation_history (id, conversation_id, user_id, sender_role, message_content, timestamp, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Pool.Exec(ctx, sql,
		message.ID,
		message.ConversationID,
		message.UserID, // user_id 用于数据隔离
		message.SenderRole,
		message.Content,
		message.Timestamp,
		message.Metadata, // 确保存储的是 JSONB 兼容的类型 (json.RawMessage 应该可以)
	)
	if err != nil {
		logger.ErrorContext(ctx, "保存消息到数据库失败", "error", err, "message_id", message.ID)
		return apperr.Wrap(err, apperr.CodeInternal, "无法保存对话消息") // Use CodeInternal
	}
	logger.InfoContext(ctx, "消息成功保存到数据库", "message_id", message.ID, "conversation_id", message.ConversationID)
	return nil
}

// GetMessagesByConversationID 获取指定对话的所有消息，按时间戳升序排列。
// 强制使用 ctx 中的 user_id 进行过滤。
func (r *postgresChatRepository) GetMessagesByConversationID(ctx context.Context, conversationID uuid.UUID, limit int, offset int) ([]*entity.Message, error) {
	userID, err := GetUserIDFromCtx(ctx) // 强制获取 UserID
	if err != nil {
		return nil, err // GetUserIDFromCtx 已经包装了错误
	}

	const sql = `
		SELECT id, conversation_id, user_id, sender_role, message_content, timestamp, metadata
		FROM conversation_history
		WHERE conversation_id = $1 AND user_id = $2
		ORDER BY timestamp ASC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.Pool.Query(ctx, sql, conversationID, userID, limit, offset)
	if err != nil {
		logger.ErrorContext(ctx, "从数据库获取对话消息失败", "error", err, "conversation_id", conversationID, "user_id", userID)
		return nil, apperr.Wrap(err, apperr.CodeInternal, "无法获取对话消息") // Use CodeInternal
	}
	defer rows.Close()

	messages := make([]*entity.Message, 0)
	for rows.Next() {
		var msg entity.Message
		// 注意：扫描 metadata (JSONB) 到 json.RawMessage
		if err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.UserID, &msg.SenderRole, &msg.Content, &msg.Timestamp, &msg.Metadata); err != nil {
			logger.ErrorContext(ctx, "扫描数据库行失败", "error", err)
			// 决定是继续处理部分结果还是返回错误
			return nil, apperr.Wrap(err, apperr.CodeInternal, "处理数据库结果时出错")
		}
		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		logger.ErrorContext(ctx, "处理数据库结果集时出错", "error", err)
		return nil, apperr.Wrap(err, apperr.CodeInternal, "处理数据库结果时出错")
	}

	return messages, nil
}

// GetConversationHistory 获取指定对话的最近 N 条消息，按时间戳降序排列。
// 强制使用 ctx 中的 user_id 进行过滤。
func (r *postgresChatRepository) GetConversationHistory(ctx context.Context, conversationID uuid.UUID, lastN int) ([]*entity.Message, error) {
	userID, err := GetUserIDFromCtx(ctx) // 强制获取 UserID
	if err != nil {
		return nil, err
	}

	// lastN <= 0 表示获取所有历史记录（或者可以设定一个合理的默认值/上限）
	if lastN <= 0 {
		// 可以选择返回错误，或者获取所有记录（可能需要调整查询）
		// 这里我们暂时返回空切片，表示不获取
		logger.WarnContext(ctx, "GetConversationHistory 请求 lastN <= 0", "conversation_id", conversationID, "user_id", userID, "lastN", lastN)
		return []*entity.Message{}, nil
		// 或者调整 SQL 去掉 LIMIT
	}

	const sql = `
		SELECT id, conversation_id, user_id, sender_role, message_content, timestamp, metadata
		FROM conversation_history
		WHERE conversation_id = $1 AND user_id = $2
		ORDER BY timestamp DESC
		LIMIT $3
	`
	rows, err := r.db.Pool.Query(ctx, sql, conversationID, userID, lastN)
	if err != nil {
		logger.ErrorContext(ctx, "从数据库获取最近对话历史失败", "error", err, "conversation_id", conversationID, "user_id", userID, "lastN", lastN)
		return nil, apperr.Wrap(err, apperr.CodeInternal, "无法获取最近对话历史") // Use CodeInternal
	}
	defer rows.Close()

	messages := make([]*entity.Message, 0, lastN) // 预分配容量
	for rows.Next() {
		var msg entity.Message
		if err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.UserID, &msg.SenderRole, &msg.Content, &msg.Timestamp, &msg.Metadata); err != nil {
			logger.ErrorContext(ctx, "扫描数据库行失败 (历史记录)", "error", err)
			return nil, apperr.Wrap(err, apperr.CodeInternal, "处理数据库结果时出错")
		}
		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		logger.ErrorContext(ctx, "处理数据库结果集时出错 (历史记录)", "error", err)
		return nil, apperr.Wrap(err, apperr.CodeInternal, "处理数据库结果时出错")
	}

	// 因为查询是 DESC，但通常上下文需要按时间顺序 (ASC)，所以反转切片
	reverseMessages(messages)

	return messages, nil
}

// reverseMessages 原地反转消息切片。
func reverseMessages(s []*entity.Message) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// GetUserConversations 获取指定用户的所有对话基本信息，按最后更新时间降序排序。
// 注意：当前实现从 conversation_history 推断信息，Title 字段将为空。
func (r *postgresChatRepository) GetUserConversations(ctx context.Context, userID string) ([]*entity.Conversation, error) {
	const sql = `
		SELECT
			conversation_id AS id,
			user_id,
			'' AS title, -- Title 暂时无法从 history 表高效获取
			MIN(timestamp) AS created_at,
			MAX(timestamp) AS last_updated_at
		FROM conversation_history
		WHERE user_id = $1
		GROUP BY conversation_id, user_id
		ORDER BY last_updated_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, sql, userID)
	if err != nil {
		logger.ErrorContext(ctx, "从数据库获取用户对话列表失败", "error", err, "user_id", userID)
		return nil, apperr.Wrap(err, apperr.CodeInternal, "无法获取用户对话列表")
	}
	defer rows.Close()

	conversations := make([]*entity.Conversation, 0)
	for rows.Next() {
		var conv entity.Conversation
		// 注意：确保 entity.Conversation 中的字段类型与查询结果匹配
		if err := rows.Scan(&conv.ID, &conv.UserID, &conv.Title, &conv.CreatedAt, &conv.LastUpdatedAt); err != nil {
			logger.ErrorContext(ctx, "扫描数据库行失败 (对话列表)", "error", err)
			// 决定是继续处理部分结果还是返回错误
			return nil, apperr.Wrap(err, apperr.CodeInternal, "处理数据库结果时出错")
		}
		// 确保 UserID 匹配 (虽然 WHERE 子句已过滤，但作为安全检查)
		if conv.UserID != userID {
			logger.WarnContext(ctx, "查询结果中的 UserID 与请求不匹配", "expected_user_id", userID, "actual_user_id", conv.UserID, "conversation_id", conv.ID)
			continue // 跳过不匹配的记录
		}
		conversations = append(conversations, &conv)
	}

	if err := rows.Err(); err != nil {
		logger.ErrorContext(ctx, "处理数据库结果集时出错 (对话列表)", "error", err)
		return nil, apperr.Wrap(err, apperr.CodeInternal, "处理数据库结果时出错")
	}

	return conversations, nil
}
