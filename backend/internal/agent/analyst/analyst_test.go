package analyst_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/masterfabric-go/masterfabric/internal/agent/analyst"
	"github.com/masterfabric-go/masterfabric/internal/agent/llmclient"
	docModel "github.com/masterfabric-go/masterfabric/internal/domain/document/model"
	oblModel "github.com/masterfabric-go/masterfabric/internal/domain/obligation/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProvider implements llmclient.Provider for testing
type mockProvider struct {
	response string
}

func (m *mockProvider) Complete(_ context.Context, _ *llmclient.CompletionRequest) (*llmclient.CompletionResponse, error) {
	return &llmclient.CompletionResponse{
		Content:    m.response,
		Model:      "mock-gemma",
		TokensUsed: 150,
	}, nil
}

func (m *mockProvider) StreamComplete(_ context.Context, _ *llmclient.CompletionRequest) (<-chan llmclient.StreamChunk, error) {
	ch := make(chan llmclient.StreamChunk, 1)
	ch <- llmclient.StreamChunk{Content: m.response, Done: true}
	close(ch)
	return ch, nil
}

func (m *mockProvider) Name() string                     { return "mock" }
func (m *mockProvider) Model() string                    { return "mock-gemma" }
func (m *mockProvider) HealthCheck(_ context.Context) error { return nil }

func TestAnalyst_AnalyzeContract(t *testing.T) {
	contractJSON := `{"extraction":{"fields":[{"field_name":"vendor_name","value":"Adobe","confidence":0.98},{"field_name":"auto_renewal","value":"true","confidence":0.95}],"summary":"Adobe Enterprise Contract"},"obligations":[{"type":"RENEWAL","title":"Adobe Renewal Notice","description":"Must cancel before Sept 18","due_date":"2026-09-18","risk_level":"HIGH","source_ref":"Clause 2.2","confidence":0.92}],"opportunities":[{"category":"saas","current_vendor_name":"Adobe","estimated_cost":24000.0,"currency":"USD","reason":"Low utilization 41%"}]}`

	// Test with markdown fence wrapping (common LLM pattern)
	wrappedResponse := "```json\n" + contractJSON + "\n```"

	mock := &mockProvider{response: wrappedResponse}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	a := analyst.New(mock, logger)

	ctx := context.Background()
	res, err := a.AnalyzeDocument(ctx, "Sample contract text", docModel.DocumentTypeContract)
	require.NoError(t, err)

	assert.NotNil(t, res.Extraction)
	assert.Equal(t, "Adobe Enterprise Contract", res.Extraction.Summary)
	assert.Len(t, res.Obligations, 1)
	assert.Equal(t, oblModel.ObligationTypeRenewal, res.Obligations[0].Type)
	assert.Equal(t, "2026-09-18", res.Obligations[0].DueDate)
	assert.Len(t, res.Opportunities, 1)
	assert.Equal(t, "Adobe", res.Opportunities[0].CurrentVendorName)
}

func TestAnalyst_AnalyzeInvoice(t *testing.T) {
	invoiceJSON := `{"extraction":{"fields":[{"field_name":"invoice_number","value":"INV-AWS-2026-9812","confidence":0.99},{"field_name":"total_amount","value":"3450.00","confidence":0.98}],"summary":"AWS Hosting Invoice"},"obligations":[{"type":"PAYMENT","title":"AWS Cloud Invoice Payment","description":"Pay 3,450 EUR","due_date":"2026-09-15","risk_level":"HIGH","source_ref":"Invoice Header","confidence":0.95}],"opportunities":[]}`

	mock := &mockProvider{response: invoiceJSON}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	a := analyst.New(mock, logger)

	ctx := context.Background()
	res, err := a.AnalyzeDocument(ctx, "AWS Invoice text", docModel.DocumentTypeInvoice)
	require.NoError(t, err)

	assert.NotNil(t, res.Extraction)
	assert.Len(t, res.Obligations, 1)
	assert.Equal(t, oblModel.ObligationTypePayment, res.Obligations[0].Type)
	assert.Equal(t, "2026-09-15", res.Obligations[0].DueDate)
}

func TestAnalyst_AnalyzeLicense(t *testing.T) {
	licenseJSON := `{"extraction":{"fields":[{"field_name":"software_name","value":"Datadog Pro","confidence":0.98},{"field_name":"utilization_rate","value":"38%","confidence":0.90}],"summary":"Datadog Monitoring License"},"obligations":[{"type":"RENEWAL","title":"Datadog License Expiry","description":"Annual renewal for 100 hosts","due_date":"2026-10-31","risk_level":"MEDIUM","source_ref":"Section 4.1","confidence":0.92}],"opportunities":[{"category":"saas","current_vendor_name":"Datadog","estimated_cost":21600.0,"currency":"USD","reason":"38% utilization rate - reduce committed seats"}]}`

	mock := &mockProvider{response: licenseJSON}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	a := analyst.New(mock, logger)

	ctx := context.Background()
	res, err := a.AnalyzeDocument(ctx, "Datadog License text", docModel.DocumentTypeLicense)
	require.NoError(t, err)

	assert.NotNil(t, res.Extraction)
	assert.Len(t, res.Obligations, 1)
	assert.Equal(t, oblModel.ObligationTypeRenewal, res.Obligations[0].Type)
	assert.Len(t, res.Opportunities, 1)
	assert.Contains(t, res.Opportunities[0].Reason, "38%")
}
