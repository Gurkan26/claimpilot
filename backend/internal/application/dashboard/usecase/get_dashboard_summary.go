package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/agent/llmclient"
	"github.com/masterfabric-go/masterfabric/internal/application/dashboard/dto"
	oblDto "github.com/masterfabric-go/masterfabric/internal/application/obligation/dto"
	mktRepo "github.com/masterfabric-go/masterfabric/internal/domain/marketplace/repository"
	oblModel "github.com/masterfabric-go/masterfabric/internal/domain/obligation/model"
	oblRepo "github.com/masterfabric-go/masterfabric/internal/domain/obligation/repository"
	"github.com/masterfabric-go/masterfabric/internal/mcp"
)

// GetDashboardSummaryUseCase compiles the "Good Morning" executive dashboard.
type GetDashboardSummaryUseCase struct {
	oblRepo  oblRepo.ObligationRepository
	mktRepo  mktRepo.MarketplaceRepository
	mcpReg   *mcp.Registry
	analyst  llmclient.Provider
	verifier llmclient.Provider
	logger   *slog.Logger
}

// Config holds dependencies for GetDashboardSummaryUseCase.
type Config struct {
	OblRepo  oblRepo.ObligationRepository
	MktRepo  mktRepo.MarketplaceRepository
	MCPReg   *mcp.Registry
	Analyst  llmclient.Provider
	Verifier llmclient.Provider
	Logger   *slog.Logger
}

// NewGetDashboardSummaryUseCase creates a new GetDashboardSummaryUseCase.
func NewGetDashboardSummaryUseCase(cfg Config) *GetDashboardSummaryUseCase {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &GetDashboardSummaryUseCase{
		oblRepo:  cfg.OblRepo,
		mktRepo:  cfg.MktRepo,
		mcpReg:   cfg.MCPReg,
		analyst:  cfg.Analyst,
		verifier: cfg.Verifier,
		logger:   logger.With("usecase", "dashboard_summary"),
	}
}

// Execute compiles the dashboard summary for the given user.
func (uc *GetDashboardSummaryUseCase) Execute(ctx context.Context, userID uuid.UUID) (*dto.DashboardSummaryResponse, error) {
	// 1. Fetch all active obligations
	allObls, err := uc.oblRepo.FindByUserID(ctx, userID, nil, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("fetch obligations for dashboard: %w", err)
	}

	totalCount := len(allObls)
	pendingApprovals := 0
	overdueCount := 0
	upcoming7Days := 0
	upcoming30Days := 0

	var riskBreakdown dto.RiskBreakdownDTO
	var criticalDeadlines []*oblDto.ObligationResponse
	var recentList []*oblDto.ObligationResponse

	now := time.Now().UTC()
	sevenDaysAhead := now.Add(7 * 24 * time.Hour)
	thirtyDaysAhead := now.Add(30 * 24 * time.Hour)

	for _, o := range allObls {
		dtoItem := oblDto.ToObligationResponse(o)
		recentList = append(recentList, dtoItem)

		// Status breakdown
		if o.Status == oblModel.ObligationStatusPendingApproval {
			pendingApprovals++
		}

		// Overdue check
		if now.After(o.DueDate) && o.Status != oblModel.ObligationStatusResolved && o.Status != oblModel.ObligationStatusDismissed {
			overdueCount++
		}

		// Deadlines buckets
		if o.DueDate.After(now) && o.DueDate.Before(sevenDaysAhead) {
			upcoming7Days++
			if o.RiskLevel == oblModel.RiskLevelHigh || o.RiskLevel == oblModel.RiskLevelCritical {
				criticalDeadlines = append(criticalDeadlines, dtoItem)
			}
		} else if o.DueDate.After(sevenDaysAhead) && o.DueDate.Before(thirtyDaysAhead) {
			upcoming30Days++
		}

		// Risk breakdown
		switch o.RiskLevel {
		case oblModel.RiskLevelCritical:
			riskBreakdown.Critical++
		case oblModel.RiskLevelHigh:
			riskBreakdown.High++
		case oblModel.RiskLevelMedium:
			riskBreakdown.Medium++
		case oblModel.RiskLevelLow:
			riskBreakdown.Low++
		}
	}

	// 2. Financial Metrics (GMV & Savings)
	totalGMV, _ := uc.mktRepo.TotalGMV(ctx, userID)
	totalSavings, _ := uc.mktRepo.TotalSavings(ctx, userID)

	// Fetch open marketplace opportunities
	opps, _ := uc.mktRepo.FindOpportunitiesByUserID(ctx, userID, nil, 100, 0)
	activeOppsCount := len(opps)

	// 3. Agent & MCP Health
	analystHealthy := false
	analystModel := "Not Configured"
	if uc.analyst != nil {
		analystHealthy = uc.analyst.HealthCheck(ctx) == nil
		analystModel = uc.analyst.Model()
	}

	verifierHealthy := false
	verifierModel := "Not Configured"
	if uc.verifier != nil {
		verifierHealthy = uc.verifier.HealthCheck(ctx) == nil
		verifierModel = uc.verifier.Model()
	}

	var mcpList []string
	if uc.mcpReg != nil {
		mcpList = uc.mcpReg.List()
	}

	// 4. Synthesize AI Executive Briefing
	greeting := "Good day! Welcome to ClaimPilot Obligation Engine."
	hour := time.Now().Hour()
	if hour < 12 {
		greeting = "Good morning! Welcome to ClaimPilot."
	} else if hour < 18 {
		greeting = "Good afternoon! Welcome to ClaimPilot."
	}

	briefing := uc.generateBriefing(totalCount, pendingApprovals, upcoming7Days, overdueCount, riskBreakdown)

	return &dto.DashboardSummaryResponse{
		Greeting:          greeting,
		AIBriefing:        briefing,
		TotalObligations:  totalCount,
		PendingApprovals:  pendingApprovals,
		OverdueCount:      overdueCount,
		Upcoming7Days:     upcoming7Days,
		Upcoming30Days:    upcoming30Days,
		RiskBreakdown:     riskBreakdown,
		FinancialSummary: dto.FinancialSummaryDTO{
			TotalGMV:            totalGMV,
			TotalSavings:        totalSavings,
			ActiveOpportunities: activeOppsCount,
		},
		AgentStatus: dto.AgentStatusDTO{
			AnalystHealthy:    analystHealthy,
			VerifierHealthy:   verifierHealthy,
			AnalystModel:      analystModel,
			VerifierModel:     verifierModel,
			ActiveMCPAdapters: mcpList,
		},
		CriticalDeadlines: criticalDeadlines,
		RecentObligations: recentList,
	}, nil
}

func (uc *GetDashboardSummaryUseCase) generateBriefing(total, pending, upcoming7, overdue int, risk dto.RiskBreakdownDTO) string {
	if total == 0 {
		return "All obligations are clear. No pending actions required today. Upload a document to start autonomous tracking."
	}

	brief := fmt.Sprintf("You currently have %d tracked obligations.", total)
	if overdue > 0 {
		brief += fmt.Sprintf(" ⚠️ Attention required: %d obligation(s) are currently overdue!", overdue)
	}
	if upcoming7 > 0 {
		brief += fmt.Sprintf(" %d critical deadline(s) are approaching within the next 7 days.", upcoming7)
	}
	if pending > 0 {
		brief += fmt.Sprintf(" %d pending agent action(s) are awaiting your approval for autonomous dispatch via Gmail/Calendar/Slack.", pending)
	}
	if risk.Critical > 0 || risk.High > 0 {
		brief += fmt.Sprintf(" High/Critical risk items identified: %d. Recommend immediate review.", risk.Critical+risk.High)
	}

	return brief
}
