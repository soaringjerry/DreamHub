package repository

import (
	"context"

	// "github.com/google/uuid" // Removed unused import
	"github.com/soaringjerry/dreamhub/internal/entity"
)

// DocumentRepository 定义了与文档元数据存储交互的方法。
// 这通常对应于关系型数据库中的 documents 表。
type DocumentRepository interface {
	// SaveDocument 保存一个新的文档元数据记录。
	SaveDocument(ctx context.Context, doc *entity.Document) error

	// GetDocumentByID 根据 ID 获取文档元数据。
	// Added userID string parameter, changed docID to string
	GetDocumentByID(ctx context.Context, userID string, docID string) (*entity.Document, error)

	// GetDocumentsByUser 获取指定用户的所有文档元数据。
	// userID is already string
	// 可以添加分页、排序等参数。
	GetDocumentsByUser(ctx context.Context, userID string, limit int, offset int) ([]*entity.Document, error)

	// UpdateDocumentStatus 更新文档的处理状态和关联的任务 ID。
	// Added userID string parameter, changed docID to string, changed taskID to *string
	UpdateDocumentStatus(ctx context.Context, userID string, docID string, status entity.TaskStatus, taskID *string, errMsg string) error

	// DeleteDocument 删除文档元数据（可能还需要触发关联向量的删除）。
	// Added userID string parameter, changed docID to string
	DeleteDocument(ctx context.Context, userID string, docID string) error

	// TODO: 可能需要添加其他方法，例如：
	// GetDocumentByHash(ctx context.Context, userID string, fileHash string) (*entity.Document, error) // 用于去重
}
