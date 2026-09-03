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
	"github.com/masterfabric-go/masterfabric/internal/mcp"
	"github.com/masterfabric-go/masterfabric/internal/shared/events"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ApproveObligationUseCase handles user approval of obligation actions and dispatches MCP tools.
type ApproveObligationUseCase struct {
	oblRepo   oblRepo.ObligationRepository
	auditRepo auditRepo.AgentAuditRepository
	mcpReg    *mcp.Registry
	eventBus  events.EventBus
	logger    *slog.Logger
}

// ApproveConfig holds dependencies for ApproveObligationUseCase.
type ApproveConfig struct {
	OblRepo   oblRepo.ObligationRepository
	AuditRepo auditRepo.AgentAuditRepository
	MCPReg    *mcp.Registry
	EventBus  events.EventBus
	Logger    *slog.Logger
}

// NewApproveObligationUseCase creates a new ApproveObligationUseCase.
func NewApproveObligationUseCase(cfg ApproveConfig) *ApproveObligationUseCase {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &ApproveObligationUseCase{
		oblRepo:   cfg.OblRepo,
		auditRepo: cfg.AuditRepo,
		mcpReg:    cfg.MCPReg,
		eventBus:  cfg.EventBus,
		logger:    logger.With("usecase", "approve_obligation"),
	}
}

// Execute approves the obligation and optionally dispatches the associated MCP action.
func (uc *ApproveObligationUseCase) Execute(ctx context.Context, req dto.ApproveObligationRequest) (*dto.ApproveObligationResponse, error) {
	id, err := bson.ObjectIDFromHex(req.ObligationID)
	if err != nil {
		return nil, fmt.Errorf("invalid obligation ID: %w", err)
	}

	obl, err := uc.oblRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find obligation: %w", err)
	}

	// Determine action details
	adapterName := "slack"
	actionType := mcp.ActionTypeNotify
	actionDesc := fmt.Sprintf("Approved action for obligation: %s", obl.Title)

	if obl.SuggestedAction != nil && obl.SuggestedAction.MCPAdapter != "" {
		adapterName = obl.SuggestedAction.MCPAdapter
		actionDesc = obl.SuggestedAction.Description
		if obl.SuggestedAction.Type != "" {
			actionType = mcp.ActionType(obl.SuggestedAction.Type)
		}
	} else {
		switch obl.Type {
		case oblModel.ObligationTypeRenewal, oblModel.ObligationTypeCancellation:
			adapterName = "gmail"
			actionType = mcp.ActionTypeSendEmail
		case oblModel.ObligationTypePayment:
			adapterName = "slack"
			actionType = mcp.ActionTypeSendSlack
		default:
			adapterName = "calendar"
			actionType = mcp.ActionTypeCreateEvent
		}
	}

	// Update approved action on obligation
	approvedAction := &oblModel.Action{
		Type:        string(actionType),
		Description: actionDesc,
		MCPAdapter:  adapterName,
	}
	if err := uc.oblRepo.SetApprovedAction(ctx, id, approvedAction); err != nil {
		return nil, fmt.Errorf("record approved action: %w", err)
	}

	var mcpResult *mcp.ActionResult
	auditStatus := "SUCCESS"
	var auditErr string

	// Dispatch MCP Action
	if req.ExecuteMCP && uc.mcpReg != nil {
		if adapter, ok := uc.mcpReg.Get(adapterName); ok {
			params := req.Parameters
			if params == nil {
				params = make(map[string]any)
			}
			params["title"] = obl.Title
			params["due_date"] = obl.DueDate.Format("2006-01-02")
			params["description"] = obl.Description
			params["risk_level"] = string(obl.RiskLevel)

			mcpAction := &mcp.Action{
				Type:       actionType,
				AdapterID:  adapterName,
				Parameters: params,
			}

			result, err := adapter.Execute(ctx, mcpAction)
			if err != nil {
				uc.logger.Error("mcp action execution failed", "error", err, "adapter", adapterName)
				auditStatus = "FAILED"
				auditErr = err.Error()
			} else {
				mcpResult = result
				uc.logger.Info("mcp action executed successfully",
					"adapter", adapterName,
					"message", result.Message,
				)
				// Transition status to RESOLVED upon successful MCP execution
				_ = uc.oblRepo.UpdateStatus(ctx, id, oblModel.ObligationStatusResolved)
			}
		} else {
			uc.logger.Warn("mcp adapter not found in registry", "adapter", adapterName)
		}
	}

	// Record immutable Audit Log entry
	if uc.auditRepo != nil {
		auditEntry := &auditModel.AgentAuditEntry{
			UserID:         req.UserID,
			OrganizationID: obl.OrganizationID,
			ActionType:     auditModel.AuditActionObligationApproved,
			ObligationID:   &obl.ID,
			DocumentID:     &obl.SourceDocumentID,
			AdapterID:      adapterName,
			ApprovalType:   auditModel.ApprovalTypeManual,
			Details: bson.M{
				"title":       obl.Title,
				"action_type": string(actionType),
				"mcp_result":  mcpResult,
			},
			Status:       auditStatus,
			ErrorMessage: auditErr,
			CreatedAt:    time.Now().UTC(),
		}
		if err := uc.auditRepo.Create(ctx, auditEntry); err != nil {
			uc.logger.Error("failed to record audit log", "error", err)
		}
	}

	// Publish domain event
	if uc.eventBus != nil {
		_ = uc.eventBus.Publish(ctx, events.TopicTenant, map[string]interface{}{
			"type":          "obligation.approved",
			"obligation_id": obl.ID.Hex(),
			"user_id":       req.UserID.String(),
			"mcp_adapter":   adapterName,
			"mcp_success":   auditStatus == "SUCCESS",
		})
	}

	// Reload updated obligation
	updatedObl, _ := uc.oblRepo.FindByID(ctx, id)
	if updatedObl != nil {
		obl = updatedObl
	}

	return &dto.ApproveObligationResponse{
		Obligation: dto.ToObligationResponse(obl),
		MCPResult:  mcpResult,
	}, nil
}
