package marketplace_test

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/agent/llmclient"
	"github.com/masterfabric-go/masterfabric/internal/agent/marketplace"
	"github.com/masterfabric-go/masterfabric/internal/agent/verifier"
	mktModel "github.com/masterfabric-go/masterfabric/internal/domain/marketplace/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type mockLLM struct{}

func (m *mockLLM) Complete(_ context.Context, _ *llmclient.CompletionRequest) (*llmclient.CompletionResponse, error) {
	return &llmclient.CompletionResponse{
		Content:    `{"trustworthy":true,"risk_score":0.12,"notes":"Low risk European hosting provider"}`,
		Model:      "mock-verifier",
		TokensUsed: 90,
	}, nil
}
func (m *mockLLM) StreamComplete(_ context.Context, _ *llmclient.CompletionRequest) (<-chan llmclient.StreamChunk, error) {
	ch := make(chan llmclient.StreamChunk, 1)
	ch <- llmclient.StreamChunk{Content: `{"trustworthy":true,"risk_score":0.12}`, Done: true}
	close(ch)
	return ch, nil
}
func (m *mockLLM) Name() string                     { return "mock" }
func (m *mockLLM) Model() string                    { return "mock-verifier" }
func (m *mockLLM) HealthCheck(_ context.Context) error { return nil }

type mockMarketplaceRepo struct {
	mu   sync.Mutex
	opps map[bson.ObjectID]*mktModel.MarketplaceOpportunity
	bids map[bson.ObjectID]*mktModel.Bid
}

func newMockMarketplaceRepo() *mockMarketplaceRepo {
	return &mockMarketplaceRepo{
		opps: make(map[bson.ObjectID]*mktModel.MarketplaceOpportunity),
		bids: make(map[bson.ObjectID]*mktModel.Bid),
	}
}

func (r *mockMarketplaceRepo) CreateOpportunity(_ context.Context, opp *mktModel.MarketplaceOpportunity) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if opp.ID.IsZero() {
		opp.ID = bson.NewObjectID()
	}
	r.opps[opp.ID] = opp
	return nil
}
func (r *mockMarketplaceRepo) FindOpportunityByID(_ context.Context, id bson.ObjectID) (*mktModel.MarketplaceOpportunity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.opps[id], nil
}
func (r *mockMarketplaceRepo) FindOpportunitiesByUserID(_ context.Context, _ uuid.UUID, _ []mktModel.OpportunityStatus, _, _ int64) ([]*mktModel.MarketplaceOpportunity, error) {
	return nil, nil
}
func (r *mockMarketplaceRepo) UpdateOpportunityStatus(_ context.Context, id bson.ObjectID, status mktModel.OpportunityStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if opp, ok := r.opps[id]; ok {
		opp.Status = status
	}
	return nil
}
func (r *mockMarketplaceRepo) CreateBid(_ context.Context, bid *mktModel.Bid) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if bid.ID.IsZero() {
		bid.ID = bson.NewObjectID()
	}
	r.bids[bid.ID] = bid
	return nil
}
func (r *mockMarketplaceRepo) FindBidByID(_ context.Context, id bson.ObjectID) (*mktModel.Bid, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.bids[id], nil
}
func (r *mockMarketplaceRepo) FindBidsByOpportunityID(_ context.Context, oppID bson.ObjectID) ([]*mktModel.Bid, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var list []*mktModel.Bid
	for _, b := range r.bids {
		if b.OpportunityID == oppID {
			list = append(list, b)
		}
	}
	return list, nil
}
func (r *mockMarketplaceRepo) UpdateBidStatus(_ context.Context, id bson.ObjectID, status mktModel.BidStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if b, ok := r.bids[id]; ok {
		b.Status = status
	}
	return nil
}
func (r *mockMarketplaceRepo) UpdateBidRiskScore(_ context.Context, id bson.ObjectID, score float64, notes string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if b, ok := r.bids[id]; ok {
		b.RiskScore = &score
		b.RiskNotes = notes
	}
	return nil
}
func (r *mockMarketplaceRepo) CreateTransaction(_ context.Context, _ *mktModel.Transaction) error {
	return nil
}
func (r *mockMarketplaceRepo) FindTransactionByID(_ context.Context, _ bson.ObjectID) (*mktModel.Transaction, error) {
	return nil, nil
}
func (r *mockMarketplaceRepo) FindTransactionsByUserID(_ context.Context, _ uuid.UUID, _, _ int64) ([]*mktModel.Transaction, error) {
	return nil, nil
}
func (r *mockMarketplaceRepo) UpdateTransactionStatus(_ context.Context, _ bson.ObjectID, _ mktModel.TransactionStatus) error {
	return nil
}
func (r *mockMarketplaceRepo) TotalGMV(_ context.Context, _ uuid.UUID) (*mktModel.Money, error) {
	return &mktModel.Money{Amount: 0, Currency: "USD"}, nil
}
func (r *mockMarketplaceRepo) TotalSavings(_ context.Context, _ uuid.UUID) (*mktModel.Money, error) {
	return &mktModel.Money{Amount: 0, Currency: "USD"}, nil
}
func (r *mockMarketplaceRepo) DeleteByUserID(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}

func TestRFQEngine_CollectBids(t *testing.T) {
	repo := newMockMarketplaceRepo()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	v := verifier.New(&mockLLM{}, logger)

	rfq := marketplace.NewRFQEngine(repo, v, logger)

	oppID := bson.NewObjectID()
	opp := &mktModel.MarketplaceOpportunity{
		ID:                oppID,
		ObligationID:      bson.NewObjectID(),
		UserID:            uuid.New(),
		Category:          "cloud-hosting",
		CurrentVendorName: "AWS EMEA SARL",
		CurrentVendorCost: &mktModel.Money{Amount: 3450.0, Currency: "EUR"},
		Status:            mktModel.OpportunityStatusOpen,
	}
	require.NoError(t, repo.CreateOpportunity(context.Background(), opp))

	ctx := context.Background()
	bids, err := rfq.CollectBids(ctx, oppID)
	require.NoError(t, err)

	assert.NotEmpty(t, bids)
	assert.GreaterOrEqual(t, len(bids), 2)

	// Check updated status
	updatedOpp, err := repo.FindOpportunityByID(ctx, oppID)
	require.NoError(t, err)
	assert.Equal(t, mktModel.OpportunityStatusMatched, updatedOpp.Status)
	assert.NotNil(t, updatedOpp.BestBidID)

	for _, b := range bids {
		assert.Equal(t, oppID, b.OpportunityID)
		assert.NotEmpty(t, b.Vendor.Name)
		assert.NotEqual(t, "AWS EMEA SARL", b.Vendor.Name)
		assert.Greater(t, b.SavingsAmount.Amount, 0.0)
		assert.NotNil(t, b.RiskScore)
		assert.InDelta(t, 0.12, *b.RiskScore, 0.01)
	}
}
