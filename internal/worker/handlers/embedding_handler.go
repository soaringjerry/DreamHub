package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	// Import strings

	"github.com/hibiken/asynq"
	"github.com/pgvector/pgvector-go" // Import pgvector
	"github.com/soaringjerry/dreamhub/internal/entity"
	"github.com/soaringjerry/dreamhub/internal/repository"
	"github.com/soaringjerry/dreamhub/internal/service"
	"github.com/soaringjerry/dreamhub/pkg/apperr"  // Import apperr
	"github.com/soaringjerry/dreamhub/pkg/ctxutil" // Import ctxutil
	"github.com/soaringjerry/dreamhub/pkg/logger"
	"github.com/tmc/langchaingo/textsplitter" // Import textsplitter
)

// EmbeddingTaskPayload 定义了 embedding:generate 任务的 payload 结构。
type EmbeddingTaskPayload struct {
	UserID     string `json:"user_id"`
	DocumentID string `json:"document_id"` // UUID as string
	// TaskID      string `json:"task_id"` // REMOVED: TaskID should be obtained from task context if needed, not payload.
	FilePath    string `json:"file_path"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
}

// EmbeddingTaskHandler 处理文件 Embedding 任务。
type EmbeddingTaskHandler struct {
	fileStorage   service.FileStorage
	docRepo       repository.DocumentRepository
	vectorRepo    repository.VectorRepository
	taskRepo      repository.TaskRepository // Optional: for detailed task status updates
	embedProvider service.EmbeddingProvider
	textSplitter  textsplitter.TextSplitter // Add text splitter field
}

// NewEmbeddingTaskHandler 创建一个新的 EmbeddingTaskHandler 实例。
func NewEmbeddingTaskHandler(
	fs service.FileStorage,
	dr repository.DocumentRepository,
	vr repository.VectorRepository,
	tr repository.TaskRepository, // Can be nil if not updating Task entity
	ep service.EmbeddingProvider,
	ts textsplitter.TextSplitter, // Accept text splitter in constructor
) *EmbeddingTaskHandler {
	return &EmbeddingTaskHandler{
		fileStorage:   fs,
		docRepo:       dr,
		vectorRepo:    vr,
		taskRepo:      tr,
		embedProvider: ep,
		textSplitter:  ts, // Store text splitter
	}
}

// ProcessTask 实现 asynq.Handler 接口。
func (h *EmbeddingTaskHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload EmbeddingTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		logger.ErrorContext(ctx, "无法解析 Embedding 任务 payload", "error", err, "payload", string(t.Payload()))
		// 返回错误，但不重试，因为 payload 格式错误
		return fmt.Errorf("无效的任务 payload: %w", err) // No retry for invalid payload
	}

	// docID is now string, no need to parse. Use payload.DocumentID directly.
	docID := payload.DocumentID
	if docID == "" { // Add basic validation
		logger.ErrorContext(ctx, "任务 payload 中缺少 document_id", "payload", string(t.Payload()))
		return fmt.Errorf("无效的任务 payload: 缺少 document_id") // No retry
	}

	// 使用 payload 中的 user_id 设置上下文，以便后续操作使用
	taskCtx := context.WithValue(ctx, ctxutil.UserIDKey, payload.UserID) // Use ctxutil after importing

	logger.InfoContext(taskCtx, "开始处理 Embedding 任务", "document_id", docID, "filename", payload.Filename)

	// 更新文档状态为 Processing
	// 注意：taskRepo 可能为 nil
	if h.taskRepo != nil {
		// TODO: Update Task entity status if needed
	}
	// 更新 Document 实体状态 - Pass string docID, userID. Pass nil for taskID.
	// Assuming the 5th argument (*string) is an optional internal task ID, not Asynq's ID.
	if err := h.docRepo.UpdateDocumentStatus(taskCtx, payload.UserID, docID, entity.TaskStatusProcessing, nil, ""); err != nil {
		// 如果更新状态失败，可能是文档已被删除，记录警告并可能停止处理
		logger.WarnContext(taskCtx, "更新文档状态为 Processing 失败 (可能已被删除?)", "error", err, "document_id", docID)
		// 根据错误类型决定是否继续
		if apperr.Is(err, apperr.CodeNotFound) {
			return nil // Document not found, nothing to process
		}
		// 对于其他错误，可能需要重试
		return fmt.Errorf("更新文档状态失败: %w", err) // Retry for other errors
	}

	// 1. 读取文件内容
	fileReader, err := h.fileStorage.GetFileReader(taskCtx, payload.FilePath)
	if err != nil {
		logger.ErrorContext(taskCtx, "无法读取文件", "error", err, "path", payload.FilePath, "document_id", docID)
		h.markDocumentAsFailed(taskCtx, docID, "无法读取文件") // Pass string docID
		return fmt.Errorf("读取文件失败: %w", err)             // Retry might help if it's a temporary issue
	}
	defer fileReader.Close()

	fileContentBytes, err := io.ReadAll(fileReader)
	if err != nil {
		logger.ErrorContext(taskCtx, "读取文件内容到内存失败", "error", err, "path", payload.FilePath, "document_id", docID)
		h.markDocumentAsFailed(taskCtx, docID, "读取文件内容失败") // Pass string docID
		return fmt.Errorf("读取文件内容失败: %w", err)             // Retry might help
	}
	fileContent := string(fileContentBytes)
	logger.InfoContext(taskCtx, "文件内容读取成功", "document_id", docID, "size", len(fileContentBytes))

	// 2. 文本分块
	// Use the injected text splitter
	chunksContent, err := h.textSplitter.SplitText(fileContent) // Remove taskCtx argument
	if err != nil {
		logger.ErrorContext(taskCtx, "文本分块失败", "error", err, "document_id", docID)
		h.markDocumentAsFailed(taskCtx, docID, "文本分块失败") // Pass string docID
		return fmt.Errorf("文本分块失败: %w", err)             // Consider retry? Depends on splitter error type.
	}
	if len(chunksContent) == 0 {
		logger.WarnContext(taskCtx, "文件分块后内容为空", "document_id", docID, "filename", payload.Filename)
		// 文件内容为空，标记为完成
		// Pass string docID, userID. Pass nil for taskID.
		errMsgEmpty := "文件内容为空"
		if err := h.docRepo.UpdateDocumentStatus(taskCtx, payload.UserID, docID, entity.TaskStatusCompleted, nil, errMsgEmpty); err != nil {
			logger.ErrorContext(taskCtx, "更新空文件状态为 Completed 失败", "error", err, "document_id", docID)
			return fmt.Errorf("更新空文件状态失败: %w", err)
		}
		return nil // No chunks to process
	}
	logger.InfoContext(taskCtx, "文本分块完成", "document_id", docID, "chunk_count", len(chunksContent))

	// 3. 生成 Embeddings
	// TODO: 处理大量 chunks 的情况，可能需要分批调用 Embedding API
	embeddings, err := h.embedProvider.CreateEmbeddings(taskCtx, chunksContent)
	if err != nil {
		logger.ErrorContext(taskCtx, "生成 Embeddings 失败", "error", err, "document_id", docID)
		h.markDocumentAsFailed(taskCtx, docID, "生成 Embeddings 失败") // Pass string docID
		// 根据错误类型决定是否重试 (e.g., rate limit vs invalid input)
		if apperr.Is(err, apperr.CodeRateLimited) || apperr.Is(err, apperr.CodeUnavailable) {
			return fmt.Errorf("生成 Embeddings 失败 (可重试): %w", err)
		}
		return fmt.Errorf("生成 Embeddings 失败 (不可重试): %w", err) // No retry for potentially permanent errors
	}
	logger.InfoContext(taskCtx, "Embeddings 生成成功", "document_id", docID, "embedding_count", len(embeddings))

	// 4. 创建 DocumentChunk 实体
	docChunks := make([]*entity.DocumentChunk, len(chunksContent))
	embeddingDim := h.embedProvider.GetEmbeddingDimension()
	for i, content := range chunksContent {
		if len(embeddings[i]) != embeddingDim {
			errMsg := fmt.Sprintf("块 %d 的 Embedding 维度不匹配 (预期 %d, 得到 %d)", i, embeddingDim, len(embeddings[i]))
			logger.ErrorContext(taskCtx, errMsg, "document_id", docID)
			h.markDocumentAsFailed(taskCtx, docID, "Embedding 维度不匹配") // Pass string docID
			return fmt.Errorf(errMsg)                                 // No retry
		}
		// 创建元数据
		metadata := map[string]any{
			// "page_number": ... // 如果能从分块器获取
		}
		// Pass string docID to NewDocumentChunk
		docChunks[i] = entity.NewDocumentChunk(
			docID,
			payload.UserID,
			i, // chunk index
			content,
			pgvector.NewVector(embeddings[i]),
			metadata,
		)
	}

	// 5. 批量保存 Chunks 到 VectorRepository
	if err := h.vectorRepo.AddChunks(taskCtx, docChunks); err != nil {
		logger.ErrorContext(taskCtx, "保存向量块到数据库失败", "error", err, "document_id", docID)
		h.markDocumentAsFailed(taskCtx, docID, "保存向量数据失败") // Pass string docID
		// 数据库错误通常可以重试
		return fmt.Errorf("保存向量块失败: %w", err) // Retry
	}
	logger.InfoContext(taskCtx, "向量块保存成功", "document_id", docID, "chunk_count", len(docChunks))

	// 6. 更新文档状态为 Completed
	// Pass string docID, userID. Pass nil for taskID.
	if err := h.docRepo.UpdateDocumentStatus(taskCtx, payload.UserID, docID, entity.TaskStatusCompleted, nil, ""); err != nil {
		logger.ErrorContext(taskCtx, "更新文档状态为 Completed 失败", "error", err, "document_id", docID)
		// 即使状态更新失败，主要工作已完成，可能需要重试或手动修复
		return fmt.Errorf("更新最终文档状态失败: %w", err) // Retry
	}

	// 7. (可选) 更新 Task 实体状态
	if h.taskRepo != nil {
		// TODO: Update Task entity status if needed
	}

	logger.InfoContext(taskCtx, "Embedding 任务成功完成", "document_id", docID, "filename", payload.Filename)
	return nil // 任务成功完成
}

// markDocumentAsFailed 是一个辅助函数，用于更新文档状态为失败。
// docID is now string
func (h *EmbeddingTaskHandler) markDocumentAsFailed(ctx context.Context, docID string, errMsg string) {
	// Need UserID and TaskID here. We can get UserID from ctx.
	// TaskID is not available directly in this helper function's scope.
	// Option 1: Pass payload to this function (cleaner).
	// Option 2: Extract UserID from ctx, TaskID remains unknown here.
	// Let's assume UserID is needed and extract it. TaskID update might need refactoring if required here.
	userID, ok := ctx.Value(ctxutil.UserIDKey).(string)
	if !ok {
		// Log error, but proceed with updating status without user/task context if possible
		logger.ErrorContext(ctx, "无法从上下文中获取 UserID (markDocumentAsFailed)", "document_id", docID)
		// Attempt update without UserID/TaskID if the repo allows, or handle error
		// For now, let's assume UpdateDocumentStatus requires UserID.
		// We cannot reliably get TaskID here without passing the payload.
		// Let's log and skip the update if UserID is missing.
		return
	}

	// We don't have payload.TaskID here. Pass nil for taskID.
	// This might be acceptable if the repo handles nil taskID gracefully,
	// or if updating the taskID is not critical in the failure case from here.
	// Pass nil for taskID.
	if err := h.docRepo.UpdateDocumentStatus(ctx, userID, docID, entity.TaskStatusFailed, nil, errMsg); err != nil {
		logger.ErrorContext(ctx, "标记文档为失败状态时出错", "error", err, "document_id", docID, "user_id", userID, "original_error", errMsg)
	}
	// TODO: Update Task entity status if needed (using h.taskRepo) - would also need TaskID here.
}

// Removed splitTextSimple function as it's replaced by textSplitter field.
