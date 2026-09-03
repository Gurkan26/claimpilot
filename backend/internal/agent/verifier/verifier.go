package verifier

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	oblModel "github.com/masterfabric-go/masterfabric/internal/domain/obligation/model"
	"github.com/masterfabric-go/masterfabric/internal/agent/analyst"
	"github.com/masterfabric-go/masterfabric/internal/agent/llmclient"
)

// Verifier is the AI agent responsible for verifying analyst outputs,
// assessing risk levels, and controlling autonomous action thresholds.
type Verifier struct {
	provider llmclient.Provider
	logger   *slog.Logger
}

// New creates a Verifier agent backed by the given LLM provider.
// The provider is configured via environment variables (LLM_VERIFIER_*).
func New(provider llmclient.Provider, logger *slog.Logger) *Verifier {
	return &Verifier{
		provider: provider,
		logger:   logger.With("agent", "verifier", "model", provider.Model()),
	}
}

// VerificationResult holds the Verifier's assessment of an obligation candidate.
type VerificationResult struct {
	Verified       bool               `json:"verified"`
	RiskLevel      oblModel.RiskLevel `json:"risk_level"`
	Confidence     float64            `json:"confidence"`
	SourceVerified bool               `json:"source_verified"` // was the source reference confirmed?
	Notes          string             `json:"notes"`
	SourceRefs     []string           `json:"source_refs"`     // verified page/clause references
	AutoApprove    bool               `json:"auto_approve"`    // within autonomous action threshold?
}

// BidVerificationResult holds the Verifier's assessment of a marketplace bid.
type BidVerificationResult struct {
	Trustworthy bool    `json:"trustworthy"`
	RiskScore   float64 `json:"risk_score"` // 0.0 (safe) to 1.0 (risky)
	Notes       string  `json:"notes"`
}

const verificationSystemPrompt = `You are a verification AI agent for ClaimPilot.
Your job is to verify the accuracy of obligation detections by cross-referencing
with the original document content. You must:
1. Confirm or deny each detected obligation
2. Verify source references (page, clause)
3. Assess risk level independently
4. Determine if the obligation falls within autonomous action thresholds

Respond ONLY with valid JSON:
{
  "verified": true,
  "risk_level": "LOW|MEDIUM|HIGH|CRITICAL",
  "confidence": 0.95,
  "source_verified": true,
  "notes": "Explanation of verification",
  "source_refs": ["Page 3, Clause 5.2"],
  "auto_approve": false
}`

const bidVerificationSystemPrompt = `You are a verification AI agent for ClaimPilot marketplace.
Assess the trustworthiness and risk of a vendor bid. Consider:
1. Price reasonableness (suspiciously low?)
2. Vendor reputation indicators
3. Contract terms fairness
4. Hidden costs or lock-in clauses

Respond ONLY with valid JSON:
{
  "trustworthy": true,
  "risk_score": 0.2,
  "notes": "Explanation"
}`

// cleanJSONContent strips markdown code fences if emitted by the LLM.
func cleanJSONContent(content string) string {
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "```json") {
		trimmed = strings.TrimPrefix(trimmed, "```json")
	} else if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```")
	}
	if strings.HasSuffix(trimmed, "```") {
		trimmed = strings.TrimSuffix(trimmed, "```")
	}
	return strings.TrimSpace(trimmed)
}

// VerifyObligation checks an analyst-detected obligation against the source content.
func (v *Verifier) VerifyObligation(ctx context.Context, candidate *analyst.ObligationCandidate, sourceContent string) (*VerificationResult, error) {
	v.logger.Info("verifying obligation",
		"type", candidate.Type,
		"title", candidate.Title,
	)

	candidateJSON, err := json.Marshal(candidate)
	if err != nil {
		return nil, fmt.Errorf("marshal obligation candidate: %w", err)
	}

	req := &llmclient.CompletionRequest{
		SystemPrompt: verificationSystemPrompt,
		Messages: []llmclient.Message{
			{
				Role: llmclient.RoleUser,
				Content: fmt.Sprintf(
					"Verify this detected obligation against the source document.\n\nDetected obligation:\n%s\n\nSource document content:\n%s",
					string(candidateJSON),
					sourceContent,
				),
			},
		},
		Temperature: 0.0, // zero temperature for strict verification
		MaxTokens:   2048,
	}

	resp, err := v.provider.Complete(ctx, req)
	if err != nil {
		v.logger.Error("obligation verification failed", "error", err)
		return nil, fmt.Errorf("verifier LLM call failed: %w", err)
	}

	v.logger.Info("obligation verification completed",
		"tokens_used", resp.TokensUsed,
		"duration", resp.Duration,
	)

	cleaned := cleanJSONContent(resp.Content)

	var result VerificationResult
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		v.logger.Error("failed to parse verification result", "error", err, "raw_content", resp.Content)
		return nil, fmt.Errorf("parse verifier output: %w", err)
	}

	// Rule-based safety guard: evaluate due date urgency
	if candidate.DueDate != "" {
		if dueDate, err := time.Parse("2006-01-02", candidate.DueDate); err == nil {
			daysUntil := int(time.Until(dueDate).Hours() / 24)
			if daysUntil <= 7 && daysUntil >= 0 {
				if result.RiskLevel == oblModel.RiskLevelLow || result.RiskLevel == oblModel.RiskLevelMedium {
					result.RiskLevel = oblModel.RiskLevelHigh
					result.Notes += fmt.Sprintf(" [Urgency Escalation: Only %d days remaining until deadline]", daysUntil)
				}
			} else if daysUntil < 0 {
				result.RiskLevel = oblModel.RiskLevelCritical
				result.Notes += " [Urgency Escalation: Deadline has already passed!]"
			}
		}
	}

	return &result, nil
}

// VerifyBid assesses the trustworthiness of a marketplace bid.
func (v *Verifier) VerifyBid(ctx context.Context, bidDescription string) (*BidVerificationResult, error) {
	v.logger.Info("verifying marketplace bid")

	req := &llmclient.CompletionRequest{
		SystemPrompt: bidVerificationSystemPrompt,
		Messages: []llmclient.Message{
			{
				Role:    llmclient.RoleUser,
				Content: fmt.Sprintf("Assess this vendor bid:\n\n%s", bidDescription),
			},
		},
		Temperature: 0.0,
		MaxTokens:   1024,
	}

	resp, err := v.provider.Complete(ctx, req)
	if err != nil {
		v.logger.Error("bid verification failed", "error", err)
		return nil, fmt.Errorf("verifier bid check failed: %w", err)
	}

	cleaned := cleanJSONContent(resp.Content)

	var result BidVerificationResult
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, fmt.Errorf("parse bid verification output: %w", err)
	}

	return &result, nil
}

// ProviderName returns the underlying LLM provider name (for audit trail).
func (v *Verifier) ProviderName() string {
	return v.provider.Name()
}

// AssessRisk performs an independent risk assessment on an obligation.
func (v *Verifier) AssessRisk(ctx context.Context, obligationDesc string) (oblModel.RiskLevel, string, error) {
	result, err := v.VerifyObligation(ctx, &analyst.ObligationCandidate{
		Description: obligationDesc,
	}, "")
	if err != nil {
		return oblModel.RiskLevelMedium, "", err
	}
	return result.RiskLevel, result.Notes, nil
}
