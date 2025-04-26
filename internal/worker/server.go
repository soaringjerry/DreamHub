package worker

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/soaringjerry/dreamhub/internal/entity"
	"github.com/soaringjerry/dreamhub/internal/repository"
	"github.com/soaringjerry/dreamhub/pkg/logger"
	"github.com/tmc/langchaingo/embeddings"
)

// RunServer starts the Asynq worker server.
// It takes the Redis connection options, worker concurrency, and necessary dependencies for handlers.
func RunServer(redisOpt asynq.RedisConnOpt, concurrency int, docRepo repository.DocumentRepository, vectorRepo repository.VectorRepository, embedder embeddings.Embedder) error {
	// Create a new Asynq server instance.
	if concurrency <= 0 {
		concurrency = 10 // Default concurrency if invalid value provided
		logger.Warn("RunServer: Invalid concurrency value provided, using default", "defaultConcurrency", concurrency)
	}
	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: concurrency, // Use concurrency from config
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				retried, _ := asynq.GetRetryCount(ctx)
				maxRetry, _ := asynq.GetMaxRetry(ctx)
				taskID, _ := asynq.GetTaskID(ctx) // Get task ID from context if available
				logger.Error("Worker: Task processing failed",
					"taskID", taskID, // Use task ID from context
					"type", task.Type(),
					"payload", string(task.Payload()), // Log payload cautiously
					"error", err,
					"retry", retried,
					"maxRetry", maxRetry,
				)
			}),
			// Queues specifies the list of queues to process tasks from, and their priorities.
			// Queues: map[string]int{
			//     "critical": 6,
			//     "default":  3,
			//     "low":      1,
			// },
			// Logger: logger.Logger, // TODO: Adapt pkg/logger to asynq.Logger interface if needed
		},
	)

	// Create a new ServeMux to route tasks to handlers.
	mux := asynq.NewServeMux()

	// Create handlers with their dependencies.
	embeddingHandler := NewEmbeddingTaskHandler(docRepo, vectorRepo, embedder) // Pass vectorRepo

	// Register handlers for specific task types.
	mux.Handle(entity.TaskTypeEmbedding, embeddingHandler)
	// mux.HandleFunc(AnotherTaskType, HandleAnotherTaskFunc) // Example for function handlers

	logger.Info("Starting Asynq worker server...")

	// Start the server.
	// Run blocks until the server is shut down.
	if err := srv.Run(mux); err != nil {
		return fmt.Errorf("could not run Asynq server: %w", err)
	}

	logger.Info("Asynq worker server shut down.")
	return nil
}
