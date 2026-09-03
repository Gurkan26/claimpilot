package parser_test

import (
	"context"
	"strings"
	"testing"

	docModel "github.com/masterfabric-go/masterfabric/internal/domain/document/model"
	"github.com/masterfabric-go/masterfabric/internal/domain/document/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextParser_Parse(t *testing.T) {
	p := &parser.TextParser{}
	ctx := context.Background()

	raw := "This is a plain text contract agreement.\nEffective Date: 2026-01-01"
	r := strings.NewReader(raw)

	result, err := p.Parse(ctx, "text/plain", r)
	require.NoError(t, err)
	assert.Equal(t, raw, result)
}

func TestCompositeParser_Default(t *testing.T) {
	cp := parser.NewDefaultParser()
	ctx := context.Background()

	assert.True(t, cp.Supports("text/plain"))
	assert.True(t, cp.Supports("text/markdown"))
	assert.True(t, cp.Supports("application/pdf"))

	raw := "Contract between Company A and Company B"
	res, err := cp.Parse(ctx, "text/plain", strings.NewReader(raw))
	require.NoError(t, err)
	assert.Equal(t, raw, res)
}

func TestDetectDocumentType(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		content  string
		expected docModel.DocumentType
	}{
		{
			name:     "Detect Contract by file name",
			fileName: "Master_Services_Agreement_2026.pdf",
			content:  "Some generic content here",
			expected: docModel.DocumentTypeContract,
		},
		{
			name:     "Detect Invoice by content",
			fileName: "doc123.pdf",
			content:  "Tax Invoice Amount Due: 500 EUR. Payment to be made on or before due date.",
			expected: docModel.DocumentTypeInvoice,
		},
		{
			name:     "Detect License by content",
			fileName: "schedule.txt",
			content:  "Software license schedule with per-seat seat allocation and active users.",
			expected: docModel.DocumentTypeLicense,
		},
		{
			name:     "Detect Subscription",
			fileName: "monthly_subscription.txt",
			content:  "Recurring cloud service subscription",
			expected: docModel.DocumentTypeSubscription,
		},
		{
			name:     "Fallback to Other",
			fileName: "notes.txt",
			content:  "Random personal meeting notes",
			expected: docModel.DocumentTypeOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detected := parser.DetectDocumentType(tt.fileName, tt.content)
			assert.Equal(t, tt.expected, detected)
		})
	}
}
