package mcp

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// ActionType defines the type of MCP action.
type ActionType string

const (
	ActionTypeSendEmail      ActionType = "send_email"
	ActionTypeCreateEvent    ActionType = "create_calendar_event"
	ActionTypeSendSlack      ActionType = "send_slack_message"
	ActionTypeCancelContract ActionType = "cancel_contract"
	ActionTypeNotify         ActionType = "notify"
)

// Action represents an action to be executed via an MCP adapter.
type Action struct {
	Type       ActionType     `json:"type"`
	AdapterID  string         `json:"adapter_id"` // gmail, calendar, slack
	Parameters map[string]any `json:"parameters"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// ActionResult represents the outcome of an MCP action execution.
type ActionResult struct {
	Success    bool           `json:"success"`
	Message    string         `json:"message"`
	ExternalID string         `json:"external_id,omitempty"` // ID from external service
	Data       map[string]any `json:"data,omitempty"`
}

// Adapter is the interface all MCP adapters must implement.
// Each adapter represents an integration with an external service
// (Gmail, Calendar, Slack, etc.) through which the agent takes actions.
type Adapter interface {
	// Name returns the adapter identifier (e.g., "gmail", "calendar", "slack").
	Name() string

	// Execute performs the given action via the external service.
	Execute(ctx context.Context, action *Action) (*ActionResult, error)

	// HealthCheck verifies the adapter's connection to the external service.
	HealthCheck(ctx context.Context) error

	// SupportedActions returns the list of action types this adapter supports.
	SupportedActions() []ActionType
}

// Registry holds all registered MCP adapters.
type Registry struct {
	adapters map[string]Adapter
}

// NewRegistry creates an empty adapter registry.
func NewRegistry() *Registry {
	return &Registry{
		adapters: make(map[string]Adapter),
	}
}

// Register adds an adapter to the registry.
func (r *Registry) Register(adapter Adapter) {
	r.adapters[adapter.Name()] = adapter
}

// Get retrieves an adapter by name.
func (r *Registry) Get(name string) (Adapter, bool) {
	a, ok := r.adapters[name]
	return a, ok
}

// List returns all registered adapter names.
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.adapters))
	for name := range r.adapters {
		names = append(names, name)
	}
	return names
}

// AuditEntry represents an auditable record of an MCP action.
type AuditEntry struct {
	ID           bson.ObjectID  `bson:"_id,omitempty" json:"id"`
	AdapterName  string         `bson:"adapter_name" json:"adapter_name"`
	ActionType   ActionType     `bson:"action_type" json:"action_type"`
	Parameters   map[string]any `bson:"parameters" json:"parameters"`
	Result       *ActionResult  `bson:"result" json:"result"`
	UserID       string         `bson:"user_id" json:"user_id"`
	ObligationID string         `bson:"obligation_id,omitempty" json:"obligation_id,omitempty"`
}
