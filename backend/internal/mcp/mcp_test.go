package mcp_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/masterfabric-go/masterfabric/internal/mcp"
	"github.com/masterfabric-go/masterfabric/internal/mcp/calendar"
	"github.com/masterfabric-go/masterfabric/internal/mcp/gmail"
	"github.com/masterfabric-go/masterfabric/internal/mcp/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGmailAdapter_Execute(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	adapter := gmail.New(logger)

	assert.Equal(t, "gmail", adapter.Name())
	assert.Contains(t, adapter.SupportedActions(), mcp.ActionTypeSendEmail)
	require.NoError(t, adapter.HealthCheck(context.Background()))

	ctx := context.Background()
	action := &mcp.Action{
		Type:      mcp.ActionTypeSendEmail,
		AdapterID: "gmail",
		Parameters: map[string]any{
			"to":           "renewals@adobe.com",
			"vendor":       "Adobe Systems",
			"contract_ref": "Adobe Enterprise MSA-2025",
		},
	}

	result, err := adapter.Execute(ctx, action)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotEmpty(t, result.ExternalID)
	assert.Contains(t, result.Message, "renewals@adobe.com")

	body := result.Data["body"].(string)
	assert.Contains(t, body, "Adobe Systems")
	assert.Contains(t, body, "notice that we do not wish to renew")
}

func TestCalendarAdapter_Execute(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	adapter := calendar.New(logger)

	assert.Equal(t, "calendar", adapter.Name())
	assert.Contains(t, adapter.SupportedActions(), mcp.ActionTypeCreateEvent)
	require.NoError(t, adapter.HealthCheck(context.Background()))

	ctx := context.Background()
	action := &mcp.Action{
		Type:      mcp.ActionTypeCreateEvent,
		AdapterID: "calendar",
		Parameters: map[string]any{
			"title":       "AWS Cloud Hosting Payment Due",
			"due_date":    "2026-09-15",
			"description": "Invoice payment deadline for AWS Luxembourg",
		},
	}

	result, err := adapter.Execute(ctx, action)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotEmpty(t, result.ExternalID)

	ical := result.Data["ical_payload"].(string)
	assert.Contains(t, ical, "BEGIN:VCALENDAR")
	assert.Contains(t, ical, "AWS Cloud Hosting Payment Due")
	assert.Contains(t, ical, "TRIGGER:-P1D")
	assert.Contains(t, ical, "END:VCALENDAR")
}

func TestSlackAdapter_Execute(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	adapter := slack.New(logger)

	assert.Equal(t, "slack", adapter.Name())
	assert.Contains(t, adapter.SupportedActions(), mcp.ActionTypeSendSlack)
	require.NoError(t, adapter.HealthCheck(context.Background()))

	ctx := context.Background()
	action := &mcp.Action{
		Type:      mcp.ActionTypeSendSlack,
		AdapterID: "slack",
		Parameters: map[string]any{
			"channel":    "#procurement",
			"title":      "Urgent Contract Notice",
			"message":    "Contract deadline within 3 days!",
			"risk_level": "CRITICAL",
		},
	}

	result, err := adapter.Execute(ctx, action)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotEmpty(t, result.ExternalID)
	assert.Contains(t, result.Message, "#procurement")
	assert.Equal(t, "#D00000", result.Data["color"]) // Red for CRITICAL
}

func TestRegistry(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	reg := mcp.NewRegistry()

	reg.Register(gmail.New(logger))
	reg.Register(calendar.New(logger))
	reg.Register(slack.New(logger))

	assert.Len(t, reg.List(), 3)

	gm, ok := reg.Get("gmail")
	assert.True(t, ok)
	assert.Equal(t, "gmail", gm.Name())

	_, ok = reg.Get("unknown")
	assert.False(t, ok)
}
