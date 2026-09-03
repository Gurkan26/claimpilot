package model

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// OpportunityStatus tracks the lifecycle of a marketplace opportunity.
type OpportunityStatus string

const (
	OpportunityStatusOpen       OpportunityStatus = "OPEN"
	OpportunityStatusBidding    OpportunityStatus = "BIDDING"
	OpportunityStatusMatched    OpportunityStatus = "MATCHED"
	OpportunityStatusTransacted OpportunityStatus = "TRANSACTED"
	OpportunityStatusExpired    OpportunityStatus = "EXPIRED"
	OpportunityStatusCancelled  OpportunityStatus = "CANCELLED"
)

// Money represents a monetary value with currency.
type Money struct {
	Amount   float64 `bson:"amount" json:"amount"`
	Currency string  `bson:"currency" json:"currency"` // USD, EUR, TRY
}

// MarketplaceOpportunity represents an opportunity to switch vendors when
// an obligation is triggered (renewal, cost increase, etc.).
type MarketplaceOpportunity struct {
	ID                bson.ObjectID     `bson:"_id,omitempty" json:"id"`
	ObligationID      bson.ObjectID     `bson:"obligation_id" json:"obligation_id"`
	UserID            uuid.UUID         `bson:"user_id" json:"user_id"`
	OrganizationID    *uuid.UUID        `bson:"organization_id,omitempty" json:"organization_id,omitempty"`
	Category          string            `bson:"category" json:"category"` // cloud-hosting, insurance, office-supplies, saas
	CurrentVendorName string            `bson:"current_vendor_name" json:"current_vendor_name"`
	CurrentVendorCost *Money            `bson:"current_vendor_cost,omitempty" json:"current_vendor_cost,omitempty"`
	Status            OpportunityStatus `bson:"status" json:"status"`
	BidCount          int               `bson:"bid_count" json:"bid_count"`
	BestBidID         *bson.ObjectID    `bson:"best_bid_id,omitempty" json:"best_bid_id,omitempty"`
	ExpiresAt         *time.Time        `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
	CreatedAt         time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time         `bson:"updated_at" json:"updated_at"`
}

// IsOpen checks if the opportunity is still accepting bids.
func (o *MarketplaceOpportunity) IsOpen() bool {
	return o.Status == OpportunityStatusOpen || o.Status == OpportunityStatusBidding
}

// HasExpired checks if the opportunity has passed its expiry.
func (o *MarketplaceOpportunity) HasExpired() bool {
	if o.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*o.ExpiresAt)
}
