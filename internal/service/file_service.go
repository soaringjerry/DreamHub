package service

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/soaringjerry/dreamhub/internal/entity"
)

// FileService 定义了处理文件上传、元数据管理和触发异步处理的业务逻辑接口。
type FileService interface {
	// UploadFile 处理文件上传。
	// 1. 保存文件到存储 (e.g., 本地磁盘, S3)。
	// 2. 保存文件元数据到数据库 (DocumentRepository)。
	// 3. 将文件处理任务 (e.g., embedding) 入队 (TaskQueueClient)。
	// 返回创建的文档实体和任务 ID。
	// 需要从 ctx 中获取 user_id。
	UploadFile(ctx context.Context, filename string, fileSize int64, contentType string, fileData io.Reader) (*entity.Document, string, error) // taskID is string as Asynq returns string ID

	// GetDocument 获取文档元数据。
	// 需要从 ctx 中获取 user_id。
	GetDocument(ctx context.Context, docID uuid.UUID) (*entity.Document, error)

	// ListUserDocuments 列出指定用户上传的文档（带分页）。
	// 需要从 ctx 中获取 user_id。
	ListUserDocuments(ctx context.Context, userID string, limit int, offset int) ([]*entity.Document, error)

	// DeleteDocument 删除文档及其关联数据（文件、向量、任务状态等）。
	// 需要从 ctx 中获取 user_id。
	DeleteDocument(ctx context.Context, docID uuid.UUID) error

	// GetTaskStatus 获取异步任务的状态 (需要 TaskRepository)。
	// 需要从 ctx 中获取 user_id (如果需要按用户隔离任务视图)。
	GetTaskStatus(ctx context.Context, taskID string) (*entity.Task, error)
}

// TaskQueueClient 定义了与任务队列交互的接口。
// 这允许我们将具体的队列实现 (如 Asynq, RabbitMQ) 解耦。
type TaskQueueClient interface {
	// EnqueueEmbeddingTask 将一个生成 Embedding 的任务放入队列。
	// payload 应包含处理任务所需的所有信息，如 user_id, document_id, file_path 等。
	// 返回由队列系统生成的任务 ID。
	EnqueueEmbeddingTask(ctx context.Context, payload map[string]interface{}) (taskID string, err error)

	// TODO: 可能需要添加其他任务类型的入队方法，例如：
	// EnqueueSummarizationTask(...)
}

// FileStorage 定义了与文件存储交互的接口 (e.g., 本地文件系统, S3)。
type FileStorage interface {
	// SaveFile 保存文件内容并返回存储路径。
	SaveFile(ctx context.Context, userID string, filename string, fileData io.Reader) (storedPath string, err error)
	// DeleteFile 删除指定路径的文件。
	DeleteFile(ctx context.Context, storedPath string) error
	// GetFileReader 获取文件的 io.ReadCloser。
	GetFileReader(ctx context.Context, storedPath string) (io.ReadCloser, error)
}
