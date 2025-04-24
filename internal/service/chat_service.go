package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/soaringjerry/dreamhub/internal/entity"
	"github.com/soaringjerry/dreamhub/internal/repository"
	"github.com/soaringjerry/dreamhub/pkg/apperr"
	"github.com/soaringjerry/dreamhub/pkg/logger"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

const maxHistoryMessages = 10 // Consider making this configurable

// ChatService defines the interface for chat operations.
type ChatService interface {
	HandleChatMessage(ctx context.Context, userID, conversationIDStr, message string) (*entity.ChatResponse, error)
}

// chatServiceImpl implements the ChatService interface.
type chatServiceImpl struct {
	dbPool  *pgxpool.Pool // Direct DB access for history (consider a HistoryRepository later)
	llm     llms.Model
	docRepo repository.DocumentRepository
	// embedder embeddings.Embedder // Embedder is used by docRepo, not directly here
}

// NewChatService creates a new ChatService implementation.
func NewChatService(db *pgxpool.Pool, llm llms.Model, dr repository.DocumentRepository) ChatService {
	if db == nil || llm == nil || dr == nil {
		panic("DB Pool, LLM, and DocumentRepository cannot be nil for ChatService")
	}
	return &chatServiceImpl{
		dbPool:  db,
		llm:     llm,
		docRepo: dr,
	}
}

// HandleChatMessage processes an incoming chat message.
func (s *chatServiceImpl) HandleChatMessage(ctx context.Context, userID, conversationIDStr, message string) (*entity.ChatResponse, error) {
	var conversationID uuid.UUID
	var err error

	// 1. Handle Conversation ID
	if conversationIDStr == "" {
		conversationID = uuid.New()
		conversationIDStr = conversationID.String()
		logger.Info("ChatService: Creating new conversation", "conversationID", conversationIDStr, "userID", userID)
	} else {
		conversationID, err = uuid.Parse(conversationIDStr)
		if err != nil {
			logger.Warn("ChatService: Invalid conversation_id format", "conversationID", conversationIDStr, "userID", userID, "error", err)
			return nil, apperr.NewValidationError("Invalid conversation_id format")
		}
	}

	// 2. Load History (using direct DB access for now)
	// TODO: Refactor history loading/saving into a ConversationRepository
	historyMessages, err := s.loadConversationHistory(ctx, conversationID, userID, maxHistoryMessages)
	if err != nil {
		// Log error but continue, allow chat without history
		logger.Error("ChatService: Failed to load conversation history", "conversationID", conversationIDStr, "userID", userID, "error", err)
	}

	// 3. Perform RAG Search via DocumentRepository
	var relevantDocs []schema.Document
	numDocsToRetrieve := 3 // Consider making this configurable
	// Pass context which should contain user info for the repository to filter
	retrievedDocs, err := s.docRepo.SimilaritySearch(ctx, message, numDocsToRetrieve)
	if err != nil {
		// Log error but continue, allow chat without RAG
		logger.Error("ChatService: RAG similarity search failed", "conversationID", conversationIDStr, "userID", userID, "error", err)
	} else {
		relevantDocs = retrievedDocs
		logger.Info("ChatService: RAG retrieved documents", "userID", userID, "retrievedDocs", len(relevantDocs), "conversationID", conversationIDStr)
	}

	// 4. Build LLM Messages
	llmMessages := s.buildLLMMessages(historyMessages, relevantDocs, message)

	// 5. Call LLM
	completion, err := s.llm.GenerateContent(ctx, llmMessages)
	if err != nil {
		logger.Error("ChatService: LLM API call failed", "conversationID", conversationIDStr, "userID", userID, "error", err)
		return nil, apperr.Wrap(apperr.LLMAPIError, "AI service request failed", err)
	}

	if len(completion.Choices) == 0 || completion.Choices[0].Content == "" {
		logger.Warn("ChatService: LLM returned empty response", "conversationID", conversationIDStr, "userID", userID)
		// Return a specific response instead of error? Or a specific error code?
		// For now, return an error indicating no reply.
		return nil, apperr.New(apperr.LLMAPIError, "AI did not provide a valid response")
	}
	reply := completion.Choices[0].Content
	logger.Info("ChatService: AI reply generated", "conversationID", conversationIDStr, "userID", userID, "replyLength", len(reply))

	// 6. Save messages to history (using direct DB access for now)
	// TODO: Refactor history loading/saving into a ConversationRepository
	errUser := s.saveMessageToHistory(ctx, conversationID, userID, "user", message)
	if errUser != nil {
		logger.Error("ChatService: Failed to save user message to history", "conversationID", conversationIDStr, "userID", userID, "error", errUser)
		// Log error but don't fail the whole request
	}
	errAI := s.saveMessageToHistory(ctx, conversationID, userID, "ai", reply)
	if errAI != nil {
		logger.Error("ChatService: Failed to save AI reply to history", "conversationID", conversationIDStr, "userID", userID, "error", errAI)
		// Log error but don't fail the whole request
	}

	// 7. Return response
	return &entity.ChatResponse{
		ConversationID: conversationIDStr,
		Reply:          reply,
	}, nil
}

// --- Helper methods (originally in main.go) ---
// TODO: Move these to a dedicated ConversationRepository implementation.

func (s *chatServiceImpl) loadConversationHistory(ctx context.Context, convID uuid.UUID, userID string, limit int) ([]llms.MessageContent, error) {
	query := `
		SELECT sender_role, message_content
		FROM conversation_history
		WHERE conversation_id = $1 AND user_id = $2
		ORDER BY timestamp DESC
		LIMIT $3
	`
	rows, err := s.dbPool.Query(ctx, query, convID, userID, limit)
	if err != nil {
		// Wrap error for context
		return nil, fmt.Errorf("failed to query conversation history: %w", err)
	}
	defer rows.Close()

	var history []llms.MessageContent
	for rows.Next() {
		var role, content string
		if err := rows.Scan(&role, &content); err != nil {
			return nil, fmt.Errorf("failed to scan history row: %w", err)
		}
		var msg llms.MessageContent
		switch role {
		case "user":
			msg = llms.MessageContent{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart(content)}}
		case "ai":
			msg = llms.MessageContent{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextPart(content)}}
		default:
			logger.Warn("loadConversationHistory: Unknown sender_role in history", "role", role, "conversationID", convID, "userID", userID)
			continue
		}
		history = append(history, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating history results: %w", err)
	}

	// Reverse history to be chronological
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	logger.Debug("loadConversationHistory: History loaded", "conversationID", convID, "userID", userID, "messageCount", len(history))
	return history, nil
}

func (s *chatServiceImpl) saveMessageToHistory(ctx context.Context, convID uuid.UUID, userID string, role string, content string) error {
	query := `
		INSERT INTO conversation_history (conversation_id, user_id, sender_role, message_content, timestamp)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := s.dbPool.Exec(ctx, query, convID, userID, role, content, time.Now())
	if err != nil {
		// Wrap error for context
		return fmt.Errorf("failed to insert message into history (UserID: %s, ConvID: %s): %w", userID, convID, err)
	}
	return nil
}

func (s *chatServiceImpl) buildLLMMessages(history []llms.MessageContent, ragDocs []schema.Document, currentMessage string) []llms.MessageContent {
	messages := []llms.MessageContent{}

	// Add RAG context as system message if available
	if len(ragDocs) > 0 {
		contextStrings := make([]string, len(ragDocs))
		for i, doc := range ragDocs {
			contextStrings[i] = doc.PageContent
		}
		contextText := strings.Join(contextStrings, "\n\n---\n\n")
		// Limit context size?
		messages = append(messages, llms.MessageContent{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(fmt.Sprintf("Please use the following context to answer the question:\n%s", contextText))},
		})
		logger.Debug("buildLLMMessages: Added RAG context to messages", "numDocs", len(ragDocs))
	}

	// Add historical messages
	messages = append(messages, history...)

	// Add current user message
	messages = append(messages, llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextPart(currentMessage)},
	})

	return messages
}
