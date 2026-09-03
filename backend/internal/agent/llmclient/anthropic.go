package llmclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/masterfabric-go/masterfabric/internal/shared/config"
)

// anthropicProvider implements the Provider interface for Anthropic's Claude API.
// Endpoint and API key come from config (env variables).
type anthropicProvider struct {
	endpoint string
	model    string
	apiKey   string
	client   *http.Client
}

// NewAnthropicProvider creates an Anthropic Claude provider from config.
func NewAnthropicProvider(cfg config.LLMConfig) Provider {
	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = "https://api.anthropic.com"
	}
	return &anthropicProvider{
		endpoint: endpoint,
		model:    cfg.Model,
		apiKey:   cfg.APIKey,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

func (p *anthropicProvider) Name() string  { return "anthropic" }
func (p *anthropicProvider) Model() string { return p.model }

// Anthropic Messages API types
type anthropicMessagesRequest struct {
	Model     string             `json:"model"`
	Messages  []anthropicMessage `json:"messages"`
	System    string             `json:"system,omitempty"`
	MaxTokens int                `json:"max_tokens"`
	Stream    bool               `json:"stream"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicMessagesResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (p *anthropicProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	start := time.Now()

	var messages []anthropicMessage
	for _, msg := range req.Messages {
		messages = append(messages, anthropicMessage{Role: string(msg.Role), Content: msg.Content})
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	anthropicReq := anthropicMessagesRequest{
		Model:     p.model,
		Messages:  messages,
		System:    req.SystemPrompt,
		MaxTokens: maxTokens,
		Stream:    false,
	}

	body, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("marshal anthropic request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create anthropic request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	if p.apiKey != "" {
		httpReq.Header.Set("x-api-key", p.apiKey)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("anthropic returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var anthropicResp anthropicMessagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return nil, fmt.Errorf("decode anthropic response: %w", err)
	}

	content := ""
	if len(anthropicResp.Content) > 0 {
		content = anthropicResp.Content[0].Text
	}

	return &CompletionResponse{
		Content:      content,
		Model:        p.model,
		TokensUsed:   anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
		FinishReason: anthropicResp.StopReason,
		Duration:     time.Since(start),
	}, nil
}

func (p *anthropicProvider) StreamComplete(ctx context.Context, req *CompletionRequest) (<-chan StreamChunk, error) {
	// Anthropic streaming uses SSE; for now return an error directing to use Complete
	return nil, fmt.Errorf("anthropic streaming not yet implemented; use Complete for synchronous requests")
}

func (p *anthropicProvider) HealthCheck(ctx context.Context) error {
	// Anthropic doesn't have a health endpoint; attempt a minimal request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, p.endpoint, nil)
	if err != nil {
		return fmt.Errorf("create health check request: %w", err)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("anthropic health check failed: %w", err)
	}
	defer resp.Body.Close()

	// Any response means the endpoint is reachable
	return nil
}
