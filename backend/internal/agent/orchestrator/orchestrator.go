package orchestrator

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/masterfabric-go/masterfabric/internal/agent/analyst"
	"github.com/masterfabric-go/masterfabric/internal/agent/verifier"
	docModel "github.com/masterfabric-go/masterfabric/internal/domain/document/model"
	docRepo "github.com/masterfabric-go/masterfabric/internal/domain/document/repository"
	mktModel "github.com/masterfabric-go/masterfabric/internal/domain/marketplace/model"
	mktRepo "github.com/masterfabric-go/masterfabric/internal/domain/marketplace/repository"
	oblModel "github.com/masterfabric-go/masterfabric/internal/domain/obligation/model"
	oblRepo "github.com/masterfabric-go/masterfabric/internal/domain/obligation/repository"
	"github.com/masterfabric-go/masterfabric/internal/pii"
	"github.com/masterfabric-go/masterfabric/internal/shared/events"
)

// Orchestrator coordinates the full agent pipeline:
// Event → Understand (extraction) → Risk/Opportunity → Decision → Verification → Approval → Action
type Orchestrator struct {
	analyst    *analyst.Analyst
	verifier   *verifier.Verifier
	redactor   pii.Redactor
	docRepo    docRepo.DocumentRepository
	oblRepo    oblRepo.ObligationRepository
	mktRepo    mktRepo.MarketplaceRepository
	eventBus   events.EventBus
	logger     *slog.Logger
}

// Config holds dependencies for creating an Orchestrator.
type Config struct {
	Analyst    *analyst.Analyst
	Verifier   *verifier.Verifier
	Redactor   pii.Redactor
	DocRepo    docRepo.DocumentRepository
	OblRepo    oblRepo.ObligationRepository
	MktRepo    mktRepo.MarketplaceRepository
	EventBus   events.EventBus
	Logger     *slog.Logger
}

// New creates an Orchestrator with all dependencies injected.
func New(cfg Config) *Orchestrator {
	return &Orchestrator{
		analyst:  cfg.Analyst,
		verifier: cfg.Verifier,
		redactor: cfg.Redactor,
		docRepo:  cfg.DocRepo,
		oblRepo:  cfg.OblRepo,
		mktRepo:  cfg.MktRepo,
		eventBus: cfg.EventBus,
		logger:   cfg.Logger.With("component", "orchestrator"),
	}
}

// ProcessDocument runs the full analysis pipeline on a document:
// 1. PII redaction
// 2. Analyst extraction
// 3. Obligation creation
// 4. Verifier validation
// 5. Marketplace opportunity creation (if applicable)
func (o *Orchestrator) ProcessDocument(ctx context.Context, docID bson.ObjectID) error {
	o.logger.Info("starting document processing pipeline", "document_id", docID.Hex())

	// 1. Fetch document
	doc, err := o.docRepo.FindByID(ctx, docID)
	if err != nil {
		return fmt.Errorf("fetch document %s: %w", docID.Hex(), err)
	}

	// Mark as processing
	if err := o.docRepo.UpdateStatus(ctx, docID, docModel.DocumentStatusProcessing); err != nil {
		return fmt.Errorf("update document status: %w", err)
	}

	// 2. PII Redaction — mask sensitive data before sending to LLM (KVKK/GDPR)
	rawContent := doc.RawContent
	if rawContent == "" {
		rawContent = doc.RedactedContent
	}

	redactedContent, _, err := o.redactor.Redact(ctx, rawContent)
	if err != nil {
		o.logger.Error("PII redaction failed", "error", err)
		_ = o.docRepo.UpdateStatus(ctx, docID, docModel.DocumentStatusFailed)
		return fmt.Errorf("pii redaction: %w", err)
	}

	// Persist redacted content
	_ = o.docRepo.UpdateContent(ctx, docID, doc.RawContent, redactedContent)

	// 3. Analyst extraction
	analysisResult, err := o.analyst.AnalyzeDocument(ctx, redactedContent, doc.FileType)
	if err != nil {
		o.logger.Error("document analysis failed", "error", err)
		_ = o.docRepo.UpdateStatus(ctx, docID, docModel.DocumentStatusFailed)
		return fmt.Errorf("analyst analysis: %w", err)
	}

	// Store extraction result
	if analysisResult.Extraction != nil {
		if err := o.docRepo.UpdateExtractionResult(ctx, docID, analysisResult.Extraction); err != nil {
			o.logger.Error("failed to store extraction result", "error", err)
		}
	}

	// 4. Create and verify obligations
	for _, candidate := range analysisResult.Obligations {
		obl, err := o.createObligation(ctx, doc, &candidate)
		if err != nil {
			o.logger.Error("failed to create obligation", "error", err, "title", candidate.Title)
			continue
		}

		// Verify with Verifier agent
		verification, err := o.verifier.VerifyObligation(ctx, &candidate, redactedContent)
		if err != nil {
			o.logger.Error("obligation verification failed", "error", err, "obligation_id", obl.ID.Hex())
			continue
		}

		// Build verifier agent output
		agentOutput := &oblModel.AgentOutput{
			AgentType:   "verifier",
			ModelUsed:   o.verifier.ProviderName(),
			Output:      verification.Notes,
			Confidence:  verification.Confidence,
			SourceRefs:  verification.SourceRefs,
			ProcessedAt: time.Now().UTC(),
		}

		status := oblModel.ObligationStatusPendingApproval
		if verification.AutoApprove {
			status = oblModel.ObligationStatusInProgress
		}

		if err := o.oblRepo.UpdateVerification(ctx, obl.ID, status, verification.RiskLevel, agentOutput); err != nil {
			o.logger.Error("failed to update obligation verification", "error", err)
		}

		if verification.Verified {
			// Publish event
			_ = o.eventBus.Publish(ctx, events.TopicTenant, map[string]interface{}{
				"type":          "obligation.verified",
				"obligation_id": obl.ID.Hex(),
				"risk_level":    verification.RiskLevel,
				"auto_approve":  verification.AutoApprove,
			})
		}
	}

	// 5. Create marketplace opportunities
	for _, oppCandidate := range analysisResult.Opportunities {
		if err := o.createMarketplaceOpportunity(ctx, doc, &oppCandidate); err != nil {
			o.logger.Error("failed to create marketplace opportunity", "error", err)
		}
	}

	// Mark document as analyzed
	if err := o.docRepo.UpdateStatus(ctx, docID, docModel.DocumentStatusAnalyzed); err != nil {
		return fmt.Errorf("update document status to analyzed: %w", err)
	}

	o.logger.Info("document processing pipeline completed",
		"document_id", docID.Hex(),
		"obligations_found", len(analysisResult.Obligations),
		"opportunities_found", len(analysisResult.Opportunities),
	)

	return nil
}

func (o *Orchestrator) createObligation(ctx context.Context, doc *docModel.Document, candidate *analyst.ObligationCandidate) (*oblModel.Obligation, error) {
	dueDate, err := time.Parse("2006-01-02", candidate.DueDate)
	if err != nil {
		dueDate = time.Now().Add(30 * 24 * time.Hour) // default to 30 days
	}

	obl := &oblModel.Obligation{
		UserID:           doc.UserID,
		OrganizationID:   doc.OrganizationID,
		SourceDocumentID: doc.ID,
		Type:             candidate.Type,
		Title:            candidate.Title,
		Description:      candidate.Description,
		DueDate:          dueDate,
		Status:           oblModel.ObligationStatusDetected,
		RiskLevel:        candidate.RiskLevel,
		AnalystOutput: &oblModel.AgentOutput{
			AgentType:   "analyst",
			Output:      candidate.Description,
			Confidence:  candidate.Confidence,
			SourceRefs:  []string{candidate.SourceRef},
			ProcessedAt: time.Now().UTC(),
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := o.oblRepo.Create(ctx, obl); err != nil {
		return nil, fmt.Errorf("create obligation: %w", err)
	}

	return obl, nil
}

func (o *Orchestrator) createMarketplaceOpportunity(ctx context.Context, doc *docModel.Document, candidate *analyst.OpportunityCandidate) error {
	// Find a matching obligation for this opportunity (first unmatched one)
	obligations, err := o.oblRepo.FindByDocumentID(ctx, doc.ID)
	if err != nil || len(obligations) == 0 {
		return fmt.Errorf("no obligations found for document %s", doc.ID.Hex())
	}

	// Use the first obligation that doesn't already have a marketplace opportunity
	var targetObl *oblModel.Obligation
	for _, obl := range obligations {
		if obl.MarketplaceOpportunityID == nil {
			targetObl = obl
			break
		}
	}
	if targetObl == nil {
		return nil // all obligations already have opportunities
	}

	var orgID *uuid.UUID
	if doc.OrganizationID != nil {
		orgID = doc.OrganizationID
	}

	opp := &mktModel.MarketplaceOpportunity{
		ObligationID:      targetObl.ID,
		UserID:            doc.UserID,
		OrganizationID:    orgID,
		Category:          candidate.Category,
		CurrentVendorName: candidate.CurrentVendorName,
		CurrentVendorCost: &mktModel.Money{
			Amount:   candidate.EstimatedCost,
			Currency: candidate.Currency,
		},
		Status:    mktModel.OpportunityStatusOpen,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := o.mktRepo.CreateOpportunity(ctx, opp); err != nil {
		return fmt.Errorf("create marketplace opportunity: %w", err)
	}

	// Link opportunity to obligation
	if err := o.oblRepo.LinkMarketplaceOpportunity(ctx, targetObl.ID, opp.ID); err != nil {
		o.logger.Error("failed to link opportunity to obligation", "error", err)
	}

	return nil
}
