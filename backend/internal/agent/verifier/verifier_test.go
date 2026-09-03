package verifier_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/masterfabric-go/masterfabric/internal/agent/analyst"
	"github.com/masterfabric-go/masterfabric/internal/agent/llmclient"
	"github.com/masterfabric-go/masterfabric/internal/agent/verifier"
	oblModel "github.com/masterfabric-go/masterfabric/internal/domain/obligation/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockProvider struct {
	response string
}

func (m *mockProvider) Complete(_ context.Context, _ *llmclient.CompletionRequest) (*llmclient.CompletionResponse, error) {
	return &llmclient.CompletionResponse{
		Content:    m.response,
		Model:      "mock-verifier",
		TokensUsed: 120,
	}, nil
}

func (m *mockProvider) StreamComplete(_ context.Context, _ *llmclient.CompletionRequest) (<-chan llmclient.StreamChunk, error) {
	ch := make(chan llmclient.StreamChunk, 1)
	ch <- llmclient.StreamChunk{Content: m.response, Done: true}
	close(ch)
	return ch, nil
}

func (m *mockProvider) Name() string                     { return "mock" }
func (m *mockProvider) Model() string                    { return "mock-verifier" }
func (m *mockProvider) HealthCheck(_ context.Context) error { return nil }

func TestVerifier_VerifyObligation(t *testing.T) {
	verifierJSON := `{"verified":true,"risk_level":"LOW","confidence":0.95,"source_verified":true,"notes":"Clause confirmed on Page 2","source_refs":["Clause 2.2"],"auto_approve":false}`

	mock := &mockProvider{response: verifierJSON}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	v := verifier.New(mock, logger)

	candidate := &analyst.ObligationCandidate{
		Type:        oblModel.ObligationTypeRenewal,
		Title:       "Renewal Notice",
		Description: "Must cancel 30 days prior",
		DueDate:     time.Now().Add(60 * 24 * time.Hour).Format("2006-01-02"), // 60 days ahead
		RiskLevel:   oblModel.RiskLevelLow,
	}

	ctx := context.Background()
	res, err := v.VerifyObligation(ctx, candidate, "Contract source content here with Clause 2.2")
	require.NoError(t, err)

	assert.True(t, res.Verified)
	assert.Equal(t, oblModel.RiskLevelLow, res.RiskLevel)
	assert.Equal(t, "mock", v.ProviderName())
}

func TestVerifier_UrgencyEscalation_NearDeadline(t *testing.T) {
	verifierJSON := `{"verified":true,"risk_level":"LOW","confidence":0.95,"source_verified":true,"notes":"Verified invoice date"}`

	mock := &mockProvider{response: verifierJSON}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	v := verifier.New(mock, logger)

	// Due in 3 days -> should automatically escalate to HIGH
	nearDueDate := time.Now().Add(3 * 24 * time.Hour).Format("2006-01-02")
	candidate := &analyst.ObligationCandidate{
		Type:      oblModel.ObligationTypePayment,
		Title:     "Urgent Invoice Payment",
		DueDate:   nearDueDate,
		RiskLevel: oblModel.RiskLevelLow,
	}

	ctx := context.Background()
	res, err := v.VerifyObligation(ctx, candidate, "Source content")
	require.NoError(t, err)

	assert.True(t, res.Verified)
	assert.Equal(t, oblModel.RiskLevelHigh, res.RiskLevel)
	assert.Contains(t, res.Notes, "Urgency Escalation")
}

func TestVerifier_UrgencyEscalation_PastDeadline(t *testing.T) {
	verifierJSON := `{"verified":true,"risk_level":"LOW","confidence":0.90,"source_verified":true,"notes":"Past invoice"}`

	mock := &mockProvider{response: verifierJSON}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	v := verifier.New(mock, logger)

	// Overdue by 2 days -> should automatically escalate to CRITICAL
	pastDueDate := time.Now().Add(-2 * 24 * time.Hour).Format("2006-01-02")
	candidate := &analyst.ObligationCandidate{
		Type:      oblModel.ObligationTypePayment,
		Title:     "Overdue Invoice",
		DueDate:   pastDueDate,
		RiskLevel: oblModel.RiskLevelLow,
	}

	ctx := context.Background()
	res, err := v.VerifyObligation(ctx, candidate, "Source content")
	require.NoError(t, err)

	assert.True(t, res.Verified)
	assert.Equal(t, oblModel.RiskLevelCritical, res.RiskLevel)
	assert.Contains(t, res.Notes, "Deadline has already passed")
}

func TestVerifier_VerifyBid(t *testing.T) {
	bidJSON := `{"trustworthy":true,"risk_score":0.15,"notes":"Reputable hosting vendor with transparent pricing"}`

	mock := &mockProvider{response: bidJSON}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	v := verifier.New(mock, logger)

	ctx := context.Background()
	res, err := v.VerifyBid(ctx, "Vendor: Hetzner Cloud, 48 EUR/month, 99.9% SLA")
	require.NoError(t, err)

	assert.True(t, res.Trustworthy)
	assert.InDelta(t, 0.15, res.RiskScore, 0.01)
	assert.Contains(t, res.Notes, "Reputable")
}
