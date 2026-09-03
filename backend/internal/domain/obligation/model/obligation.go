package model

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ObligationType categorizes the type of obligation detected.
type ObligationType string

const (
	ObligationTypeRenewal      ObligationType = "RENEWAL"
	ObligationTypePayment      ObligationType = "PAYMENT"
	ObligationTypeReport       ObligationType = "REPORT"
	ObligationTypeWarranty     ObligationType = "WARRANTY"
	ObligationTypeCancellation ObligationType = "CANCELLATION"
	ObligationTypeCompliance   ObligationType = "COMPLIANCE"
)

// ObligationStatus tracks the lifecycle of an obligation.
type ObligationStatus string

const (
	ObligationStatusDetected        ObligationStatus = "DETECTED"
	ObligationStatusPendingApproval ObligationStatus = "PENDING_APPROVAL"
	ObligationStatusInProgress      ObligationStatus = "IN_PROGRESS"
	ObligationStatusResolved        ObligationStatus = "RESOLVED"
	ObligationStatusExpired         ObligationStatus = "EXPIRED"
	ObligationStatusDismissed       ObligationStatus = "DISMISSED"
)

// RiskLevel indicates the urgency/risk of an obligation.
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "LOW"
	RiskLevelMedium   RiskLevel = "MEDIUM"
	RiskLevelHigh     RiskLevel = "HIGH"
	RiskLevelCritical RiskLevel = "CRITICAL"
)

// Action represents a suggested or approved action for an obligation.
type Action struct {
	Type        string `bson:"type" json:"type"`               // cancel, renew, switch_vendor, notify, escalate
	Description string `bson:"description" json:"description"`
	MCPAdapter  string `bson:"mcp_adapter,omitempty" json:"mcp_adapter,omitempty"` // gmail, calendar, slack
	Parameters  bson.M `bson:"parameters,omitempty" json:"parameters,omitempty"`
}

// AgentOutput captures the output from an AI agent (Analyst or Verifier).
type AgentOutput struct {
	AgentType   string    `bson:"agent_type" json:"agent_type"` // analyst, verifier
	ModelUsed   string    `bson:"model_used" json:"model_used"`
	Output      string    `bson:"output" json:"output"`
	Confidence  float64   `bson:"confidence" json:"confidence"`
	SourceRefs  []string  `bson:"source_refs,omitempty" json:"source_refs,omitempty"` // page/clause references
	ProcessedAt time.Time `bson:"processed_at" json:"processed_at"`
}

// Obligation represents a detected obligation from a document.
type Obligation struct {
	ID                       bson.ObjectID    `bson:"_id,omitempty" json:"id"`
	UserID                   uuid.UUID        `bson:"user_id" json:"user_id"`
	OrganizationID           *uuid.UUID       `bson:"organization_id,omitempty" json:"organization_id,omitempty"`
	SourceDocumentID         bson.ObjectID    `bson:"source_document_id" json:"source_document_id"`
	Type                     ObligationType   `bson:"type" json:"type"`
	Title                    string           `bson:"title" json:"title"`
	Description              string           `bson:"description" json:"description"`
	DueDate                  time.Time        `bson:"due_date" json:"due_date"`
	ReminderDate             *time.Time       `bson:"reminder_date,omitempty" json:"reminder_date,omitempty"`
	Status                   ObligationStatus `bson:"status" json:"status"`
	RiskLevel                RiskLevel        `bson:"risk_level" json:"risk_level"`
	SuggestedAction          *Action          `bson:"suggested_action,omitempty" json:"suggested_action,omitempty"`
	ApprovedAction           *Action          `bson:"approved_action,omitempty" json:"approved_action,omitempty"`
	MarketplaceOpportunityID *bson.ObjectID   `bson:"marketplace_opportunity_id,omitempty" json:"marketplace_opportunity_id,omitempty"`
	AnalystOutput            *AgentOutput     `bson:"analyst_output,omitempty" json:"analyst_output,omitempty"`
	VerifierOutput           *AgentOutput     `bson:"verifier_output,omitempty" json:"verifier_output,omitempty"`
	AutoApproveThreshold     *float64         `bson:"auto_approve_threshold,omitempty" json:"auto_approve_threshold,omitempty"` // e.g. 500.00 EUR
	CreatedAt                time.Time        `bson:"created_at" json:"created_at"`
	UpdatedAt                time.Time        `bson:"updated_at" json:"updated_at"`
}

// IsOverdue checks if the obligation has passed its due date.
func (o *Obligation) IsOverdue() bool {
	return time.Now().After(o.DueDate) && o.Status != ObligationStatusResolved
}

// DaysUntilDue returns the number of days until the due date.
func (o *Obligation) DaysUntilDue() int {
	return int(time.Until(o.DueDate).Hours() / 24)
}

// IsActionable checks if the obligation can still be acted upon.
func (o *Obligation) IsActionable() bool {
	return o.Status == ObligationStatusDetected ||
		o.Status == ObligationStatusPendingApproval
}
