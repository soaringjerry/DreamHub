package main

import (
	"context"
	"fmt" // 移到这里
	"os"

	"github.com/hibiken/asynq"
	// "github.com/joho/godotenv" // Not used directly, handled by config
	repoPgvector "github.com/soaringjerry/dreamhub/internal/repository/pgvector"
	"github.com/soaringjerry/dreamhub/internal/worker"
	"github.com/soaringjerry/dreamhub/pkg/config" // Import config package
	"github.com/soaringjerry/dreamhub/pkg/logger"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/vectorstores/pgvector"
)

// --- Copied Embedder Wrapper ---
// TODO: Move this wrapper to a shared internal package (e.g., internal/infra/openai)
// to avoid duplication between cmd/server and cmd/worker.
type openAIEmbedderWrapper struct {
	client *openai.LLM
}

func (w *openAIEmbedderWrapper) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	return w.client.CreateEmbedding(ctx, texts)
}
func (w *openAIEmbedderWrapper) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := w.client.CreateEmbedding(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	} // Simplified error
	return embeddings[0], nil
}

// --- End Copied Wrapper ---

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Worker: Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger (potentially reconfigure based on cfg.Environment)
	// For now, logger is initialized in its init()

	// --- Dependency Initialization ---
	ctx := context.Background()

	// Config values are now in cfg
	// openaiAPIKey := cfg.OpenAIAPIKey
	// databaseURL := cfg.DatabaseURL
	// redisAddr := cfg.RedisAddr

	// OpenAI Client (for Embedder)
	// TODO: Refactor OpenAI client creation into a shared function/package.
	// Use config values for OpenAI client specific to embedding
	llmClient, err := openai.New(
		// openai.WithModel(cfg.OpenAIModel), // Model might not be needed for embedding
		openai.WithEmbeddingModel(cfg.OpenAIEmbeddingModel),
		openai.WithToken(cfg.OpenAIAPIKey),
	)
	if err != nil {
		logger.Error("Worker: Failed to init OpenAI client", "error", err)
		os.Exit(1)
	}
	logger.Info("Worker: OpenAI client initialized")
	embedderWrapper := &openAIEmbedderWrapper{client: llmClient}

	// PGVector Store
	// TODO: Refactor PGVector store creation into a shared function/package.
	// Use config value for DatabaseURL
	vectorStore, err := pgvector.New(
		ctx,
		pgvector.WithConnectionURL(cfg.DatabaseURL),
		pgvector.WithEmbedder(embedderWrapper), // Pass the embedder wrapper
	)
	if err != nil {
		logger.Error("Worker: Failed to init pgvector store", "error", err)
		os.Exit(1)
	}
	logger.Info("Worker: pgvector store initialized")

	// Document Repository
	docRepo := repoPgvector.New(vectorStore)
	logger.Info("Worker: Document repository initialized")

	// Redis Connection Options for Asynq using config value
	redisOpt := asynq.RedisClientOpt{Addr: cfg.RedisAddr}

	// --- Start Worker Server ---
	logger.Info("Worker: Starting worker server...", "concurrency", cfg.WorkerConcurrency)
	// Pass concurrency from config to RunServer
	if err := worker.RunServer(redisOpt, cfg.WorkerConcurrency, docRepo, embedderWrapper); err != nil {
		logger.Error("Worker: Could not start worker server", "error", err)
		os.Exit(1)
	}
}

// Note: We are missing the *pgxpool.Pool dependency here, which might be needed
// if the worker needs to interact with the main database (e.g., updating task status).
// This needs to be added if required by future worker tasks or status updates.
// Also, the embedder wrapper and potentially client/store creation logic should be refactored
// into shared internal packages to avoid code duplication between cmd/server and cmd/worker.

// Need to import "fmt" for the simplified error message in EmbedQuery
// import "fmt" // Duplicate removed, already imported at the top
