package usecase_test

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/application/obligation/dto"
	"github.com/masterfabric-go/masterfabric/internal/application/obligation/usecase"
	auditModel "github.com/masterfabric-go/masterfabric/internal/domain/audit/model"
	oblModel "github.com/masterfabric-go/masterfabric/internal/domain/obligation/model"
	"github.com/masterfabric-go/masterfabric/internal/mcp"
	"github.com/masterfabric-go/masterfabric/internal/mcp/calendar"
	"github.com/masterfabric-go/masterfabric/internal/mcp/gmail"
	"github.com/masterfabric-go/masterfabric/internal/mcp/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type inMemoryOblRepo struct {
	mu   sync.Mutex
	obls map[bson.ObjectID]*oblModel.Obligation
}

func newInMemoryOblRepo() *inMemoryOblRepo {
	return &inMemoryOblRepo{obls: make(map[bson.ObjectID]*oblModel.Obligation)}
}

func (r *inMemoryOblRepo) Create(_ context.Context, obl *oblModel.Obligation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if obl.ID.IsZero() {
		obl.ID = bson.NewObjectID()
	}
	r.obls[obl.ID] = obl
	return nil
}

func (r *inMemoryOblRepo) FindByID(_ context.Context, id bson.ObjectID) (*oblModel.Obligation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	o, ok := r.obls[id]
	if !ok {
		return nil, os.ErrNotExist
	}
	return o, nil
}

func (r *inMemoryOblRepo) FindByUserID(_ context.Context, userID uuid.UUID, statuses []oblModel.ObligationStatus, _, _ int64) ([]*oblModel.Obligation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var list []*oblModel.Obligation
	for _, o := range r.obls {
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

func (r *inMemoryOblRepo) FindByDocumentID(_ context.Context, docID bson.ObjectID) ([]*oblModel.Obligation, error) {
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

func (r *inMemoryOblRepo) FindUpcoming(_ context.Context, userID uuid.UUID, daysAhead int) ([]*oblModel.Obligation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	deadline := time.Now().Add(time.Duration(daysAhead) * 24 * time.Hour)
	var list []*oblModel.Obligation
	for _, o := range r.obls {
		if o.UserID == userID && o.DueDate.Before(deadline) && o.DueDate.After(time.Now().Add(-24*time.Hour)) {
			list = append(list, o)
		}
	}
	return list, nil
}

func (r *inMemoryOblRepo) UpdateStatus(_ context.Context, id bson.ObjectID, status oblModel.ObligationStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if o, ok := r.obls[id]; ok {
		o.Status = status
	}
	return nil
}

func (r *inMemoryOblRepo) UpdateAgentOutput(_ context.Context, id bson.ObjectID, output *oblModel.AgentOutput) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if o, ok := r.obls[id]; ok {
		if output.AgentType == "verifier" {
			o.VerifierOutput = output
		} else {
			o.AnalystOutput = output
		}
	}
	return nil
}

func (r *inMemoryOblRepo) UpdateVerification(_ context.Context, id bson.ObjectID, status oblModel.ObligationStatus, risk oblModel.RiskLevel, output *oblModel.AgentOutput) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if o, ok := r.obls[id]; ok {
		o.Status = status
		o.RiskLevel = risk
		o.VerifierOutput = output
	}
	return nil
}

func (r *inMemoryOblRepo) SetApprovedAction(_ context.Context, id bson.ObjectID, action *oblModel.Action) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if o, ok := r.obls[id]; ok {
		o.ApprovedAction = action
		o.Status = oblModel.ObligationStatusInProgress
	}
	return nil
}

func (r *inMemoryOblRepo) LinkMarketplaceOpportunity(_ context.Context, oblID, oppID bson.ObjectID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if o, ok := r.obls[oblID]; ok {
		o.MarketplaceOpportunityID = &oppID
	}
	return nil
}

func (r *inMemoryOblRepo) Delete(_ context.Context, id bson.ObjectID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.obls, id)
	return nil
}

func (r *inMemoryOblRepo) DeleteByUserID(_ context.Context, userID uuid.UUID) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var count int64
	for id, o := range r.obls {
		if o.UserID == userID {
			delete(r.obls, id)
			count++
		}
	}
	return count, nil
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
	if entry.ID.IsZero() {
		entry.ID = bson.NewObjectID()
	}
	r.entries = append(r.entries, entry)
	return nil
}

func (r *inMemoryAuditRepo) FindByUserID(_ context.Context, userID uuid.UUID, _, _ int64) ([]*auditModel.AgentAuditEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var list []*auditModel.AgentAuditEntry
	for _, e := range r.entries {
		if e.UserID == userID {
			list = append(list, e)
		}
	}
	return list, nil
}

func (r *inMemoryAuditRepo) FindByObligationID(_ context.Context, oblID bson.ObjectID) ([]*auditModel.AgentAuditEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var list []*auditModel.AgentAuditEntry
	for _, e := range r.entries {
		if e.ObligationID != nil && *e.ObligationID == oblID {
			list = append(list, e)
		}
	}
	return list, nil
}

func (r *inMemoryAuditRepo) DeleteByUserID(_ context.Context, userID uuid.UUID) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var count int64
	var filtered []*auditModel.AgentAuditEntry
	for _, e := range r.entries {
		if e.UserID == userID {
			count++
		} else {
			filtered = append(filtered, e)
		}
	}
	r.entries = filtered
	return count, nil
}

func TestApproveObligationUseCase_Execute(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	oblRepo := newInMemoryOblRepo()
	auditRepo := newInMemoryAuditRepo()

	mcpReg := mcp.NewRegistry()
	mcpReg.Register(gmail.New(logger))
	mcpReg.Register(calendar.New(logger))
	mcpReg.Register(slack.New(logger))

	approveUC := usecase.NewApproveObligationUseCase(usecase.ApproveConfig{
		OblRepo:   oblRepo,
		AuditRepo: auditRepo,
		MCPReg:    mcpReg,
		Logger:    logger,
	})

	userID := uuid.New()
	oblID := bson.NewObjectID()
	obl := &oblModel.Obligation{
		ID:               oblID,
		UserID:           userID,
		SourceDocumentID: bson.NewObjectID(),
		Type:             oblModel.ObligationTypeRenewal,
		Title:            "Adobe Enterprise Renewal",
		Description:      "Cancel 30 days prior to Oct 18",
		DueDate:          time.Now().Add(10 * 24 * time.Hour),
		Status:           oblModel.ObligationStatusPendingApproval,
		RiskLevel:        oblModel.RiskLevelHigh,
		SuggestedAction: &oblModel.Action{
			Type:        "send_email",
			Description: "Send notice to renewals@adobe.com",
			MCPAdapter:  "gmail",
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	require.NoError(t, oblRepo.Create(context.Background(), obl))

	ctx := context.Background()
	req := dto.ApproveObligationRequest{
		ObligationID: oblID.Hex(),
		UserID:       userID,
		ExecuteMCP:   true,
		Parameters: map[string]any{
			"to": "renewals@adobe.com",
		},
	}

	resp, err := approveUC.Execute(ctx, req)
	require.NoError(t, err)

	assert.NotNil(t, resp.Obligation)
	assert.Equal(t, string(oblModel.ObligationStatusResolved), resp.Obligation.Status)
	assert.NotNil(t, resp.MCPResult)
	assert.True(t, resp.MCPResult.Success)
	assert.Contains(t, resp.MCPResult.Message, "renewals@adobe.com")

	// Verify Audit Log was recorded
	require.Len(t, auditRepo.entries, 1)
	assert.Equal(t, auditModel.AuditActionObligationApproved, auditRepo.entries[0].ActionType)
	assert.Equal(t, "gmail", auditRepo.entries[0].AdapterID)
	assert.Equal(t, "SUCCESS", auditRepo.entries[0].Status)
}

func TestDismissObligationUseCase_Execute(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	oblRepo := newInMemoryOblRepo()
	auditRepo := newInMemoryAuditRepo()

	dismissUC := usecase.NewDismissObligationUseCase(oblRepo, auditRepo, logger)

	userID := uuid.New()
	oblID := bson.NewObjectID()
	obl := &oblModel.Obligation{
		ID:               oblID,
		UserID:           userID,
		SourceDocumentID: bson.NewObjectID(),
		Type:             oblModel.ObligationTypeReport,
		Title:            "Quarterly Compliance Report",
		Status:           oblModel.ObligationStatusPendingApproval,
		CreatedAt:        time.Now().UTC(),
	}
	require.NoError(t, oblRepo.Create(context.Background(), obl))

	ctx := context.Background()
	req := dto.DismissObligationRequest{
		ObligationID: oblID.Hex(),
		UserID:       userID,
		Reason:       "Already renewed manually",
	}

	resp, err := dismissUC.Execute(ctx, req)
	require.NoError(t, err)

	assert.Equal(t, string(oblModel.ObligationStatusDismissed), resp.Status)
	require.Len(t, auditRepo.entries, 1)
	assert.Equal(t, auditModel.AuditActionObligationDismissed, auditRepo.entries[0].ActionType)
}

func TestListObligationsUseCase_Execute(t *testing.T) {
	oblRepo := newInMemoryOblRepo()
	listUC := usecase.NewListObligationsUseCase(oblRepo)

	userID := uuid.New()
	obl1 := &oblModel.Obligation{
		ID:        bson.NewObjectID(),
		UserID:    userID,
		Type:      oblModel.ObligationTypeRenewal,
		Title:     "Renewal 1",
		DueDate:   time.Now().Add(5 * 24 * time.Hour),
		Status:    oblModel.ObligationStatusPendingApproval,
		RiskLevel: oblModel.RiskLevelHigh,
	}
	obl2 := &oblModel.Obligation{
		ID:        bson.NewObjectID(),
		UserID:    userID,
		Type:      oblModel.ObligationTypePayment,
		Title:     "Payment 1",
		DueDate:   time.Now().Add(25 * 24 * time.Hour),
		Status:    oblModel.ObligationStatusInProgress,
		RiskLevel: oblModel.RiskLevelLow,
	}
	require.NoError(t, oblRepo.Create(context.Background(), obl1))
	require.NoError(t, oblRepo.Create(context.Background(), obl2))

	ctx := context.Background()

	// Query upcoming in 7 days
	resUpcoming, err := listUC.Execute(ctx, usecase.FilterParams{
		UserID:    userID,
		DaysAhead: 7,
	})
	require.NoError(t, err)
	assert.Len(t, resUpcoming.Items, 1)
	assert.Equal(t, "Renewal 1", resUpcoming.Items[0].Title)

	// Query by status PENDING_APPROVAL
	resStatus, err := listUC.Execute(ctx, usecase.FilterParams{
		UserID:   userID,
		Statuses: []oblModel.ObligationStatus{oblModel.ObligationStatusPendingApproval},
	})
	require.NoError(t, err)
	assert.Len(t, resStatus.Items, 1)

	// Query single by ID
	resSingle, err := listUC.GetByID(ctx, obl1.ID.Hex())
	require.NoError(t, err)
	assert.Equal(t, obl1.ID.Hex(), resSingle.ID)
}
