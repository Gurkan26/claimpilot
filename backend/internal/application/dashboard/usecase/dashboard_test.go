package usecase_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/application/dashboard/usecase"
	mktModel "github.com/masterfabric-go/masterfabric/internal/domain/marketplace/model"
	oblModel "github.com/masterfabric-go/masterfabric/internal/domain/obligation/model"
	"github.com/masterfabric-go/masterfabric/internal/mcp"
	"github.com/masterfabric-go/masterfabric/internal/mcp/calendar"
	"github.com/masterfabric-go/masterfabric/internal/mcp/gmail"
	"github.com/masterfabric-go/masterfabric/internal/mcp/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type mockDashboardOblRepo struct {
	obls []*oblModel.Obligation
}

func (r *mockDashboardOblRepo) Create(_ context.Context, _ *oblModel.Obligation) error { return nil }
func (r *mockDashboardOblRepo) FindByID(_ context.Context, _ bson.ObjectID) (*oblModel.Obligation, error) {
	return nil, nil
}
func (r *mockDashboardOblRepo) FindByUserID(_ context.Context, _ uuid.UUID, _ []oblModel.ObligationStatus, _, _ int64) ([]*oblModel.Obligation, error) {
	return r.obls, nil
}
func (r *mockDashboardOblRepo) FindByDocumentID(_ context.Context, _ bson.ObjectID) ([]*oblModel.Obligation, error) {
	return nil, nil
}
func (r *mockDashboardOblRepo) FindUpcoming(_ context.Context, _ uuid.UUID, _ int) ([]*oblModel.Obligation, error) {
	return nil, nil
}
func (r *mockDashboardOblRepo) UpdateStatus(_ context.Context, _ bson.ObjectID, _ oblModel.ObligationStatus) error {
	return nil
}
func (r *mockDashboardOblRepo) UpdateAgentOutput(_ context.Context, _ bson.ObjectID, _ *oblModel.AgentOutput) error {
	return nil
}
func (r *mockDashboardOblRepo) UpdateVerification(_ context.Context, _ bson.ObjectID, _ oblModel.ObligationStatus, _ oblModel.RiskLevel, _ *oblModel.AgentOutput) error {
	return nil
}
func (r *mockDashboardOblRepo) SetApprovedAction(_ context.Context, _ bson.ObjectID, _ *oblModel.Action) error {
	return nil
}
func (r *mockDashboardOblRepo) LinkMarketplaceOpportunity(_ context.Context, _, _ bson.ObjectID) error {
	return nil
}
func (r *mockDashboardOblRepo) Delete(_ context.Context, _ bson.ObjectID) error { return nil }
func (r *mockDashboardOblRepo) DeleteByUserID(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}

type mockDashboardMktRepo struct {
	mu sync.Mutex
}

func (r *mockDashboardMktRepo) CreateOpportunity(_ context.Context, _ *mktModel.MarketplaceOpportunity) error {
	return nil
}
func (r *mockDashboardMktRepo) FindOpportunityByID(_ context.Context, _ bson.ObjectID) (*mktModel.MarketplaceOpportunity, error) {
	return nil, nil
}
func (r *mockDashboardMktRepo) FindOpportunitiesByUserID(_ context.Context, _ uuid.UUID, _ []mktModel.OpportunityStatus, _, _ int64) ([]*mktModel.MarketplaceOpportunity, error) {
	return []*mktModel.MarketplaceOpportunity{
		{ID: bson.NewObjectID(), Category: "cloud-hosting", Status: mktModel.OpportunityStatusOpen},
	}, nil
}
func (r *mockDashboardMktRepo) UpdateOpportunityStatus(_ context.Context, _ bson.ObjectID, _ mktModel.OpportunityStatus) error {
	return nil
}
func (r *mockDashboardMktRepo) CreateBid(_ context.Context, _ *mktModel.Bid) error { return nil }
func (r *mockDashboardMktRepo) FindBidByID(_ context.Context, _ bson.ObjectID) (*mktModel.Bid, error) {
	return nil, nil
}
func (r *mockDashboardMktRepo) FindBidsByOpportunityID(_ context.Context, _ bson.ObjectID) ([]*mktModel.Bid, error) {
	return nil, nil
}
func (r *mockDashboardMktRepo) UpdateBidStatus(_ context.Context, _ bson.ObjectID, _ mktModel.BidStatus) error {
	return nil
}
func (r *mockDashboardMktRepo) UpdateBidRiskScore(_ context.Context, _ bson.ObjectID, _ float64, _ string) error {
	return nil
}
func (r *mockDashboardMktRepo) CreateTransaction(_ context.Context, _ *mktModel.Transaction) error {
	return nil
}
func (r *mockDashboardMktRepo) FindTransactionByID(_ context.Context, _ bson.ObjectID) (*mktModel.Transaction, error) {
	return nil, nil
}
func (r *mockDashboardMktRepo) FindTransactionsByUserID(_ context.Context, _ uuid.UUID, _, _ int64) ([]*mktModel.Transaction, error) {
	return nil, nil
}
func (r *mockDashboardMktRepo) UpdateTransactionStatus(_ context.Context, _ bson.ObjectID, _ mktModel.TransactionStatus) error {
	return nil
}
func (r *mockDashboardMktRepo) TotalGMV(_ context.Context, _ uuid.UUID) (*mktModel.Money, error) {
	return &mktModel.Money{Amount: 24000.0, Currency: "USD"}, nil
}
func (r *mockDashboardMktRepo) TotalSavings(_ context.Context, _ uuid.UUID) (*mktModel.Money, error) {
	return &mktModel.Money{Amount: 9800.0, Currency: "USD"}, nil
}
func (r *mockDashboardMktRepo) DeleteByUserID(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}

func TestGetDashboardSummaryUseCase_Execute(t *testing.T) {
	userID := uuid.New()
	now := time.Now().UTC()

	obls := []*oblModel.Obligation{
		{
			ID:        bson.NewObjectID(),
			UserID:    userID,
			Type:      oblModel.ObligationTypeRenewal,
			Title:     "Adobe Creative Cloud Renewal",
			DueDate:   now.Add(3 * 24 * time.Hour), // 3 days ahead -> upcoming 7 days
			Status:    oblModel.ObligationStatusPendingApproval,
			RiskLevel: oblModel.RiskLevelCritical,
		},
		{
			ID:        bson.NewObjectID(),
			UserID:    userID,
			Type:      oblModel.ObligationTypePayment,
			Title:     "AWS EMEA Hosting Invoice",
			DueDate:   now.Add(14 * 24 * time.Hour), // 14 days ahead -> upcoming 30 days
			Status:    oblModel.ObligationStatusPendingApproval,
			RiskLevel: oblModel.RiskLevelHigh,
		},
		{
			ID:        bson.NewObjectID(),
			UserID:    userID,
			Type:      oblModel.ObligationTypeReport,
			Title:     "Overdue Annual Audit",
			DueDate:   now.Add(-2 * 24 * time.Hour), // overdue
			Status:    oblModel.ObligationStatusDetected,
			RiskLevel: oblModel.RiskLevelMedium,
		},
	}

	oblRepo := &mockDashboardOblRepo{obls: obls}
	mktRepo := &mockDashboardMktRepo{}

	mcpReg := mcp.NewRegistry()
	mcpReg.Register(gmail.New(nil))
	mcpReg.Register(calendar.New(nil))
	mcpReg.Register(slack.New(nil))

	uc := usecase.NewGetDashboardSummaryUseCase(usecase.Config{
		OblRepo: oblRepo,
		MktRepo: mktRepo,
		MCPReg:  mcpReg,
	})

	ctx := context.Background()
	summary, err := uc.Execute(ctx, userID)
	require.NoError(t, err)

	assert.Equal(t, 3, summary.TotalObligations)
	assert.Equal(t, 2, summary.PendingApprovals)
	assert.Equal(t, 1, summary.OverdueCount)
	assert.Equal(t, 1, summary.Upcoming7Days)
	assert.Equal(t, 1, summary.Upcoming30Days)

	// Risk breakdown
	assert.Equal(t, 1, summary.RiskBreakdown.Critical)
	assert.Equal(t, 1, summary.RiskBreakdown.High)
	assert.Equal(t, 1, summary.RiskBreakdown.Medium)
	assert.Equal(t, 0, summary.RiskBreakdown.Low)

	// Critical deadlines list
	assert.Len(t, summary.CriticalDeadlines, 1)
	assert.Equal(t, "Adobe Creative Cloud Renewal", summary.CriticalDeadlines[0].Title)

	// Financials
	assert.Equal(t, 24000.0, summary.FinancialSummary.TotalGMV.Amount)
	assert.Equal(t, 9800.0, summary.FinancialSummary.TotalSavings.Amount)
	assert.Equal(t, 1, summary.FinancialSummary.ActiveOpportunities)

	// MCP status
	assert.Len(t, summary.AgentStatus.ActiveMCPAdapters, 3)

	// AI Executive briefing
	assert.NotEmpty(t, summary.AIBriefing)
	assert.Contains(t, summary.AIBriefing, "3 tracked obligations")
	assert.Contains(t, summary.AIBriefing, "overdue")
	assert.NotEmpty(t, summary.Greeting)
}
