package config

import (
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv" // 用于加载 .env 文件
)

// Config 存储应用程序的所有配置。
type Config struct {
	ServerPort    string // API 服务器监听端口
	DatabaseURL   string // PostgreSQL 连接字符串
	RedisAddr     string // Redis 服务器地址
	RedisPassword string // Redis 密码 (新增)
	OpenAIAPIKey  string // OpenAI API 密钥
	OpenAIModel   string // OpenAI 聊天模型名称 (新增)
	// OpenAIEmbeddingModel string // TODO: Add if needed for embedding model config
	UploadDir         string // 文件上传目录
	LogLevel          string // 日志级别 (e.g., "debug", "info", "warn", "error")
	WorkerConcurrency int    // Worker 并发数
	// JWT 相关配置
	JWTSecret            string // 用于签名 JWT 的密钥
	JWTExpirationMinutes int    // JWT 过期时间（分钟）
	// 可以根据需要添加更多配置项...
	// 例如：FrontendURL, VectorDBAddr 等
	UserAPIKeyEncryptionSecret string // 用于加密用户 API Key 的密钥
}

var (
	cfg  *Config
	once sync.Once
)

// LoadConfig 加载配置。
// 它首先尝试从 .env 文件加载（如果存在），然后从环境变量加载。
// 环境变量会覆盖 .env 文件中的值。
// 这个函数是幂等的，只会加载一次配置。
func LoadConfig() *Config {
	once.Do(func() {
		// 尝试加载 .env 文件，忽略错误（可能文件不存在）
		_ = godotenv.Load() // godotenv.Load(".env") 也可以

		workerConcurrencyStr := getEnv("WORKER_CONCURRENCY", "10") // 默认并发数为 10
		workerConcurrency, err := strconv.Atoi(workerConcurrencyStr)
		if err != nil {
			log.Printf("警告: 无效的 WORKER_CONCURRENCY 值 '%s'，将使用默认值 10。错误: %v", workerConcurrencyStr, err)
			workerConcurrency = 10
		}

		jwtExpirationMinutesStr := getEnv("JWT_EXPIRATION_MINUTES", "60") // 默认 60 分钟
		jwtExpirationMinutes, err := strconv.Atoi(jwtExpirationMinutesStr)
		if err != nil {
			log.Printf("警告: 无效的 JWT_EXPIRATION_MINUTES 值 '%s'，将使用默认值 60。错误: %v", jwtExpirationMinutesStr, err)
			jwtExpirationMinutes = 60
		}

		cfg = &Config{
			ServerPort:    getEnv("SERVER_PORT", "8080"),          // 默认端口 8080
			DatabaseURL:   getEnv("DATABASE_URL", ""),             // 没有默认值，必须提供
			RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"), // 默认 Redis 地址
			RedisPassword: getEnv("REDIS_PASSWORD", ""),           // 加载 Redis 密码，默认为空
			OpenAIAPIKey:  getEnv("OPENAI_API_KEY", ""),           // 没有默认值，必须提供
			OpenAIModel:   getEnv("OPENAI_MODEL", ""),             // 新增：加载聊天模型名称，默认为空
			// OpenAIEmbeddingModel: getEnv("OPENAI_EMBEDDING_MODEL", ""), // TODO: Add if needed
			UploadDir:                  getEnv("UPLOAD_DIR", "./uploads"), // 默认上传目录
			LogLevel:                   getEnv("LOG_LEVEL", "info"),       // 默认日志级别 info
			WorkerConcurrency:          workerConcurrency,
			JWTSecret:                  getEnv("JWT_SECRET", ""), // 没有默认值，必须提供
			JWTExpirationMinutes:       jwtExpirationMinutes,
			UserAPIKeyEncryptionSecret: getEnv("USER_API_KEY_ENCRYPTION_SECRET", ""), // 没有默认值，必须提供
		}

		// 可以在这里添加对必要配置项的检查
		if cfg.DatabaseURL == "" {
			log.Fatal("错误: 环境变量 DATABASE_URL 未设置。")
		}
		if cfg.OpenAIAPIKey == "" {
			log.Fatal("错误: 环境变量 OPENAI_API_KEY 未设置。")
		}
		if cfg.JWTSecret == "" {
			// 在生产环境中，JWT 密钥是必需的
			log.Println("警告: 环境变量 JWT_SECRET 未设置。在生产环境中，请务必设置一个强密钥。")
			// 可以选择在这里 Fatal 退出，或者在 auth service 中使用默认值（如已实现）
			// log.Fatal("错误: 环境变量 JWT_SECRET 未设置。")
		}
		if cfg.UserAPIKeyEncryptionSecret == "" {
			log.Fatal("错误: 环境变量 USER_API_KEY_ENCRYPTION_SECRET 未设置。这是加密用户 API Key 所必需的。")
		}
	})
	return cfg
}

// Get 返回已加载的配置实例。
func Get() *Config {
	if cfg == nil {
		// 如果尚未加载，则加载配置
		// 这在直接调用 Get() 而非先调用 LoadConfig() 的情况下很有用
		return LoadConfig()
	}
	return cfg
}

// getEnv 获取环境变量的值，如果环境变量未设置，则返回默认值。
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
