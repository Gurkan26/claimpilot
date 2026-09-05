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

// openaiProvider implements the Provider interface for OpenAI-compatible APIs.
// Endpoint and API key come from config (env variables).
type openaiProvider struct {
	name     string
	endpoint string
	model    string
	apiKey   string
	client   *http.Client
}

// NewOpenAIProvider creates an OpenAI-compatible provider from config.
func NewOpenAIProvider(cfg config.LLMConfig) Provider {
	providerName := cfg.Provider
	if providerName == "" {
		providerName = "openai"
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	return &openaiProvider{
		name:     providerName,
		endpoint: cfg.Endpoint,
		model:    cfg.Model,
		apiKey:   cfg.APIKey,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (p *openaiProvider) Name() string  { return p.name }
func (p *openaiProvider) Model() string { return p.model }

// OpenAI API request/response types
type openaiChatRequest struct {
	Model       string          `json:"model"`
	Messages    []openaiMessage `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	TopP        float64         `json:"top_p,omitempty"`
	Stream      bool            `json:"stream"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiChatResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Content   string `json:"content"`
			Reasoning string `json:"reasoning,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

type openaiStreamResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

func (p *openaiProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	start := time.Now()

	maxTokens := req.MaxTokens
	if maxTokens <= 0 || maxTokens < 2048 {
		maxTokens = 2048
	}

	messages := p.buildMessages(req)
	oaiReq := openaiChatRequest{
		Model:       p.model,
		Messages:    messages,
		Temperature: req.Temperature,
		MaxTokens:   maxTokens,
		TopP:        req.TopP,
		Stream:      false,
	}

	body, err := json.Marshal(oaiReq)
	if err != nil {
		return nil, fmt.Errorf("marshal openai request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create openai request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var oaiResp openaiChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&oaiResp); err != nil {
		return nil, fmt.Errorf("decode openai response: %w", err)
	}

	if len(oaiResp.Choices) == 0 {
		return nil, fmt.Errorf("openai returned no choices")
	}

	choice := oaiResp.Choices[0]
	content := choice.Message.Content
	if content == "" && choice.Message.Reasoning != "" {
		content = choice.Message.Reasoning
	}

	return &CompletionResponse{
		Content:      content,
		Model:        p.model,
		TokensUsed:   oaiResp.Usage.TotalTokens,
		FinishReason: choice.FinishReason,
		Duration:     time.Since(start),
	}, nil
}

func (p *openaiProvider) StreamComplete(ctx context.Context, req *CompletionRequest) (<-chan StreamChunk, error) {
	messages := p.buildMessages(req)
	oaiReq := openaiChatRequest{
		Model:       p.model,
		Messages:    messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		TopP:        req.TopP,
		Stream:      true,
	}

	body, err := json.Marshal(oaiReq)
	if err != nil {
		return nil, fmt.Errorf("marshal openai stream request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create openai stream request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai stream request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("openai stream returned status %d", resp.StatusCode)
	}

	ch := make(chan StreamChunk, 64)
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		for {
			var line json.RawMessage
			if err := decoder.Decode(&line); err != nil {
				if err != io.EOF {
					ch <- StreamChunk{Error: err}
				}
				return
			}

			var chunk openaiStreamResponse
			if err := json.Unmarshal(line, &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) == 0 {
				continue
			}

			done := chunk.Choices[0].FinishReason != nil
			ch <- StreamChunk{
				Content: chunk.Choices[0].Delta.Content,
				Done:    done,
			}
			if done {
				return
			}
		}
	}()

	return ch, nil
}

func (p *openaiProvider) HealthCheck(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, p.endpoint+"/models", nil)
	if err != nil {
		return fmt.Errorf("create health check request: %w", err)
	}
	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("openai health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("openai health check returned status %d", resp.StatusCode)
	}
	return nil
}

func (p *openaiProvider) buildMessages(req *CompletionRequest) []openaiMessage {
	var messages []openaiMessage
	if req.SystemPrompt != "" {
		messages = append(messages, openaiMessage{Role: "system", Content: req.SystemPrompt})
	}
	for _, msg := range req.Messages {
		messages = append(messages, openaiMessage{Role: string(msg.Role), Content: msg.Content})
	}
	return messages
}
