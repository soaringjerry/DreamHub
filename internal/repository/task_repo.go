package repository

import (
	"context"

	"github.com/soaringjerry/dreamhub/internal/entity"
)

// TaskRepository 定义了与异步任务状态存储交互的方法。
// 这通常对应于关系型数据库中的 tasks 表。
type TaskRepository interface {
	// CreateTask 创建一个新的任务记录。
	CreateTask(ctx context.Context, task *entity.Task) error

	// GetTaskByID 根据 ID 获取任务信息。
	// 需要确保实现中根据 ctx 中的 user_id 进行了过滤 (如果需要按用户隔离任务视图)。
	// taskID is now string
	GetTaskByID(ctx context.Context, taskID string) (*entity.Task, error)

	// UpdateTaskStatus 更新任务的状态、进度和错误信息。
	// taskID is now string
	UpdateTaskStatus(ctx context.Context, taskID string, status entity.TaskStatus, progress float64, errMsg string) error

	// UpdateTaskResult 更新任务成功时的结果。
	// taskID is now string
	UpdateTaskResult(ctx context.Context, taskID string, result map[string]interface{}) error

	// IncrementRetryCount 增加任务的重试次数。
	// taskID is now string
	IncrementRetryCount(ctx context.Context, taskID string) error

	// GetPendingTasks 获取处于 Pending 状态的任务 (可能用于 Worker 恢复或检查)。
	// 可以添加过滤条件，例如按创建时间、优先级等。
	GetPendingTasks(ctx context.Context, limit int) ([]*entity.Task, error)

	// TODO: 可能需要添加其他方法，例如：
	// GetTasksByUser(ctx context.Context, userID string, limit int, offset int) ([]*entity.Task, error)
	// DeleteTask(ctx context.Context, taskID uuid.UUID) error
	// FindTasksByStatus(ctx context.Context, status entity.TaskStatus, limit int) ([]*entity.Task, error)
}
