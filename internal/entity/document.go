package entity

import (
	"time"

	"github.com/google/uuid"          // Keep uuid for New()
	"github.com/pgvector/pgvector-go" // 引入 pgvector 类型
)

// Document 代表用户上传的原始文件信息。
// 这可以映射到数据库中的一个表，用于存储文件元数据。
type Document struct {
	ID               string     `json:"id"`                 // 文件的唯一 ID (string UUID)
	UserID           string     `json:"user_id"`            // 上传文件的用户 ID
	OriginalFilename string     `json:"original_filename"`  // 原始文件名
	StoredPath       string     `json:"stored_path"`        // 文件在服务器上的存储路径
	FileSize         int64      `json:"file_size"`          // 文件大小 (bytes)
	ContentType      string     `json:"content_type"`       // 文件 MIME 类型
	UploadTime       time.Time  `json:"upload_time"`        // 文件上传时间
	ProcessingStatus TaskStatus `json:"processing_status"`  // 文件处理状态 (关联 Task)
	ProcessingTaskID *string    `json:"processing_task_id"` // 关联的处理任务 ID (string, e.g., Asynq ID)
	ErrorMessage     string     `json:"error_message"`      // 处理失败时的错误信息
	// 可以添加文件哈希等字段用于去重
	// FileHash         string     `json:"file_hash"`
}

// DocumentChunk 代表文档被分割后的一个块及其向量。
// 这通常对应于向量数据库中的一条记录。
type DocumentChunk struct {
	ID         string          `json:"id"`          // 块的唯一 ID (string UUID)
	DocumentID string          `json:"document_id"` // 所属文档的 ID (string UUID)
	UserID     string          `json:"user_id"`     // 所属用户 ID (用于数据隔离)
	ChunkIndex int             `json:"chunk_index"` // 块在文档中的顺序索引
	Content    string          `json:"content"`     // 块的文本内容
	Embedding  pgvector.Vector `json:"-"`           // 块内容的向量表示 (不在 JSON 中序列化)
	Metadata   map[string]any  `json:"metadata"`    // 附加元数据 (例如页码、来源等)
	CreatedAt  time.Time       `json:"created_at"`  // 创建时间
	// 可以添加 chunk_hash 用于幂等性检查
	// ChunkHash  string          `json:"chunk_hash"`
}

// NewDocument 创建一个新的 Document 实例。
func NewDocument(userID, originalFilename, storedPath string, fileSize int64, contentType string) *Document {
	return &Document{
		ID:               uuid.New().String(), // Generate string UUID
		UserID:           userID,
		OriginalFilename: originalFilename,
		StoredPath:       storedPath,
		FileSize:         fileSize,
		ContentType:      contentType,
		UploadTime:       time.Now(),
		ProcessingStatus: TaskStatusPending, // 初始状态为待处理
	}
}

// NewDocumentChunk 创建一个新的 DocumentChunk 实例。
// Changed documentID parameter to string
func NewDocumentChunk(documentID string, userID string, chunkIndex int, content string, embedding pgvector.Vector, metadata map[string]any) *DocumentChunk {
	return &DocumentChunk{
		ID:         uuid.New().String(), // Generate string UUID
		DocumentID: documentID,          // Assign string documentID
		UserID:     userID,
		ChunkIndex: chunkIndex,
		Content:    content,
		Embedding:  embedding,
		Metadata:   metadata,
		CreatedAt:  time.Now(),
	}
}
