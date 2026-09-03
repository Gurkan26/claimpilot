package dto

import (
	oblDto "github.com/masterfabric-go/masterfabric/internal/application/obligation/dto"
	mktModel "github.com/masterfabric-go/masterfabric/internal/domain/marketplace/model"
)

// AgentStatusDTO represents health and active models of the AI agents and MCP tools.
type AgentStatusDTO struct {
	AnalystHealthy    bool     `json:"analyst_healthy"`
	VerifierHealthy   bool     `json:"verifier_healthy"`
	AnalystModel      string   `json:"analyst_model"`
	VerifierModel     string   `json:"verifier_model"`
	ActiveMCPAdapters []string `json:"active_mcp_adapters"`
}

// RiskBreakdownDTO categorizes obligations by their verifier-assessed risk level.
type RiskBreakdownDTO struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
}

// FinancialSummaryDTO summarizes GMV, realized savings, and open opportunities.
type FinancialSummaryDTO struct {
	TotalGMV            *mktModel.Money `json:"total_gmv"`
	TotalSavings        *mktModel.Money `json:"total_savings"`
	ActiveOpportunities int             `json:"active_opportunities"`
}

// DashboardSummaryResponse represents the complete "Good Morning" dashboard payload.
type DashboardSummaryResponse struct {
	Greeting          string                    `json:"greeting"`
	AIBriefing        string                    `json:"ai_briefing"`
	TotalObligations  int                       `json:"total_obligations"`
	PendingApprovals  int                       `json:"pending_approvals"`
	OverdueCount      int                       `json:"overdue_count"`
	Upcoming7Days     int                       `json:"upcoming_7_days"`
	Upcoming30Days    int                       `json:"upcoming_30_days"`
	RiskBreakdown     RiskBreakdownDTO          `json:"risk_breakdown"`
	FinancialSummary  FinancialSummaryDTO       `json:"financial_summary"`
	AgentStatus       AgentStatusDTO            `json:"agent_status"`
	CriticalDeadlines []*oblDto.ObligationResponse `json:"critical_deadlines"`
	RecentObligations []*oblDto.ObligationResponse `json:"recent_obligations"`
}
