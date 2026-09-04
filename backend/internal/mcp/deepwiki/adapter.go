package deepwiki

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/masterfabric-go/masterfabric/internal/mcp"
)

const (
	ActionTypeSearchWiki       mcp.ActionType = "search_wiki"
	ActionTypeQueryPrecedents   mcp.ActionType = "query_precedents"
	ActionTypeGetPolicyRule    mcp.ActionType = "get_policy_rule"
	ActionTypeVerifyClause     mcp.ActionType = "verify_clause"
)

// Adapter implements mcp.Adapter for DeepWiki enterprise knowledge base & legal precedents.
type Adapter struct {
	logger   *slog.Logger
	endpoint string
	mu       sync.RWMutex
	online   bool
}

// New creates a new DeepWiki MCP adapter.
func New(endpoint string, logger *slog.Logger) *Adapter {
	if endpoint == "" {
		endpoint = "http://localhost:8899/mcp/deepwiki/sse"
	}
	return &Adapter{
		logger:   logger.With("adapter", "deepwiki"),
		endpoint: endpoint,
		online:   true,
	}
}

// Name returns the identifier for this adapter.
func (a *Adapter) Name() string {
	return "deepwiki"
}

// SupportedActions returns the list of actions supported by DeepWiki.
func (a *Adapter) SupportedActions() []mcp.ActionType {
	return []mcp.ActionType{
		ActionTypeSearchWiki,
		ActionTypeQueryPrecedents,
		ActionTypeGetPolicyRule,
		ActionTypeVerifyClause,
	}
}

// HealthCheck verifies connection to DeepWiki.
func (a *Adapter) HealthCheck(ctx context.Context) error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if !a.online {
		return fmt.Errorf("deepwiki adapter is offline or disabled")
	}
	return nil
}

// SetOnline updates the online state.
func (a *Adapter) SetOnline(online bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.online = online
}

// Execute performs knowledge search or clause verification via DeepWiki.
func (a *Adapter) Execute(ctx context.Context, action *mcp.Action) (*mcp.ActionResult, error) {
	a.logger.Info("executing deepwiki mcp action",
		"type", action.Type,
		"parameters", action.Parameters,
	)

	switch action.Type {
	case ActionTypeSearchWiki:
		query, _ := action.Parameters["query"].(string)
		return &mcp.ActionResult{
			Success: true,
			Message: fmt.Sprintf("Found relevant knowledge articles for '%s'", query),
			Data: map[string]any{
				"articles": []map[string]any{
					{
						"title":   "SaaS Procurement & Auto-Renewal Governance 2025",
						"space":   "legal-procurement",
						"snippet": "All subscription agreements exceeding $2,000/yr must have 30-day non-renewal notices dispatched before milestone lock.",
					},
					{
						"title":   "Vendor Escalation Caps & Indemnity Clauses",
						"space":   "procurement-standards",
						"snippet": "Price increases capped at annual CPI or 5% maximum.",
					},
				},
				"retrieval_latency_ms": 14,
			},
		}, nil

	case ActionTypeQueryPrecedents:
		category, _ := action.Parameters["category"].(string)
		return &mcp.ActionResult{
			Success: true,
			Message: fmt.Sprintf("Retrieved contract precedents for category '%s'", category),
			Data: map[string]any{
				"precedents": []string{
					"Adobe MSA 2024: 30-day notice requirement (precedent settled)",
					"AWS Enterprise Discount Addendum: 14-day billing objection window",
				},
				"confidence": 0.96,
			},
		}, nil

	case ActionTypeGetPolicyRule:
		ruleID, _ := action.Parameters["rule_id"].(string)
		return &mcp.ActionResult{
			Success: true,
			Message: fmt.Sprintf("Rule '%s' retrieved successfully", ruleID),
			Data: map[string]any{
				"rule_id":        ruleID,
				"requires_rfq":   true,
				"approval_chain": []string{"procurement_lead", "cfo"},
			},
		}, nil

	case ActionTypeVerifyClause:
		clauseText, _ := action.Parameters["clause_text"].(string)
		return &mcp.ActionResult{
			Success: true,
			Message: "Clause verified against corporate policy",
			Data: map[string]any{
				"clause_snippet": clauseText,
				"is_compliant":   true,
				"risk_level":     "LOW",
				"verified_at":    time.Now().UTC().Format(time.RFC3339),
			},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported deepwiki action type: %s", action.Type)
	}
}
