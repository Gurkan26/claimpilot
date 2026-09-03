package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// BidSource indicates how the bid was submitted.
type BidSource string

const (
	BidSourceAgentAutomated BidSource = "AGENT_AUTOMATED"
	BidSourceVendorManual   BidSource = "VENDOR_MANUAL"
	BidSourceAPIIntegration BidSource = "API_INTEGRATION"
)

// BidStatus tracks the state of a bid.
type BidStatus string

const (
	BidStatusPending  BidStatus = "PENDING"
	BidStatusAccepted BidStatus = "ACCEPTED"
	BidStatusRejected BidStatus = "REJECTED"
	BidStatusExpired  BidStatus = "EXPIRED"
)

// Vendor represents a vendor/supplier in the marketplace.
type Vendor struct {
	ID          string  `bson:"id" json:"id"`
	Name        string  `bson:"name" json:"name"`
	Category    string  `bson:"category" json:"category"`
	Rating      float64 `bson:"rating,omitempty" json:"rating,omitempty"`
	Website     string  `bson:"website,omitempty" json:"website,omitempty"`
	Description string  `bson:"description,omitempty" json:"description,omitempty"`
	Verified    bool    `bson:"verified" json:"verified"`
}

// Bid represents a vendor's bid/offer for a marketplace opportunity.
type Bid struct {
	ID            bson.ObjectID `bson:"_id,omitempty" json:"id"`
	OpportunityID bson.ObjectID `bson:"opportunity_id" json:"opportunity_id"`
	Vendor        Vendor        `bson:"vendor" json:"vendor"`
	Price         Money         `bson:"price" json:"price"`
	Terms         string        `bson:"terms,omitempty" json:"terms,omitempty"`
	ContractURL   string        `bson:"contract_url,omitempty" json:"contract_url,omitempty"`
	SubmittedBy   BidSource     `bson:"submitted_by" json:"submitted_by"`
	Status        BidStatus     `bson:"status" json:"status"`
	RiskScore     *float64      `bson:"risk_score,omitempty" json:"risk_score,omitempty"`       // set by Verifier Agent
	RiskNotes     string        `bson:"risk_notes,omitempty" json:"risk_notes,omitempty"`       // verifier explanation
	SavingsAmount *Money        `bson:"savings_amount,omitempty" json:"savings_amount,omitempty"` // compared to current vendor
	ValidUntil    *time.Time    `bson:"valid_until,omitempty" json:"valid_until,omitempty"`
	CreatedAt     time.Time     `bson:"created_at" json:"created_at"`
}

// IsBetterThan checks if this bid offers a lower price than another bid.
func (b *Bid) IsBetterThan(other *Bid) bool {
	if b.Price.Currency != other.Price.Currency {
		return false // cannot compare different currencies without conversion
	}
	return b.Price.Amount < other.Price.Amount
}
