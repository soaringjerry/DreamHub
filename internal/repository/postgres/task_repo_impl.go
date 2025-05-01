package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/soaringjerry/dreamhub/internal/entity"
	"github.com/soaringjerry/dreamhub/internal/repository"
	"github.com/soaringjerry/dreamhub/pkg/apperr"
	"github.com/soaringjerry/dreamhub/pkg/logger"
)

// postgresTaskRepository 是 TaskRepository 接口的 PostgreSQL 实现。
type postgresTaskRepository struct {
	db *DB // 嵌入 DB 连接池
}

// NewPostgresTaskRepository 创建一个新的 postgresTaskRepository 实例。
func NewPostgresTaskRepository(db *DB) repository.TaskRepository {
	return &postgresTaskRepository{db: db}
}

// CreateTask 创建一个新的任务记录到 tasks 表。
func (r *postgresTaskRepository) CreateTask(ctx context.Context, task *entity.Task) error {
	const sql = `
		INSERT INTO tasks (id, type, payload, status, user_id, file_id, original_filename, progress, result, error_message, retry_count, max_retries, created_at, started_at, completed_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`
	_, err := r.db.Pool.Exec(ctx, sql,
		task.ID,
		task.Type,
		task.Payload,
		task.Status,
		task.UserID,
		task.FileID,
		task.OriginalFilename,
		task.Progress,
		task.Result,
		task.ErrorMessage,
		task.RetryCount,
		task.MaxRetries,
		task.CreatedAt,
		task.StartedAt,
		task.CompletedAt,
		task.UpdatedAt,
	)
	if err != nil {
		logger.ErrorContext(ctx, "创建任务记录失败", "error", err, "task_id", task.ID, "type", task.Type)
		return apperr.Wrap(err, apperr.CodeInternal, "无法创建任务记录")
	}
	logger.InfoContext(ctx, "任务记录创建成功", "task_id", task.ID, "type", task.Type)
	return nil
}

// GetTaskByID 根据 ID 获取任务信息。
// 注意：目前未强制按 user_id 过滤，因为任务状态查询可能需要跨用户（例如管理员视图）。
// 如果需要用户隔离，应添加 user_id 条件。
// taskID is now string
func (r *postgresTaskRepository) GetTaskByID(ctx context.Context, taskID string) (*entity.Task, error) {
	const sql = `
		SELECT id, type, payload, status, user_id, file_id, original_filename, progress, result, error_message, retry_count, max_retries, created_at, started_at, completed_at, updated_at
		FROM tasks
		WHERE id = $1
	`
	// Pass string taskID
	row := r.db.Pool.QueryRow(ctx, sql, taskID)
	var task entity.Task
	err := row.Scan(
		&task.ID, &task.Type, &task.Payload, &task.Status, &task.UserID, &task.FileID, &task.OriginalFilename,
		&task.Progress, &task.Result, &task.ErrorMessage, &task.RetryCount, &task.MaxRetries,
		&task.CreatedAt, &task.StartedAt, &task.CompletedAt, &task.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			logger.WarnContext(ctx, "未找到指定的任务", "task_id", taskID)
			return nil, apperr.ErrNotFound("任务未找到")
		}
		logger.ErrorContext(ctx, "从数据库获取任务信息失败", "error", err, "task_id", taskID)
		return nil, apperr.Wrap(err, apperr.CodeInternal, "无法获取任务信息")
	}
	return &task, nil
}

// UpdateTaskStatus 更新任务的状态、进度、错误信息和更新时间。
// 如果状态是 Processing，则更新 StartedAt (如果尚未设置)。
// 如果状态是 Completed 或 Failed，则更新 CompletedAt。
// taskID is now string
func (r *postgresTaskRepository) UpdateTaskStatus(ctx context.Context, taskID string, status entity.TaskStatus, progress float64, errMsg string) error {
	now := time.Now()
	// var startedAtExpr, completedAtExpr string // Unused variables
	var args []interface{}

	sqlBase := `UPDATE tasks SET status = $1, progress = $2, error_message = $3, updated_at = $4`
	args = append(args, status, progress, errMsg, now)
	argCounter := 5 // Start from $5

	if status == entity.TaskStatusProcessing {
		// 仅在 started_at 为 NULL 时设置
		sqlBase += fmt.Sprintf(", started_at = COALESCE(started_at, $%d)", argCounter)
		args = append(args, now)
		argCounter++
	} else if status == entity.TaskStatusCompleted || status == entity.TaskStatusFailed {
		sqlBase += fmt.Sprintf(", completed_at = $%d", argCounter)
		args = append(args, now)
		argCounter++
	}

	sqlWhere := fmt.Sprintf(" WHERE id = $%d", argCounter)
	// Append string taskID
	args = append(args, taskID)

	sql := sqlBase + sqlWhere

	cmdTag, err := r.db.Pool.Exec(ctx, sql, args...)
	if err != nil {
		logger.ErrorContext(ctx, "更新任务状态失败", "error", err, "task_id", taskID, "status", status)
		return apperr.Wrap(err, apperr.CodeInternal, "无法更新任务状态")
	}

	if cmdTag.RowsAffected() == 0 {
		logger.WarnContext(ctx, "尝试更新不存在的任务状态", "task_id", taskID)
		return apperr.ErrNotFound("任务未找到或无法更新")
	}

	logger.InfoContext(ctx, "任务状态更新成功", "task_id", taskID, "status", status, "progress", progress)
	return nil
}

// UpdateTaskResult 更新任务成功时的结果，并将状态设置为 Completed。
// taskID is now string
func (r *postgresTaskRepository) UpdateTaskResult(ctx context.Context, taskID string, result map[string]interface{}) error {
	resultBytes, err := json.Marshal(result)
	if err != nil {
		logger.ErrorContext(ctx, "序列化任务结果失败", "error", err, "task_id", taskID)
		return apperr.Wrap(err, apperr.CodeInternal, "无法序列化任务结果")
	}

	now := time.Now()
	const sql = `
		UPDATE tasks
		SET status = $1, result = $2, error_message = '', progress = 100, completed_at = $3, updated_at = $3
		WHERE id = $4
	`
	// Pass string taskID
	cmdTag, err := r.db.Pool.Exec(ctx, sql, entity.TaskStatusCompleted, resultBytes, now, taskID)
	if err != nil {
		logger.ErrorContext(ctx, "更新任务结果失败", "error", err, "task_id", taskID)
		return apperr.Wrap(err, apperr.CodeInternal, "无法更新任务结果")
	}

	if cmdTag.RowsAffected() == 0 {
		logger.WarnContext(ctx, "尝试更新不存在的任务结果", "task_id", taskID)
		return apperr.ErrNotFound("任务未找到或无法更新")
	}

	logger.InfoContext(ctx, "任务结果更新成功", "task_id", taskID)
	return nil
}

// IncrementRetryCount 增加任务的重试次数和更新时间。
// taskID is now string
func (r *postgresTaskRepository) IncrementRetryCount(ctx context.Context, taskID string) error {
	const sql = `
		UPDATE tasks
		SET retry_count = retry_count + 1, updated_at = $1
		WHERE id = $2
	`
	// Pass string taskID
	cmdTag, err := r.db.Pool.Exec(ctx, sql, time.Now(), taskID)
	if err != nil {
		logger.ErrorContext(ctx, "增加任务重试次数失败", "error", err, "task_id", taskID)
		return apperr.Wrap(err, apperr.CodeInternal, "无法增加任务重试次数")
	}

	if cmdTag.RowsAffected() == 0 {
		logger.WarnContext(ctx, "尝试增加不存在任务的重试次数", "task_id", taskID)
		return apperr.ErrNotFound("任务未找到或无法更新")
	}
	// 不需要记录 Info 日志，因为这通常在 MarkAsFailed 之后调用
	return nil
}

// GetPendingTasks 获取处于 Pending 状态的任务，按创建时间升序排列。
func (r *postgresTaskRepository) GetPendingTasks(ctx context.Context, limit int) ([]*entity.Task, error) {
	const sql = `
		SELECT id, type, payload, status, user_id, file_id, original_filename, progress, result, error_message, retry_count, max_retries, created_at, started_at, completed_at, updated_at
		FROM tasks
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
	`
	rows, err := r.db.Pool.Query(ctx, sql, entity.TaskStatusPending, limit)
	if err != nil {
		logger.ErrorContext(ctx, "获取待处理任务失败", "error", err)
		return nil, apperr.Wrap(err, apperr.CodeInternal, "无法获取待处理任务")
	}
	defer rows.Close()

	tasks := make([]*entity.Task, 0)
	for rows.Next() {
		var task entity.Task
		err := rows.Scan( // Corrected: use rows.Scan instead of row.Scan
			&task.ID, &task.Type, &task.Payload, &task.Status, &task.UserID, &task.FileID, &task.OriginalFilename,
			&task.Progress, &task.Result, &task.ErrorMessage, &task.RetryCount, &task.MaxRetries,
			&task.CreatedAt, &task.StartedAt, &task.CompletedAt, &task.UpdatedAt,
		)
		if err != nil {
			logger.ErrorContext(ctx, "扫描待处理任务行失败", "error", err)
			return nil, apperr.Wrap(err, apperr.CodeInternal, "处理数据库结果时出错")
		}
		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		logger.ErrorContext(ctx, "处理待处理任务结果集时出错", "error", err)
		return nil, apperr.Wrap(err, apperr.CodeInternal, "处理数据库结果时出错")
	}

	return tasks, nil
}
