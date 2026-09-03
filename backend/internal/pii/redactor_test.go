package pii_test

import (
	"context"
	"strings"
	"testing"

	"github.com/masterfabric-go/masterfabric/internal/pii"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatternRedactor_RedactAndRestore(t *testing.T) {
	redactor := pii.NewPatternRedactor()
	ctx := context.Background()

	originalText := `Müşteri Bilgileri:
TCKN: 12345678901
Telefon: +90 532 123 45 67
E-posta: ahmet.yilmaz@acme.com
IBAN: TR330006100511123456789012`

	redacted, mapping, err := redactor.Redact(ctx, originalText)
	require.NoError(t, err)

	// Verify all sensitive items were masked
	assert.False(t, strings.Contains(redacted, "12345678901"))
	assert.False(t, strings.Contains(redacted, "+90 532 123 45 67"))
	assert.False(t, strings.Contains(redacted, "ahmet.yilmaz@acme.com"))
	assert.False(t, strings.Contains(redacted, "TR330006100511123456789012"))

	assert.Contains(t, redacted, "[TCKN_REDACTED")
	assert.Contains(t, redacted, "[PHONE_REDACTED")
	assert.Contains(t, redacted, "[EMAIL_REDACTED")
	assert.Contains(t, redacted, "[IBAN_REDACTED")

	// Verify restoration restores exact text
	restored, err := redactor.Restore(ctx, redacted, mapping)
	require.NoError(t, err)
	assert.Equal(t, originalText, restored)
}
