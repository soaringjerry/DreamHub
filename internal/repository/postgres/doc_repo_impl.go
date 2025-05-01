package postgres

import (
	"context"
	"time"

	// "github.com/google/uuid" // Removed unused import
	"github.com/jackc/pgx/v5"
	"github.com/soaringjerry/dreamhub/internal/entity"
	"github.com/soaringjerry/dreamhub/internal/repository"
	"github.com/soaringjerry/dreamhub/pkg/apperr"
	"github.com/soaringjerry/dreamhub/pkg/logger"
)

// postgresDocumentRepository 是 DocumentRepository 接口的 PostgreSQL 实现。
type postgresDocumentRepository struct {
	db *DB // 嵌入 DB 连接池
}

// NewPostgresDocumentRepository 创建一个新的 postgresDocumentRepository 实例。
func NewPostgresDocumentRepository(db *DB) repository.DocumentRepository {
	return &postgresDocumentRepository{db: db}
}

// SaveDocument 保存一个新的文档元数据记录到 documents 表。
func (r *postgresDocumentRepository) SaveDocument(ctx context.Context, doc *entity.Document) error {
	const sql = `
		INSERT INTO documents (id, user_id, original_filename, stored_path, file_size, content_type, upload_time, processing_status, processing_task_id, error_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Pool.Exec(ctx, sql,
		doc.ID,
		doc.UserID,
		doc.OriginalFilename,
		doc.StoredPath,
		doc.FileSize,
		doc.ContentType,
		doc.UploadTime,
		doc.ProcessingStatus,
		doc.ProcessingTaskID,
		doc.ErrorMessage,
	)
	if err != nil {
		logger.ErrorContext(ctx, "保存文档元数据到数据库失败", "error", err, "doc_id", doc.ID, "filename", doc.OriginalFilename)
		// 考虑是否因为唯一约束冲突 (e.g., file_hash) 返回 CodeAlreadyExists
		return apperr.Wrap(err, apperr.CodeInternal, "无法保存文档元数据")
	}
	logger.InfoContext(ctx, "文档元数据成功保存", "doc_id", doc.ID, "filename", doc.OriginalFilename)
	return nil
}

// GetDocumentByID 根据 ID 获取文档元数据。
// Added userID string parameter, changed docID to string
func (r *postgresDocumentRepository) GetDocumentByID(ctx context.Context, userID string, docID string) (*entity.Document, error) {
	// userID is now passed explicitly.
	// userID, err := GetUserIDFromCtx(ctx) // REMOVED
	// if err != nil {
	// 	return nil, err
	// }

	const sql = `
		SELECT id, user_id, original_filename, stored_path, file_size, content_type, upload_time, processing_status, processing_task_id, error_message
		FROM documents
		WHERE id = $1 AND user_id = $2
	`
	// Pass string docID and userID
	row := r.db.Pool.QueryRow(ctx, sql, docID, userID)
	var doc entity.Document
	// Assuming entity.Document fields (ID, UserID, ProcessingTaskID) are now string or *string
	err := row.Scan(
		&doc.ID, &doc.UserID, &doc.OriginalFilename, &doc.StoredPath, &doc.FileSize,
		&doc.ContentType, &doc.UploadTime, &doc.ProcessingStatus, &doc.ProcessingTaskID, &doc.ErrorMessage,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			logger.WarnContext(ctx, "未找到指定的文档", "doc_id", docID, "user_id", userID)
			return nil, apperr.ErrNotFound("文档未找到")
		}
		logger.ErrorContext(ctx, "从数据库获取文档元数据失败", "error", err, "doc_id", docID, "user_id", userID)
		return nil, apperr.Wrap(err, apperr.CodeInternal, "无法获取文档元数据")
	}
	return &doc, nil
}

// GetDocumentsByUser 获取指定用户的所有文档元数据，按上传时间降序排列。
// userID is already string
func (r *postgresDocumentRepository) GetDocumentsByUser(ctx context.Context, userID string, limit int, offset int) ([]*entity.Document, error) {
	// Optional: Keep context check for defense-in-depth or admin roles, but primary filtering uses passed userID.
	// ctxUserID, err := GetUserIDFromCtx(ctx)
	// if err != nil {
	// 	return nil, err
	// }
	// if userID != ctxUserID {
	// 	logger.WarnContext(ctx, "Attempting to list documents for a different user", "request_user_id", userID, "context_user_id", ctxUserID)
	// 	return nil, apperr.ErrPermissionDenied("Permission denied to list documents for this user")
	// }

	const sql = `
		SELECT id, user_id, original_filename, stored_path, file_size, content_type, upload_time, processing_status, processing_task_id, error_message
		FROM documents
		WHERE user_id = $1
		ORDER BY upload_time DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Pool.Query(ctx, sql, userID, limit, offset)
	if err != nil {
		logger.ErrorContext(ctx, "从数据库获取用户文档列表失败", "error", err, "user_id", userID)
		return nil, apperr.Wrap(err, apperr.CodeInternal, "无法获取用户文档列表")
	}
	defer rows.Close()

	documents := make([]*entity.Document, 0)
	for rows.Next() {
		var doc entity.Document
		err := rows.Scan(
			&doc.ID, &doc.UserID, &doc.OriginalFilename, &doc.StoredPath, &doc.FileSize,
			&doc.ContentType, &doc.UploadTime, &doc.ProcessingStatus, &doc.ProcessingTaskID, &doc.ErrorMessage,
		)
		if err != nil {
			logger.ErrorContext(ctx, "扫描文档行失败", "error", err)
			return nil, apperr.Wrap(err, apperr.CodeInternal, "处理数据库结果时出错")
		}
		documents = append(documents, &doc)
	}

	if err := rows.Err(); err != nil {
		logger.ErrorContext(ctx, "处理文档结果集时出错", "error", err)
		return nil, apperr.Wrap(err, apperr.CodeInternal, "处理数据库结果时出错")
	}

	return documents, nil
}

// UpdateDocumentStatus 更新文档的处理状态、关联任务 ID 和错误信息。
// Added userID string, changed docID to string, changed taskID to *string
func (r *postgresDocumentRepository) UpdateDocumentStatus(ctx context.Context, userID string, docID string, status entity.TaskStatus, taskID *string, errMsg string) error {
	// userID is now passed explicitly.
	// userID, err := GetUserIDFromCtx(ctx) // REMOVED
	// if err != nil {
	// 	return err
	// }

	const sql = `
		UPDATE documents
		SET processing_status = $1, processing_task_id = $2, error_message = $3, upload_time = $4 -- 更新 upload_time 作为 updated_at
		WHERE id = $5 AND user_id = $6
	`
	// 注意：这里使用 upload_time 作为 updated_at 的替代，如果需要精确的 updated_at，应添加该列
	// Pass string docID, userID, and *string taskID
	cmdTag, err := r.db.Pool.Exec(ctx, sql, status, taskID, errMsg, time.Now(), docID, userID)
	if err != nil {
		logger.ErrorContext(ctx, "更新文档状态失败", "error", err, "doc_id", docID, "user_id", userID, "status", status) // Log string IDs
		return apperr.Wrap(err, apperr.CodeInternal, "无法更新文档状态")
	}

	if cmdTag.RowsAffected() == 0 {
		logger.WarnContext(ctx, "尝试更新不存在或不属于该用户的文档状态", "doc_id", docID, "user_id", userID)
		// 返回 NotFound 错误，因为 WHERE 条件 (id 和 user_id) 未匹配
		return apperr.ErrNotFound("文档未找到或无权更新")
	}

	logger.InfoContext(ctx, "文档状态更新成功", "doc_id", docID, "status", status)
	return nil
}

// DeleteDocument 删除文档元数据。
// 注意：此操作通常应在 Service 层协调，确保关联的向量数据也被删除。
// Added userID string, changed docID to string
func (r *postgresDocumentRepository) DeleteDocument(ctx context.Context, userID string, docID string) error {
	// userID is now passed explicitly.
	// userID, err := GetUserIDFromCtx(ctx) // REMOVED
	// if err != nil {
	// 	return err
	// }

	const sql = `DELETE FROM documents WHERE id = $1 AND user_id = $2`
	// Pass string docID and userID
	cmdTag, err := r.db.Pool.Exec(ctx, sql, docID, userID)
	if err != nil {
		logger.ErrorContext(ctx, "删除文档元数据失败", "error", err, "doc_id", docID, "user_id", userID) // Log string IDs
		return apperr.Wrap(err, apperr.CodeInternal, "无法删除文档元数据")
	}

	if cmdTag.RowsAffected() == 0 {
		logger.WarnContext(ctx, "尝试删除不存在或不属于该用户的文档", "doc_id", docID, "user_id", userID)
		return apperr.ErrNotFound("文档未找到或无权删除")
	}

	logger.InfoContext(ctx, "文档元数据删除成功", "doc_id", docID)
	return nil
}
