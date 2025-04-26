package queue

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/soaringjerry/dreamhub/internal/service" // 引入 service 包以引用接口
	"github.com/soaringjerry/dreamhub/pkg/apperr"
	"github.com/soaringjerry/dreamhub/pkg/config"
	"github.com/soaringjerry/dreamhub/pkg/logger"
)

// 定义任务类型常量，应与 Worker 端保持一致
// TODO: 将这些常量移到共享位置 (e.g., internal/tasks)
const (
	TypeEmbedding = "embedding:generate"
)

// asynqClient 是 TaskQueueClient 接口的 Asynq 实现。
type asynqClient struct {
	client *asynq.Client
}

// NewAsynqClient 创建一个新的 Asynq 客户端实例。
func NewAsynqClient(cfg *config.Config) service.TaskQueueClient {
	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword, // Use the password from config
		// DB:       cfg.RedisDB,
	}
	client := asynq.NewClient(redisOpt)
	// 注意：这里没有显式的 Connect 方法，连接是在第一次操作时建立的。
	// 可以考虑添加一个 Ping 或 Info 调用来验证连接，但这通常不是必需的。
	logger.Info("Asynq 客户端初始化完成。", "redis_addr", cfg.RedisAddr)
	return &asynqClient{client: client}
}

// EnqueueEmbeddingTask 将生成 Embedding 的任务放入 Asynq 队列。
func (c *asynqClient) EnqueueEmbeddingTask(ctx context.Context, payload map[string]interface{}) (taskID string, err error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.ErrorContext(ctx, "序列化 Embedding 任务 payload 失败", "error", err)
		return "", apperr.Wrap(err, apperr.CodeInternal, "无法序列化任务 payload")
	}

	// 创建一个新的 Asynq 任务
	// TypeEmbedding 是任务类型，payloadBytes 是任务数据
	task := asynq.NewTask(TypeEmbedding, payloadBytes)

	// 将任务入队
	// 可以指定队列名称和选项，例如延迟、重试次数等
	// asynq.Queue("embeddings"), asynq.MaxRetry(5), asynq.Timeout(10*time.Minute)
	taskInfo, err := c.client.EnqueueContext(ctx, task)
	if err != nil {
		logger.ErrorContext(ctx, "将 Embedding 任务入队失败", "error", err)
		// 考虑根据错误类型返回不同的 AppError Code (e.g., CodeUnavailable if Redis is down)
		return "", apperr.Wrap(err, apperr.CodeUnavailable, "无法将任务入队")
	}

	logger.InfoContext(ctx, "Embedding 任务成功入队", "task_id", taskInfo.ID, "queue", taskInfo.Queue)
	return taskInfo.ID, nil
}

// Close 关闭 Asynq 客户端连接。
func (c *asynqClient) Close() error {
	if c.client != nil {
		logger.Info("正在关闭 Asynq 客户端...")
		err := c.client.Close()
		if err != nil {
			logger.Error("关闭 Asynq 客户端失败", "error", err)
			return err
		}
		logger.Info("Asynq 客户端已关闭。")
	}
	return nil
}
