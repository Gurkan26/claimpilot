package model

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// AuditActionType classifies the audit log action.
type AuditActionType string

const (
	AuditActionObligationDetected  AuditActionType = "OBLIGATION_DETECTED"
	AuditActionObligationVerified  AuditActionType = "OBLIGATION_VERIFIED"
	AuditActionObligationApproved  AuditActionType = "OBLIGATION_APPROVED"
	AuditActionObligationDismissed AuditActionType = "OBLIGATION_DISMISSED"
	AuditActionMCPActionExecuted   AuditActionType = "MCP_ACTION_EXECUTED"
)

// ApprovalType denotes whether the action was approved manually or autonomously.
type ApprovalType string

const (
	ApprovalTypeManual     ApprovalType = "MANUAL"
	ApprovalTypeAutonomous ApprovalType = "AUTONOMOUS"
)

// AgentAuditEntry represents an immutable audit trail entry for agent decisions,
// user approvals, and MCP tool execution in compliance with KVKK and enterprise governance.
type AgentAuditEntry struct {
	ID             bson.ObjectID   `bson:"_id,omitempty" json:"id"`
	UserID         uuid.UUID       `bson:"user_id" json:"user_id"`
	OrganizationID *uuid.UUID      `bson:"organization_id,omitempty" json:"organization_id,omitempty"`
	ActionType     AuditActionType `bson:"action_type" json:"action_type"`
	ObligationID   *bson.ObjectID  `bson:"obligation_id,omitempty" json:"obligation_id,omitempty"`
	DocumentID     *bson.ObjectID  `bson:"document_id,omitempty" json:"document_id,omitempty"`
	AdapterID      string          `bson:"adapter_id,omitempty" json:"adapter_id,omitempty"` // gmail, calendar, slack
	ApprovalType   ApprovalType    `bson:"approval_type,omitempty" json:"approval_type,omitempty"`
	Details        bson.M          `bson:"details,omitempty" json:"details,omitempty"`
	Status         string          `bson:"status" json:"status"` // SUCCESS, FAILED
	ErrorMessage   string          `bson:"error_message,omitempty" json:"error_message,omitempty"`
	CreatedAt      time.Time       `bson:"created_at" json:"created_at"`
}
