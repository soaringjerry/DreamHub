package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"                // 引入 uuid 包
	"github.com/jackc/pgx/v5/pgxpool"       // 引入 pgx 连接池
	"github.com/joho/godotenv"              // 引入 godotenv 包
	"github.com/tmc/langchaingo/embeddings" // 引入 embeddings 接口包
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"           // 引入 openai 实现包
	"github.com/tmc/langchaingo/schema"                // 引入 schema 包 (用于 Document)
	"github.com/tmc/langchaingo/textsplitter"          // 引入 textsplitter 包
	"github.com/tmc/langchaingo/vectorstores"          // 引入 vectorstores 包 (用于过滤选项)
	"github.com/tmc/langchaingo/vectorstores/pgvector" // 引入 pgvector 包
)

const uploadDir = "uploads"
const maxHistoryMessages = 10 // 限制加载的对话历史数量

// ChatMessage 定义接收聊天消息的结构体
type ChatMessage struct {
	ConversationID string `json:"conversation_id"`            // 对话 ID (可选，为空则新建)
	UserID         string `json:"user_id" binding:"required"` // 用户 ID (临时添加, 设为必填)
	Message        string `json:"message" binding:"required"` // 用户消息
}

// ChatResponse 定义返回给客户端的结构体
type ChatResponse struct {
	ConversationID string `json:"conversation_id"` // 返回当前或新的对话 ID
	Reply          string `json:"reply"`           // AI 回复
}

// AppContext 包含应用共享的资源
type AppContext struct {
	dbPool      *pgxpool.Pool
	llm         llms.Model
	embedder    embeddings.Embedder
	vectorStore pgvector.Store
}

// --- Wrapper to satisfy embeddings.Embedder interface ---
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
	// ... (main 函数前半部分保持不变) ...
	err := godotenv.Load()
	if err != nil {
		log.Println("注意: 未找到 .env 文件或加载失败, 将依赖系统环境变量:", err)
	} else {
		log.Println(".env 文件加载成功")
	}

	ctx := context.Background()

	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		log.Fatal("错误: 环境变量 OPENAI_API_KEY 未设置")
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("错误: 环境变量 DATABASE_URL 未设置")
	}

	dbPool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("无法连接到数据库: %v\n", err)
	}
	defer dbPool.Close()
	log.Println("数据库连接池初始化成功")

	client, err := openai.New(
		openai.WithModel("gpt-4o"),
		openai.WithEmbeddingModel("text-embedding-3-large"),
		openai.WithToken(openaiAPIKey),
	)
	if err != nil {
		log.Fatalf("初始化 OpenAI 客户端失败: %v", err)
	}
	log.Println("OpenAI 客户端初始化成功")

	embedderWrapper := &openAIEmbedderWrapper{client: client}

	vectorStore, err := pgvector.New(
		ctx,
		pgvector.WithConnectionURL(databaseURL),
		pgvector.WithEmbedder(embedderWrapper),
	)
	if err != nil {
		log.Fatalf("初始化 pgvector 存储失败: %v", err)
	}
	log.Println("pgvector 存储初始化成功")

	appCtx := &AppContext{
		dbPool:      dbPool,
		llm:         client,
		embedder:    embedderWrapper,
		vectorStore: vectorStore,
	}

	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	apiV1 := router.Group("/api/v1")
	{
		apiV1.POST("/upload", func(c *gin.Context) { handleUpload(c, appCtx) })
		apiV1.POST("/chat", func(c *gin.Context) { handleChat(c, appCtx) })
	}
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	log.Println("Server starting on :8080")
	router.Run(":8080")
}

// handleUpload 处理文件上传请求
func handleUpload(c *gin.Context, appCtx *AppContext) {
	// ... (handleUpload 函数前半部分保持不变) ...
	userID := c.PostForm("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 user_id 字段"})
		return
	}
	log.Printf("收到来自 UserID '%s' 的文件上传请求\n", userID)

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("获取文件失败: %s", err.Error())})
		return
	}
	filename := filepath.Base(file.Filename)
	userUploadDir := filepath.Join(uploadDir, userID)
	if err := os.MkdirAll(userUploadDir, os.ModePerm); err != nil {
		log.Printf("错误: 创建用户上传目录 '%s' 失败: %v\n", userUploadDir, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建用户目录"})
		return
	}
	dst := filepath.Join(userUploadDir, filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("保存文件失败: %s", err.Error())})
		return
	}
	log.Printf("文件 '%s' 已成功上传到 '%s' (UserID: %s)\n", filename, dst, userID)

	contentBytes, err := os.ReadFile(dst)
	if err != nil {
		log.Printf("错误: 读取文件 '%s' 失败: %v\n", dst, err)
		c.JSON(http.StatusOK, gin.H{"message": "文件上传成功，但后续处理失败", "filename": filename, "error": fmt.Sprintf("读取文件内容失败: %v", err)})
		return
	}
	content := string(contentBytes)

	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(1000),
		textsplitter.WithChunkOverlap(200),
	)
	chunks, err := splitter.SplitText(content)
	if err != nil {
		log.Printf("错误: 分割文件 '%s' 失败: %v\n", filename, err)
		c.JSON(http.StatusOK, gin.H{"message": "文件上传成功，但文本分割失败", "filename": filename, "error": fmt.Sprintf("文本分割失败: %v", err)})
		return
	}
	log.Printf("文件 '%s' 被分割成 %d 个块 (UserID: %s)\n", filename, len(chunks), userID)

	if appCtx.embedder == nil {
		log.Println("严重错误: Embedder 未初始化")
		c.JSON(http.StatusInternalServerError, gin.H{"message": "文件处理失败 (服务内部错误)", "filename": filename})
		return
	}

	docs := make([]schema.Document, len(chunks))
	for i, chunk := range chunks {
		docs[i] = schema.Document{
			PageContent: chunk,
			Metadata: map[string]any{
				"source":   filename,
				"chunk_id": i,
				"user_id":  userID, // 添加 user_id 到元数据
			},
		}
	}

	_, err = appCtx.vectorStore.AddDocuments(context.Background(), docs)
	if err != nil {
		log.Printf("错误: 添加文档向量到 pgvector 失败 (UserID: %s): %v\n", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":  "文件上传分块成功，但存入向量数据库失败",
			"filename": filename,
			"error":    fmt.Sprintf("存储向量失败: %v", err),
		})
		return
	}

	log.Printf("文件 '%s' 的 %d 个块已成功向量化并存入数据库 (UserID: %s)\n", filename, len(chunks), userID)

	c.JSON(http.StatusOK, gin.H{
		"message":  "文件上传、分块、向量化并存储成功",
		"filename": filename,
		"chunks":   len(chunks),
	})
}

// handleChat 处理聊天消息请求 (包含对话历史和用户隔离)
func handleChat(c *gin.Context, appCtx *AppContext) {
	var req ChatMessage
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("无效的请求: %s", err.Error())})
		return
	}
	if req.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 user_id 字段"})
		return
	}
	log.Printf("收到来自 UserID '%s' 的聊天消息: %s (ConvID: %s)\n", req.UserID, req.Message, req.ConversationID)

	ctx := context.Background()
	conversationIDStr := req.ConversationID
	var conversationID uuid.UUID
	var err error

	if conversationIDStr == "" {
		conversationID = uuid.New()
		conversationIDStr = conversationID.String()
		log.Printf("创建新对话 ID: %s (UserID: %s)\n", conversationIDStr, req.UserID)
	} else {
		conversationID, err = uuid.Parse(conversationIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 conversation_id 格式"})
			return
		}
	}

	historyMessages, err := loadConversationHistory(ctx, appCtx.dbPool, conversationID, req.UserID, maxHistoryMessages) // Pass userID
	if err != nil {
		log.Printf("错误: 加载对话历史失败 (ConvID: %s, UserID: %s): %v\n", conversationIDStr, req.UserID, err)
	}

	var relevantDocs []schema.Document
	if appCtx.embedder == nil {
		log.Println("严重错误: Embedder 未初始化，无法执行 RAG 检索")
	} else {
		numDocsToRetrieve := 3
		// --- 实现基于 user_id 的向量搜索过滤 ---
		// 构建元数据过滤器
		filter := map[string]any{"user_id": req.UserID}

		// 假设使用 vectorstores.WithFilters 选项 (需要验证是否存在及用法)
		options := []vectorstores.Option{
			// vectorstores.WithScoreThreshold(0.7), // 可以添加其他选项，如分数阈值
			vectorstores.WithFilters(filter), // 传入过滤器
		}

		// 执行带有过滤器的相似性搜索
		retrievedDocs, err := appCtx.vectorStore.SimilaritySearch(ctx, req.Message, numDocsToRetrieve, options...)

		if err != nil {
			log.Printf("错误: RAG 相似性搜索失败 (ConvID: %s, UserID: %s): %v\n", conversationIDStr, req.UserID, err)
		} else {
			relevantDocs = retrievedDocs
			log.Printf("RAG (UserID: %s) 检索到 %d 个相关文档块 (ConvID: %s)\n", req.UserID, len(relevantDocs), conversationIDStr)
		}
	}

	llmMessages := buildLLMMessages(historyMessages, relevantDocs, req.Message)

	completion, err := appCtx.llm.GenerateContent(ctx, llmMessages)
	if err != nil {
		log.Printf("错误: 调用 OpenAI API 失败 (ConvID: %s, UserID: %s): %v\n", conversationIDStr, req.UserID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI 服务响应错误"})
		return
	}

	if len(completion.Choices) == 0 || completion.Choices[0].Content == "" {
		log.Println("警告: OpenAI API 返回了空的回复 (ConvID: %s, UserID: %s)\n", conversationIDStr, req.UserID)
		c.JSON(http.StatusOK, ChatResponse{Reply: "(AI 没有返回有效回复)", ConversationID: conversationIDStr})
		return
	}

	reply := completion.Choices[0].Content
	log.Printf("AI 回复 (ConvID: %s, UserID: %s): %s\n", conversationIDStr, req.UserID, reply)

	err = saveMessageToHistory(ctx, appCtx.dbPool, conversationID, req.UserID, "user", req.Message)
	if err != nil {
		log.Printf("错误: 保存用户消息到历史失败 (ConvID: %s, UserID: %s): %v\n", conversationIDStr, req.UserID, err)
	}
	err = saveMessageToHistory(ctx, appCtx.dbPool, conversationID, req.UserID, "ai", reply)
	if err != nil {
		log.Printf("错误: 保存 AI 回复到历史失败 (ConvID: %s, UserID: %s): %v\n", conversationIDStr, req.UserID, err)
	}

	c.JSON(http.StatusOK, ChatResponse{
		ConversationID: conversationIDStr,
		Reply:          reply,
	})
}

// loadConversationHistory 从数据库加载特定用户的对话历史
func loadConversationHistory(ctx context.Context, dbPool *pgxpool.Pool, convID uuid.UUID, userID string, limit int) ([]llms.MessageContent, error) {
	query := `
		SELECT sender_role, message_content
		FROM conversation_history
		WHERE conversation_id = $1 AND user_id = $2 -- Filter by both conversation and user ID
		ORDER BY timestamp DESC
		LIMIT $3 -- Parameter index shifted due to adding user_id
	`
	rows, err := dbPool.Query(ctx, query, convID, userID, limit) // Pass userID to the query
	// ... (rest of function unchanged) ...
	if err != nil {
		return nil, fmt.Errorf("查询对话历史失败: %w", err)
	}
	defer rows.Close()

	var history []llms.MessageContent
	for rows.Next() {
		var role, content string
		if err := rows.Scan(&role, &content); err != nil {
			return nil, fmt.Errorf("扫描对话历史行失败: %w", err)
		}
		var msg llms.MessageContent
		switch role {
		case "user":
			msg = llms.MessageContent{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextPart(content)},
			}
		case "ai":
			msg = llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{llms.TextPart(content)},
			}
		default:
			log.Printf("警告: 未知的 sender_role '%s' 在对话历史中\n", role)
			continue
		}
		history = append(history, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("处理对话历史结果集时出错: %w", err)
	}

	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	log.Printf("为 ConvID %s 加载了 %d 条历史消息\n", convID, len(history))
	return history, nil
}

// saveMessageToHistory 将消息保存到数据库 (包含 user_id)
func saveMessageToHistory(ctx context.Context, dbPool *pgxpool.Pool, convID uuid.UUID, userID string, role string, content string) error {
	query := `
		INSERT INTO conversation_history (conversation_id, user_id, sender_role, message_content, timestamp)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := dbPool.Exec(ctx, query, convID, userID, role, content, time.Now())
	if err != nil {
		// 在错误日志中也包含 userID
		return fmt.Errorf("插入对话历史失败 (UserID: %s, ConvID: %s): %w", userID, convID, err)
	}
	return nil
}

// buildLLMMessages 构建发送给 LLM 的消息列表
func buildLLMMessages(history []llms.MessageContent, ragDocs []schema.Document, currentMessage string) []llms.MessageContent {
	// ... (function unchanged) ...
	messages := []llms.MessageContent{}

	if len(ragDocs) > 0 {
		contextStrings := make([]string, len(ragDocs))
		for i, doc := range ragDocs {
			contextStrings[i] = doc.PageContent
		}
		contextText := strings.Join(contextStrings, "\n\n---\n\n")
		messages = append(messages, llms.MessageContent{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(fmt.Sprintf("请参考以下信息回答问题:\n%s", contextText))},
		})
	}

	messages = append(messages, history...)

	messages = append(messages, llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextPart(currentMessage)},
	})

	return messages
}
