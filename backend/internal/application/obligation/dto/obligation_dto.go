package dto

import (
	"time"

	"github.com/google/uuid"
	oblModel "github.com/masterfabric-go/masterfabric/internal/domain/obligation/model"
	"github.com/masterfabric-go/masterfabric/internal/mcp"
)

// ActionDTO represents an action suggestion or approved execution.
type ActionDTO struct {
	Type        string         `json:"type"`
	Description string         `json:"description"`
	MCPAdapter  string         `json:"mcp_adapter,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

// AgentOutputDTO represents analyst or verifier remarks.
type AgentOutputDTO struct {
	AgentType   string    `json:"agent_type"`
	ModelUsed   string    `json:"model_used"`
	Output      string    `json:"output"`
	Confidence  float64   `json:"confidence"`
	SourceRefs  []string  `json:"source_refs,omitempty"`
	ProcessedAt time.Time `json:"processed_at"`
}

// ObligationResponse represents obligation details returned to client.
type ObligationResponse struct {
	ID                       string          `json:"id"`
	UserID                   string          `json:"user_id"`
	OrganizationID           string          `json:"organization_id,omitempty"`
	SourceDocumentID         string          `json:"source_document_id"`
	Type                     string          `json:"type"`
	Title                    string          `json:"title"`
	Description              string          `json:"description"`
	DueDate                  time.Time       `json:"due_date"`
	DaysUntilDue             int             `json:"days_until_due"`
	IsOverdue                bool            `json:"is_overdue"`
	Status                   string          `json:"status"`
	RiskLevel                string          `json:"risk_level"`
	SuggestedAction          *ActionDTO      `json:"suggested_action,omitempty"`
	ApprovedAction           *ActionDTO      `json:"approved_action,omitempty"`
	AnalystOutput            *AgentOutputDTO `json:"analyst_output,omitempty"`
	VerifierOutput           *AgentOutputDTO `json:"verifier_output,omitempty"`
	MarketplaceOpportunityID string          `json:"marketplace_opportunity_id,omitempty"`
	CreatedAt                time.Time       `json:"created_at"`
	UpdatedAt                time.Time       `json:"updated_at"`
}

// ObligationListResponse represents paginated list of obligations.
type ObligationListResponse struct {
	Items      []*ObligationResponse `json:"items"`
	TotalCount int                   `json:"total_count"`
	Limit      int64                 `json:"limit"`
	Offset     int64                 `json:"offset"`
}

// ApproveObligationRequest represents approval parameters.
type ApproveObligationRequest struct {
	ObligationID string
	UserID       uuid.UUID
	ExecuteMCP   bool
	Parameters   map[string]any
}

// ApproveObligationResponse represents approval outcome with MCP dispatch details.
type ApproveObligationResponse struct {
	Obligation *ObligationResponse `json:"obligation"`
	MCPResult  *mcp.ActionResult   `json:"mcp_result,omitempty"`
}

// DismissObligationRequest represents dismissal input.
type DismissObligationRequest struct {
	ObligationID string
	UserID       uuid.UUID
	Reason       string
}

// ToObligationResponse transforms an Obligation domain model to DTO.
func ToObligationResponse(o *oblModel.Obligation) *ObligationResponse {
	if o == nil {
		return nil
	}

	daysUntil := int(time.Until(o.DueDate).Hours() / 24)
	isOverdue := time.Now().After(o.DueDate)

	resp := &ObligationResponse{
		ID:               o.ID.Hex(),
		UserID:           o.UserID.String(),
		SourceDocumentID: o.SourceDocumentID.Hex(),
		Type:             string(o.Type),
		Title:            o.Title,
		Description:      o.Description,
		DueDate:          o.DueDate,
		DaysUntilDue:     daysUntil,
		IsOverdue:        isOverdue,
		Status:           string(o.Status),
		RiskLevel:        string(o.RiskLevel),
		CreatedAt:        o.CreatedAt,
		UpdatedAt:        o.UpdatedAt,
	}

	if o.OrganizationID != nil {
		resp.OrganizationID = o.OrganizationID.String()
	}

	if o.MarketplaceOpportunityID != nil {
		resp.MarketplaceOpportunityID = o.MarketplaceOpportunityID.Hex()
	}

	if o.SuggestedAction != nil {
		resp.SuggestedAction = &ActionDTO{
			Type:        o.SuggestedAction.Type,
			Description: o.SuggestedAction.Description,
			MCPAdapter:  o.SuggestedAction.MCPAdapter,
		}
	}

	if o.ApprovedAction != nil {
		resp.ApprovedAction = &ActionDTO{
			Type:        o.ApprovedAction.Type,
			Description: o.ApprovedAction.Description,
			MCPAdapter:  o.ApprovedAction.MCPAdapter,
		}
	}

	if o.AnalystOutput != nil {
		resp.AnalystOutput = &AgentOutputDTO{
			AgentType:   o.AnalystOutput.AgentType,
			ModelUsed:   o.AnalystOutput.ModelUsed,
			Output:      o.AnalystOutput.Output,
			Confidence:  o.AnalystOutput.Confidence,
			SourceRefs:  o.AnalystOutput.SourceRefs,
			ProcessedAt: o.AnalystOutput.ProcessedAt,
		}
	}

	if o.VerifierOutput != nil {
		resp.VerifierOutput = &AgentOutputDTO{
			AgentType:   o.VerifierOutput.AgentType,
			ModelUsed:   o.VerifierOutput.ModelUsed,
			Output:      o.VerifierOutput.Output,
			Confidence:  o.VerifierOutput.Confidence,
			SourceRefs:  o.VerifierOutput.SourceRefs,
			ProcessedAt: o.VerifierOutput.ProcessedAt,
		}
	}

	return resp
}
