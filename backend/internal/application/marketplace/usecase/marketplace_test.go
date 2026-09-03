package usecase_test

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	mktAgent "github.com/masterfabric-go/masterfabric/internal/agent/marketplace"
	"github.com/masterfabric-go/masterfabric/internal/application/marketplace/dto"
	"github.com/masterfabric-go/masterfabric/internal/application/marketplace/usecase"
	auditModel "github.com/masterfabric-go/masterfabric/internal/domain/audit/model"
	mktModel "github.com/masterfabric-go/masterfabric/internal/domain/marketplace/model"
	"github.com/masterfabric-go/masterfabric/internal/mcp"
	"github.com/masterfabric-go/masterfabric/internal/mcp/gmail"
	"github.com/masterfabric-go/masterfabric/internal/mcp/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type inMemoryMarketplaceRepo struct {
	mu    sync.Mutex
	opps  map[bson.ObjectID]*mktModel.MarketplaceOpportunity
	bids  map[bson.ObjectID]*mktModel.Bid
	txns  map[bson.ObjectID]*mktModel.Transaction
}

func newInMemoryMarketplaceRepo() *inMemoryMarketplaceRepo {
	return &inMemoryMarketplaceRepo{
		opps: make(map[bson.ObjectID]*mktModel.MarketplaceOpportunity),
		bids: make(map[bson.ObjectID]*mktModel.Bid),
		txns: make(map[bson.ObjectID]*mktModel.Transaction),
	}
}

func (r *inMemoryMarketplaceRepo) CreateOpportunity(_ context.Context, opp *mktModel.MarketplaceOpportunity) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if opp.ID.IsZero() {
		opp.ID = bson.NewObjectID()
	}
	r.opps[opp.ID] = opp
	return nil
}

func (r *inMemoryMarketplaceRepo) FindOpportunityByID(_ context.Context, id bson.ObjectID) (*mktModel.MarketplaceOpportunity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	o, ok := r.opps[id]
	if !ok {
		return nil, os.ErrNotExist
	}
	return o, nil
}

func (r *inMemoryMarketplaceRepo) FindOpportunitiesByUserID(_ context.Context, userID uuid.UUID, statuses []mktModel.OpportunityStatus, _, _ int64) ([]*mktModel.MarketplaceOpportunity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var list []*mktModel.MarketplaceOpportunity
	for _, o := range r.opps {
		if o.UserID == userID {
			if len(statuses) > 0 {
				matched := false
				for _, s := range statuses {
					if o.Status == s {
						matched = true
						break
					}
				}
				if !matched {
					continue
				}
			}
			list = append(list, o)
		}
	}
	return list, nil
}

func (r *inMemoryMarketplaceRepo) UpdateOpportunityStatus(_ context.Context, id bson.ObjectID, status mktModel.OpportunityStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if o, ok := r.opps[id]; ok {
		o.Status = status
	}
	return nil
}

func (r *inMemoryMarketplaceRepo) CreateBid(_ context.Context, bid *mktModel.Bid) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if bid.ID.IsZero() {
		bid.ID = bson.NewObjectID()
	}
	r.bids[bid.ID] = bid
	return nil
}

func (r *inMemoryMarketplaceRepo) FindBidByID(_ context.Context, id bson.ObjectID) (*mktModel.Bid, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.bids[id]
	if !ok {
		return nil, os.ErrNotExist
	}
	return b, nil
}

func (r *inMemoryMarketplaceRepo) FindBidsByOpportunityID(_ context.Context, oppID bson.ObjectID) ([]*mktModel.Bid, error) {
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

func (r *inMemoryMarketplaceRepo) UpdateBidStatus(_ context.Context, id bson.ObjectID, status mktModel.BidStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if b, ok := r.bids[id]; ok {
		b.Status = status
	}
	return nil
}

func (r *inMemoryMarketplaceRepo) UpdateBidRiskScore(_ context.Context, id bson.ObjectID, score float64, notes string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if b, ok := r.bids[id]; ok {
		b.RiskScore = &score
		b.RiskNotes = notes
	}
	return nil
}

func (r *inMemoryMarketplaceRepo) CreateTransaction(_ context.Context, txn *mktModel.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if txn.ID.IsZero() {
		txn.ID = bson.NewObjectID()
	}
	r.txns[txn.ID] = txn
	return nil
}

func (r *inMemoryMarketplaceRepo) FindTransactionByID(_ context.Context, id bson.ObjectID) (*mktModel.Transaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.txns[id]
	if !ok {
		return nil, os.ErrNotExist
	}
	return t, nil
}

func (r *inMemoryMarketplaceRepo) FindTransactionsByUserID(_ context.Context, userID uuid.UUID, _, _ int64) ([]*mktModel.Transaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var list []*mktModel.Transaction
	for _, t := range r.txns {
		if t.UserID == userID {
			list = append(list, t)
		}
	}
	return list, nil
}

func (r *inMemoryMarketplaceRepo) UpdateTransactionStatus(_ context.Context, id bson.ObjectID, status mktModel.TransactionStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t, ok := r.txns[id]; ok {
		t.Status = status
	}
	return nil
}

func (r *inMemoryMarketplaceRepo) TotalGMV(_ context.Context, userID uuid.UUID) (*mktModel.Money, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	total := 0.0
	currency := "USD"
	for _, t := range r.txns {
		if t.UserID == userID && t.Status == mktModel.TransactionStatusCompleted {
			total += t.Amount.Amount
			currency = t.Amount.Currency
		}
	}
	return &mktModel.Money{Amount: total, Currency: currency}, nil
}

func (r *inMemoryMarketplaceRepo) TotalSavings(_ context.Context, userID uuid.UUID) (*mktModel.Money, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	total := 0.0
	currency := "USD"
	for _, t := range r.txns {
		if t.UserID == userID && t.Status == mktModel.TransactionStatusCompleted && t.SavingsRealized != nil {
			total += t.SavingsRealized.Amount
			currency = t.SavingsRealized.Currency
		}
	}
	return &mktModel.Money{Amount: total, Currency: currency}, nil
}

func (r *inMemoryMarketplaceRepo) DeleteByUserID(_ context.Context, userID uuid.UUID) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return 0, nil
}

type inMemoryAuditRepo struct {
	mu      sync.Mutex
	entries []*auditModel.AgentAuditEntry
}

func newInMemoryAuditRepo() *inMemoryAuditRepo {
	return &inMemoryAuditRepo{}
}

func (r *inMemoryAuditRepo) Create(_ context.Context, entry *auditModel.AgentAuditEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, entry)
	return nil
}

func (r *inMemoryAuditRepo) FindByUserID(_ context.Context, _ uuid.UUID, _, _ int64) ([]*auditModel.AgentAuditEntry, error) {
	return nil, nil
}

func (r *inMemoryAuditRepo) FindByObligationID(_ context.Context, _ bson.ObjectID) ([]*auditModel.AgentAuditEntry, error) {
	return nil, nil
}

func (r *inMemoryAuditRepo) DeleteByUserID(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}

func TestTriggerRFQUseCase_Execute(t *testing.T) {
	repo := newInMemoryMarketplaceRepo()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	rfqEngine := mktAgent.NewRFQEngine(repo, nil, logger)

	uc := usecase.NewTriggerRFQUseCase(rfqEngine, repo, nil, logger)

	userID := uuid.New()
	oppID := bson.NewObjectID()
	opp := &mktModel.MarketplaceOpportunity{
		ID:                oppID,
		ObligationID:      bson.NewObjectID(),
		UserID:            userID,
		Category:          "cloud-hosting",
		CurrentVendorName: "AWS EMEA SARL",
		CurrentVendorCost: &mktModel.Money{Amount: 3450.0, Currency: "EUR"},
		Status:            mktModel.OpportunityStatusOpen,
	}
	require.NoError(t, repo.CreateOpportunity(context.Background(), opp))

	ctx := context.Background()
	resp, err := uc.Execute(ctx, oppID.Hex())
	require.NoError(t, err)

	assert.Equal(t, oppID.Hex(), resp.ID)
	assert.GreaterOrEqual(t, len(resp.Bids), 2)
	assert.Equal(t, string(mktModel.OpportunityStatusMatched), resp.Status)
}

func TestAcceptBidUseCase_Execute(t *testing.T) {
	repo := newInMemoryMarketplaceRepo()
	auditRepo := newInMemoryAuditRepo()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	mcpReg := mcp.NewRegistry()
	mcpReg.Register(gmail.New(logger))
	mcpReg.Register(slack.New(logger))

	acceptUC := usecase.NewAcceptBidUseCase(usecase.AcceptBidConfig{
		MktRepo:   repo,
		AuditRepo: auditRepo,
		MCPReg:    mcpReg,
		Logger:    logger,
	})

	userID := uuid.New()
	oppID := bson.NewObjectID()
	bidID := bson.NewObjectID()
	otherBidID := bson.NewObjectID()

	opp := &mktModel.MarketplaceOpportunity{
		ID:                oppID,
		ObligationID:      bson.NewObjectID(),
		UserID:            userID,
		Category:          "cloud-hosting",
		CurrentVendorName: "AWS EMEA SARL",
		CurrentVendorCost: &mktModel.Money{Amount: 3450.0, Currency: "EUR"},
		Status:            mktModel.OpportunityStatusMatched,
	}
	require.NoError(t, repo.CreateOpportunity(context.Background(), opp))

	selectedBid := &mktModel.Bid{
		ID:            bidID,
		OpportunityID: oppID,
		Vendor:        mktModel.Vendor{ID: "v1", Name: "Hetzner Cloud", Category: "cloud-hosting"},
		Price:         mktModel.Money{Amount: 1656.0, Currency: "EUR"},
		SavingsAmount: &mktModel.Money{Amount: 1794.0, Currency: "EUR"},
		Status:        mktModel.BidStatusPending,
	}
	require.NoError(t, repo.CreateBid(context.Background(), selectedBid))

	otherBid := &mktModel.Bid{
		ID:            otherBidID,
		OpportunityID: oppID,
		Vendor:        mktModel.Vendor{ID: "v2", Name: "GCP", Category: "cloud-hosting"},
		Price:         mktModel.Money{Amount: 2484.0, Currency: "EUR"},
		Status:        mktModel.BidStatusPending,
	}
	require.NoError(t, repo.CreateBid(context.Background(), otherBid))

	ctx := context.Background()
	req := dto.AcceptBidRequest{
		OpportunityID: oppID.Hex(),
		BidID:         bidID.Hex(),
		UserID:        userID,
	}

	resp, err := acceptUC.Execute(ctx, req)
	require.NoError(t, err)

	assert.NotNil(t, resp.Transaction)
	assert.Equal(t, 1656.0, resp.Transaction.Amount.Amount)
	// 4% take rate of 1656.0 = 66.24
	assert.Equal(t, 66.24, resp.Transaction.Commission.Amount)
	assert.Equal(t, 0.04, resp.Transaction.CommissionRate)
	assert.Equal(t, 1794.0, resp.Transaction.SavingsRealized.Amount)
	assert.Equal(t, "Hetzner Cloud", resp.Transaction.VendorName)

	// Check bid statuses
	b1, _ := repo.FindBidByID(ctx, bidID)
	assert.Equal(t, mktModel.BidStatusAccepted, b1.Status)
	b2, _ := repo.FindBidByID(ctx, otherBidID)
	assert.Equal(t, mktModel.BidStatusRejected, b2.Status)

	// Check opportunity status
	updatedOpp, _ := repo.FindOpportunityByID(ctx, oppID)
	assert.Equal(t, mktModel.OpportunityStatusTransacted, updatedOpp.Status)

	// Check audit entry
	require.Len(t, auditRepo.entries, 1)
	assert.Equal(t, "marketplace", auditRepo.entries[0].AdapterID)
}

func TestGetMetricsUseCase_Execute(t *testing.T) {
	repo := newInMemoryMarketplaceRepo()
	metricsUC := usecase.NewGetMetricsUseCase(repo)

	userID := uuid.New()
	now := time.Now().UTC()

	txn := &mktModel.Transaction{
		ID:             bson.NewObjectID(),
		OpportunityID:  bson.NewObjectID(),
		BidID:          bson.NewObjectID(),
		UserID:         userID,
		Amount:         mktModel.Money{Amount: 1656.0, Currency: "EUR"},
		Commission:     mktModel.Money{Amount: 66.24, Currency: "EUR"},
		CommissionRate: 0.04,
		SavingsRealized: &mktModel.Money{Amount: 1794.0, Currency: "EUR"},
		Status:         mktModel.TransactionStatusCompleted,
		VendorName:     "Hetzner Cloud",
		CompletedAt:    &now,
	}
	require.NoError(t, repo.CreateTransaction(context.Background(), txn))

	ctx := context.Background()
	metrics, err := metricsUC.Execute(ctx, userID)
	require.NoError(t, err)

	assert.Equal(t, 1656.0, metrics.TotalGMV.Amount)
	assert.Equal(t, 1794.0, metrics.TotalSavings.Amount)
	assert.Equal(t, 66.24, metrics.EstimatedCommission.Amount)
	assert.Equal(t, 1, metrics.CompletedDeals)
	assert.Greater(t, metrics.AverageSavingsRate, 50.0) // 1794 / 3450 = 52%
}
