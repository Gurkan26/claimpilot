package orchestrator_test

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/agent/analyst"
	"github.com/masterfabric-go/masterfabric/internal/agent/llmclient"
	"github.com/masterfabric-go/masterfabric/internal/agent/orchestrator"
	"github.com/masterfabric-go/masterfabric/internal/agent/verifier"
	docModel "github.com/masterfabric-go/masterfabric/internal/domain/document/model"
	mktModel "github.com/masterfabric-go/masterfabric/internal/domain/marketplace/model"
	oblModel "github.com/masterfabric-go/masterfabric/internal/domain/obligation/model"
	"github.com/masterfabric-go/masterfabric/internal/pii"
	"github.com/masterfabric-go/masterfabric/internal/shared/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type mockLLM struct {
	response string
}

func (m *mockLLM) Complete(_ context.Context, _ *llmclient.CompletionRequest) (*llmclient.CompletionResponse, error) {
	return &llmclient.CompletionResponse{Content: m.response, Model: "mock-model", TokensUsed: 200}, nil
}
func (m *mockLLM) StreamComplete(_ context.Context, _ *llmclient.CompletionRequest) (<-chan llmclient.StreamChunk, error) {
	ch := make(chan llmclient.StreamChunk, 1)
	ch <- llmclient.StreamChunk{Content: m.response, Done: true}
	close(ch)
	return ch, nil
}
func (m *mockLLM) Name() string                     { return "mock" }
func (m *mockLLM) Model() string                    { return "mock-model" }
func (m *mockLLM) HealthCheck(_ context.Context) error { return nil }

// in-memory test repositories
type mockDocRepo struct {
	mu   sync.Mutex
	docs map[bson.ObjectID]*docModel.Document
}

func (r *mockDocRepo) Create(_ context.Context, doc *docModel.Document) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.docs[doc.ID] = doc
	return nil
}
func (r *mockDocRepo) FindByID(_ context.Context, id bson.ObjectID) (*docModel.Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.docs[id], nil
}
func (r *mockDocRepo) FindByUserID(_ context.Context, _ uuid.UUID, _, _ int64) ([]*docModel.Document, error) {
	return nil, nil
}
func (r *mockDocRepo) FindByOrganizationID(_ context.Context, _ uuid.UUID, _, _ int64) ([]*docModel.Document, error) {
	return nil, nil
}
func (r *mockDocRepo) UpdateStatus(_ context.Context, id bson.ObjectID, status docModel.DocumentStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if d, ok := r.docs[id]; ok {
		d.Status = status
	}
	return nil
}
func (r *mockDocRepo) UpdateExtractionResult(_ context.Context, id bson.ObjectID, result *docModel.ExtractionResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if d, ok := r.docs[id]; ok {
		d.ExtractionResult = result
	}
	return nil
}
func (r *mockDocRepo) UpdateContent(_ context.Context, id bson.ObjectID, raw, red string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if d, ok := r.docs[id]; ok {
		d.RawContent = raw
		d.RedactedContent = red
	}
	return nil
}
func (r *mockDocRepo) Delete(_ context.Context, _ bson.ObjectID) error { return nil }
func (r *mockDocRepo) DeleteByUserID(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}

type mockOblRepo struct {
	mu   sync.Mutex
	obls map[bson.ObjectID]*oblModel.Obligation
}

func (r *mockOblRepo) Create(_ context.Context, obl *oblModel.Obligation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if obl.ID.IsZero() {
		obl.ID = bson.NewObjectID()
	}
	r.obls[obl.ID] = obl
	return nil
}
func (r *mockOblRepo) FindByID(_ context.Context, id bson.ObjectID) (*oblModel.Obligation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.obls[id], nil
}
func (r *mockOblRepo) FindByUserID(_ context.Context, _ uuid.UUID, _ []oblModel.ObligationStatus, _, _ int64) ([]*oblModel.Obligation, error) {
	return nil, nil
}
func (r *mockOblRepo) FindByDocumentID(_ context.Context, docID bson.ObjectID) ([]*oblModel.Obligation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var list []*oblModel.Obligation
	for _, o := range r.obls {
		if o.SourceDocumentID == docID {
			list = append(list, o)
		}
	}
	return list, nil
}
func (r *mockOblRepo) FindUpcoming(_ context.Context, _ uuid.UUID, _ int) ([]*oblModel.Obligation, error) {
	return nil, nil
}
func (r *mockOblRepo) UpdateStatus(_ context.Context, id bson.ObjectID, status oblModel.ObligationStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if o, ok := r.obls[id]; ok {
		o.Status = status
	}
	return nil
}
func (r *mockOblRepo) UpdateAgentOutput(_ context.Context, id bson.ObjectID, out *oblModel.AgentOutput) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if o, ok := r.obls[id]; ok {
		if out.AgentType == "verifier" {
			o.VerifierOutput = out
		} else {
			o.AnalystOutput = out
		}
	}
	return nil
}
func (r *mockOblRepo) UpdateVerification(_ context.Context, id bson.ObjectID, status oblModel.ObligationStatus, risk oblModel.RiskLevel, out *oblModel.AgentOutput) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if o, ok := r.obls[id]; ok {
		o.Status = status
		o.RiskLevel = risk
		o.VerifierOutput = out
	}
	return nil
}
func (r *mockOblRepo) SetApprovedAction(_ context.Context, _ bson.ObjectID, _ *oblModel.Action) error {
	return nil
}
func (r *mockOblRepo) LinkMarketplaceOpportunity(_ context.Context, oblID, oppID bson.ObjectID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if o, ok := r.obls[oblID]; ok {
		o.MarketplaceOpportunityID = &oppID
	}
	return nil
}
func (r *mockOblRepo) Delete(_ context.Context, _ bson.ObjectID) error { return nil }
func (r *mockOblRepo) DeleteByUserID(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}

type mockMktRepo struct {
	mu   sync.Mutex
	opps map[bson.ObjectID]*mktModel.MarketplaceOpportunity
}

func (r *mockMktRepo) CreateOpportunity(_ context.Context, opp *mktModel.MarketplaceOpportunity) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if opp.ID.IsZero() {
		opp.ID = bson.NewObjectID()
	}
	r.opps[opp.ID] = opp
	return nil
}
func (r *mockMktRepo) FindOpportunityByID(_ context.Context, id bson.ObjectID) (*mktModel.MarketplaceOpportunity, error) {
	return r.opps[id], nil
}
func (r *mockMktRepo) FindOpportunitiesByUserID(_ context.Context, _ uuid.UUID, _ []mktModel.OpportunityStatus, _, _ int64) ([]*mktModel.MarketplaceOpportunity, error) {
	return nil, nil
}
func (r *mockMktRepo) UpdateOpportunityStatus(_ context.Context, _ bson.ObjectID, _ mktModel.OpportunityStatus) error {
	return nil
}
func (r *mockMktRepo) CreateBid(_ context.Context, _ *mktModel.Bid) error { return nil }
func (r *mockMktRepo) FindBidByID(_ context.Context, _ bson.ObjectID) (*mktModel.Bid, error) {
	return nil, nil
}
func (r *mockMktRepo) FindBidsByOpportunityID(_ context.Context, _ bson.ObjectID) ([]*mktModel.Bid, error) {
	return nil, nil
}
func (r *mockMktRepo) UpdateBidStatus(_ context.Context, _ bson.ObjectID, _ mktModel.BidStatus) error {
	return nil
}
func (r *mockMktRepo) UpdateBidRiskScore(_ context.Context, _ bson.ObjectID, _ float64, _ string) error {
	return nil
}
func (r *mockMktRepo) CreateTransaction(_ context.Context, _ *mktModel.Transaction) error {
	return nil
}
func (r *mockMktRepo) FindTransactionByID(_ context.Context, _ bson.ObjectID) (*mktModel.Transaction, error) {
	return nil, nil
}
func (r *mockMktRepo) FindTransactionsByUserID(_ context.Context, _ uuid.UUID, _, _ int64) ([]*mktModel.Transaction, error) {
	return nil, nil
}
func (r *mockMktRepo) UpdateTransactionStatus(_ context.Context, _ bson.ObjectID, _ mktModel.TransactionStatus) error {
	return nil
}
func (r *mockMktRepo) TotalGMV(_ context.Context, _ uuid.UUID) (*mktModel.Money, error) {
	return &mktModel.Money{Amount: 0, Currency: "USD"}, nil
}
func (r *mockMktRepo) TotalSavings(_ context.Context, _ uuid.UUID) (*mktModel.Money, error) {
	return &mktModel.Money{Amount: 0, Currency: "USD"}, nil
}
func (r *mockMktRepo) DeleteByUserID(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}

type modelStatus = mktModel.OpportunityStatus

func TestOrchestrator_ProcessDocument_EndToEnd(t *testing.T) {
	analystJSON := `{"extraction":{"fields":[{"field_name":"vendor","value":"Adobe"},{"field_name":"annual_fee","value":"24000 USD"}],"summary":"Adobe agreement"},"obligations":[{"type":"RENEWAL","title":"Adobe Renewal","description":"Renews Oct 18","due_date":"2026-09-18","risk_level":"HIGH","source_ref":"Clause 2","confidence":0.95}],"opportunities":[{"category":"cloud-hosting","current_vendor_name":"Adobe","estimated_cost":24000.0,"currency":"USD","reason":"Low usage"}]}`
	verifierJSON := `{"verified":true,"risk_level":"HIGH","confidence":0.96,"source_verified":true,"notes":"Clause verified","source_refs":["Clause 2"],"auto_approve":false}`

	analystLLM := &mockLLM{response: analystJSON}
	verifierLLM := &mockLLM{response: verifierJSON}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	a := analyst.New(analystLLM, logger)
	v := verifier.New(verifierLLM, logger)
	redactor := pii.NewPatternRedactor()
	eventBus := events.NewInProcessBus(logger, 64)
	defer eventBus.Close()

	docRepo := &mockDocRepo{docs: make(map[bson.ObjectID]*docModel.Document)}
	oblRepo := &mockOblRepo{obls: make(map[bson.ObjectID]*oblModel.Obligation)}
	mktRepo := &mockMktRepo{opps: make(map[bson.ObjectID]*mktModel.MarketplaceOpportunity)}

	orch := orchestrator.New(orchestrator.Config{
		Analyst:  a,
		Verifier: v,
		Redactor: redactor,
		DocRepo:  docRepo,
		OblRepo:  oblRepo,
		MktRepo:  mktRepo,
		EventBus: eventBus,
		Logger:   logger,
	})

	docID := bson.NewObjectID()
	rawDocText := `Customer TCKN: 12345678901, Phone: +90 532 123 45 67. Contract with Adobe renewing on 2026-09-18.`
	doc := &docModel.Document{
		ID:         docID,
		UserID:     uuid.New(),
		FileName:   "adobe_contract.txt",
		FileType:   docModel.DocumentTypeContract,
		Status:     docModel.DocumentStatusUploaded,
		RawContent: rawDocText,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	require.NoError(t, docRepo.Create(context.Background(), doc))

	ctx := context.Background()
	err := orch.ProcessDocument(ctx, docID)
	require.NoError(t, err)

	// Verify Document status
	updatedDoc, err := docRepo.FindByID(ctx, docID)
	require.NoError(t, err)
	assert.Equal(t, docModel.DocumentStatusAnalyzed, updatedDoc.Status)
	assert.NotEmpty(t, updatedDoc.RedactedContent)
	assert.False(t, assert.ObjectsAreEqual(rawDocText, updatedDoc.RedactedContent))
	assert.NotNil(t, updatedDoc.ExtractionResult)

	// Verify Obligation created & verified
	obls, err := oblRepo.FindByDocumentID(ctx, docID)
	require.NoError(t, err)
	require.Len(t, obls, 1)
	assert.Equal(t, oblModel.ObligationTypeRenewal, obls[0].Type)
	assert.Equal(t, oblModel.ObligationStatusPendingApproval, obls[0].Status)
	assert.Equal(t, oblModel.RiskLevelHigh, obls[0].RiskLevel)
	assert.NotNil(t, obls[0].VerifierOutput)
	assert.NotNil(t, obls[0].MarketplaceOpportunityID)

	// Verify Marketplace Opportunity created
	assert.Len(t, mktRepo.opps, 1)
	for _, opp := range mktRepo.opps {
		assert.Equal(t, "Adobe", opp.CurrentVendorName)
		assert.Equal(t, obls[0].ID, opp.ObligationID)
	}
}
