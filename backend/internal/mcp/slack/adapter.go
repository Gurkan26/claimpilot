package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/masterfabric-go/masterfabric/internal/mcp"
)

// Adapter is the Slack MCP adapter.
// Dispatches urgent obligation reminders and approval notifications to Slack channels.
type Adapter struct {
	webhookURL     string
	defaultChannel string
	client         *http.Client
	logger         *slog.Logger
}

// New creates a new Slack adapter.
func New(logger *slog.Logger) *Adapter {
	if logger == nil {
		logger = slog.Default()
	}
	webhookURL := os.Getenv("MCP_SLACK_WEBHOOK_URL")
	defaultChannel := os.Getenv("MCP_SLACK_DEFAULT_CHANNEL")
	if defaultChannel == "" {
		defaultChannel = "#claimpilot-alerts"
	}

	return &Adapter{
		webhookURL:     webhookURL,
		defaultChannel: defaultChannel,
		client:         &http.Client{Timeout: 10 * time.Second},
		logger:         logger.With("mcp_adapter", "slack"),
	}
}

func (a *Adapter) Name() string { return "slack" }

func (a *Adapter) SupportedActions() []mcp.ActionType {
	return []mcp.ActionType{
		mcp.ActionTypeSendSlack,
		mcp.ActionTypeNotify,
	}
}

func (a *Adapter) Execute(ctx context.Context, action *mcp.Action) (*mcp.ActionResult, error) {
	a.logger.Info("executing slack action",
		"action_type", action.Type,
		"parameters", action.Parameters,
	)

	switch action.Type {
	case mcp.ActionTypeSendSlack, mcp.ActionTypeNotify:
		return a.sendMessage(ctx, action.Parameters)
	default:
		return nil, fmt.Errorf("slack adapter does not support action type %q", action.Type)
	}
}

// SlackAttachment defines attachment formatting for Slack messages.
type SlackAttachment struct {
	Color  string `json:"color"`
	Title  string `json:"title"`
	Text   string `json:"text"`
	Footer string `json:"footer"`
	Ts     int64  `json:"ts"`
}

// SlackPayload represents outgoing message body.
type SlackPayload struct {
	Channel     string            `json:"channel,omitempty"`
	Text        string            `json:"text"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

func (a *Adapter) sendMessage(ctx context.Context, params map[string]any) (*mcp.ActionResult, error) {
	channel, _ := params["channel"].(string)
	if channel == "" {
		channel = a.defaultChannel
	}

	message, _ := params["message"].(string)
	if message == "" {
		message = "ClaimPilot Autonomous Obligation Notification"
	}

	title, _ := params["title"].(string)
	if title == "" {
		title = "Obligation Alert"
	}

	riskLevel, _ := params["risk_level"].(string)
	color := "#3AA3E3" // Blue (default/low/medium)
	switch riskLevel {
	case "CRITICAL":
		color = "#D00000" // Red
	case "HIGH":
		color = "#FFA500" // Orange/Amber
	}

	payload := SlackPayload{
		Channel: channel,
		Text:    fmt.Sprintf("*[ClaimPilot]* %s", title),
		Attachments: []SlackAttachment{
			{
				Color:  color,
				Title:  title,
				Text:   message,
				Footer: "ClaimPilot Autonomous Agent Engine",
				Ts:     time.Now().Unix(),
			},
		},
	}

	externalID := fmt.Sprintf("slack-msg-%d", time.Now().UnixNano())

	// If webhook is provided, dispatch actual HTTP call
	if a.webhookURL != "" {
		body, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal slack payload: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.webhookURL, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("create slack request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := a.client.Do(req)
		if err != nil {
			a.logger.Warn("slack webhook delivery failed, logging fallback", "error", err)
		} else {
			_ = resp.Body.Close()
		}
	}

	a.logger.Info("slack notification dispatched",
		"channel", channel,
		"title", title,
		"external_id", externalID,
	)

	return &mcp.ActionResult{
		Success:    true,
		Message:    fmt.Sprintf("Slack alert dispatched to %s: %q", channel, title),
		ExternalID: externalID,
		Data: map[string]any{
			"channel":      channel,
			"title":        title,
			"message":      message,
			"risk_level":   riskLevel,
			"color":        color,
			"is_simulated": a.webhookURL == "",
		},
	}, nil
}

func (a *Adapter) HealthCheck(_ context.Context) error {
	return nil
}
