package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/masterfabric-go/masterfabric/internal/application/obligation/dto"
	auditModel "github.com/masterfabric-go/masterfabric/internal/domain/audit/model"
	auditRepo "github.com/masterfabric-go/masterfabric/internal/domain/audit/repository"
	oblModel "github.com/masterfabric-go/masterfabric/internal/domain/obligation/model"
	oblRepo "github.com/masterfabric-go/masterfabric/internal/domain/obligation/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// DismissObligationUseCase marks an obligation as dismissed by the user.
type DismissObligationUseCase struct {
	oblRepo   oblRepo.ObligationRepository
	auditRepo auditRepo.AgentAuditRepository
	logger    *slog.Logger
}

// NewDismissObligationUseCase creates a new DismissObligationUseCase.
func NewDismissObligationUseCase(oblRepo oblRepo.ObligationRepository, auditRepo auditRepo.AgentAuditRepository, logger *slog.Logger) *DismissObligationUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &DismissObligationUseCase{
		oblRepo:   oblRepo,
		auditRepo: auditRepo,
		logger:    logger.With("usecase", "dismiss_obligation"),
	}
}

// Execute dismisses the obligation.
func (uc *DismissObligationUseCase) Execute(ctx context.Context, req dto.DismissObligationRequest) (*dto.ObligationResponse, error) {
	id, err := bson.ObjectIDFromHex(req.ObligationID)
	if err != nil {
		return nil, fmt.Errorf("invalid obligation ID: %w", err)
	}

	obl, err := uc.oblRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find obligation: %w", err)
	}

	if err := uc.oblRepo.UpdateStatus(ctx, id, oblModel.ObligationStatusDismissed); err != nil {
		return nil, fmt.Errorf("update status to dismissed: %w", err)
	}

	// Record audit log entry
	if uc.auditRepo != nil {
		auditEntry := &auditModel.AgentAuditEntry{
			UserID:         req.UserID,
			OrganizationID: obl.OrganizationID,
			ActionType:     auditModel.AuditActionObligationDismissed,
			ObligationID:   &obl.ID,
			DocumentID:     &obl.SourceDocumentID,
			ApprovalType:   auditModel.ApprovalTypeManual,
			Details: bson.M{
				"reason": req.Reason,
				"title":  obl.Title,
			},
			Status:    "SUCCESS",
			CreatedAt: time.Now().UTC(),
		}
		_ = uc.auditRepo.Create(ctx, auditEntry)
	}

	obl.Status = oblModel.ObligationStatusDismissed
	return dto.ToObligationResponse(obl), nil
}
