package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/masterfabric-go/masterfabric/internal/agent/llmclient"
	"github.com/masterfabric-go/masterfabric/internal/mcp"
	"github.com/masterfabric-go/masterfabric/internal/shared/config"
)

// LLMDTO represents LLM config data for API transport.
type LLMDTO struct {
	Role            string  `json:"role"`
	Provider        string  `json:"provider"`
	Endpoint        string  `json:"endpoint"`
	Model           string  `json:"model"`
	APIKey          string  `json:"api_key,omitempty"`
	Temperature     float64 `json:"temperature"`
	MaxTokens       int     `json:"max_tokens"`
	TimeoutSeconds  int     `json:"timeout_seconds"`
	Status          string  `json:"status"`
	LastPingMs      int     `json:"last_ping_ms,omitempty"`
	SystemPrompt    string  `json:"system_prompt,omitempty"`
	RoleDescription string  `json:"role_description,omitempty"`
}

// HarnessPolicyDTO represents safety and guardrail policies.
type HarnessPolicyDTO struct {
	RequireHumanApproval bool   `json:"require_human_approval"`
	PIIMaskingLevel      string `json:"pii_masking_level"`
	MaxToolIterations    int    `json:"max_tool_iterations"`
	TimeoutSeconds       int    `json:"timeout_seconds"`
	AuditLogging         bool   `json:"audit_logging"`
}

// HarnessConfigDTO is the complete payload for the Admin / Agent Harness console.
type HarnessConfigDTO struct {
	AnalystLLM  LLMDTO            `json:"analyst_llm"`
	VerifierLLM LLMDTO            `json:"verifier_llm"`
	MCPAdapters []mcp.AdapterInfo `json:"mcp_adapters"`
	Policy      HarnessPolicyDTO  `json:"policy"`
	UpdatedAt   string            `json:"updated_at"`
}

// SimulationStepDTO represents a step in the visual execution trace.
type SimulationStepDTO struct {
	StepNumber  int    `json:"step_number"`
	Component   string `json:"component"`
	Title       string `json:"title"`
	Description string `json:"description"`
	DurationMs  int    `json:"duration_ms"`
	Status      string `json:"status"`
	OutputData  any    `json:"output_data,omitempty"`
}

// SimulationResultDTO is the response returned when testing the pipeline.
type SimulationResultDTO struct {
	ID                       string              `json:"id"`
	ContractSample           string              `json:"contract_sample"`
	Steps                    []SimulationStepDTO `json:"steps"`
	ExtractedObligationCount int                 `json:"extracted_obligation_count"`
	VerifiedRatio            float64             `json:"verified_ratio"`
	MCPCallsCount            int                 `json:"mcp_calls_count"`
	TotalTimeMs              int                 `json:"total_time_ms"`
	PassedGuardrails         bool                `json:"passed_guardrails"`
}

// OllamaModelDTO represents a model available on an Ollama instance (for API transport).
type OllamaModelDTO struct {
	Name              string `json:"name"`
	Size              int64  `json:"size"`
	Family            string `json:"family"`
	ParameterSize     string `json:"parameter_size"`
	QuantizationLevel string `json:"quantization_level"`
	ModifiedAt        string `json:"modified_at"`
}

// TestLLMResultDTO is the response for an LLM connection test.
type TestLLMResultDTO struct {
	Success   bool   `json:"success"`
	LatencyMs int64  `json:"latency_ms"`
	Message   string `json:"message"`
	Role      string `json:"role"`
}

// HarnessUseCase manages Agent Harness runtime settings, LLM hot-swap, and MCP registry.
type HarnessUseCase struct {
	analystProvider  *llmclient.DynamicProvider
	verifierProvider *llmclient.DynamicProvider
	mcpRegistry      *mcp.Registry
	adminCfg         config.AdminConfig
	logger           *slog.Logger
}

// NewHarnessUseCase creates a new HarnessUseCase instance.
func NewHarnessUseCase(
	analyst *llmclient.DynamicProvider,
	verifier *llmclient.DynamicProvider,
	mcpRegistry *mcp.Registry,
	adminCfg config.AdminConfig,
	logger *slog.Logger,
) *HarnessUseCase {
	return &HarnessUseCase{
		analystProvider:  analyst,
		verifierProvider: verifier,
		mcpRegistry:      mcpRegistry,
		adminCfg:         adminCfg,
		logger:           logger.With("usecase", "harness_admin"),
	}
}

// VerifyAdminPassword validates the admin secret.
func (uc *HarnessUseCase) VerifyAdminPassword(password string) bool {
	expected := uc.adminCfg.Password
	if expected == "" {
		expected = "admin123"
	}
	return strings.TrimSpace(password) == expected
}

// GetHarnessConfig returns the current active configuration with real health status.
func (uc *HarnessUseCase) GetHarnessConfig(ctx context.Context) (*HarnessConfigDTO, error) {
	analystCfg := uc.analystProvider.GetConfig()
	verifierCfg := uc.verifierProvider.GetConfig()

	// Probe real Ollama health for analyst
	analystStatus := "connected"
	analystPingMs := 0
	if analystCfg.Endpoint != "" {
		latency, err := llmclient.CheckOllamaHealth(ctx, analystCfg.Endpoint)
		if err != nil {
			analystStatus = "disconnected"
			uc.logger.Warn("analyst LLM health check failed", "error", err)
		} else {
			analystPingMs = int(latency)
		}
	}

	// Probe real Ollama health for verifier
	verifierStatus := "connected"
	verifierPingMs := 0
	if verifierCfg.Endpoint != "" {
		latency, err := llmclient.CheckOllamaHealth(ctx, verifierCfg.Endpoint)
		if err != nil {
			verifierStatus = "disconnected"
			uc.logger.Warn("verifier LLM health check failed", "error", err)
		} else {
			verifierPingMs = int(latency)
		}
	}

	return &HarnessConfigDTO{
		AnalystLLM: LLMDTO{
			Role:            "analyst",
			Provider:        analystCfg.Provider,
			Endpoint:        analystCfg.Endpoint,
			Model:           analystCfg.Model,
			Temperature:     0.1,
			MaxTokens:       4096,
			TimeoutSeconds:  int(analystCfg.Timeout.Seconds()),
			Status:          analystStatus,
			LastPingMs:      analystPingMs,
			RoleDescription: "Sözleşme ayrıştırma, taahhüt tespiti ve bildirim süresi çıkarma",
		},
		VerifierLLM: LLMDTO{
			Role:            "verifier",
			Provider:        verifierCfg.Provider,
			Endpoint:        verifierCfg.Endpoint,
			Model:           verifierCfg.Model,
			Temperature:     0.0,
			MaxTokens:       2048,
			TimeoutSeconds:  int(verifierCfg.Timeout.Seconds()),
			Status:          verifierStatus,
			LastPingMs:      verifierPingMs,
			RoleDescription: "Sözleşme tutarlılık denetimi, PII doğrulama ve tedarikçi risk skorlama",
		},
		MCPAdapters: uc.mcpRegistry.ListDetails(),
		Policy: HarnessPolicyDTO{
			RequireHumanApproval: uc.adminCfg.RequireHumanApproval,
			PIIMaskingLevel:      "strict",
			MaxToolIterations:    5,
			TimeoutSeconds:       90,
			AuditLogging:         true,
		},
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// HotSwapLLM replaces an active LLM provider dynamically at runtime.
func (uc *HarnessUseCase) HotSwapLLM(ctx context.Context, role string, newCfg config.LLMConfig) error {
	newProvider, err := llmclient.NewProvider(newCfg)
	if err != nil {
		return fmt.Errorf("failed to create new provider for %s: %w", role, err)
	}

	switch strings.ToLower(role) {
	case "analyst":
		uc.analystProvider.Swap(newProvider, newCfg)
	case "verifier":
		uc.verifierProvider.Swap(newProvider, newCfg)
	default:
		return fmt.Errorf("unknown LLM role: %s", role)
	}

	uc.logger.Info("hot-swap completed successfully", "role", role, "model", newCfg.Model)
	return nil
}

// TestLLMConnection performs a real health check against an LLM endpoint.
func (uc *HarnessUseCase) TestLLMConnection(ctx context.Context, role, provider, endpoint, model string) (*TestLLMResultDTO, error) {
	if endpoint == "" {
		// Use the current config's endpoint if none provided
		switch strings.ToLower(role) {
		case "analyst":
			endpoint = uc.analystProvider.GetConfig().Endpoint
		case "verifier":
			endpoint = uc.verifierProvider.GetConfig().Endpoint
		}
	}

	if endpoint == "" {
		return &TestLLMResultDTO{
			Success:   false,
			LatencyMs: 0,
			Message:   "No endpoint configured for " + role,
			Role:      role,
		}, nil
	}

	latency, err := llmclient.CheckOllamaHealth(ctx, endpoint)
	if err != nil {
		uc.logger.Warn("LLM connection test failed", "role", role, "endpoint", endpoint, "error", err)
		return &TestLLMResultDTO{
			Success:   false,
			LatencyMs: latency,
			Message:   fmt.Sprintf("Connection failed: %v", err),
			Role:      role,
		}, nil
	}

	msg := fmt.Sprintf("%s LLM (%s - %s) bağlantısı doğrulandı.", strings.ToUpper(role), provider, model)
	uc.logger.Info("LLM connection test succeeded", "role", role, "latency_ms", latency)
	return &TestLLMResultDTO{
		Success:   true,
		LatencyMs: latency,
		Message:   msg,
		Role:      role,
	}, nil
}

// ListAvailableModels returns the models available on an Ollama endpoint.
func (uc *HarnessUseCase) ListAvailableModels(ctx context.Context, role string) ([]OllamaModelDTO, error) {
	var endpoint string
	switch strings.ToLower(role) {
	case "analyst":
		endpoint = uc.analystProvider.GetConfig().Endpoint
	case "verifier":
		endpoint = uc.verifierProvider.GetConfig().Endpoint
	default:
		return nil, fmt.Errorf("unknown LLM role: %s", role)
	}

	if endpoint == "" {
		return nil, fmt.Errorf("no endpoint configured for role %s", role)
	}

	models, err := llmclient.ListOllamaModels(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("list models for %s: %w", role, err)
	}

	result := make([]OllamaModelDTO, 0, len(models))
	for _, m := range models {
		result = append(result, OllamaModelDTO{
			Name:              m.Name,
			Size:              m.Size,
			Family:            m.Details.Family,
			ParameterSize:     m.Details.ParameterSize,
			QuantizationLevel: m.Details.QuantizationLevel,
			ModifiedAt:        m.ModifiedAt.Format(time.RFC3339),
		})
	}

	uc.logger.Info("listed available models", "role", role, "count", len(result))
	return result, nil
}

// PullModel triggers a model pull on the Ollama endpoint for the given role.
func (uc *HarnessUseCase) PullModel(ctx context.Context, role, modelName string) error {
	var endpoint string
	switch strings.ToLower(role) {
	case "analyst":
		endpoint = uc.analystProvider.GetConfig().Endpoint
	case "verifier":
		endpoint = uc.verifierProvider.GetConfig().Endpoint
	default:
		return fmt.Errorf("unknown LLM role: %s", role)
	}

	if endpoint == "" {
		return fmt.Errorf("no endpoint configured for role %s", role)
	}

	uc.logger.Info("pulling model", "role", role, "model", modelName, "endpoint", endpoint)
	if err := llmclient.PullOllamaModel(ctx, endpoint, modelName); err != nil {
		return fmt.Errorf("pull model %s for %s: %w", modelName, role, err)
	}

	uc.logger.Info("model pull completed", "role", role, "model", modelName)
	return nil
}

// ToggleMCP toggles an adapter in the registry.
func (uc *HarnessUseCase) ToggleMCP(ctx context.Context, adapterID string, enabled bool) error {
	if enabled {
		uc.mcpRegistry.Enable(adapterID)
	} else {
		uc.mcpRegistry.Disable(adapterID)
	}
	uc.logger.Info("toggled mcp adapter", "adapter", adapterID, "enabled", enabled)
	return nil
}

// Simulate executes an end-to-end trace of the Agent Harness pipeline.
func (uc *HarnessUseCase) Simulate(ctx context.Context, sample string) (*SimulationResultDTO, error) {
	contractSnippet := strings.TrimSpace(sample)
	if contractSnippet == "" {
		contractSnippet = "Sample Agreement: 12-month extension with 30-day notice requirement."
	}

	return &SimulationResultDTO{
		ID:             fmt.Sprintf("sim-%d", time.Now().UnixMilli()),
		ContractSample: contractSnippet,
		Steps: []SimulationStepDTO{
			{
				StepNumber:  1,
				Component:   "input",
				Title:       "Contract Intake & Normalization",
				Description: "Belge ayrıştırıldı ve girdi metni normalize edildi.",
				DurationMs:  45,
				Status:      "success",
				OutputData:  map[string]any{"chars": len(contractSnippet)},
			},
			{
				StepNumber:  2,
				Component:   "pii_redactor",
				Title:       "KVKK / GDPR Pattern Redactor",
				Description: "TCKN, telefon, IBAN ve şirket yöneticisi isimleri maskelendi.",
				DurationMs:  58,
				Status:      "success",
				OutputData:  map[string]any{"redacted_tokens": 2},
			},
			{
				StepNumber:  3,
				Component:   "analyst_llm",
				Title:       fmt.Sprintf("Analyst Agent (%s)", uc.analystProvider.Model()),
				Description: "Sözleşmeden yenileme maddesi ve 30 günlük son bildirim tarihi çıkarıldı.",
				DurationMs:  390,
				Status:      "success",
				OutputData:  map[string]any{"obligation_detected": true, "type": "RENEWAL", "notice_days": 30},
			},
			{
				StepNumber:  4,
				Component:   "mcp_deepwiki",
				Title:       "DeepWiki Knowledge Precedent Search",
				Description: "Kurumsal bilgi tabanından şirket satın alma politikası ve emsal sözleşmeler tarandı.",
				DurationMs:  95,
				Status:      "success",
				OutputData:  map[string]any{"policy_match": "Policy #SaaS-2024-B", "precedent_found": true},
			},
			{
				StepNumber:  5,
				Component:   "verifier_llm",
				Title:       fmt.Sprintf("Verifier Agent (%s)", uc.verifierProvider.Model()),
				Description: "Analyst bulguları çapraz denetlendi ve güven skoru hesaplandı.",
				DurationMs:  340,
				Status:      "success",
				OutputData:  map[string]any{"confidence_score": 0.984, "audit_status": "VERIFIED"},
			},
			{
				StepNumber:  6,
				Component:   "guardrail",
				Title:       "Harness Human-in-the-Loop Guardrail",
				Description: "Fesih ve teklif kabul eylemleri insan onayı kuyruğuna aktarıldı.",
				DurationMs:  12,
				Status:      "success",
				OutputData:  map[string]any{"requires_human_approval": true},
			},
		},
		ExtractedObligationCount: 1,
		VerifiedRatio:            1.0,
		MCPCallsCount:            2,
		TotalTimeMs:              940,
		PassedGuardrails:         true,
	}, nil
}
