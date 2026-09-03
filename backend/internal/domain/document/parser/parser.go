package parser

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	docModel "github.com/masterfabric-go/masterfabric/internal/domain/document/model"
)

// Parser extracts plain text content from document streams.
type Parser interface {
	Parse(ctx context.Context, mimeType string, r io.Reader) (string, error)
	Supports(mimeType string) bool
}

// CompositeParser dispatches extraction to registered format-specific parsers.
type CompositeParser struct {
	parsers []Parser
}

// NewDefaultParser creates a composite parser supporting text, markdown, json, csv, and pdf.
func NewDefaultParser() *CompositeParser {
	return &CompositeParser{
		parsers: []Parser{
			&TextParser{},
			&PDFParser{},
		},
	}
}

func (cp *CompositeParser) Supports(mimeType string) bool {
	for _, p := range cp.parsers {
		if p.Supports(mimeType) {
			return true
		}
	}
	return false
}

func (cp *CompositeParser) Parse(ctx context.Context, mimeType string, r io.Reader) (string, error) {
	for _, p := range cp.parsers {
		if p.Supports(mimeType) {
			return p.Parse(ctx, mimeType, r)
		}
	}
	// Fallback to text parser
	return (&TextParser{}).Parse(ctx, mimeType, r)
}

// TextParser handles plain text, markdown, JSON, and CSV documents.
type TextParser struct{}

func (p *TextParser) Supports(mimeType string) bool {
	mime := strings.ToLower(mimeType)
	return strings.HasPrefix(mime, "text/") ||
		strings.Contains(mime, "json") ||
		strings.Contains(mime, "csv") ||
		mime == "application/octet-stream" ||
		mime == ""
}

func (p *TextParser) Parse(_ context.Context, _ string, r io.Reader) (string, error) {
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return "", fmt.Errorf("read text content: %w", err)
	}
	text := strings.TrimSpace(buf.String())
	return text, nil
}

// PDFParser extracts text streams from standard PDF files.
type PDFParser struct{}

func (p *PDFParser) Supports(mimeType string) bool {
	return strings.EqualFold(mimeType, "application/pdf")
}

func (p *PDFParser) Parse(_ context.Context, _ string, r io.Reader) (string, error) {
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return "", fmt.Errorf("read pdf stream: %w", err)
	}

	content := buf.String()

	// If the file is plain text content uploaded as PDF (common in test fixtures), return directly
	if !strings.HasPrefix(content, "%PDF-") {
		return strings.TrimSpace(content), nil
	}

	// Extract standard PDF text objects (between BT and ET operators)
	btEtRegex := regexp.MustCompile(`(?s)BT(.*?)ET`)
	tjRegex := regexp.MustCompile(`\((.*?)\)\s*Tj|\[(.*?)\]\s*TJ`)

	matches := btEtRegex.FindAllStringSubmatch(content, -1)
	var sb strings.Builder

	for _, match := range matches {
		if len(match) > 1 {
			tjMatches := tjRegex.FindAllStringSubmatch(match[1], -1)
			for _, tj := range tjMatches {
				if len(tj) > 1 && tj[1] != "" {
					sb.WriteString(tj[1])
					sb.WriteString(" ")
				} else if len(tj) > 2 && tj[2] != "" {
					// Clean up TJ array brackets
					cleaned := regexp.MustCompile(`\((.*?)\)`).FindAllStringSubmatch(tj[2], -1)
					for _, c := range cleaned {
						if len(c) > 1 {
							sb.WriteString(c[1])
							sb.WriteString(" ")
						}
					}
				}
			}
			sb.WriteString("\n")
		}
	}

	extracted := strings.TrimSpace(sb.String())
	if extracted != "" {
		return extracted, nil
	}

	// Fallback: search for readable ascii lines of reasonable length
	asciiRegex := regexp.MustCompile(`[a-zA-Z0-9\s.,:;/%$€£¥\-_'"]{10,}`)
	asciiMatches := asciiRegex.FindAllString(content, -1)
	if len(asciiMatches) > 0 {
		return strings.TrimSpace(strings.Join(asciiMatches, " ")), nil
	}

	return "", fmt.Errorf("unable to extract text from PDF (document may be scanned image without OCR)")
}

// DetectDocumentType infers DocumentType from file name and raw text keywords.
func DetectDocumentType(fileName, content string) docModel.DocumentType {
	lowerName := strings.ToLower(fileName)
	lowerContent := strings.ToLower(content)

	// Invoice keywords
	invoiceKeywords := []string{"invoice", "fatura", "bill", "billing", "odeme", "tahsilat", "tax invoice", "subtotal", "amount due"}
	for _, kw := range invoiceKeywords {
		if strings.Contains(lowerName, kw) || strings.Contains(lowerContent, kw) {
			return docModel.DocumentTypeInvoice
		}
	}

	// License keywords
	licenseKeywords := []string{"license", "lisans", "seats", "per-seat", "seat allocation", "authorized users", "eula", "software license"}
	for _, kw := range licenseKeywords {
		if strings.Contains(lowerName, kw) || strings.Contains(lowerContent, kw) {
			return docModel.DocumentTypeLicense
		}
	}

	// Contract keywords
	contractKeywords := []string{"contract", "sozlesme", "agreement", "master services", "terms and conditions", "nda", "service level", "auto-renew"}
	for _, kw := range contractKeywords {
		if strings.Contains(lowerName, kw) || strings.Contains(lowerContent, kw) {
			return docModel.DocumentTypeContract
		}
	}

	// Subscription keywords
	subKeywords := []string{"subscription", "abonelik", "membership", "recurring"}
	for _, kw := range subKeywords {
		if strings.Contains(lowerName, kw) || strings.Contains(lowerContent, kw) {
			return docModel.DocumentTypeSubscription
		}
	}

	return docModel.DocumentTypeOther
}
