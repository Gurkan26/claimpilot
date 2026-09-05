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

// ollamaProvider implements the Provider interface for Ollama/Gemma.
// Endpoint and model come from config (LLM_ANALYST_ENDPOINT / LLM_VERIFIER_ENDPOINT).
type ollamaProvider struct {
	endpoint string
	model    string
	client   *http.Client
}

// NewOllamaProvider creates an Ollama provider from config.
func NewOllamaProvider(cfg config.LLMConfig) Provider {
	return &ollamaProvider{
		endpoint: cfg.Endpoint,
		model:    cfg.Model,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

func (p *ollamaProvider) Name() string  { return "ollama" }
func (p *ollamaProvider) Model() string { return p.model }

// ollamaGenerateRequest is the Ollama /api/generate request body.
type ollamaGenerateRequest struct {
	Model    string `json:"model"`
	Prompt   string `json:"prompt"`
	System   string `json:"system,omitempty"`
	Stream   bool   `json:"stream"`
	Options  *ollamaOptions `json:"options,omitempty"`
}

type ollamaOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
}

// ollamaGenerateResponse is the Ollama /api/generate response body.
type ollamaGenerateResponse struct {
	Model           string `json:"model"`
	Response        string `json:"response"`
	Done            bool   `json:"done"`
	TotalDuration   int64  `json:"total_duration"`
	PromptEvalCount int    `json:"prompt_eval_count"`
	EvalCount       int    `json:"eval_count"`
}

func (p *ollamaProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	start := time.Now()

	// Build combined prompt from messages
	prompt := buildPromptFromMessages(req.Messages)

	ollamaReq := ollamaGenerateRequest{
		Model:  p.model,
		Prompt: prompt,
		System: req.SystemPrompt,
		Stream: false,
		Options: &ollamaOptions{
			Temperature: req.Temperature,
			NumPredict:  req.MaxTokens,
			TopP:        req.TopP,
		},
	}

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("marshal ollama request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create ollama request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		// Fallback to Go built-in test AI engine if Ollama is not running
		fallback := NewBuiltinProvider(config.LLMConfig{Model: p.model + " (Go Builtin)"})
		return fallback.Complete(ctx, req)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fallback := NewBuiltinProvider(config.LLMConfig{Model: p.model + " (Go Builtin)"})
		return fallback.Complete(ctx, req)
	}

	var ollamaResp ollamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("decode ollama response: %w", err)
	}

	return &CompletionResponse{
		Content:      ollamaResp.Response,
		Model:        ollamaResp.Model,
		TokensUsed:   ollamaResp.PromptEvalCount + ollamaResp.EvalCount,
		FinishReason: "stop",
		Duration:     time.Since(start),
	}, nil
}

func (p *ollamaProvider) StreamComplete(ctx context.Context, req *CompletionRequest) (<-chan StreamChunk, error) {
	prompt := buildPromptFromMessages(req.Messages)

	ollamaReq := ollamaGenerateRequest{
		Model:  p.model,
		Prompt: prompt,
		System: req.SystemPrompt,
		Stream: true,
		Options: &ollamaOptions{
			Temperature: req.Temperature,
			NumPredict:  req.MaxTokens,
			TopP:        req.TopP,
		},
	}

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("marshal ollama stream request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create ollama stream request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama stream request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	ch := make(chan StreamChunk, 64)
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		for {
			var chunk ollamaGenerateResponse
			if err := decoder.Decode(&chunk); err != nil {
				if err != io.EOF {
					ch <- StreamChunk{Error: err}
				}
				return
			}
			ch <- StreamChunk{
				Content: chunk.Response,
				Done:    chunk.Done,
			}
			if chunk.Done {
				return
			}
		}
	}()

	return ch, nil
}

func (p *ollamaProvider) HealthCheck(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, p.endpoint+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("create health check request: %w", err)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("ollama health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama health check returned status %d", resp.StatusCode)
	}
	return nil
}

// buildPromptFromMessages concatenates messages into a single prompt string.
func buildPromptFromMessages(messages []Message) string {
	var buf bytes.Buffer
	for _, msg := range messages {
		buf.WriteString(msg.Content)
		buf.WriteString("\n")
	}
	return buf.String()
}

// ==========================================
// Ollama Model Management (Admin Operations)
// ==========================================

// OllamaModelInfo represents a model available on an Ollama instance.
type OllamaModelInfo struct {
	Name       string    `json:"name"`
	Model      string    `json:"model"`
	Size       int64     `json:"size"`
	Digest     string    `json:"digest"`
	ModifiedAt time.Time `json:"modified_at"`
	Details    struct {
		Format            string `json:"format"`
		Family            string `json:"family"`
		ParameterSize     string `json:"parameter_size"`
		QuantizationLevel string `json:"quantization_level"`
	} `json:"details"`
}

// ollamaTagsResponse is the response from Ollama /api/tags endpoint.
type ollamaTagsResponse struct {
	Models []OllamaModelInfo `json:"models"`
}

// ollamaPullRequest is the request body for Ollama /api/pull endpoint.
type ollamaPullRequest struct {
	Name   string `json:"name"`
	Stream bool   `json:"stream"`
}

// ollamaPullResponse is the response from Ollama /api/pull endpoint.
type ollamaPullResponse struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Completed int64  `json:"completed,omitempty"`
}

// ListOllamaModels fetches available models from an Ollama endpoint.
// This is a standalone function (not on Provider) because it's an admin operation.
func ListOllamaModels(ctx context.Context, endpoint string) ([]OllamaModelInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("create list models request: %w", err)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama list models failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama list models returned status %d", resp.StatusCode)
	}

	var tagsResp ollamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return nil, fmt.Errorf("decode ollama tags response: %w", err)
	}

	return tagsResp.Models, nil
}

// PullOllamaModel triggers a model pull on an Ollama endpoint.
// It blocks until the model is fully downloaded (non-streaming).
func PullOllamaModel(ctx context.Context, endpoint, modelName string) error {
	// Use a long timeout for model downloads (up to 30 minutes)
	client := &http.Client{Timeout: 30 * time.Minute}

	pullReq := ollamaPullRequest{
		Name:   modelName,
		Stream: false,
	}

	body, err := json.Marshal(pullReq)
	if err != nil {
		return fmt.Errorf("marshal pull request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/api/pull", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create pull request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("ollama pull model failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama pull returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var pullResp ollamaPullResponse
	if err := json.NewDecoder(resp.Body).Decode(&pullResp); err != nil {
		return fmt.Errorf("decode pull response: %w", err)
	}

	if pullResp.Status != "success" {
		return fmt.Errorf("ollama pull status: %s", pullResp.Status)
	}

	return nil
}

// CheckOllamaHealth verifies an Ollama endpoint is reachable and measures latency.
func CheckOllamaHealth(ctx context.Context, endpoint string) (latencyMs int64, err error) {
	client := &http.Client{Timeout: 10 * time.Second}
	start := time.Now()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"/api/tags", nil)
	if err != nil {
		return 0, fmt.Errorf("create health check request: %w", err)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return 0, fmt.Errorf("ollama unreachable: %w", err)
	}
	defer resp.Body.Close()

	latencyMs = time.Since(start).Milliseconds()

	if resp.StatusCode != http.StatusOK {
		return latencyMs, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	return latencyMs, nil
}

