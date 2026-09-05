package llmclient

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/masterfabric-go/masterfabric/internal/shared/config"
)

// builtinProvider implements the Provider interface completely in-process in Go.
// It requires NO external Ollama, OpenAI, or API keys, and provides realistic,
// deterministic contract analysis, obligation extraction, and verifier responses.
type builtinProvider struct {
	model string
}

// NewBuiltinProvider creates a Go built-in test AI provider.
func NewBuiltinProvider(cfg config.LLMConfig) Provider {
	model := cfg.Model
	if model == "" {
		model = "go-builtin-ai-v1"
	}
	return &builtinProvider{model: model}
}

func (p *builtinProvider) Name() string {
	return "builtin"
}

func (p *builtinProvider) Model() string {
	return p.model
}

func (p *builtinProvider) HealthCheck(_ context.Context) error {
	return nil
}

func (p *builtinProvider) Complete(_ context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	start := time.Now()

	// Gather user input text from messages
	var userText string
	for _, m := range req.Messages {
		if m.Role == RoleUser {
			userText += m.Content + "\n"
		}
	}

	responseContent := p.generateResponse(req.SystemPrompt, userText)

	return &CompletionResponse{
		Content:      responseContent,
		Model:        p.model,
		TokensUsed:   len(responseContent) / 4,
		FinishReason: "stop",
		Duration:     time.Since(start),
	}, nil
}

func (p *builtinProvider) StreamComplete(_ context.Context, req *CompletionRequest) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 2)
	go func() {
		defer close(ch)
		res, err := p.Complete(context.Background(), req)
		if err != nil {
			ch <- StreamChunk{Error: err, Done: true}
			return
		}
		ch <- StreamChunk{Content: res.Content, Done: false}
		ch <- StreamChunk{Done: true}
	}()
	return ch, nil
}

func (p *builtinProvider) generateResponse(systemPrompt, userText string) string {
	lowerSys := strings.ToLower(systemPrompt)
	lowerUser := strings.ToLower(userText)

	// 1. Verifier request check
	if strings.Contains(lowerSys, "verifier") || strings.Contains(lowerSys, "verify") || strings.Contains(lowerUser, "verify") {
		return `{
  "verified": true,
  "confidence_score": 0.985,
  "audit_status": "VERIFIED",
  "reasoning": "Go Dahili Denetçi Motoru: Sözleşmedeki 30 günlük fesih ihbar süresi ve mali yükümlülük maddeleri doğrulandı.",
  "action_required": "HUMAN_APPROVAL"
}`
	}

	// 2. Document Analysis / Contract Extraction check
	if strings.Contains(lowerSys, "document analysis") || strings.Contains(lowerSys, "contract analysis") || strings.Contains(lowerUser, "analyze this") {
		return p.extractContractJSON(userText)
	}

	// 3. Conversational / Chat query
	if strings.Contains(lowerUser, "merhaba") || strings.Contains(lowerUser, "selam") || strings.Contains(lowerUser, "hello") || strings.Contains(lowerUser, "kimsin") {
		return "Merhaba! Ben ClaimPilot Spark dahili AI çalışanıyım. Sözleşmelerinizi, ihtarname sürelerinizi, yaklaşan taahhütlerinizi ve tasarruf fırsatlarınızı Go dahili test motoru üzerinden anında yönetebilirsiniz. 'help' yazarak tüm komutları görebilirsiniz."
	}

	if strings.Contains(lowerUser, "taahhüt") || strings.Contains(lowerUser, "obligation") || strings.Contains(lowerUser, "sözleşme") {
		return "ClaimPilot şu anda portföyünüzdeki sözleşmeleri aktif olarak denetliyor. Yaklaşan ihtarname süreleri için 'obligations' yazabilir, kurumsal tasarruf fırsatları için 'marketplace' komutunu kullanabilirsiniz."
	}

	return fmt.Sprintf("ClaimPilot Spark (Go Dahili AI Motoru): '%s' talebiniz analiz edildi. İlgili sözleşme şartları ve yükümlülükler doğrulanmıştır. Detaylar için 'briefing' veya 'obligations' yazabilirsiniz.", strings.TrimSpace(userText))
}

// extractContractJSON parses realistic contract fields from user text via heuristic rules.
func (p *builtinProvider) extractContractJSON(text string) string {
	vendor := "Kurumsal Hizmet Sağlayıcısı"
	if matched := regexp.MustCompile(`(?i)(adobe|aws|amazon|microsoft|google|salesforce|oracle|turkcell|vodafone|acme)`).FindString(text); matched != "" {
		vendor = strings.Title(matched)
	}

	noticeDays := 30
	if matched := regexp.MustCompile(`(?i)(\d{1,3})\s*[- ]*(?:day|gün|gun)`).FindStringSubmatch(text); len(matched) > 1 {
		if d, err := strconv.Atoi(matched[1]); err == nil && d > 0 {
			noticeDays = d
		}
	}

	cost := 3600.0
	currency := "USD"
	if strings.Contains(text, "EUR") || strings.Contains(text, "€") {
		currency = "EUR"
		cost = 2450.0
	} else if strings.Contains(text, "TL") || strings.Contains(text, "₺") {
		currency = "TRY"
		cost = 45000.0
	}

	// Calculate due date (target: 30 days from now, or end of next month)
	targetDate := time.Now().AddDate(0, 1, 0).Format("2006-01-02")

	result := map[string]any{
		"extraction": map[string]any{
			"fields": []map[string]any{
				{"field_name": "vendor_name", "value": vendor, "confidence": 0.98},
				{"field_name": "notice_period_days", "value": fmt.Sprintf("%d Gün", noticeDays), "confidence": 0.96},
				{"field_name": "annual_commitment", "value": fmt.Sprintf("%.2f %s", cost, currency), "confidence": 0.95},
				{"field_name": "auto_renewal", "value": "Otomatik Yenileme (Aktif)", "confidence": 0.99},
			},
			"summary": fmt.Sprintf("%s sözleşmesi: %d günlük son fesih/ihbar bildirim süresi ve %.2f %s yıllık taahhüt içerir.", vendor, noticeDays, cost, currency),
		},
		"obligations": []map[string]any{
			{
				"type":        "RENEWAL",
				"title":       fmt.Sprintf("%s Sözleşme Yenileme İhbarnamesi", vendor),
				"description": fmt.Sprintf("%d gün önceden yazılı ihtarname verilmediğinde sözleşme otomatik yenilenecektir. Yıllık maliyet: %.2f %s.", noticeDays, cost, currency),
				"due_date":    targetDate,
				"risk_level":  "HIGH",
				"source_ref":  "Madde 3.1 (Fesih ve Otomatik Uzama)",
				"confidence":  0.96,
			},
		},
		"opportunities": []map[string]any{
			{
				"category":            "saas-optimization",
				"current_vendor_name": vendor,
				"estimated_cost":      cost,
				"currency":            currency,
				"reason":              fmt.Sprintf("%s pazar alternatifleri değerlendirilerek %s25%% maliyet optimizasyonu sağlanabilir.", vendor, ""),
			},
		},
	}

	bytesData, _ := json.MarshalIndent(result, "", "  ")
	return string(bytesData)
}
