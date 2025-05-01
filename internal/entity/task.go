package entity

import (
	"encoding/json"
	"time"
)

// TaskStatus 定义了异步任务的状态。
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"    // 任务等待处理
	TaskStatusProcessing TaskStatus = "processing" // 任务正在处理中
	TaskStatusCompleted  TaskStatus = "completed"  // 任务成功完成
	TaskStatusFailed     TaskStatus = "failed"     // 任务处理失败
)

// Task 代表一个异步处理任务。
// 这可以用于跟踪文件处理（如 Embedding）等后台操作的状态。
// 这个结构可以映射到一个数据库表，用于持久化任务状态。
type Task struct {
	ID               string          `json:"id"`                // 任务的唯一 ID (string, e.g., Asynq ID or generated UUID string)
	Type             string          `json:"type"`              // 任务类型 (e.g., "embedding:generate")
	Payload          json.RawMessage `json:"payload"`           // 任务的输入数据 (JSON)
	Status           TaskStatus      `json:"status"`            // 任务当前状态
	UserID           string          `json:"user_id"`           // 关联的用户 ID (用于数据隔离和通知)
	FileID           *string         `json:"file_id"`           // 关联的文件 ID (string UUID, if task relates to a file)
	OriginalFilename string          `json:"original_filename"` // 原始文件名 (方便追踪)
	Progress         float64         `json:"progress"`          // 处理进度 (0.0 to 100.0)
	Result           json.RawMessage `json:"result"`            // 任务成功时的结果 (JSON)
	ErrorMessage     string          `json:"error_message"`     // 任务失败时的错误信息
	RetryCount       int             `json:"retry_count"`       // 重试次数
	MaxRetries       int             `json:"max_retries"`       // 最大允许重试次数
	CreatedAt        time.Time       `json:"created_at"`        // 任务创建时间
	StartedAt        *time.Time      `json:"started_at"`        // 任务开始处理时间
	CompletedAt      *time.Time      `json:"completed_at"`      // 任务完成时间
	UpdatedAt        time.Time       `json:"updated_at"`        // 任务最后更新时间
	// 可以添加 Priority, Queue 等字段
}

// NewTask 创建一个新的 Task 实例。
// taskID is now string
func NewTask(taskID string, taskType string, userID string, payload map[string]interface{}) (*Task, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err // 或者返回 apperr.Wrap
	}

	// Extract file_id from payload if it exists and is a string
	var fileIDPtr *string
	if fileIDVal, ok := payload["document_id"]; ok {
		if fileIDStr, ok := fileIDVal.(string); ok {
			fileIDPtr = &fileIDStr
		}
	}
	// Extract original_filename from payload if it exists
	var originalFilename string
	if filenameVal, ok := payload["filename"]; ok {
		if filenameStr, ok := filenameVal.(string); ok {
			originalFilename = filenameStr
		}
	}

	return &Task{
		ID:               taskID,
		Type:             taskType,
		Payload:          payloadBytes,
		Status:           TaskStatusPending, // 初始状态为 Pending
		UserID:           userID,
		FileID:           fileIDPtr,        // Set FileID from payload
		OriginalFilename: originalFilename, // Set OriginalFilename from payload
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		MaxRetries:       3, // 默认最大重试次数
	}, nil
}

// SetPayload 设置任务的 Payload。
func (t *Task) SetPayload(data map[string]interface{}) error {
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	t.Payload = payloadBytes
	t.UpdatedAt = time.Now()
	return nil
}

// GetPayload 将 Payload 解析到提供的 map 中。
func (t *Task) GetPayload(target map[string]interface{}) error {
	if t.Payload == nil {
		return nil // 没有 Payload
	}
	return json.Unmarshal(t.Payload, &target)
}

// SetResult 设置任务成功的结果。
func (t *Task) SetResult(data map[string]interface{}) error {
	resultBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	t.Result = resultBytes
	t.Status = TaskStatusCompleted
	now := time.Now()
	t.CompletedAt = &now
	t.UpdatedAt = now
	return nil
}

// GetResult 将 Result 解析到提供的 map 中。
func (t *Task) GetResult(target map[string]interface{}) error {
	if t.Result == nil {
		return nil // 没有结果
	}
	return json.Unmarshal(t.Result, &target)
}

// MarkAsProcessing 标记任务为正在处理。
func (t *Task) MarkAsProcessing() {
	if t.Status == TaskStatusPending || t.Status == TaskStatusFailed { // 允许从 Pending 或 Failed 状态重试
		t.Status = TaskStatusProcessing
		now := time.Now()
		if t.StartedAt == nil { // 只记录第一次开始的时间
			t.StartedAt = &now
		}
		t.UpdatedAt = now
		t.ErrorMessage = "" // 清除之前的错误信息
	}
}

// MarkAsFailed 标记任务为失败。
func (t *Task) MarkAsFailed(errMsg string) {
	t.Status = TaskStatusFailed
	t.ErrorMessage = errMsg
	now := time.Now()
	t.CompletedAt = &now // 失败也算完成处理周期
	t.UpdatedAt = now
	t.RetryCount++
}

// UpdateProgress 更新任务进度。
func (t *Task) UpdateProgress(progress float64) {
	if t.Status == TaskStatusProcessing {
		if progress < 0 {
			progress = 0
		}
		if progress > 100 {
			progress = 100
		}
		t.Progress = progress
		t.UpdatedAt = time.Now()
	}
}

// CanRetry 检查任务是否还可以重试。
func (t *Task) CanRetry() bool {
	return t.Status == TaskStatusFailed && t.RetryCount < t.MaxRetries
}
