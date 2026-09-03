package calendar

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/masterfabric-go/masterfabric/internal/mcp"
)

// Adapter is the Google Calendar MCP adapter.
// Manages deadline scheduling, renewal calendar invites, and iCal output.
type Adapter struct {
	logger *slog.Logger
}

// New creates a new Calendar adapter.
func New(logger *slog.Logger) *Adapter {
	if logger == nil {
		logger = slog.Default()
	}
	return &Adapter{
		logger: logger.With("mcp_adapter", "calendar"),
	}
}

func (a *Adapter) Name() string { return "calendar" }

func (a *Adapter) SupportedActions() []mcp.ActionType {
	return []mcp.ActionType{
		mcp.ActionTypeCreateEvent,
	}
}

func (a *Adapter) Execute(ctx context.Context, action *mcp.Action) (*mcp.ActionResult, error) {
	a.logger.Info("executing calendar action",
		"action_type", action.Type,
		"parameters", action.Parameters,
	)

	switch action.Type {
	case mcp.ActionTypeCreateEvent:
		return a.createEvent(ctx, action.Parameters)
	default:
		return nil, fmt.Errorf("calendar adapter does not support action type %q", action.Type)
	}
}

func (a *Adapter) createEvent(_ context.Context, params map[string]any) (*mcp.ActionResult, error) {
	title, _ := params["title"].(string)
	if title == "" {
		title = "ClaimPilot Obligation Deadline"
	}

	dateStr, _ := params["due_date"].(string)
	if dateStr == "" {
		dateStr, _ = params["date"].(string)
	}

	var startTime, endTime time.Time
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			startTime = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 9, 0, 0, 0, time.UTC)
			endTime = startTime.Add(1 * time.Hour)
		} else if parsedWithTime, err := time.Parse(time.RFC3339, dateStr); err == nil {
			startTime = parsedWithTime
			endTime = startTime.Add(1 * time.Hour)
		}
	}
	if startTime.IsZero() {
		startTime = time.Now().Add(7 * 24 * time.Hour)
		endTime = startTime.Add(1 * time.Hour)
	}

	description, _ := params["description"].(string)
	if description == "" {
		description = "Scheduled obligation deadline tracked autonomously by ClaimPilot."
	}

	externalID := fmt.Sprintf("cal-evt-%d", time.Now().UnixNano())
	icalData := generateICalString(externalID, title, description, startTime, endTime)

	a.logger.Info("calendar event scheduled",
		"title", title,
		"start_time", startTime.Format(time.RFC3339),
		"external_id", externalID,
	)

	return &mcp.ActionResult{
		Success:    true,
		Message:    fmt.Sprintf("Calendar deadline event scheduled: %q on %s", title, startTime.Format("2006-01-02")),
		ExternalID: externalID,
		Data: map[string]any{
			"title":        title,
			"start_time":   startTime.Format(time.RFC3339),
			"end_time":     endTime.Format(time.RFC3339),
			"description":  description,
			"ical_payload": icalData,
			"is_simulated": true,
		},
	}, nil
}

func generateICalString(uid, title, description string, start, end time.Time) string {
	return strings.TrimSpace(fmt.Sprintf(`BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//ClaimPilot//Autonomous AI Obligations//EN
CALSCALE:GREGORIAN
METHOD:REQUEST
BEGIN:VEVENT
UID:%s
DTSTAMP:%s
DTSTART:%s
DTEND:%s
SUMMARY:%s
DESCRIPTION:%s
STATUS:CONFIRMED
BEGIN:VALARM
TRIGGER:-P1D
ACTION:DISPLAY
DESCRIPTION:ClaimPilot 24-Hour Obligation Reminder
END:VALARM
END:VEVENT
END:VCALENDAR`,
		uid,
		time.Now().UTC().Format("20060102T150405Z"),
		start.Format("20060102T150405Z"),
		end.Format("20060102T150405Z"),
		title,
		description,
	))
}

func (a *Adapter) HealthCheck(_ context.Context) error {
	return nil
}
