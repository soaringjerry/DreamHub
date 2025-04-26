package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/soaringjerry/dreamhub/internal/repository/pgvector" // Import pgvector repo impl
	"github.com/soaringjerry/dreamhub/internal/repository/postgres" // Import postgres repo impl
	"github.com/soaringjerry/dreamhub/internal/service/embedding"   // Import embedding provider impl
	"github.com/soaringjerry/dreamhub/internal/service/storage"     // Import storage impl
	"github.com/soaringjerry/dreamhub/internal/worker/handlers"     // Import handlers
	"github.com/soaringjerry/dreamhub/pkg/config"                   // 导入 config 包
	"github.com/soaringjerry/dreamhub/pkg/logger"                   // 导入 logger 包
	"github.com/tmc/langchaingo/textsplitter"                       // Import textsplitter
)

// TODO: 将任务类型定义移到更合适的位置 (e.g., internal/tasks or internal/entity)
const (
	// TypeEmbedding 是用于生成 embedding 的任务类型。
	TypeEmbedding = "embedding:generate"
)

func main() {
	// --- 1. Initialization Phase ---
	logger.Info("Worker 初始化开始...")
	ctx, cancel := context.WithCancel(context.Background()) // Context for initialization and shutdown signals
	defer cancel()

	// Load Config
	cfg := config.LoadConfig()
	logger.Info("Worker 配置加载完成。", "redis", cfg.RedisAddr, "concurrency", cfg.WorkerConcurrency, "logLevel", cfg.LogLevel)

	// Initialize Logger
	logger.Info("Worker 日志系统初始化完成。")

	// Initialize Database Connection Pool
	dbPool, err := postgres.NewDB(ctx, cfg)
	if err != nil {
		logger.Error("数据库连接池初始化失败", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// Initialize File Storage
	fileStorage, err := storage.NewLocalStorage(cfg)
	if err != nil {
		logger.Error("本地文件存储初始化失败", "error", err)
		os.Exit(1)
	}

	// Initialize Embedding Provider
	embeddingProvider, err := embedding.NewOpenAIEmbeddingProvider(cfg)
	if err != nil {
		logger.Error("Embedding Provider 初始化失败", "error", err)
		os.Exit(1)
	}

	// Initialize Repositories
	docRepo := postgres.NewPostgresDocumentRepository(dbPool)
	vectorRepo := pgvector.NewPGVectorRepository(dbPool)
	taskRepo := postgres.NewPostgresTaskRepository(dbPool) // Initialize TaskRepo

	// Initialize Text Splitter
	// TODO: Read ChunkSize and ChunkOverlap from config if defined
	chunkSize := 1000   // Default chunk size
	chunkOverlap := 200 // Default chunk overlap
	// if cfg.SplitterChunkSize > 0 { chunkSize = cfg.SplitterChunkSize }
	// if cfg.SplitterChunkOverlap >= 0 { chunkOverlap = cfg.SplitterChunkOverlap }
	textSplitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(chunkSize),
		textsplitter.WithChunkOverlap(chunkOverlap),
	)
	logger.Info("文本分割器初始化完成。", "chunk_size", chunkSize, "chunk_overlap", chunkOverlap)

	// Initialize Task Handler
	embeddingHandler := handlers.NewEmbeddingTaskHandler(
		fileStorage,
		docRepo,
		vectorRepo,
		taskRepo, // Pass TaskRepo
		embeddingProvider,
		textSplitter, // Pass TextSplitter
	)

	logger.Info("Worker 依赖初始化完成。")

	// --- 2. Setup Asynq Server ---
	redisConnOpt := asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword, // Uncommented to include password
		// DB:       cfg.RedisDB,
	}

	srv := asynq.NewServer(
		redisConnOpt,
		asynq.Config{
			Concurrency: cfg.WorkerConcurrency,
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				logger.ErrorContext(ctx, "Worker 处理任务失败",
					"type", task.Type(),
					"error", err,
				)
			}),
			// Logger: logger.NewAsynqLogger(logger.GetLogger()), // Optional custom logger
		},
	)

	// --- 3. Register Task Handlers ---
	mux := asynq.NewServeMux()
	// Register the actual handler
	mux.Handle(TypeEmbedding, embeddingHandler)
	// Register other handlers here...
	logger.Info("Asynq 任务处理器注册完成。")

	// --- 4. Start Asynq Server ---
	logger.Info("Worker 准备启动...", "concurrency", cfg.WorkerConcurrency)
	if err := srv.Start(mux); err != nil {
		logger.Error("无法启动 Worker 服务器", "error", err)
		cancel() // Signal shutdown
		os.Exit(1)
	}

	// --- 5. Graceful Shutdown ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logger.Info("收到关闭信号", "signal", sig.String())
	case <-ctx.Done():
		logger.Info("上下文取消，开始关闭 Worker...")
	}

	logger.Info("正在优雅地关闭 Worker...")
	// Asynq's Shutdown waits for active tasks to complete.
	srv.Shutdown()
	logger.Info("Worker 已成功关闭。")
}

// Placeholder function removed as it's no longer used.

// TODO: Implement Asynq logger adapter if needed.
