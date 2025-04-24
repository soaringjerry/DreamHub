package config

import (
	"fmt" // 移到这里
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/soaringjerry/dreamhub/pkg/logger" // Use our logger
)

// Config holds all configuration for the application.
type Config struct {
	// Environment specific
	Environment string `env:"ENVIRONMENT,default=development"` // e.g., development, staging, production

	// Server configuration
	ServerPort string `env:"SERVER_PORT,default=8080"`

	// Database configuration
	DatabaseURL string `env:"DATABASE_URL,required"`

	// Redis configuration
	RedisAddr string `env:"REDIS_ADDR,default=localhost:6379"`

	// OpenAI configuration
	OpenAIAPIKey         string `env:"OPENAI_API_KEY,required"`
	OpenAIModel          string `env:"OPENAI_MODEL,default=gpt-4o"`
	OpenAIEmbeddingModel string `env:"OPENAI_EMBEDDING_MODEL,default=text-embedding-3-large"`

	// File Upload configuration
	UploadDir string `env:"UPLOAD_DIR,default=uploads"`

	// Chat configuration
	MaxHistoryMessages int `env:"MAX_HISTORY_MESSAGES,default=10"`

	// Worker / Embedding configuration
	WorkerConcurrency    int           `env:"WORKER_CONCURRENCY,default=10"`
	SplitterChunkSize    int           `env:"SPLITTER_CHUNK_SIZE,default=1000"`
	SplitterChunkOverlap int           `env:"SPLITTER_CHUNK_OVERLAP,default=200"`
	EmbeddingTimeout     time.Duration `env:"EMBEDDING_TIMEOUT,default=5m"` // Example timeout for embedding

	// Add other configurations as needed (e.g., JWT secrets, CORS origins)
}

// Load loads configuration from environment variables.
// It loads .env file first if it exists.
func Load() (*Config, error) {
	// Load .env file, ignore error if it doesn't exist
	err := godotenv.Load()
	if err != nil {
		logger.Warn("Config: .env file not found or failed to load, relying on system env vars", "error", err)
	} else {
		logger.Info("Config: .env file loaded successfully")
	}

	cfg := &Config{}

	// Manually load env vars into struct fields, checking required ones
	// A library like 'cleanenv' or 'viper' could automate this, but manual is fine for now.

	cfg.Environment = getEnv("ENVIRONMENT", "development")
	cfg.ServerPort = getEnv("SERVER_PORT", "8080")

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		logger.Error("Config error: DATABASE_URL environment variable is required")
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	cfg.RedisAddr = getEnv("REDIS_ADDR", "localhost:6379")

	cfg.OpenAIAPIKey = os.Getenv("OPENAI_API_KEY")
	if cfg.OpenAIAPIKey == "" {
		logger.Error("Config error: OPENAI_API_KEY environment variable is required")
		return nil, fmt.Errorf("OPENAI_API_KEY is required")
	}
	cfg.OpenAIModel = getEnv("OPENAI_MODEL", "gpt-4o")
	cfg.OpenAIEmbeddingModel = getEnv("OPENAI_EMBEDDING_MODEL", "text-embedding-3-large")

	cfg.UploadDir = getEnv("UPLOAD_DIR", "uploads")

	cfg.MaxHistoryMessages = getEnvAsInt("MAX_HISTORY_MESSAGES", 10)
	cfg.WorkerConcurrency = getEnvAsInt("WORKER_CONCURRENCY", 10)
	cfg.SplitterChunkSize = getEnvAsInt("SPLITTER_CHUNK_SIZE", 1000)
	cfg.SplitterChunkOverlap = getEnvAsInt("SPLITTER_CHUNK_OVERLAP", 200)
	cfg.EmbeddingTimeout = getEnvAsDuration("EMBEDDING_TIMEOUT", 5*time.Minute)

	logger.Info("Configuration loaded successfully")
	return cfg, nil
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt retrieves an environment variable as an integer or returns a default value.
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	logger.Warn("Config: Invalid or missing integer value for env var, using default", "key", key, "default", defaultValue)
	return defaultValue
}

// getEnvAsDuration retrieves an environment variable as a duration or returns a default value.
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	logger.Warn("Config: Invalid or missing duration value for env var, using default", "key", key, "default", defaultValue)
	return defaultValue
}

// Need to import fmt for error messages
// import "fmt" // Duplicate removed, already imported at the top
