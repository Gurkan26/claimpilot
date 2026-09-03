package llmclient

import (
	"context"
	"time"
)

// Role defines the role of a message sender.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message represents a single message in a conversation.
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// CompletionRequest represents a request to the LLM.
type CompletionRequest struct {
	SystemPrompt string    `json:"system_prompt,omitempty"`
	Messages     []Message `json:"messages"`
	Temperature  float64   `json:"temperature,omitempty"`
	MaxTokens    int       `json:"max_tokens,omitempty"`
	TopP         float64   `json:"top_p,omitempty"`
}

// CompletionResponse represents the LLM's response.
type CompletionResponse struct {
	Content      string        `json:"content"`
	Model        string        `json:"model"`
	TokensUsed   int           `json:"tokens_used"`
	FinishReason string        `json:"finish_reason"`
	Duration     time.Duration `json:"duration"`
}

// StreamChunk represents a single chunk in a streaming response.
type StreamChunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
	Error   error  `json:"error,omitempty"`
}

// Provider is the LLM provider interface.
// All LLM interactions go through this interface, ensuring the endpoint
// and model configuration come from config, not hardcoded values.
type Provider interface {
	// Complete sends a completion request and returns the full response.
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

	// StreamComplete sends a completion request and returns a streaming channel.
	StreamComplete(ctx context.Context, req *CompletionRequest) (<-chan StreamChunk, error)

	// Name returns the provider name (for logging and audit).
	Name() string

	// Model returns the configured model name (for audit trail).
	Model() string

	// HealthCheck verifies the provider is reachable.
	HealthCheck(ctx context.Context) error
}
