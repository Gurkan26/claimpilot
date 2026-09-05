package llmclient

import (
	"github.com/masterfabric-go/masterfabric/internal/shared/config"
)

// NewProvider creates the appropriate LLM Provider based on config.
// The config determines which provider, endpoint, model, and API key to use.
// Nothing is hardcoded — all values come from environment variables via config.
func NewProvider(cfg config.LLMConfig) (Provider, error) {
	switch cfg.Provider {
	case "builtin", "internal", "mock":
		return NewBuiltinProvider(cfg), nil
	case "groq":
		if cfg.Endpoint == "" {
			cfg.Endpoint = "https://api.groq.com/openai/v1"
		}
		if cfg.Model == "" {
			cfg.Model = "openai/gpt-oss-120b"
		}
		return NewOpenAIProvider(cfg), nil
	case "openrouter":
		if cfg.Endpoint == "" {
			cfg.Endpoint = "https://openrouter.ai/api/v1"
		}
		if cfg.Model == "" {
			cfg.Model = "meta-llama/llama-3.3-70b-instruct:free"
		}
		return NewOpenAIProvider(cfg), nil
	case "ollama", "gemma":
		if cfg.Endpoint == "" {
			return NewBuiltinProvider(cfg), nil
		}
		return NewOllamaProvider(cfg), nil
	case "openai":
		if cfg.Endpoint == "" {
			return NewBuiltinProvider(cfg), nil
		}
		return NewOpenAIProvider(cfg), nil
	case "anthropic":
		if cfg.Endpoint == "" {
			return NewBuiltinProvider(cfg), nil
		}
		return NewAnthropicProvider(cfg), nil
	default:
		return NewBuiltinProvider(cfg), nil
	}
}
