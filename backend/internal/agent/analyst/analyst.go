package analyst

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	docModel "github.com/masterfabric-go/masterfabric/internal/domain/document/model"
	oblModel "github.com/masterfabric-go/masterfabric/internal/domain/obligation/model"
	"github.com/masterfabric-go/masterfabric/internal/agent/llmclient"
)

// Analyst is the AI agent responsible for document analysis, field extraction,
// and obligation/opportunity detection.
type Analyst struct {
	provider llmclient.Provider
	logger   *slog.Logger
}

// New creates an Analyst agent backed by the given LLM provider.
// The provider is configured via environment variables (LLM_ANALYST_*).
func New(provider llmclient.Provider, logger *slog.Logger) *Analyst {
	return &Analyst{
		provider: provider,
		logger:   logger.With("agent", "analyst", "model", provider.Model()),
	}
}

// AnalysisResult holds the complete analysis of a document.
type AnalysisResult struct {
	Extraction   *docModel.ExtractionResult `json:"extraction"`
	Obligations  []ObligationCandidate      `json:"obligations"`
	Opportunities []OpportunityCandidate    `json:"opportunities"`
}

// ObligationCandidate is a potential obligation detected by the analyst.
type ObligationCandidate struct {
	Type        oblModel.ObligationType `json:"type"`
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	DueDate     string                  `json:"due_date"` // ISO 8601 string from LLM
	RiskLevel   oblModel.RiskLevel      `json:"risk_level"`
	SourceRef   string                  `json:"source_ref"` // page/clause reference
	Confidence  float64                 `json:"confidence"`
}

// OpportunityCandidate is a potential marketplace opportunity detected by the analyst.
type OpportunityCandidate struct {
	Category          string  `json:"category"`
	CurrentVendorName string  `json:"current_vendor_name"`
	EstimatedCost     float64 `json:"estimated_cost"`
	Currency          string  `json:"currency"`
	Reason            string  `json:"reason"` // why this is an opportunity
}

// Specialized prompts for different document types
const genericAnalysisPrompt = `You are a document analysis AI agent for ClaimPilot.
Your job is to analyze documents (contracts, invoices, licenses, policies, subscriptions) and extract:
1. Key fields (parties, dates, amounts, terms)
2. Obligations (renewals, payments, deadlines, compliance requirements)
3. Marketplace opportunities (vendor switches, cost savings)

Respond ONLY with valid JSON in the following format:
{
  "extraction": {
    "fields": [{"field_name": "...", "value": "...", "confidence": 0.95, "page_ref": 1}],
    "summary": "..."
  },
  "obligations": [
    {
      "type": "RENEWAL|PAYMENT|REPORT|WARRANTY|CANCELLATION|COMPLIANCE",
      "title": "...",
      "description": "...",
      "due_date": "2024-12-31",
      "risk_level": "LOW|MEDIUM|HIGH|CRITICAL",
      "source_ref": "Page 3, Clause 5.2",
      "confidence": 0.9
    }
  ],
  "opportunities": [
    {
      "category": "cloud-hosting",
      "current_vendor_name": "...",
      "estimated_cost": 1200.00,
      "currency": "USD",
      "reason": "..."
    }
  ]
}`

const contractAnalysisPrompt = `You are a contract analysis AI agent for ClaimPilot.
You specialize in parsing commercial contracts, SaaS agreements, service level agreements (SLAs), and non-disclosure agreements.
Extract:
1. Core contract terms: Parties (vendor/customer), Effective Date, Expiration/End Date, Auto-Renewal Clause (Yes/No), Notice Period for Termination (e.g. 30 days prior), Total Contract Value, Currency, Usage/Commitment tiers.
2. Obligations:
   - RENEWAL: Mark due date as (End Date - Notice Period Days) if cancellation/non-renewal notice is required!
   - CANCELLATION: Notice requirements and penalty terms.
   - COMPLIANCE / AUDIT: Any reporting or audit obligations.
3. Marketplace Opportunities:
   - If contract contains recurring SaaS/hosting fees, propose an opportunity for vendor comparison or optimization.

Respond ONLY with valid JSON:
{
  "extraction": {
    "fields": [
      {"field_name": "vendor_name", "value": "...", "confidence": 0.98},
      {"field_name": "customer_name", "value": "...", "confidence": 0.98},
      {"field_name": "effective_date", "value": "YYYY-MM-DD", "confidence": 0.95},
      {"field_name": "expiration_date", "value": "YYYY-MM-DD", "confidence": 0.95},
      {"field_name": "auto_renewal", "value": "true|false", "confidence": 0.95},
      {"field_name": "notice_period_days", "value": "30", "confidence": 0.90},
      {"field_name": "total_value", "value": "12000.00", "confidence": 0.95},
      {"field_name": "currency", "value": "USD", "confidence": 0.95},
      {"field_name": "utilization_rate", "value": "41%", "confidence": 0.85}
    ],
    "summary": "High-level contract summary"
  },
  "obligations": [
    {
      "type": "RENEWAL",
      "title": "Contract Renewal / Termination Notice Deadline",
      "description": "...",
      "due_date": "YYYY-MM-DD",
      "risk_level": "HIGH",
      "source_ref": "Clause X.Y",
      "confidence": 0.92
    }
  ],
  "opportunities": [
    {
      "category": "saas|cloud-hosting|vendor-switch",
      "current_vendor_name": "...",
      "estimated_cost": 12000.00,
      "currency": "USD",
      "reason": "Contract renewal approaching with low utilization or cost saving potential"
    }
  ]
}`

const invoiceAnalysisPrompt = `You are a financial invoice analysis AI agent for ClaimPilot.
You specialize in extracting billing, payment, and expense obligations from invoices, bills, and payment notices.
Extract:
1. Core invoice details: Vendor/Issuer Name, Recipient/Customer Name, Invoice Number, Invoice Date, Payment Due Date, Subtotal, Tax/VAT, Total Amount, Currency, Payment Terms (Net 15, Net 30, etc.), Late Payment Penalty Terms.
2. Obligations:
   - PAYMENT: Create an obligation for the invoice payment due date with amount and payment details.
     Risk Level is HIGH or CRITICAL if payment due date is within 7-14 days.
3. Marketplace Opportunities:
   - If recurring hosting/infrastructure/SaaS bill (e.g. AWS, GCP, Datadog), suggest competitive vendor review if costs are high.

Respond ONLY with valid JSON:
{
  "extraction": {
    "fields": [
      {"field_name": "vendor_name", "value": "...", "confidence": 0.98},
      {"field_name": "invoice_number", "value": "...", "confidence": 0.99},
      {"field_name": "invoice_date", "value": "YYYY-MM-DD", "confidence": 0.95},
      {"field_name": "due_date", "value": "YYYY-MM-DD", "confidence": 0.98},
      {"field_name": "total_amount", "value": "3450.00", "confidence": 0.98},
      {"field_name": "currency", "value": "EUR", "confidence": 0.99},
      {"field_name": "payment_terms", "value": "Net 14", "confidence": 0.90}
    ],
    "summary": "Invoice summary and items"
  },
  "obligations": [
    {
      "type": "PAYMENT",
      "title": "Invoice Payment Due: [Vendor Name]",
      "description": "Payment of 3450.00 EUR due on YYYY-MM-DD",
      "due_date": "YYYY-MM-DD",
      "risk_level": "HIGH",
      "source_ref": "Invoice Header / Terms",
      "confidence": 0.95
    }
  ],
  "opportunities": [
    {
      "category": "cloud-hosting",
      "current_vendor_name": "...",
      "estimated_cost": 3450.00,
      "currency": "EUR",
      "reason": "Monthly cloud infrastructure expense review"
    }
  ]
}`

const licenseAnalysisPrompt = `You are a software license and asset compliance AI agent for ClaimPilot.
You specialize in software subscriptions, seat allocations, EULAs, and cloud software license renewals.
Extract:
1. Core license details: Software/Vendor Name, License Metric (per-seat, core, user tier, enterprise), Total Seats Purchased, Active/Assigned Seats, Seat Utilization Rate, Renewal/Expiration Date, Cost per Seat, Total Annual Cost, Currency.
2. Obligations:
   - RENEWAL: License renewal deadline.
   - COMPLIANCE: Seat true-up or audit obligations.
3. Marketplace Opportunities:
   - Underutilized licenses (e.g. < 60% active seats) represent immediate downgrade or renegotiation opportunity.
   - Alternative vendor options in the same category.

Respond ONLY with valid JSON:
{
  "extraction": {
    "fields": [
      {"field_name": "software_name", "value": "...", "confidence": 0.98},
      {"field_name": "vendor_name", "value": "...", "confidence": 0.95},
      {"field_name": "license_type", "value": "Per-Seat Enterprise", "confidence": 0.92},
      {"field_name": "total_seats", "value": "100", "confidence": 0.95},
      {"field_name": "active_seats", "value": "38", "confidence": 0.90},
      {"field_name": "utilization_rate", "value": "38%", "confidence": 0.90},
      {"field_name": "renewal_date", "value": "YYYY-MM-DD", "confidence": 0.95},
      {"field_name": "annual_cost", "value": "21600.00", "confidence": 0.95},
      {"field_name": "currency", "value": "USD", "confidence": 0.98}
    ],
    "summary": "License tier and utilization summary"
  },
  "obligations": [
    {
      "type": "RENEWAL",
      "title": "Software License Renewal: [Software Name]",
      "description": "Annual renewal for 100 seats at 21,600 USD",
      "due_date": "YYYY-MM-DD",
      "risk_level": "MEDIUM",
      "source_ref": "Section 4.1",
      "confidence": 0.92
    }
  ],
  "opportunities": [
    {
      "category": "saas",
      "current_vendor_name": "...",
      "estimated_cost": 21600.00,
      "currency": "USD",
      "reason": "Active utilization is only 38%; reduce seats from 100 to 45 for 55% cost savings"
    }
  ]
}`

// getPromptForType selects the specialized extraction prompt based on document type.
func getPromptForType(fileType docModel.DocumentType) string {
	switch fileType {
	case docModel.DocumentTypeContract:
		return contractAnalysisPrompt
	case docModel.DocumentTypeInvoice:
		return invoiceAnalysisPrompt
	case docModel.DocumentTypeLicense:
		return licenseAnalysisPrompt
	default:
		return genericAnalysisPrompt
	}
}

// cleanJSONContent strips markdown code fences if emitted by the LLM.
func cleanJSONContent(content string) string {
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "```json") {
		trimmed = strings.TrimPrefix(trimmed, "```json")
	} else if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```")
	}
	if strings.HasSuffix(trimmed, "```") {
		trimmed = strings.TrimSuffix(trimmed, "```")
	}
	return strings.TrimSpace(trimmed)
}

// AnalyzeDocument performs full analysis on a document's text content.
func (a *Analyst) AnalyzeDocument(ctx context.Context, redactedContent string, fileType docModel.DocumentType) (*AnalysisResult, error) {
	a.logger.Info("starting document analysis",
		"content_length", len(redactedContent),
		"file_type", fileType,
	)

	prompt := getPromptForType(fileType)

	req := &llmclient.CompletionRequest{
		SystemPrompt: prompt,
		Messages: []llmclient.Message{
			{
				Role:    llmclient.RoleUser,
				Content: fmt.Sprintf("Analyze this %s document:\n\n%s", fileType, redactedContent),
			},
		},
		Temperature: 0.1, // low temperature for factual extraction
		MaxTokens:   4096,
	}

	resp, err := a.provider.Complete(ctx, req)
	if err != nil {
		a.logger.Error("document analysis failed", "error", err)
		return nil, fmt.Errorf("analyst LLM call failed: %w", err)
	}

	a.logger.Info("document analysis completed",
		"tokens_used", resp.TokensUsed,
		"duration", resp.Duration,
	)

	cleaned := cleanJSONContent(resp.Content)

	var result AnalysisResult
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		a.logger.Error("failed to parse analysis result", "error", err, "raw_content", resp.Content)
		return nil, fmt.Errorf("parse analyst output: %w", err)
	}

	// Enrich extraction result with metadata
	if result.Extraction != nil {
		result.Extraction.ProcessedAt = time.Now().UTC()
		result.Extraction.ModelUsed = resp.Model
		result.Extraction.RawText = redactedContent
	}

	return &result, nil
}

// DetectObligations performs obligation-only detection on already-extracted content.
func (a *Analyst) DetectObligations(ctx context.Context, content string) ([]ObligationCandidate, error) {
	result, err := a.AnalyzeDocument(ctx, content, docModel.DocumentTypeOther)
	if err != nil {
		return nil, err
	}
	return result.Obligations, nil
}
