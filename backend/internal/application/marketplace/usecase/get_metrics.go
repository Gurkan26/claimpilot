package usecase

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/application/marketplace/dto"
	mktModel "github.com/masterfabric-go/masterfabric/internal/domain/marketplace/model"
	mktRepo "github.com/masterfabric-go/masterfabric/internal/domain/marketplace/repository"
)

// GetMetricsUseCase computes aggregated marketplace business metrics.
type GetMetricsUseCase struct {
	mktRepo mktRepo.MarketplaceRepository
}

// NewGetMetricsUseCase creates a new GetMetricsUseCase.
func NewGetMetricsUseCase(mktRepo mktRepo.MarketplaceRepository) *GetMetricsUseCase {
	return &GetMetricsUseCase{mktRepo: mktRepo}
}

// Execute aggregates GMV, savings, and platform revenue.
func (uc *GetMetricsUseCase) Execute(ctx context.Context, userID uuid.UUID) (*dto.MarketplaceMetricsResponse, error) {
	gmv, err := uc.mktRepo.TotalGMV(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("calculate total GMV: %w", err)
	}

	savings, err := uc.mktRepo.TotalSavings(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("calculate total savings: %w", err)
	}

	txns, err := uc.mktRepo.FindTransactionsByUserID(ctx, userID, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("fetch user transactions: %w", err)
	}

	currency := "USD"
	if gmv != nil && gmv.Currency != "" {
		currency = gmv.Currency
	}

	totalCommission := 0.0
	completedDeals := 0
	totalBaseline := 0.0
	totalSaved := 0.0

	for _, t := range txns {
		if t.Status == mktModel.TransactionStatusCompleted {
			completedDeals++
			totalCommission += t.Commission.Amount
			if t.SavingsRealized != nil {
				totalSaved += t.SavingsRealized.Amount
				totalBaseline += t.Amount.Amount + t.SavingsRealized.Amount
			}
		}
	}

	avgSavingsRate := 0.0
	if totalBaseline > 0 {
		avgSavingsRate = math.Round((totalSaved/totalBaseline)*1000) / 10
	}

	return &dto.MarketplaceMetricsResponse{
		TotalGMV:            gmv,
		TotalSavings:        savings,
		EstimatedCommission: &mktModel.Money{Amount: math.Round(totalCommission*100) / 100, Currency: currency},
		CompletedDeals:      completedDeals,
		AverageSavingsRate:  avgSavingsRate,
	}, nil
}
