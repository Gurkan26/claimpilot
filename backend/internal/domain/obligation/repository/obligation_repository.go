package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/domain/obligation/model"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ObligationRepository defines the persistence contract for obligations.
type ObligationRepository interface {
	// Create inserts a new obligation record.
	Create(ctx context.Context, obl *model.Obligation) error

	// FindByID retrieves an obligation by its ObjectID.
	FindByID(ctx context.Context, id bson.ObjectID) (*model.Obligation, error)

	// FindByUserID retrieves obligations for a user with optional status filter.
	FindByUserID(ctx context.Context, userID uuid.UUID, statusFilter []model.ObligationStatus, limit, offset int64) ([]*model.Obligation, error)

	// FindByDocumentID retrieves obligations extracted from a specific document.
	FindByDocumentID(ctx context.Context, docID bson.ObjectID) ([]*model.Obligation, error)

	// FindUpcoming retrieves obligations due within the given number of days.
	FindUpcoming(ctx context.Context, userID uuid.UUID, daysAhead int) ([]*model.Obligation, error)

	// UpdateStatus transitions the obligation to a new status.
	UpdateStatus(ctx context.Context, id bson.ObjectID, status model.ObligationStatus) error

	// UpdateAgentOutput stores analyst or verifier output.
	UpdateAgentOutput(ctx context.Context, id bson.ObjectID, output *model.AgentOutput) error

	// UpdateVerification stores verifier output and updates status and risk level.
	UpdateVerification(ctx context.Context, id bson.ObjectID, status model.ObligationStatus, riskLevel model.RiskLevel, output *model.AgentOutput) error

	// SetApprovedAction records the user-approved action.
	SetApprovedAction(ctx context.Context, id bson.ObjectID, action *model.Action) error

	// LinkMarketplaceOpportunity associates an obligation with a marketplace opportunity.
	LinkMarketplaceOpportunity(ctx context.Context, oblID, oppID bson.ObjectID) error

	// Delete removes an obligation (GDPR/KVKK).
	Delete(ctx context.Context, id bson.ObjectID) error

	// DeleteByUserID removes all obligations for a user (account deletion).
	DeleteByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}
