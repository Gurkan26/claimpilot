package pii

import (
	"context"
)

// RedactionMap holds the mapping between original PII values and their masked
// replacements. This map is stored separately with restricted access (KVKK/GDPR).
type RedactionMap struct {
	Mappings []RedactionEntry `bson:"mappings" json:"mappings"`
}

// RedactionEntry is a single PII value → placeholder mapping.
type RedactionEntry struct {
	Original    string `bson:"original" json:"-"`     // never serialized to JSON
	Placeholder string `bson:"placeholder" json:"placeholder"`
	Category    string `bson:"category" json:"category"` // name, tckn, phone, iban, email
	StartPos    int    `bson:"start_pos" json:"start_pos"`
	EndPos      int    `bson:"end_pos" json:"end_pos"`
}

// Redactor masks PII from text before sending to LLM and can restore it afterward.
type Redactor interface {
	// Redact replaces PII in the given text with placeholders.
	// Returns redacted text and a mapping for later restoration.
	Redact(ctx context.Context, text string) (redactedText string, mapping *RedactionMap, err error)

	// Restore replaces placeholders back with original PII values.
	Restore(ctx context.Context, redactedText string, mapping *RedactionMap) (string, error)
}
