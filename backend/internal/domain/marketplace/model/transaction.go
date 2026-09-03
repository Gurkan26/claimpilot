package model

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// TransactionStatus tracks the state of a marketplace transaction.
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "PENDING"
	TransactionStatusCompleted TransactionStatus = "COMPLETED"
	TransactionStatusFailed    TransactionStatus = "FAILED"
	TransactionStatusRefunded  TransactionStatus = "REFUNDED"
)

// Transaction represents a completed marketplace deal — a vendor switch or new contract.
// Each transaction contributes to platform GMV and generates commission.
type Transaction struct {
	ID                bson.ObjectID     `bson:"_id,omitempty" json:"id"`
	OpportunityID     bson.ObjectID     `bson:"opportunity_id" json:"opportunity_id"`
	BidID             bson.ObjectID     `bson:"bid_id" json:"bid_id"`
	UserID            uuid.UUID         `bson:"user_id" json:"user_id"`
	OrganizationID    *uuid.UUID        `bson:"organization_id,omitempty" json:"organization_id,omitempty"`
	Amount            Money             `bson:"amount" json:"amount"`               // deal amount
	Commission        Money             `bson:"commission" json:"commission"`         // platform commission
	CommissionRate    float64           `bson:"commission_rate" json:"commission_rate"` // e.g. 0.05 for 5%
	SavingsRealized   *Money            `bson:"savings_realized,omitempty" json:"savings_realized,omitempty"`
	Status            TransactionStatus `bson:"status" json:"status"`
	VendorName        string            `bson:"vendor_name" json:"vendor_name"`
	ContractDocumentID *bson.ObjectID   `bson:"contract_document_id,omitempty" json:"contract_document_id,omitempty"`
	Notes             string            `bson:"notes,omitempty" json:"notes,omitempty"`
	CompletedAt       *time.Time        `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
	CreatedAt         time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time         `bson:"updated_at" json:"updated_at"`
}

// IsCompleted checks if the transaction has been finalized.
func (t *Transaction) IsCompleted() bool {
	return t.Status == TransactionStatusCompleted
}

// GMVContribution returns the deal amount as the GMV contribution for reporting.
func (t *Transaction) GMVContribution() Money {
	return t.Amount
}
