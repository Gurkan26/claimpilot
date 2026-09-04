package mcp

import (
	"context"
	"sync"

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

// AdapterInfo provides runtime metadata for admin console.
type AdapterInfo struct {
	Name             string       `json:"name"`
	Enabled          bool         `json:"enabled"`
	SupportedActions []ActionType `json:"supported_actions"`
}

// Registry holds all registered MCP adapters.
type Registry struct {
	adapters map[string]Adapter
	disabled map[string]bool
	mu       sync.RWMutex
}

// NewRegistry creates an empty adapter registry.
func NewRegistry() *Registry {
	return &Registry{
		adapters: make(map[string]Adapter),
		disabled: make(map[string]bool),
	}
}

// Register adds an adapter to the registry.
func (r *Registry) Register(adapter Adapter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[adapter.Name()] = adapter
}

// Get retrieves an adapter by name. Returns false if not registered or disabled.
func (r *Registry) Get(name string) (Adapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.disabled[name] {
		return nil, false
	}
	a, ok := r.adapters[name]
	return a, ok
}

// Enable activates an adapter.
func (r *Registry) Enable(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.disabled, name)
}

// Disable deactivates an adapter.
func (r *Registry) Disable(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.disabled[name] = true
}

// IsEnabled returns true if adapter exists and is not disabled.
func (r *Registry) IsEnabled(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.adapters[name]
	return exists && !r.disabled[name]
}

// List returns all registered adapter names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.adapters))
	for name := range r.adapters {
		names = append(names, name)
	}
	return names
}

// ListDetails returns details and enabled status for all adapters.
func (r *Registry) ListDetails() []AdapterInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]AdapterInfo, 0, len(r.adapters))
	for name, a := range r.adapters {
		list = append(list, AdapterInfo{
			Name:             name,
			Enabled:          !r.disabled[name],
			SupportedActions: a.SupportedActions(),
		})
	}
	return list
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
