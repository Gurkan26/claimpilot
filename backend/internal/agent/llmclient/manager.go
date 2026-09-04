package llmclient

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/masterfabric-go/masterfabric/internal/shared/config"
)

// DynamicProvider wraps a Provider with thread-safe hot-swap capability.
// It implements the Provider interface so consumers (Analyst, Verifier, Orchestrator)
// do not need to restart when the admin switches models or providers.
type DynamicProvider struct {
	mu       sync.RWMutex
	provider Provider
	cfg      config.LLMConfig
	role     string
	logger   *slog.Logger
}

// NewDynamicProvider creates a new hot-swappable provider wrapper.
func NewDynamicProvider(role string, initial Provider, cfg config.LLMConfig, logger *slog.Logger) *DynamicProvider {
	return &DynamicProvider{
		role:     role,
		provider: initial,
		cfg:      cfg,
		logger:   logger.With("dynamic_llm_role", role),
	}
}

// Swap atomically replaces the active LLM provider and its configuration.
func (d *DynamicProvider) Swap(newProvider Provider, newCfg config.LLMConfig) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.provider = newProvider
	d.cfg = newCfg
	d.logger.Info("hot-swapped LLM provider at runtime",
		"role", d.role,
		"provider", newCfg.Provider,
		"model", newCfg.Model,
		"endpoint", newCfg.Endpoint,
	)
}

// GetConfig returns a copy of the current LLM configuration.
func (d *DynamicProvider) GetConfig() config.LLMConfig {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.cfg
}

// Complete executes a completion request against the current active provider.
func (d *DynamicProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	d.mu.RLock()
	p := d.provider
	d.mu.RUnlock()
	if p == nil {
		return nil, fmt.Errorf("llm provider for role %q is not initialized", d.role)
	}
	return p.Complete(ctx, req)
}

// StreamComplete initiates streaming completion against current active provider.
func (d *DynamicProvider) StreamComplete(ctx context.Context, req *CompletionRequest) (<-chan StreamChunk, error) {
	d.mu.RLock()
	p := d.provider
	d.mu.RUnlock()
	if p == nil {
		return nil, fmt.Errorf("llm provider for role %q is not initialized", d.role)
	}
	return p.StreamComplete(ctx, req)
}

// Name returns the provider name.
func (d *DynamicProvider) Name() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.provider != nil {
		return d.provider.Name()
	}
	return d.cfg.Provider
}

// Model returns the active model name.
func (d *DynamicProvider) Model() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.provider != nil {
		return d.provider.Model()
	}
	return d.cfg.Model
}

// HealthCheck verifies connectivity to the underlying model endpoint.
func (d *DynamicProvider) HealthCheck(ctx context.Context) error {
	d.mu.RLock()
	p := d.provider
	d.mu.RUnlock()
	if p == nil {
		return fmt.Errorf("llm provider %q is not configured", d.role)
	}
	return p.HealthCheck(ctx)
}
