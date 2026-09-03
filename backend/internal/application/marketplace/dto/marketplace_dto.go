package dto

import (
	"time"

	"github.com/google/uuid"
	mktModel "github.com/masterfabric-go/masterfabric/internal/domain/marketplace/model"
)

// VendorDTO represents vendor information.
type VendorDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Rating      float64 `json:"rating"`
	Website     string  `json:"website"`
	Description string  `json:"description"`
	Verified    bool    `json:"verified"`
}

// BidResponse represents an individual vendor bid.
type BidResponse struct {
	ID            string          `json:"id"`
	OpportunityID string          `json:"opportunity_id"`
	Vendor        VendorDTO       `json:"vendor"`
	Price         mktModel.Money  `json:"price"`
	SavingsAmount *mktModel.Money `json:"savings_amount,omitempty"`
	SavingsRate   float64         `json:"savings_rate"` // percentage e.g. 35.0
	Terms         string          `json:"terms"`
	ContractURL   string          `json:"contract_url"`
	Status        string          `json:"status"`
	RiskScore     *float64        `json:"risk_score,omitempty"`
	RiskNotes     string          `json:"risk_notes,omitempty"`
	ValidUntil    *time.Time      `json:"valid_until,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

// OpportunityResponse represents a marketplace opportunity with associated bids.
type OpportunityResponse struct {
	ID                string          `json:"id"`
	ObligationID      string          `json:"obligation_id"`
	UserID            string          `json:"user_id"`
	OrganizationID    string          `json:"organization_id,omitempty"`
	Category          string          `json:"category"`
	CurrentVendorName string          `json:"current_vendor_name"`
	CurrentVendorCost *mktModel.Money `json:"current_vendor_cost,omitempty"`
	Status            string          `json:"status"`
	BidCount          int             `json:"bid_count"`
	BestBidID         string          `json:"best_bid_id,omitempty"`
	Bids              []*BidResponse  `json:"bids,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// TransactionResponse represents a completed deal.
type TransactionResponse struct {
	ID              string          `json:"id"`
	OpportunityID   string          `json:"opportunity_id"`
	BidID           string          `json:"bid_id"`
	UserID          string          `json:"user_id"`
	Amount          mktModel.Money  `json:"amount"`           // Gross deal value (GMV)
	Commission      mktModel.Money  `json:"commission"`       // Platform fee
	CommissionRate  float64         `json:"commission_rate"`  // 0.04 for 4%
	SavingsRealized *mktModel.Money `json:"savings_realized,omitempty"`
	Status          string          `json:"status"`
	VendorName      string          `json:"vendor_name"`
	Notes           string          `json:"notes,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
}

// MarketplaceMetricsResponse represents aggregated platform business metrics.
type MarketplaceMetricsResponse struct {
	TotalGMV            *mktModel.Money `json:"total_gmv"`
	TotalSavings        *mktModel.Money `json:"total_savings"`
	EstimatedCommission *mktModel.Money `json:"estimated_commission"`
	CompletedDeals      int             `json:"completed_deals"`
	AverageSavingsRate  float64         `json:"average_savings_rate"`
}

// AcceptBidRequest represents request to accept a bid.
type AcceptBidRequest struct {
	OpportunityID string
	BidID         string
	UserID        uuid.UUID
}

// AcceptBidResponse represents outcome of accepting a bid.
type AcceptBidResponse struct {
	Transaction *TransactionResponse `json:"transaction"`
	Opportunity *OpportunityResponse `json:"opportunity"`
	Message     string               `json:"message"`
}

// ToBidResponse transforms a Bid model to DTO.
func ToBidResponse(b *mktModel.Bid, baseCost float64) *BidResponse {
	if b == nil {
		return nil
	}

	savingsRate := 0.0
	if baseCost > 0 && b.SavingsAmount != nil {
		savingsRate = (b.SavingsAmount.Amount / baseCost) * 100.0
	}

	return &BidResponse{
		ID:            b.ID.Hex(),
		OpportunityID: b.OpportunityID.Hex(),
		Vendor: VendorDTO{
			ID:          b.Vendor.ID,
			Name:        b.Vendor.Name,
			Category:    b.Vendor.Category,
			Rating:      b.Vendor.Rating,
			Website:     b.Vendor.Website,
			Description: b.Vendor.Description,
			Verified:    b.Vendor.Verified,
		},
		Price:         b.Price,
		SavingsAmount: b.SavingsAmount,
		SavingsRate:   savingsRate,
		Terms:         b.Terms,
		ContractURL:   b.ContractURL,
		Status:        string(b.Status),
		RiskScore:     b.RiskScore,
		RiskNotes:     b.RiskNotes,
		ValidUntil:    b.ValidUntil,
		CreatedAt:     b.CreatedAt,
	}
}

// ToOpportunityResponse transforms an Opportunity model to DTO.
func ToOpportunityResponse(o *mktModel.MarketplaceOpportunity, bids []*mktModel.Bid) *OpportunityResponse {
	if o == nil {
		return nil
	}

	baseCost := 0.0
	if o.CurrentVendorCost != nil {
		baseCost = o.CurrentVendorCost.Amount
	}

	resp := &OpportunityResponse{
		ID:                o.ID.Hex(),
		ObligationID:      o.ObligationID.Hex(),
		UserID:            o.UserID.String(),
		Category:          o.Category,
		CurrentVendorName: o.CurrentVendorName,
		CurrentVendorCost: o.CurrentVendorCost,
		Status:            string(o.Status),
		BidCount:          o.BidCount,
		CreatedAt:         o.CreatedAt,
		UpdatedAt:         o.UpdatedAt,
	}

	if o.OrganizationID != nil {
		resp.OrganizationID = o.OrganizationID.String()
	}

	if o.BestBidID != nil {
		resp.BestBidID = o.BestBidID.Hex()
	}

	if len(bids) > 0 {
		resp.Bids = make([]*BidResponse, len(bids))
		for i, b := range bids {
			resp.Bids[i] = ToBidResponse(b, baseCost)
		}
		resp.BidCount = len(bids)
	}

	return resp
}

// ToTransactionResponse transforms a Transaction model to DTO.
func ToTransactionResponse(t *mktModel.Transaction) *TransactionResponse {
	if t == nil {
		return nil
	}

	return &TransactionResponse{
		ID:              t.ID.Hex(),
		OpportunityID:   t.OpportunityID.Hex(),
		BidID:           t.BidID.Hex(),
		UserID:          t.UserID.String(),
		Amount:          t.Amount,
		Commission:      t.Commission,
		CommissionRate:  t.CommissionRate,
		SavingsRealized: t.SavingsRealized,
		Status:          string(t.Status),
		VendorName:      t.VendorName,
		Notes:           t.Notes,
		CreatedAt:       t.CreatedAt,
	}
}
