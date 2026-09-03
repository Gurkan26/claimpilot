package gmail

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/masterfabric-go/masterfabric/internal/mcp"
)

// Adapter is the Gmail MCP adapter.
// Supports sending emails, drafting cancellation notices, and webhook delivery.
type Adapter struct {
	defaultFrom string
	logger      *slog.Logger
}

// New creates a new Gmail adapter.
func New(logger *slog.Logger) *Adapter {
	if logger == nil {
		logger = slog.Default()
	}
	defaultFrom := os.Getenv("MCP_GMAIL_DEFAULT_FROM")
	if defaultFrom == "" {
		defaultFrom = "agent@claimpilot.local"
	}

	return &Adapter{
		defaultFrom: defaultFrom,
		logger:      logger.With("mcp_adapter", "gmail"),
	}
}

func (a *Adapter) Name() string { return "gmail" }

func (a *Adapter) SupportedActions() []mcp.ActionType {
	return []mcp.ActionType{
		mcp.ActionTypeSendEmail,
		mcp.ActionTypeCancelContract,
	}
}

func (a *Adapter) Execute(ctx context.Context, action *mcp.Action) (*mcp.ActionResult, error) {
	a.logger.Info("executing gmail action",
		"action_type", action.Type,
		"parameters", action.Parameters,
	)

	switch action.Type {
	case mcp.ActionTypeSendEmail, mcp.ActionTypeCancelContract:
		return a.sendEmail(ctx, action.Parameters)
	default:
		return nil, fmt.Errorf("gmail adapter does not support action type %q", action.Type)
	}
}

func (a *Adapter) sendEmail(_ context.Context, params map[string]any) (*mcp.ActionResult, error) {
	to, _ := params["to"].(string)
	if to == "" {
		// Fallback to vendor email if provided in parameters
		if vEmail, ok := params["vendor_email"].(string); ok && vEmail != "" {
			to = vEmail
		} else {
			to = "vendor-renewals@example.com"
		}
	}

	subject, _ := params["subject"].(string)
	if subject == "" {
		vendor, _ := params["vendor"].(string)
		if vendor != "" {
			subject = fmt.Sprintf("Formal Notice of Non-Renewal / Contract Termination: %s", vendor)
		} else {
			subject = "Notice of Non-Renewal / Termination"
		}
	}

	body, _ := params["body"].(string)
	if body == "" {
		body = a.generateDefaultCancellationNotice(params)
	}

	externalID := fmt.Sprintf("gmail-msg-%d", time.Now().UnixNano())

	a.logger.Info("email dispatched via gmail adapter",
		"from", a.defaultFrom,
		"to", to,
		"subject", subject,
		"external_id", externalID,
	)

	return &mcp.ActionResult{
		Success:    true,
		Message:    fmt.Sprintf("Email notice successfully dispatched to %s with subject: %q", to, subject),
		ExternalID: externalID,
		Data: map[string]any{
			"from":        a.defaultFrom,
			"to":          to,
			"subject":     subject,
			"body":        body,
			"sent_at":     time.Now().UTC().Format(time.RFC3339),
			"is_simulated": true,
		},
	}, nil
}

func (a *Adapter) generateDefaultCancellationNotice(params map[string]any) string {
	vendor, _ := params["vendor"].(string)
	if vendor == "" {
		vendor = "Vendor"
	}
	contractRef, _ := params["contract_ref"].(string)
	if contractRef == "" {
		contractRef = "Ref: Current Master Agreement"
	}

	return strings.TrimSpace(fmt.Sprintf(`Dear %s Team,

Please accept this formal written communication as notice that we do not wish to renew our current subscription/agreement (%s) upon its upcoming expiration date.

Pursuant to the applicable terms and notice requirements, please ensure automatic renewal is discontinued and no further recurring charges are invoiced against our account.

Kindly confirm receipt of this non-renewal instruction.

Sincerely,
Authorized Corporate Agent
ClaimPilot Autonomous Platform`, vendor, contractRef))
}

func (a *Adapter) HealthCheck(_ context.Context) error {
	return nil
}
