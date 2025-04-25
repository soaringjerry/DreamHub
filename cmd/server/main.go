package main

import (
	"context" // 引入 json 包用于序列化任务载荷
	"fmt"

	// "log" // 使用 pkg/logger 替换
	"net/http"
	"os"

	// "strings" // Will be unused after refactor
	// "time" // Will be unused after refactor

	"github.com/gin-gonic/gin"
	// "github.com/google/uuid" // Will be unused after refactor
	"github.com/jackc/pgx/v5/pgxpool" // 引入 pgx 连接池 (needed for ChatService init)
	// "github.com/joho/godotenv"           // 不再直接使用，由 config 包处理
	// 引入 embeddings 接口包 (needed for AppContext embedder field) - TODO: Refactor embedder wrapper
	// Needed for ChatService init
	"github.com/tmc/langchaingo/llms/openai" // 引入 openai 实现包

	// "github.com/tmc/langchaingo/schema" // Will be unused after refactor

	// 引入 textsplitter 包
	// 引入 vectorstores 包 (用于过滤选项)
	"github.com/tmc/langchaingo/vectorstores/pgvector" // 引入 pgvector 包

	"github.com/hibiken/asynq"                      // 引入 asynq 任务队列包
	"github.com/soaringjerry/dreamhub/internal/api" // 引入 API Handler 包

	// 引入内部实体包
	// 引入仓库接口包
	repoPgvector "github.com/soaringjerry/dreamhub/internal/repository/pgvector" // 引入 pgvector 仓库实现包 (带别名)
	"github.com/soaringjerry/dreamhub/internal/service"                          // 引入服务包
	"github.com/soaringjerry/dreamhub/pkg/config"                                // 引入配置包
	"github.com/soaringjerry/dreamhub/pkg/logger"                                // 引入日志包
)

// Constants moved to config
// const uploadDir = "uploads"
// const maxHistoryMessages = 10

// ChatMessage 定义接收聊天消息的结构体
// ChatMessage 和 ChatResponse 已移动到 internal/entity/chat.go

// AppContext is no longer needed as dependencies are injected directly.
// type AppContext struct { ... }

// --- Wrapper to satisfy embeddings.Embedder interface ---
// TODO: Move this wrapper to a shared internal package
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
		return nil, fmt.Errorf("no embedding returned for query")
	}
	return embeddings[0], nil
}

// --- End Wrapper ---

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		// Logger might not be fully initialized here if config loading fails early
		// Use standard log for this specific critical error
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger (potentially reconfigure based on cfg.Environment)
	// For now, logger is initialized in its init()

	ctx := context.Background()

	// Dependencies are now loaded from cfg
	// openaiAPIKey := cfg.OpenAIAPIKey
	// databaseURL := cfg.DatabaseURL
	// redisAddr := cfg.RedisAddr

	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("无法连接到数据库", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()
	logger.Info("数据库连接池初始化成功")

	// Use config values for OpenAI client
	client, err := openai.New(
		openai.WithModel(cfg.OpenAIModel),
		openai.WithEmbeddingModel(cfg.OpenAIEmbeddingModel),
		openai.WithToken(cfg.OpenAIAPIKey),
	)
	if err != nil {
		logger.Error("初始化 OpenAI 客户端失败", "error", err)
		os.Exit(1)
	}
	logger.Info("OpenAI 客户端初始化成功")

	embedderWrapper := &openAIEmbedderWrapper{client: client}

	// Use config value for DatabaseURL
	vectorStore, err := pgvector.New(
		ctx,
		pgvector.WithConnectionURL(cfg.DatabaseURL),
		pgvector.WithEmbedder(embedderWrapper),
		pgvector.WithCollectionName("documents"), // Match worker's collection name
	)
	if err != nil {
		logger.Error("初始化 pgvector 存储失败", "error", err)
		os.Exit(1)
	}
	logger.Info("pgvector 存储初始化成功")

	// 创建 DocumentRepository 实例
	docRepo := repoPgvector.New(vectorStore)

	// 创建 Asynq Client 实例 using config value
	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword, // Add Redis password from config
	}
	asynqClient := asynq.NewClient(redisOpt)
	defer asynqClient.Close()
	logger.Info("Asynq Client 初始化成功", "redisAddr", cfg.RedisAddr)
	// --- Instantiate Services ---
	fileService := service.NewFileService(asynqClient)
	// Instantiate ChatService (Injecting llm client directly, not the wrapper)
	chatService := service.NewChatService(dbPool, client, docRepo)

	// --- Instantiate Handlers ---
	uploadHandler := api.NewUploadHandler(fileService)
	chatHandler := api.NewChatHandler(chatService)

	// AppContext is removed.

	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	apiV1 := router.Group("/api/v1")
	{
		// Use the new handler methods
		apiV1.POST("/upload", uploadHandler.HandleUpload)
		apiV1.POST("/chat", chatHandler.HandleChat)
	}
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// Use config value for server port
	serverAddr := ":" + cfg.ServerPort
	logger.Info("Server starting", "address", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		logger.Error("Failed to run server", "error", err)
		os.Exit(1)
	}
}

// handleUpload function removed as its logic is now in api.UploadHandler and service.FileService

// handleChat function and its helpers (loadConversationHistory, saveMessageToHistory, buildLLMMessages)
// removed as their logic is now in api.ChatHandler and service.ChatService
