package pii

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// RedactionPattern defines a regex pattern for a PII category.
type RedactionPattern struct {
	Category    string
	Pattern     *regexp.Regexp
	Placeholder string // e.g. "[TCKN]", "[PHONE]"
}

// PatternRedactor implements Redactor using regex-based PII detection.
// Supports Turkish (TCKN, Turkish phone) and international (IBAN, email, credit card) patterns.
type PatternRedactor struct {
	patterns []RedactionPattern
}

// NewPatternRedactor creates a Redactor with default PII patterns for
// Turkey (KVKK) and international (GDPR) compliance.
func NewPatternRedactor() *PatternRedactor {
	return &PatternRedactor{
		patterns: defaultPatterns(),
	}
}

// NewPatternRedactorWithPatterns creates a Redactor with custom patterns.
func NewPatternRedactorWithPatterns(patterns []RedactionPattern) *PatternRedactor {
	return &PatternRedactor{patterns: patterns}
}

func defaultPatterns() []RedactionPattern {
	return []RedactionPattern{
		{
			Category:    "email",
			Pattern:     regexp.MustCompile(`\b[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b`),
			Placeholder: "[EMAIL_REDACTED]",
		},
		{
			Category:    "iban",
			Pattern:     regexp.MustCompile(`\b[A-Z]{2}\d{2}(?:\s?[0-9A-Za-z]){12,30}\b`),
			Placeholder: "[IBAN_REDACTED]",
		},
		{
			Category:    "tckn",
			Pattern:     regexp.MustCompile(`\b[1-9]\d{10}\b`), // Turkish TCKN: 11 digits, not starting with 0
			Placeholder: "[TCKN_REDACTED]",
		},
		{
			Category:    "phone_tr",
			Pattern:     regexp.MustCompile(`(?:\+90|0090|\b0)\s*(?:\(\d{3}\)|\d{3})[\s.-]?\d{3}[\s.-]?\d{2}[\s.-]?\d{2}\b`),
			Placeholder: "[PHONE_REDACTED]",
		},
		{
			Category:    "phone_intl",
			Pattern:     regexp.MustCompile(`\+\d{1,3}[\s.-]?\(?\d{2,4}\)?[\s.-]?\d{3,4}[\s.-]?\d{3,4}\b`),
			Placeholder: "[PHONE_REDACTED]",
		},
		{
			Category:    "credit_card",
			Pattern:     regexp.MustCompile(`\b\d{4}[\s.-]?\d{4}[\s.-]?\d{4}[\s.-]?\d{4}\b`),
			Placeholder: "[CC_REDACTED]",
		},
		{
			Category:    "ssn_us",
			Pattern:     regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
			Placeholder: "[SSN_REDACTED]",
		},
	}
}

func (r *PatternRedactor) Redact(_ context.Context, text string) (string, *RedactionMap, error) {
	if text == "" {
		return "", &RedactionMap{}, nil
	}

	var rawEntries []RedactionEntry
	counter := make(map[string]int)

	for _, p := range r.patterns {
		matches := p.Pattern.FindAllStringIndex(text, -1)
		for _, match := range matches {
			original := text[match[0]:match[1]]
			counter[p.Category]++
			placeholder := fmt.Sprintf("%s_%d", p.Placeholder, counter[p.Category])

			rawEntries = append(rawEntries, RedactionEntry{
				Original:    original,
				Placeholder: placeholder,
				Category:    p.Category,
				StartPos:    match[0],
				EndPos:      match[1],
			})
		}
	}

	if len(rawEntries) == 0 {
		return text, &RedactionMap{}, nil
	}

	// Sort ascending by StartPos to resolve overlaps
	sort.Slice(rawEntries, func(i, j int) bool {
		if rawEntries[i].StartPos == rawEntries[j].StartPos {
			return rawEntries[i].EndPos > rawEntries[j].EndPos
		}
		return rawEntries[i].StartPos < rawEntries[j].StartPos
	})

	// Filter out overlapping matches
	var nonOverlapping []RedactionEntry
	lastEnd := 0
	for _, entry := range rawEntries {
		if entry.StartPos >= lastEnd {
			nonOverlapping = append(nonOverlapping, entry)
			lastEnd = entry.EndPos
		}
	}

	// Apply replacements in reverse order (highest StartPos first) to keep earlier offsets valid
	redactedText := text
	for i := len(nonOverlapping) - 1; i >= 0; i-- {
		entry := nonOverlapping[i]
		redactedText = redactedText[:entry.StartPos] + entry.Placeholder + redactedText[entry.EndPos:]
	}

	return redactedText, &RedactionMap{Mappings: nonOverlapping}, nil
}

func (r *PatternRedactor) Restore(_ context.Context, redactedText string, mapping *RedactionMap) (string, error) {
	if mapping == nil || len(mapping.Mappings) == 0 {
		return redactedText, nil
	}

	result := redactedText
	for _, entry := range mapping.Mappings {
		result = strings.Replace(result, entry.Placeholder, entry.Original, 1)
	}

	return result, nil
}
