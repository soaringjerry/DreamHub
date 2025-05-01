package service

import (
	"context"
	"io"

	// "github.com/google/uuid" // Removed unused import
	"github.com/soaringjerry/dreamhub/internal/entity"
	"github.com/soaringjerry/dreamhub/internal/repository"

	// "github.com/soaringjerry/dreamhub/internal/repository/postgres" // Avoid dependency on specific implementation details like GetUserIDFromCtx
	"github.com/soaringjerry/dreamhub/pkg/apperr"
	"github.com/soaringjerry/dreamhub/pkg/logger"
)

// fileServiceImpl 是 FileService 接口的实现。
type fileServiceImpl struct {
	fileStorage FileStorage                   // 文件存储接口
	docRepo     repository.DocumentRepository // 文档元数据仓库
	taskRepo    repository.TaskRepository     // 任务状态仓库
	taskQueue   TaskQueueClient               // 任务队列客户端
	vectorRepo  repository.VectorRepository   // 向量仓库 (用于删除)
}

// NewFileService 创建一个新的 fileServiceImpl 实例。
func NewFileService(
	fs FileStorage,
	dr repository.DocumentRepository,
	tr repository.TaskRepository,
	tq TaskQueueClient,
	vr repository.VectorRepository,
) FileService {
	return &fileServiceImpl{
		fileStorage: fs,
		docRepo:     dr,
		taskRepo:    tr,
		taskQueue:   tq,
		vectorRepo:  vr,
	}
}

// UploadFile 处理文件上传，保存文件和元数据，并触发 Embedding 任务。
// Added userID string parameter
func (s *fileServiceImpl) UploadFile(ctx context.Context, userID string, filename string, fileSize int64, contentType string, fileData io.Reader) (*entity.Document, string, error) {
	// userID is now passed explicitly, no need to extract from context here.
	// userID, err := postgres.GetUserIDFromCtx(ctx) // REMOVED
	// if err != nil {
	// 	return nil, "", err
	// }

	// 1. 保存文件到存储 (SaveFile already accepts userID string)
	storedPath, err := s.fileStorage.SaveFile(ctx, userID, filename, fileData)
	if err != nil {
		// SaveFile 内部已经记录了错误日志
		return nil, "", err // 错误已经被包装
	}

	// 2. 创建并保存文件元数据
	doc := entity.NewDocument(userID, filename, storedPath, fileSize, contentType)
	// 可以在这里计算文件哈希用于去重 (可选)
	// doc.FileHash = calculateHash(storedPath)
	// existingDoc, _ := s.docRepo.GetDocumentByHash(ctx, userID, doc.FileHash)
	// if existingDoc != nil { /* 处理重复文件逻辑 */ }

	if err := s.docRepo.SaveDocument(ctx, doc); err != nil {
		// 如果保存元数据失败，尝试删除已上传的文件以保持一致性
		logger.WarnContext(ctx, "保存文档元数据失败，尝试回滚删除已上传的文件", "error", err, "stored_path", storedPath)
		if delErr := s.fileStorage.DeleteFile(ctx, storedPath); delErr != nil {
			logger.ErrorContext(ctx, "回滚删除文件失败", "delete_error", delErr, "original_error", err, "stored_path", storedPath)
			// 返回原始的保存错误，但记录了删除失败
		}
		return nil, "", err // 返回保存元数据的错误
	}

	// 3. 将 Embedding 任务入队
	// 准备任务 payload
	taskPayload := map[string]interface{}{
		"user_id":      userID, // Use the passed userID string
		"document_id":  doc.ID, // Assuming doc.ID is now string UUID
		"file_path":    storedPath,
		"filename":     filename,
		"content_type": contentType,
	}

	// 使用 TaskQueueClient 入队
	// TODO: 从 queue 包引入 TypeEmbedding 常量
	taskID, err := s.taskQueue.EnqueueEmbeddingTask(ctx, taskPayload)
	if err != nil {
		// 如果入队失败，这是一个严重问题，可能需要标记文档状态为错误
		// 或者尝试回滚数据库记录和文件删除 (更复杂)
		logger.ErrorContext(ctx, "将 Embedding 任务入队失败", "error", err, "document_id", doc.ID)
		// 尝试更新文档状态为错误 (assuming doc.ID is string)
		// Call UpdateDocumentStatus with correct signature (taskID is nil here)
		updateErr := s.docRepo.UpdateDocumentStatus(ctx, userID, doc.ID, entity.TaskStatusFailed, nil, "Failed to enqueue processing task")
		if updateErr != nil {
			logger.ErrorContext(ctx, "入队失败后更新文档状态也失败", "update_error", updateErr, "original_error", err, "document_id", doc.ID, "user_id", userID) // Log string doc.ID and userID
		}
		// 返回入队错误
		return doc, "", err // 返回文档信息和入队错误
	}

	// 4. (可选) 创建 Task 实体并保存到数据库 (如果需要持久化任务状态)
	// 注意：Asynq 自身会管理任务状态，这里创建 Task 实体是为了 API 查询
	// Asynq ID 通常不是 UUID，这里需要调整 Task 实体的 ID 处理方式。
	// 我们可以使用 Document ID 作为 Task ID 的一部分，或者让 TaskRepository 生成 ID。
	// 暂时不创建 Task 实体，依赖 Asynq 的状态管理，并通过 Document 关联。
	// var taskEntityID *uuid.UUID = nil // Task ID is likely string now, or not used here
	// var taskEntityID *string = nil // REMOVED: Unused variable

	// 5. 更新文档状态为 Pending 并关联 Task ID (如果 Task 实体创建成功)
	// 由于 Asynq 返回的 taskID 不是 UUID，我们暂时不直接关联 Task ID 到 Document。
	// TODO: Check signature of UpdateDocumentStatus, it might expect *uuid.UUID for task ID
	// Worker 在处理时可以通过 payload 中的 document_id 来更新 Document 状态。
	// 我们只更新状态为 Pending，并关联 Asynq 返回的 taskID。
	// Pass string doc.ID, userID, and the actual taskID from the queue
	var returnedTaskIDPtr *string
	if taskID != "" {
		returnedTaskIDPtr = &taskID // Use the taskID returned from EnqueueEmbeddingTask
	}
	if updateErr := s.docRepo.UpdateDocumentStatus(ctx, userID, doc.ID, entity.TaskStatusPending, returnedTaskIDPtr, ""); updateErr != nil {
		logger.ErrorContext(ctx, "更新文档状态为 Pending 并关联 TaskID 失败", "error", updateErr, "document_id", doc.ID, "user_id", userID, "task_id", taskID) // Log string doc.ID and userID
		// 记录错误，但不阻塞返回
	}

	logger.InfoContext(ctx, "文件上传处理完成，任务已入队", "document_id", doc.ID, "task_id", taskID) // Log string doc.ID
	return doc, taskID, nil
}

// GetDocument 获取文档元数据。
// Added userID string, changed docID to string
func (s *fileServiceImpl) GetDocument(ctx context.Context, userID string, docID string) (*entity.Document, error) {
	// Pass userID and string docID explicitly to repository
	// TODO: Update docRepo.GetDocumentByID signature to accept userID and string docID
	doc, err := s.docRepo.GetDocumentByID(ctx, userID, docID) // Assuming repo method is updated
	if err != nil {
		// GetDocumentByID 内部应记录日志和包装错误
		return nil, err
	}
	return doc, nil
}

// ListUserDocuments 列出用户文档。
func (s *fileServiceImpl) ListUserDocuments(ctx context.Context, userID string, limit int, offset int) ([]*entity.Document, error) {
	// GetDocumentsByUser 内部会根据 ctx 中的 user_id 过滤和验证
	docs, err := s.docRepo.GetDocumentsByUser(ctx, userID, limit, offset)
	if err != nil {
		// GetDocumentsByUser 内部已记录日志和包装错误
		return nil, err
	}
	return docs, nil
}

// DeleteDocument 删除文档及其关联数据。
// Added userID string, changed docID to string
func (s *fileServiceImpl) DeleteDocument(ctx context.Context, userID string, docID string) error {
	// userID is now passed explicitly.
	// userID, err := postgres.GetUserIDFromCtx(ctx) // REMOVED
	// if err != nil {
	// 	return err
	// }

	// 1. 获取文档信息，特别是存储路径，并验证用户权限
	// Pass userID and string docID explicitly to repository
	// TODO: Update docRepo.GetDocumentByID signature to accept userID and string docID
	doc, err := s.docRepo.GetDocumentByID(ctx, userID, docID) // Assuming repo method is updated
	if err != nil {
		return err // Not found or other error
	}
	// Assuming GetDocumentByID now performs the user check

	// 2. 删除向量数据 (VectorRepository)
	// 即使失败也继续尝试删除其他数据，但记录错误
	var vectorErr error
	// Pass userID and string docID to vectorRepo
	if err := s.vectorRepo.DeleteChunksByDocumentID(ctx, userID, doc.ID); err != nil { // Use doc.ID (string)
		vectorErr = err                                                                                   // 保存错误稍后处理
		logger.ErrorContext(ctx, "删除文档关联的向量数据失败", "error", err, "document_id", doc.ID, "user_id", userID) // Log string doc.ID
		// 不立即返回，继续删除文件和元数据
	}

	// 3. 删除物理文件 (FileStorage)
	// 即使失败也继续尝试删除其他数据，但记录错误
	var fileErr error
	if doc.StoredPath != "" {
		if err := s.fileStorage.DeleteFile(ctx, doc.StoredPath); err != nil {
			// 忽略文件未找到的错误，因为可能已被删除或从未成功保存
			if !apperr.Is(err, apperr.CodeNotFound) {
				fileErr = err // 保存错误稍后处理
				logger.ErrorContext(ctx, "删除物理文件失败", "error", err, "document_id", docID, "user_id", userID, "path", doc.StoredPath)
			} else {
				logger.WarnContext(ctx, "尝试删除的物理文件未找到（可能已被删除）", "document_id", docID, "user_id", userID, "path", doc.StoredPath)
			}
		}
	}

	// 4. 删除文档元数据 (DocumentRepository)
	// 这是最后一步，如果之前的步骤有失败，这个操作也可能失败
	// Pass userID and string docID explicitly to repository
	// TODO: Update docRepo.DeleteDocument signature to accept userID and string docID
	metaErr := s.docRepo.DeleteDocument(ctx, userID, doc.ID) // Assuming repo method is updated
	if metaErr != nil {
		logger.ErrorContext(ctx, "删除文档元数据失败", "error", metaErr, "document_id", doc.ID, "user_id", userID) // Log string doc.ID
		// 如果元数据删除失败，之前的删除操作可能部分成功，状态不一致
		// 返回元数据删除错误，因为它阻止了记录的完全清除
		return metaErr
	}

	// 5. 处理之前的错误
	// 如果向量或文件删除失败，但元数据删除成功，我们应该报告哪个错误？
	// 优先报告向量删除错误，因为它可能更关键
	if vectorErr != nil {
		return vectorErr // 返回向量删除错误
	}
	if fileErr != nil {
		return fileErr // 返回文件删除错误
	}

	// 6. (可选) 删除关联的任务记录 (TaskRepository)
	// 如果我们决定不持久化 Task 实体，则此步骤不需要。
	// 如果持久化了，需要找到与 docID 关联的 Task ID 并删除。
	// taskIDStr := ""
	// if doc.ProcessingTaskID != nil {
	//  taskIDStr = doc.ProcessingTaskID.String()
	//  if taskErr := s.taskRepo.DeleteTask(ctx, *doc.ProcessingTaskID); taskErr != nil {
	//      logger.ErrorContext(ctx, "删除关联的任务记录失败", "error", taskErr, "task_id", taskIDStr, "document_id", docID)
	//      // 通常不阻塞删除流程
	//  }
	// }

	logger.InfoContext(ctx, "文档及其关联数据删除成功", "document_id", docID, "user_id", userID)
	return nil
}

// GetTaskStatus 获取异步任务的状态。
// Added userID string parameter
func (s *fileServiceImpl) GetTaskStatus(ctx context.Context, userID string, taskID string) (*entity.Task, error) {
	// userID is now passed explicitly.
	// 注意：taskID 来自 Asynq，不是我们 Task 实体的 UUID。
	// 我们需要一种方法来通过 Asynq task ID 查询我们的 Task 实体状态，
	// 或者直接查询 Asynq 的任务状态（如果 Asynq Client API 支持）。
	// 目前的 TaskRepository 是基于 UUID 的。
	// 解决方案：
	// 1. 修改 Task 实体，使其 ID 可以是 string 类型，并存储 Asynq ID。
	// 2. 或者，在 TaskRepository 中添加 FindByAsynqID 方法。
	// 3. 或者，不持久化 Task 实体，直接查询 Asynq 状态（需要 Asynq Inspector）。

	// 暂时返回未实现错误，因为当前 TaskRepository 不支持按 Asynq ID 查询。
	// Use the passed userID in the log context
	logger.WarnContext(ctx, "GetTaskStatus 尚未完全实现以支持 Asynq Task ID 查询", "task_id", taskID, "user_id", userID)
	return nil, apperr.New(apperr.CodeUnimplemented, "按 Asynq 任务 ID 查询状态的功能尚未实现")

	/* 假设 TaskRepository 支持按 Asynq ID 查询 (伪代码)
	// Pass userID for potential authorization check in repo
	// TODO: Update taskRepo.GetTaskByAsynqID signature if implemented
	task, err := s.taskRepo.GetTaskByAsynqID(ctx, userID, taskID) // Assuming repo method is updated
	if err != nil {
		return nil, err
	}
	// 可能需要权限检查
	// ctxUserID, _ := postgres.GetUserIDFromCtx(ctx)
	// if task.UserID != ctxUserID { ... return permission denied ... }
	return task, nil
	*/
}
