package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/masterfabric-go/masterfabric/internal/application/marketplace/dto"
	auditModel "github.com/masterfabric-go/masterfabric/internal/domain/audit/model"
	auditRepo "github.com/masterfabric-go/masterfabric/internal/domain/audit/repository"
	mktModel "github.com/masterfabric-go/masterfabric/internal/domain/marketplace/model"
	mktRepo "github.com/masterfabric-go/masterfabric/internal/domain/marketplace/repository"
	"github.com/masterfabric-go/masterfabric/internal/mcp"
	"github.com/masterfabric-go/masterfabric/internal/shared/events"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const defaultPlatformTakeRate = 0.04 // 4% platform commission

// AcceptBidUseCase handles accepting a vendor bid, generating a transaction,
// calculating take-rate, dispatching MCP notifications, and auditing.
type AcceptBidUseCase struct {
	mktRepo   mktRepo.MarketplaceRepository
	auditRepo auditRepo.AgentAuditRepository
	mcpReg    *mcp.Registry
	eventBus  events.EventBus
	logger    *slog.Logger
}

// AcceptBidConfig holds dependencies for AcceptBidUseCase.
type AcceptBidConfig struct {
	MktRepo   mktRepo.MarketplaceRepository
	AuditRepo auditRepo.AgentAuditRepository
	MCPReg    *mcp.Registry
	EventBus  events.EventBus
	Logger    *slog.Logger
}

// NewAcceptBidUseCase creates a new AcceptBidUseCase.
func NewAcceptBidUseCase(cfg AcceptBidConfig) *AcceptBidUseCase {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &AcceptBidUseCase{
		mktRepo:   cfg.MktRepo,
		auditRepo: cfg.AuditRepo,
		mcpReg:    cfg.MCPReg,
		eventBus:  cfg.EventBus,
		logger:    logger.With("usecase", "accept_bid"),
	}
}

// Execute processes the acceptance of a vendor bid and records the transaction.
func (uc *AcceptBidUseCase) Execute(ctx context.Context, req dto.AcceptBidRequest) (*dto.AcceptBidResponse, error) {
	oppID, err := bson.ObjectIDFromHex(req.OpportunityID)
	if err != nil {
		return nil, fmt.Errorf("invalid opportunity ID: %w", err)
	}

	bidID, err := bson.ObjectIDFromHex(req.BidID)
	if err != nil {
		return nil, fmt.Errorf("invalid bid ID: %w", err)
	}

	// 1. Fetch opportunity & bid
	opp, err := uc.mktRepo.FindOpportunityByID(ctx, oppID)
	if err != nil {
		return nil, fmt.Errorf("find opportunity: %w", err)
	}

	bid, err := uc.mktRepo.FindBidByID(ctx, bidID)
	if err != nil {
		return nil, fmt.Errorf("find bid: %w", err)
	}

	if bid.OpportunityID != oppID {
		return nil, fmt.Errorf("bid does not belong to opportunity")
	}

	// 2. Update bid statuses
	if err := uc.mktRepo.UpdateBidStatus(ctx, bidID, mktModel.BidStatusAccepted); err != nil {
		return nil, fmt.Errorf("update bid status: %w", err)
	}

	allBids, _ := uc.mktRepo.FindBidsByOpportunityID(ctx, oppID)
	for _, b := range allBids {
		if b.ID != bidID && b.Status == mktModel.BidStatusPending {
			_ = uc.mktRepo.UpdateBidStatus(ctx, b.ID, mktModel.BidStatusRejected)
		}
	}

	// 3. Calculate take-rate commission
	grossAmount := bid.Price.Amount
	commissionAmount := math.Round(grossAmount*defaultPlatformTakeRate*100) / 100
	now := time.Now().UTC()

	// 4. Create Transaction record (contributes to platform GMV)
	txn := &mktModel.Transaction{
		ID:             bson.NewObjectID(),
		OpportunityID:  oppID,
		BidID:          bidID,
		UserID:         req.UserID,
		OrganizationID: opp.OrganizationID,
		Amount:         bid.Price,
		Commission: mktModel.Money{
			Amount:   commissionAmount,
			Currency: bid.Price.Currency,
		},
		CommissionRate:  defaultPlatformTakeRate,
		SavingsRealized: bid.SavingsAmount,
		Status:          mktModel.TransactionStatusCompleted,
		VendorName:      bid.Vendor.Name,
		Notes: fmt.Sprintf("Autonomous deal closed via ClaimPilot Marketplace. Switched from %s to %s.",
			opp.CurrentVendorName, bid.Vendor.Name),
		CompletedAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.mktRepo.CreateTransaction(ctx, txn); err != nil {
		return nil, fmt.Errorf("create transaction: %w", err)
	}

	// 5. Update opportunity status to TRANSACTED
	_ = uc.mktRepo.UpdateOpportunityStatus(ctx, oppID, mktModel.OpportunityStatusTransacted)

	// 6. Dispatch MCP Actions (Gmail non-renewal notice, Slack celebration)
	if uc.mcpReg != nil {
		// A. Gmail notice to previous vendor
		if gmailAdapter, ok := uc.mcpReg.Get("gmail"); ok {
			_, _ = gmailAdapter.Execute(ctx, &mcp.Action{
				Type:      mcp.ActionTypeCancelContract,
				AdapterID: "gmail",
				Parameters: map[string]any{
					"vendor":       opp.CurrentVendorName,
					"contract_ref": fmt.Sprintf("Account %s", req.UserID.String()[:8]),
				},
			})
		}

		// B. Slack announcement to procurement channel
		if slackAdapter, ok := uc.mcpReg.Get("slack"); ok {
			savingsDesc := "substantial cost savings"
			if bid.SavingsAmount != nil {
				savingsDesc = fmt.Sprintf("%.2f %s annual savings", bid.SavingsAmount.Amount, bid.SavingsAmount.Currency)
			}
			msg := fmt.Sprintf("🎉 DEAL CLOSED: Successfully negotiated switch to %s! Realized %s (Deal GMV: %.2f %s).",
				bid.Vendor.Name, savingsDesc, txn.Amount.Amount, txn.Amount.Currency)

			_, _ = slackAdapter.Execute(ctx, &mcp.Action{
				Type:      mcp.ActionTypeSendSlack,
				AdapterID: "slack",
				Parameters: map[string]any{
					"title":      "ClaimPilot Deal Closed",
					"message":    msg,
					"risk_level": "LOW",
				},
			})
		}
	}

	// 7. Audit log record
	if uc.auditRepo != nil {
		_ = uc.auditRepo.Create(ctx, &auditModel.AgentAuditEntry{
			UserID:         req.UserID,
			OrganizationID: opp.OrganizationID,
			ActionType:     auditModel.AuditActionMCPActionExecuted,
			ObligationID:   &opp.ObligationID,
			AdapterID:      "marketplace",
			ApprovalType:   auditModel.ApprovalTypeManual,
			Details: bson.M{
				"event":             "marketplace_bid_accepted",
				"transaction_id":    txn.ID.Hex(),
				"new_vendor":        bid.Vendor.Name,
				"previous_vendor":   opp.CurrentVendorName,
				"gross_deal_amount": txn.Amount.Amount,
				"commission":        commissionAmount,
				"savings":           bid.SavingsAmount,
			},
			Status:    "SUCCESS",
			CreatedAt: now,
		})
	}

	// 8. Publish domain event
	if uc.eventBus != nil {
		_ = uc.eventBus.Publish(ctx, events.TopicTenant, map[string]interface{}{
			"type":           "marketplace.deal_closed",
			"transaction_id": txn.ID.Hex(),
			"gross_amount":   txn.Amount.Amount,
			"commission":     commissionAmount,
		})
	}

	// Reload updated bids
	reloadedBids, _ := uc.mktRepo.FindBidsByOpportunityID(ctx, oppID)
	opp.Status = mktModel.OpportunityStatusTransacted

	return &dto.AcceptBidResponse{
		Transaction: dto.ToTransactionResponse(txn),
		Opportunity: dto.ToOpportunityResponse(opp, reloadedBids),
		Message: fmt.Sprintf("Deal successfully closed with %s! Realized savings recorded.",
			bid.Vendor.Name),
	}, nil
}
