package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/domain/audit/model"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// AgentAuditRepository defines the persistence contract for agent audit log entries.
type AgentAuditRepository interface {
	// Create inserts a new audit log record.
	Create(ctx context.Context, entry *model.AgentAuditEntry) error

	// FindByUserID retrieves audit log entries for a user with pagination.
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int64) ([]*model.AgentAuditEntry, error)

	// FindByObligationID retrieves all audit trail entries for a given obligation.
	FindByObligationID(ctx context.Context, oblID bson.ObjectID) ([]*model.AgentAuditEntry, error)

	// DeleteByUserID removes all audit log entries for a user (KVKK account deletion).
	DeleteByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}
