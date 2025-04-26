package main

import (
	"context"
	"errors" // Import errors package for As
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings" // Import strings package for NoRoute check
	"syscall"
	"time"

	"github.com/gin-contrib/static" // Import static file serving middleware
	"github.com/gin-gonic/gin"
	"github.com/soaringjerry/dreamhub/internal/api" // Import API handlers
	"github.com/soaringjerry/dreamhub/internal/repository/pgvector"
	"github.com/soaringjerry/dreamhub/internal/repository/postgres" // Import Postgres implementations
	"github.com/soaringjerry/dreamhub/internal/service"             // Import Service implementations
	"github.com/soaringjerry/dreamhub/internal/service/embedding"   // Import Embedding provider implementation
	"github.com/soaringjerry/dreamhub/internal/service/llm"         // Import LLM provider implementation
	"github.com/soaringjerry/dreamhub/internal/service/queue"       // Import Queue client implementation
	"github.com/soaringjerry/dreamhub/internal/service/storage"     // Import Storage implementation
	"github.com/soaringjerry/dreamhub/pkg/apperr"                   // Import apperr
	"github.com/soaringjerry/dreamhub/pkg/config"                   // 导入 config 包
	"github.com/soaringjerry/dreamhub/pkg/logger"                   // 导入 logger 包
)

func main() {
	// --- 1. Initialization Phase ---
	logger.Info("服务器初始化开始...")
	ctx, cancel := context.WithCancel(context.Background()) // Context for initialization and shutdown signals
	defer cancel()                                          // Ensure cancel is called eventually

	// Load Config
	cfg := config.LoadConfig()
	logger.Info("配置加载完成。", "port", cfg.ServerPort, "redis", cfg.RedisAddr, "logLevel", cfg.LogLevel)

	// Initialize Logger (already done implicitly by first Info call)
	logger.Info("日志系统初始化完成。")

	// Initialize Database Connection Pool
	dbPool, err := postgres.NewDB(ctx, cfg)
	if err != nil {
		logger.Error("数据库连接池初始化失败", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close() // Ensure pool is closed on exit

	// Initialize Asynq Client (Task Queue)
	taskQueueClient := queue.NewAsynqClient(cfg)
	// defer taskQueueClient.Close() // Add Close method to interface and call here if needed

	// Initialize File Storage
	fileStorage, err := storage.NewLocalStorage(cfg)
	if err != nil {
		logger.Error("本地文件存储初始化失败", "error", err)
		os.Exit(1)
	}

	// Initialize LLM Provider
	llmProvider, err := llm.NewOpenAIProvider(cfg)
	if err != nil {
		logger.Error("LLM Provider 初始化失败", "error", err)
		os.Exit(1)
	}

	// Initialize Embedding Provider
	embeddingProvider, err := embedding.NewOpenAIEmbeddingProvider(cfg)
	if err != nil {
		logger.Error("Embedding Provider 初始化失败", "error", err)
		os.Exit(1)
	}

	// Initialize Repositories
	chatRepo := postgres.NewPostgresChatRepository(dbPool)
	docRepo := postgres.NewPostgresDocumentRepository(dbPool)
	vectorRepo := pgvector.NewPGVectorRepository(dbPool)
	taskRepo := postgres.NewPostgresTaskRepository(dbPool)

	// Initialize Services
	ragService := service.NewRAGService(vectorRepo, embeddingProvider) // Initialize RAGService
	// TODO: Initialize MemoryService when available
	chatService := service.NewChatService(chatRepo, llmProvider, ragService /*, memoryService */) // Inject RAGService
	fileService := service.NewFileService(fileStorage, docRepo, taskRepo, taskQueueClient, vectorRepo)

	// Initialize API Handlers
	chatHandler := api.NewChatHandler(chatService)
	fileHandler := api.NewFileHandler(fileService)

	logger.Info("所有依赖初始化完成。")

	// --- 2. Setup Gin Engine & Middleware ---
	// gin.SetMode(gin.ReleaseMode) // Set based on config/env
	router := gin.New()

	// Middleware order matters: Recovery -> Logger -> Error Handler -> (Auth) -> Routes
	router.Use(gin.Recovery())            // Recover from panics
	router.Use(requestLoggerMiddleware()) // Log requests
	router.Use(errorHandlerMiddleware())  // Handle application errors

	// TODO: Add Authentication Middleware
	// authMiddleware := api.NewAuthMiddleware(...)
	// apiV1.Use(authMiddleware.Authenticate())

	// --- 3. Serve Static Frontend Files & Handle SPA Routing ---
	staticFilesDir := "./frontend/dist" // Relative path inside the container where 'npm run build' outputs files
	router.Use(static.Serve("/", static.LocalFile(staticFilesDir, true)))
	router.NoRoute(func(c *gin.Context) {
		// If the request is not for an API endpoint, serve the index.html for SPA routing
		if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
			// Check if the file exists first to avoid serving index.html for missing assets?
			// Or just serve index.html directly for any non-API route.
			c.File(staticFilesDir + "/index.html")
		} else {
			// Let the default 404 handler take care of non-existent API routes
			// Or return a specific API 404 JSON response
			c.JSON(http.StatusNotFound, gin.H{"code": "API_ENDPOINT_NOT_FOUND", "message": "API endpoint not found"})
		}
	})
	logger.Info("静态文件服务和 SPA 路由已配置。", "directory", staticFilesDir)

	// --- 4. Register API Routes ---
	// Health Check (outside API group, no auth needed)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API v1 Group
	apiV1 := router.Group("/api/v1")
	{
		// Register routes from handlers
		chatHandler.RegisterRoutes(apiV1)
		fileHandler.RegisterRoutes(apiV1)
		// Register other handlers here...
	}
	logger.Info("API 路由注册完成。")

	// --- 4. Start HTTP Server ---
	serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,
		// ReadTimeout:  5 * time.Second, // Add timeouts for production
		// WriteTimeout: 10 * time.Second,
		// IdleTimeout:  120 * time.Second,
	}

	logger.Info("服务器准备启动...", "address", serverAddr)

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("服务器启动失败", "error", err)
			cancel() // Signal shutdown on server start failure
		}
	}()

	// --- 5. Graceful Shutdown ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received or context is cancelled
	select {
	case sig := <-quit:
		logger.Info("收到关闭信号", "signal", sig.String())
	case <-ctx.Done():
		logger.Info("上下文取消，开始关闭服务器...")
	}

	// Create a deadline context for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second) // Increased timeout
	defer shutdownCancel()

	logger.Info("正在优雅地关闭服务器...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("服务器关闭失败", "error", err)
		os.Exit(1)
	}

	// Close other resources like DB pool (already deferred)
	// Close Asynq client if Close method is added

	logger.Info("服务器已成功关闭。")
}

// requestLoggerMiddleware logs request details using slog.
func requestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next() // Important: call Next first to get status code and errors

		// Log request details after handling
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		// Get error message if any middleware/handler set it
		errorMessage := ""
		if len(c.Errors) > 0 {
			errorMessage = c.Errors.String() // Get all errors as a string
		}

		logFields := []interface{}{
			"status", statusCode,
			"latency", latency.String(),
			"ip", clientIP,
			"method", method,
			"path", path,
		}
		if raw != "" {
			logFields = append(logFields, "query", raw)
		}
		if errorMessage != "" {
			logFields = append(logFields, "error", errorMessage)
		}

		// Use logger.Ctx to potentially include trace_id if set by another middleware
		reqCtxLogger := logger.Ctx(c.Request.Context())

		if statusCode >= 500 {
			reqCtxLogger.Error("[GIN]", logFields...)
		} else if statusCode >= 400 {
			reqCtxLogger.Warn("[GIN]", logFields...)
		} else {
			reqCtxLogger.Info("[GIN]", logFields...)
		}
	}
}

// errorHandlerMiddleware catches errors set in the context (e.g., by ShouldBindJSON)
// or returned by handlers (if using c.Error()) and converts AppErrors to JSON responses.
func errorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // Process request first

		// Check for errors after handler execution
		err := c.Errors.Last() // Get the last error
		if err == nil {
			return // No error
		}

		// Use errors.As to check if it's an AppError
		var appErr *apperr.AppError
		if errors.As(err.Err, &appErr) {
			// Log the application error details
			logger.WarnContext(c.Request.Context(), "应用程序错误",
				"code", appErr.Code,
				"message", appErr.Message,
				"details", appErr.Details,
				"original_error", appErr.Err, // Log original error if wrapped
			)
			// Return structured JSON error response
			c.AbortWithStatusJSON(appErr.HTTPStatus, gin.H{"error": appErr})
		} else {
			// If it's not an AppError, treat as internal server error
			logger.ErrorContext(c.Request.Context(), "未处理的内部错误", "error", err.Err)
			// Return generic internal server error response
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": apperr.New(apperr.CodeInternal, "发生内部服务器错误"),
			})
		}
	}
}
