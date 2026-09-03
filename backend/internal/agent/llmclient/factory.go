package llmclient

import (
	"fmt"

	"github.com/masterfabric-go/masterfabric/internal/shared/config"
)

// NewProvider creates the appropriate LLM Provider based on config.
// The config determines which provider, endpoint, model, and API key to use.
// Nothing is hardcoded — all values come from environment variables via config.
func NewProvider(cfg config.LLMConfig) (Provider, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("LLM endpoint is required (set via environment variable)")
	}

	switch cfg.Provider {
	case "ollama", "gemma":
		return NewOllamaProvider(cfg), nil
	case "openai":
		return NewOpenAIProvider(cfg), nil
	case "anthropic":
		return NewAnthropicProvider(cfg), nil
	default:
		return nil, fmt.Errorf("unknown LLM provider %q; supported: ollama, gemma, openai, anthropic", cfg.Provider)
	}
}
