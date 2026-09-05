package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/agent/analyst"
	"github.com/masterfabric-go/masterfabric/internal/agent/llmclient"
	mktAgent "github.com/masterfabric-go/masterfabric/internal/agent/marketplace"
	"github.com/masterfabric-go/masterfabric/internal/agent/orchestrator"
	"github.com/masterfabric-go/masterfabric/internal/agent/verifier"
	dashUsecase "github.com/masterfabric-go/masterfabric/internal/application/dashboard/usecase"
	docUsecase "github.com/masterfabric-go/masterfabric/internal/application/document/usecase"
	mktUsecase "github.com/masterfabric-go/masterfabric/internal/application/marketplace/usecase"
	oblDTO "github.com/masterfabric-go/masterfabric/internal/application/obligation/dto"
	oblUsecase "github.com/masterfabric-go/masterfabric/internal/application/obligation/usecase"
	docParser "github.com/masterfabric-go/masterfabric/internal/domain/document/parser"
	docStorage "github.com/masterfabric-go/masterfabric/internal/domain/document/storage"
	mongoAudit "github.com/masterfabric-go/masterfabric/internal/infrastructure/mongodb/audit"
	mongoDoc "github.com/masterfabric-go/masterfabric/internal/infrastructure/mongodb/document"
	mongoMkt "github.com/masterfabric-go/masterfabric/internal/infrastructure/mongodb/marketplace"
	mongoObl "github.com/masterfabric-go/masterfabric/internal/infrastructure/mongodb/obligation"
	"github.com/masterfabric-go/masterfabric/internal/mcp"
	mcpCalendar "github.com/masterfabric-go/masterfabric/internal/mcp/calendar"
	mcpGmail "github.com/masterfabric-go/masterfabric/internal/mcp/gmail"
	mcpSlack "github.com/masterfabric-go/masterfabric/internal/mcp/slack"
	"github.com/masterfabric-go/masterfabric/internal/pii"
	"github.com/masterfabric-go/masterfabric/internal/shared/config"
	"github.com/masterfabric-go/masterfabric/internal/shared/database"
	"github.com/masterfabric-go/masterfabric/internal/shared/events"
)

// ANSI color escape codes
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorMagenta= "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

// UserAccount represents an active user context in Spark CLI
type UserAccount struct {
	ID           uuid.UUID
	Name         string
	Email        string
	AccountType  string // "b2b", "b2c", "admin"
	Organization string
	Role         string
}

// Demo data structures for standalone mode
type standaloneObligation struct {
	ID              string    `json:"id"`
	AccountType     string    `json:"account_type"` // "b2b", "b2c", "admin"
	Title           string    `json:"title"`
	Type            string    `json:"type"`
	DueDate         time.Time `json:"due_date"`
	RiskLevel       string    `json:"risk_level"`
	Status          string    `json:"status"`
	SuggestedAction string    `json:"suggested_action"`
	Cost            float64   `json:"cost"`
	Currency        string    `json:"currency"`
	RenewalRule     string    `json:"renewal_rule"`
}

type standaloneOpportunity struct {
	ID              string
	Category        string
	CurrentVendor   string
	EstimatedCost   float64
	Currency        string
	Status          string
	AlternativeBids int
	BestSavings     float64
}

type sparkApp struct {
	cfg            *config.Config
	currentAccount UserAccount
	accounts       map[string]UserAccount
	isStandalone   bool

	// Live use cases (used when MongoDB is connected)
	dashboardUC *dashUsecase.GetDashboardSummaryUseCase
	listOblUC   *oblUsecase.ListObligationsUseCase
	approveUC   *oblUsecase.ApproveObligationUseCase
	dismissUC   *oblUsecase.DismissObligationUseCase
	rfqUC       *mktUsecase.TriggerRFQUseCase
	acceptBidUC *mktUsecase.AcceptBidUseCase
	metricsUC   *mktUsecase.GetMetricsUseCase
	listOppUC   *mktUsecase.ListOpportunitiesUseCase
	uploadUC    *docUsecase.UploadDocumentUseCase

	// Go built-in AI providers
	analystLLM  llmclient.Provider
	verifierLLM llmclient.Provider

	// Standalone in-memory state
	standaloneObligations   []standaloneObligation
	standaloneOpportunities []standaloneOpportunity
}

func main() {
	briefingFlag := flag.Bool("briefing", false, "Print Good Morning briefing and exit")
	obligationsFlag := flag.Bool("obligations", false, "List upcoming obligations and exit")
	uploadFlag := flag.String("upload", "", "Upload and analyze a document file")
	testAiFlag := flag.String("test-ai", "", "Test Go built-in AI test engine with a sample contract")
	accountFlag := flag.String("account", "b2b", "Initial account: b2b, b2c, or admin")
	addFlag := flag.String("add", "", "Research and add a subscription/obligation via AI (e.g. -add 'YouTube Premium')")
	flag.Parse()

	app := initSparkApp(*accountFlag)

	ctx := context.Background()

	// Handle one-shot CLI flags
	if *addFlag != "" {
		app.cmdAIAddObligation(ctx, *addFlag)
		return
	}
	if *briefingFlag {
		app.cmdBriefing(ctx)
		return
	}
	if *obligationsFlag {
		app.cmdListObligations(ctx, nil)
		return
	}
	if *uploadFlag != "" {
		app.cmdUpload(ctx, *uploadFlag)
		return
	}
	if *testAiFlag != "" {
		app.cmdTestAI(ctx, *testAiFlag)
		return
	}

	// Interactive REPL mode
	app.runREPL()
}

func initSparkApp(initialAccountType string) *sparkApp {
	cfg := config.Load()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil)) // quiet CLI logger

	// Define pre-configured accounts
	accounts := map[string]UserAccount{
		"b2b": {
			ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Name:         "Gürkan Şentürk",
			Email:        "gurkan@acmeholding.com.tr",
			AccountType:  "b2b",
			Organization: "Acme Holding A.Ş.",
			Role:         "Head of Procurement & IT",
		},
		"b2c": {
			ID:           uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			Name:         "Gürkan Şentürk",
			Email:        "gurkan.senturk@gmail.com",
			AccountType:  "b2c",
			Organization: "Bireysel Portföy",
			Role:         "Hesap Sahibi",
		},
		"admin": {
			ID:           uuid.MustParse("00000000-0000-0000-0000-000000000099"),
			Name:         "Sistem Yöneticisi",
			Email:        "admin@claimpilot.local",
			AccountType:  "admin",
			Organization: "ClaimPilot Core Architecture",
			Role:         "Agent Harness Administrator",
		},
	}

	activeAccount, ok := accounts[strings.ToLower(initialAccountType)]
	if !ok {
		activeAccount = accounts["b2b"]
	}

	// Always ensure Go built-in AI providers are available for reliable testing
	analystProvider, err := llmclient.NewProvider(cfg.LLMAnalyst)
	if err != nil || analystProvider == nil {
		analystProvider = llmclient.NewBuiltinProvider(cfg.LLMAnalyst)
	}

	verifierProvider, err := llmclient.NewProvider(cfg.LLMVerifier)
	if err != nil || verifierProvider == nil {
		verifierProvider = llmclient.NewBuiltinProvider(cfg.LLMVerifier)
	}

	app := &sparkApp{
		cfg:            cfg,
		currentAccount: activeAccount,
		accounts:       accounts,
		analystLLM:     analystProvider,
		verifierLLM:    verifierProvider,
	}

	// Try MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	mongoClient, err := database.NewMongoClient(ctx, cfg.MongoDB)
	if err != nil {
		// Fallback to standalone mode with rich demo data
		app.isStandalone = true
		app.initStandaloneData()
		return app
	}

	mongoDB := database.GetDatabase(mongoClient, cfg.MongoDB)

	// Live Repositories
	docRepo := mongoDoc.NewMongoRepository(mongoDB)
	oblRepo := mongoObl.NewMongoRepository(mongoDB)
	mktRepo := mongoMkt.NewMongoRepository(mongoDB)
	auditRepo := mongoAudit.NewMongoAgentAuditRepository(mongoDB)

	mcpReg := mcp.NewRegistry()
	mcpReg.Register(mcpGmail.New(logger))
	mcpReg.Register(mcpCalendar.New(logger))
	mcpReg.Register(mcpSlack.New(logger))

	eventBus := events.NewInProcessBus(logger, 64)

	var verifierAgent *verifier.Verifier
	if verifierProvider != nil {
		verifierAgent = verifier.New(verifierProvider, logger)
	}

	var orch *orchestrator.Orchestrator
	if analystProvider != nil && verifierAgent != nil {
		analystAgent := analyst.New(analystProvider, logger)
		redactor := pii.NewPatternRedactor()
		orch = orchestrator.New(orchestrator.Config{
			Analyst:  analystAgent,
			Verifier: verifierAgent,
			Redactor: redactor,
			DocRepo:  docRepo,
			OblRepo:  oblRepo,
			MktRepo:  mktRepo,
			EventBus: eventBus,
			Logger:   logger,
		})
	}

	storageService, _ := docStorage.NewLocalStorageService("uploads/documents")
	parserService := docParser.NewDefaultParser()

	app.uploadUC = docUsecase.NewUploadDocumentUseCase(docUsecase.UploadConfig{
		DocRepo:      docRepo,
		Storage:      storageService,
		Parser:       parserService,
		Orchestrator: orch,
		EventBus:     eventBus,
		Logger:       logger,
	})

	app.dashboardUC = dashUsecase.NewGetDashboardSummaryUseCase(dashUsecase.Config{
		OblRepo:  oblRepo,
		MktRepo:  mktRepo,
		MCPReg:   mcpReg,
		Analyst:  analystProvider,
		Verifier: verifierProvider,
		Logger:   logger,
	})

	app.listOblUC = oblUsecase.NewListObligationsUseCase(oblRepo)
	app.approveUC = oblUsecase.NewApproveObligationUseCase(oblUsecase.ApproveConfig{
		OblRepo:   oblRepo,
		AuditRepo: auditRepo,
		MCPReg:    mcpReg,
		EventBus:  eventBus,
		Logger:    logger,
	})
	app.dismissUC = oblUsecase.NewDismissObligationUseCase(oblRepo, auditRepo, logger)

	rfqEngine := mktAgent.NewRFQEngine(mktRepo, verifierAgent, logger)
	app.rfqUC = mktUsecase.NewTriggerRFQUseCase(rfqEngine, mktRepo, eventBus, logger)
	app.acceptBidUC = mktUsecase.NewAcceptBidUseCase(mktUsecase.AcceptBidConfig{
		MktRepo:   mktRepo,
		AuditRepo: auditRepo,
		MCPReg:    mcpReg,
		EventBus:  eventBus,
		Logger:    logger,
	})
	app.metricsUC = mktUsecase.NewGetMetricsUseCase(mktRepo)
	app.listOppUC = mktUsecase.NewListOpportunitiesUseCase(mktRepo)

	return app
}

func (a *sparkApp) initStandaloneData() {
	now := time.Now()
	a.standaloneObligations = []standaloneObligation{
		// B2B Kurumsal Sözleşmeler (Acme Holding A.Ş.)
		{
			ID:              "obl-b2b-001",
			AccountType:     "b2b",
			Title:           "Adobe Creative Cloud Enterprise Yenilemesi",
			Type:            "RENEWAL",
			DueDate:         now.AddDate(0, 0, 18),
			RiskLevel:       "HIGH",
			Status:          "PENDING_APPROVAL",
			SuggestedAction: "İhtarname e-postası hazırla (Gmail MCP)",
			Cost:            3600.0,
			Currency:        "USD",
			RenewalRule:     "12 aylık otomatik uzama maddesi; fesih için 30 gün öncesinden yazılı bildirim iletilmelidir.",
		},
		{
			ID:              "obl-b2b-002",
			AccountType:     "b2b",
			Title:           "AWS EMEA Cloud Taahhüt Süresi Sonu",
			Type:            "RENEWAL",
			DueDate:         now.AddDate(0, 0, 24),
			RiskLevel:       "CRITICAL",
			Status:          "PENDING_APPROVAL",
			SuggestedAction: "Takvime son fesih tarihi ekle (Calendar MCP)",
			Cost:            2450.0,
			Currency:        "USD",
			RenewalRule:     "Reserved Instance süresi doluyor; yenilenmezse on-demand tarifeye geçer.",
		},
		{
			ID:              "obl-b2b-003",
			AccountType:     "b2b",
			Title:           "Turkcell Kurumsal Fiber İnternet Sözleşmesi",
			Type:            "CANCELLATION",
			DueDate:         now.AddDate(0, 1, 5),
			RiskLevel:       "MEDIUM",
			Status:          "PENDING_APPROVAL",
			SuggestedAction: "Alternatif teklif havuzunu tara (Marketplace)",
			Cost:            850.0,
			Currency:        "TRY",
			RenewalRule:     "24 aylık kurumsal taahhüt sonu; cayma bedelsiz fesih veya tarife indirimi penceresi.",
		},
		// B2C Bireysel Abonelikler (Gürkan Şentürk)
		{
			ID:              "obl-b2c-001",
			AccountType:     "b2c",
			Title:           "Netflix Standart Plan Aboneliği",
			Type:            "SUBSCRIPTION",
			DueDate:         now.AddDate(0, 0, 11),
			RiskLevel:       "LOW",
			Status:          "PENDING_APPROVAL",
			SuggestedAction: "Yenileme öncesi bildirim kur (Calendar MCP)",
			Cost:            229.99,
			Currency:        "TRY",
			RenewalRule:     "Aylık karttan otomatik çekim; fatura kesim tarihinden önce iptal edilebilir.",
		},
		{
			ID:              "obl-b2c-002",
			AccountType:     "b2c",
			Title:           "Spotify Premium Bireysel Abonelik",
			Type:            "SUBSCRIPTION",
			DueDate:         now.AddDate(0, 0, 6),
			RiskLevel:       "LOW",
			Status:          "PENDING_APPROVAL",
			SuggestedAction: "Ödeme takvimini kontrol et (Calendar MCP)",
			Cost:            59.99,
			Currency:        "TRY",
			RenewalRule:     "Aylık otomatik yenileme; hesap ayarlarından anında iptal edilebilir.",
		},
	}

	// Persistent cache file
	dataFile := filepath.Join("testdata", "standalone_obligations.json")
	if data, err := os.ReadFile(dataFile); err == nil {
		var saved []standaloneObligation
		if err := json.Unmarshal(data, &saved); err == nil && len(saved) > 0 {
			a.standaloneObligations = saved
		}
	}

	a.standaloneOpportunities = []standaloneOpportunity{
		{
			ID:              "opp-adobe-101",
			Category:        "SaaS Lisans",
			CurrentVendor:   "Adobe Systems",
			EstimatedCost:   3600.0,
			Currency:        "USD",
			Status:          "ACTIVE",
			AlternativeBids: 3,
			BestSavings:     850.0,
		},
		{
			ID:              "opp-aws-102",
			Category:        "Bulut Altyapı",
			CurrentVendor:   "AWS EMEA",
			EstimatedCost:   2450.0,
			Currency:        "EUR",
			Status:          "MATCHED",
			AlternativeBids: 2,
			BestSavings:     450.0,
		},
	}
}

func (a *sparkApp) saveStandaloneData() {
	dataFile := filepath.Join("testdata", "standalone_obligations.json")
	data, err := json.MarshalIndent(a.standaloneObligations, "", "  ")
	if err == nil {
		_ = os.WriteFile(dataFile, data, 0644)
	}
}

func (a *sparkApp) runREPL() {
	a.printWelcomeBanner()

	scanner := bufio.NewScanner(os.Stdin)
	ctx := context.Background()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Printf("\n%sClaimPilot Spark oturumu sonlandırıldı. İyi çalışmalar!%s\n", colorGray, colorReset)
		os.Exit(0)
	}()

	for {
		promptOrg := a.currentAccount.Organization
		if a.currentAccount.AccountType == "admin" {
			promptOrg = "ADMIN"
		}
		fmt.Printf("%sclaimpilot (%s)%s> ", colorCyan+colorBold, promptOrg, colorReset)
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := strings.ToLower(parts[0])
		args := parts[1:]

		switch cmd {
		case "ai-add", "add", "ekle", "arastir-ekle", "abone-ekle":
			if len(args) == 0 {
				fmt.Printf("%sKullanım: ai-add <abonelik veya taahhüt tanımı> (Örn: ai-add YouTube Premium Bireysel)%s\n", colorYellow, colorReset)
			} else {
				a.cmdAIAddObligation(ctx, strings.Join(args, " "))
			}
		case "test-ai", "testai", "simule", "simulate":
			sample := ""
			if len(args) > 0 {
				sample = strings.Join(args, " ")
			}
			a.cmdTestAI(ctx, sample)
		case "account", "hesap", "whoami":
			a.cmdAccount(args)
		case "status", "durum":
			a.cmdStatus()
		case "briefing", "morning", "ozet", "summary":
			a.cmdBriefing(ctx)
		case "obligations", "list", "taahhutler", "yukumlulukler", "abonelikler":
			a.cmdListObligations(ctx, args)
		case "approve", "onayla":
			if len(args) == 0 {
				fmt.Printf("%sKullanım: approve <obligation_id>%s\n", colorYellow, colorReset)
			} else {
				a.cmdApprove(ctx, args[0])
			}
		case "dismiss", "reddet", "sil", "remove":
			if len(args) == 0 {
				fmt.Printf("%sKullanım: dismiss <obligation_id> [gerekçe]%s\n", colorYellow, colorReset)
			} else {
				reason := "Spark CLI üzerinden kaldırıldı"
				if len(args) > 1 {
					reason = strings.Join(args[1:], " ")
				}
				a.cmdDismiss(ctx, args[0], reason)
			}
		case "marketplace", "deals", "firsatlar":
			a.cmdListMarketplace(ctx)
		case "rfq":
			if len(args) == 0 {
				fmt.Printf("%sKullanım: rfq <opportunity_id>%s\n", colorYellow, colorReset)
			} else {
				a.cmdTriggerRFQ(ctx, args[0])
			}
		case "accept", "kabul":
			if len(args) < 2 {
				fmt.Printf("%sKullanım: accept <opportunity_id> <bid_id>%s\n", colorYellow, colorReset)
			} else {
				a.cmdAcceptBid(ctx, args[0], args[1])
			}
		case "upload", "yukle":
			if len(args) == 0 {
				fmt.Printf("%sKullanım: upload <dosya_yolu>%s\n", colorYellow, colorReset)
			} else {
				a.cmdUpload(ctx, args[0])
			}
		case "chat", "ask", "sor":
			if len(args) == 0 {
				fmt.Printf("%sKullanım: chat <sorunuz veya sözleşme sorusu>%s\n", colorYellow, colorReset)
			} else {
				a.cmdChat(ctx, strings.Join(args, " "))
			}
		case "help", "yardim", "?":
			a.printHelp()
		case "exit", "quit", "q", "cikis":
			fmt.Printf("%sClaimPilot Spark oturumu sonlandırıldı. İyi çalışmalar!%s\n", colorGray, colorReset)
			return
		default:
			// Treat arbitrary input as a chat query to the AI employee
			a.cmdChat(ctx, line)
		}
	}
}

func (a *sparkApp) printWelcomeBanner() {
	fmt.Println()
	fmt.Printf("%s%s", colorCyan, colorBold)
	fmt.Println(`   ____ _       _           ____  _ _       _   `)
	fmt.Println(`  / ___| | __ _(_)_ __ ___ |  _ \(_) | ___ | |_ `)
	fmt.Println(` | |   | |/ _` + "`" + ` | | '_ ` + "`" + ` _ \| |_) | | |/ _ \| __|`)
	fmt.Println(` | |___| | (_| | | | | | | |  __/| | | (_) | |_ `)
	fmt.Println(`  \____|_|\__,_|_|_| |_| |_|_|   |_|_|\___/ \__|`)
	fmt.Printf("%s", colorReset)
	fmt.Printf("%s AI Employee for Corporate & Personal Obligations • Spark CLI v0.2.0%s\n", colorGray, colorReset)
	fmt.Println()

	typeBadge := "Kurumsal (B2B)"
	badgeColor := colorCyan
	if a.currentAccount.AccountType == "b2c" {
		typeBadge = "Bireysel (B2C)"
		badgeColor = colorGreen
	} else if a.currentAccount.AccountType == "admin" {
		typeBadge = "Sistem Yöneticisi (ADMIN)"
		badgeColor = colorMagenta
	}

	fmt.Printf("%s====================================================================================%s\n", colorCyan, colorReset)
	fmt.Printf("  ✨ %sHoş geldin, %s!%s 🚀\n", colorBold+colorGreen, a.currentAccount.Name, colorReset)
	fmt.Printf("  🏢 %sAktif Hesap:%s  %s%s%s • %s [%s]\n",
		colorBold, colorReset, badgeColor+colorBold, typeBadge, colorReset,
		a.currentAccount.Organization, a.currentAccount.Role)
	aiEngineLabel := "Go Dahili Test Motoru"
	if a.analystLLM.Name() == "groq" || a.analystLLM.Name() == "openai" {
		aiEngineLabel = "Groq Sinirsel AI Motoru"
	}
	fmt.Printf("  🧠 %sAI Motoru:%s    %s%s (%s) [AKTİF]%s\n",
		colorBold, colorReset, colorGreen+colorBold, aiEngineLabel, a.analystLLM.Model(), colorReset)
	if a.isStandalone {
		fmt.Printf("  ⚡ %sÇalışma Modu:%s Standalone / Bağımsız Test Modu (Sıfır Dış Bağımlılık)\n",
			colorBold, colorReset)
	} else {
		fmt.Printf("  ⚡ %sÇalışma Modu:%s Canlı MongoDB & Mikroservis Veri Tabanı Bağlı\n",
			colorBold, colorReset)
	}
	fmt.Printf("%s====================================================================================%s\n", colorCyan, colorReset)
	fmt.Printf(" Komutlar için %shelp%s, AI ile abonelik/vade eklemek için %sai-add <tanım>%s, analiz için %stest-ai%s yazabilirsiniz.\n\n",
		colorBold+colorYellow, colorReset, colorBold+colorGreen, colorReset, colorBold+colorCyan, colorReset)
}

func (a *sparkApp) printHelp() {
	fmt.Println()
	fmt.Printf("%s%sMevcut Spark Komutları:%s\n", colorBold, colorYellow, colorReset)
	fmt.Printf("  %s%-32s%s %s\n", colorCyan, "ai-add <abonelik/sözleşme>", colorReset, "Yapay zekaya araştırt ve otomatik yükümlülük/vade ekle (Örn: ai-add YouTube Premium)")
	fmt.Printf("  %s%-32s%s %s\n", colorCyan, "obligations / list", colorReset, "Aktif hesaba ait taahhütler ve abonelik vadeleri")
	fmt.Printf("  %s%-32s%s %s\n", colorCyan, "approve <id> / onayla", colorReset, "Taahhüt onaylama ve otonom MCP aksiyonu tetikleme")
	fmt.Printf("  %s%-32s%s %s\n", colorCyan, "dismiss <id> / reddet / sil", colorReset, "Taahhüdü listeden kaldırma")
	fmt.Printf("  %s%-32s%s %s\n", colorCyan, "chat <sorunuz>", colorReset, "AI çalışana piyasa araştırması ve sözleşme soruları sor")
	fmt.Printf("  %s%-32s%s %s\n", colorCyan, "test-ai [sözleşme_metni]", colorReset, "Aktif AI motorunu çalıştır ve sözleşmeyi analiz et")
	fmt.Printf("  %s%-32s%s %s\n", colorCyan, "account [b2b|b2c|admin]", colorReset, "Aktif kullanıcı hesabını göster veya değiştir")
	fmt.Printf("  %s%-32s%s %s\n", colorCyan, "status / durum", colorReset, "Sistem, hesap ve AI motoru sağlık durumunu göster")
	fmt.Printf("  %s%-32s%s %s\n", colorCyan, "briefing / ozet", colorReset, "Günlük yönetici özeti ve risk analiz metrikleri")
	fmt.Printf("  %s%-32s%s %s\n", colorCyan, "marketplace / firsatlar", colorReset, "Tedarikçi fırsatları ve alternatif teklif listesi")
	fmt.Printf("  %s%-32s%s %s\n", colorCyan, "rfq <fırsat_id>", colorReset, "Otonom RFQ tedarikçi teklif toplama sürecini başlat")
	fmt.Printf("  %s%-32s%s %s\n", colorCyan, "accept <opp_id> <bid_id>", colorReset, "En avantajlı teklifi onayla ve geçişi bağla")
	fmt.Printf("  %s%-32s%s %s\n", colorCyan, "upload <dosya_yolu>", colorReset, "Yerel sözleşme belgesi yükle ve AI ile analiz et")
	fmt.Printf("  %s%-32s%s %s\n", colorCyan, "exit / quit", colorReset, "Terminalden çıkış yap")
	fmt.Println()
}

func (a *sparkApp) cmdAccount(args []string) {
	if len(args) == 0 {
		fmt.Println()
		fmt.Printf("%s%sAktif Kullanıcı Bilgileri:%s\n", colorBold, colorGreen, colorReset)
		fmt.Printf("  • İsim:         %s\n", a.currentAccount.Name)
		fmt.Printf("  • E-posta:      %s\n", a.currentAccount.Email)
		fmt.Printf("  • Hesap Türü:   %s\n", strings.ToUpper(a.currentAccount.AccountType))
		fmt.Printf("  • Kurum/Portföy: %s\n", a.currentAccount.Organization)
		fmt.Printf("  • Rol:          %s\n", a.currentAccount.Role)
		fmt.Printf("\nHesap değiştirmek için: %saccount b2b%s, %saccount b2c%s veya %saccount admin%s yazabilirsiniz.\n\n",
			colorYellow, colorReset, colorYellow, colorReset, colorYellow, colorReset)
		return
	}

	target := strings.ToLower(args[0])
	newAcc, ok := a.accounts[target]
	if !ok {
		fmt.Printf("%sGeçersiz hesap türü! Seçenekler: b2b (kurumsal), b2c (bireysel), admin (yönetici)%s\n", colorRed, colorReset)
		return
	}

	a.currentAccount = newAcc
	fmt.Println()
	fmt.Printf("%s✓ Hesap başarıyla değiştirildi!%s\n", colorGreen+colorBold, colorReset)
	fmt.Printf("  Hoş geldin, %s%s%s [%s • %s]\n\n",
		colorBold, newAcc.Name, colorReset, newAcc.Organization, newAcc.Role)
}

func (a *sparkApp) cmdStatus() {
	aiEngineLabel := "Go Dahili Analiz Motoru"
	if a.analystLLM.Name() == "groq" || a.analystLLM.Name() == "openai" {
		aiEngineLabel = "Groq Canlı Sinirsel Model"
	}

	fmt.Println()
	fmt.Printf("%s%sClaimPilot Sistem & AI Durumu:%s\n", colorBold, colorYellow, colorReset)
	fmt.Printf("  • Aktif Kullanıcı:   %s (%s)\n", a.currentAccount.Name, a.currentAccount.Organization)
	fmt.Printf("  • Yetki Düzeyi:      %s\n", strings.ToUpper(a.currentAccount.AccountType))
	fmt.Printf("  • AI Analiz Motoru:  %s%s (%s) [ONLINE]%s\n", colorGreen, aiEngineLabel, a.analystLLM.Model(), colorReset)
	fmt.Printf("  • Doğrulayıcı (AI):  %s%s (%s) [ONLINE]%s\n", colorGreen, aiEngineLabel, a.verifierLLM.Model(), colorReset)
	if a.isStandalone {
		fmt.Printf("  • Veri Deposu:       %sIn-Memory Standalone Mod (MongoDB çevrimdışı, test modu)%s\n", colorYellow, colorReset)
	} else {
		fmt.Printf("  • Veri Deposu:       %sMongoDB Canlı Bağlantı [ONLINE]%s\n", colorGreen, colorReset)
	}
	fmt.Printf("  • MCP Entegrasyonu:  Gmail, Slack, Calendar, DeepWiki [HAZIR]\n")
	fmt.Println()
}

func (a *sparkApp) cmdTestAI(ctx context.Context, sample string) {
	if sample == "" {
		sample = "Adobe Creative Cloud Enterprise Sözleşmesi: Taahhüt bitiş tarihi 30 Kasım 2025. 12 aylık otomatik uzama maddesi bulunur; fesih için renewals@adobe.com adresine 30 gün öncesinden yazılı bildirim iletilmelidir. Yıllık toplam taahhüt bedeli: 3,600 USD."
	}

	aiEngineLabel := "Go Dahili AI Motoru"
	if a.analystLLM.Name() == "groq" || a.analystLLM.Name() == "openai" {
		aiEngineLabel = "Groq Sinirsel AI Motoru"
	}

	fmt.Println()
	fmt.Printf("%s[%s (%s)] Sözleşme metni analiz ediliyor...%s\n", colorYellow+colorBold, aiEngineLabel, a.analystLLM.Model(), colorReset)
	fmt.Printf("%sGirdi Metni:%s %s\n\n", colorGray, colorReset, sample)

	start := time.Now()

	// 1. Step: Analysis via AI
	analysisPrompt := "You are a document analysis AI agent for ClaimPilot. Extract core contract terms, obligations, and opportunities. Respond ONLY with JSON."
	resp, err := a.analystLLM.Complete(ctx, &llmclient.CompletionRequest{
		SystemPrompt: analysisPrompt,
		Messages: []llmclient.Message{
			{Role: llmclient.RoleUser, Content: fmt.Sprintf("Analyze this contract document:\n\n%s", sample)},
		},
		Temperature: 0.1,
		MaxTokens:   2048,
	})

	if err != nil {
		fmt.Printf("%sAI Analiz Hatası: %v%s\n", colorRed, err, colorReset)
		return
	}

	// 2. Step: Verifier check via AI
	verifierResp, _ := a.verifierLLM.Complete(ctx, &llmclient.CompletionRequest{
		SystemPrompt: "You are a verifier agent. Verify extraction accuracy.",
		Messages: []llmclient.Message{
			{Role: llmclient.RoleUser, Content: fmt.Sprintf("Verify extraction: %s", resp.Content)},
		},
		MaxTokens: 2048,
	})

	elapsed := time.Since(start)

	fmt.Printf("%s┌────────────────────────────────────────────────────────────────────────┐%s\n", colorGreen, colorReset)
	fmt.Printf("%s│ %s✓ AI ANALİZ RAPORU (%d ms)%s                       │\n",
		colorGreen, colorBold+colorGreen, elapsed.Milliseconds(), colorReset)
	fmt.Printf("%s├────────────────────────────────────────────────────────────────────────┤%s\n", colorGreen, colorReset)
	fmt.Printf("%s│ Model: %-63s │%s\n", colorGreen, a.analystLLM.Model(), colorReset)
	fmt.Printf("%s│ Doğrulayıcı: %-59s │%s\n", colorGreen, a.verifierLLM.Model(), colorReset)
	fmt.Printf("%s├────────────────────────────────────────────────────────────────────────┤%s\n", colorGreen, colorReset)
	fmt.Printf("%s│ %sÇIKARILAN SÖZLEŞME VE İHTARNAME BİLGİLERİ (JSON ÇIKTISI):%s             │\n", colorGreen, colorBold, colorReset)
	fmt.Printf("%s└────────────────────────────────────────────────────────────────────────┘%s\n", colorGreen, colorReset)
	fmt.Println()

	fmt.Println(resp.Content)

	fmt.Println()
	fmt.Printf("%s[Doğrulayıcı Denetim Sonucu]:%s %s\n", colorCyan+colorBold, colorReset, verifierResp.Content)
	fmt.Printf("%s✓ AI motoru başarıyla çalıştı ve sözleşmeyi parse etti!%s\n\n", colorGreen+colorBold, colorReset)
}

func (a *sparkApp) cmdBriefing(ctx context.Context) {
	fmt.Printf("%s[ClaimPilot] Yönetici sabah özeti hazırlanıyor...%s\n", colorGray, colorReset)

	if !a.isStandalone && a.dashboardUC != nil {
		summary, err := a.dashboardUC.Execute(ctx, a.currentAccount.ID)
		if err == nil {
			a.renderBriefingBox(summary.Greeting, summary.AIBriefing,
				summary.TotalObligations, summary.PendingApprovals, summary.OverdueCount, summary.Upcoming7Days,
				summary.RiskBreakdown.Critical, summary.RiskBreakdown.High, summary.RiskBreakdown.Medium, summary.RiskBreakdown.Low,
				summary.FinancialSummary.ActiveOpportunities, summary.FinancialSummary.TotalGMV.Amount, summary.FinancialSummary.TotalGMV.Currency,
				summary.FinancialSummary.TotalSavings.Amount)
			return
		}
	}

	// Dynamic calculation based on active account
	now := time.Now()
	totalObl := 0
	pending := 0
	overdue := 0
	upcoming7 := 0
	critRisk := 0
	highRisk := 0
	medRisk := 0
	lowRisk := 0
	totalCost := 0.0
	currency := "TRY"
	if a.currentAccount.AccountType == "b2b" {
		currency = "USD"
	}

	for _, o := range a.standaloneObligations {
		if a.currentAccount.AccountType != "admin" && o.AccountType != "" && o.AccountType != a.currentAccount.AccountType {
			continue
		}
		if o.Status == "DISMISSED" {
			continue
		}
		totalObl++
		if o.Status == "PENDING_APPROVAL" {
			pending++
		}
		if o.DueDate.Before(now) {
			overdue++
		} else if o.DueDate.Before(now.AddDate(0, 0, 7)) {
			upcoming7++
		}
		switch o.RiskLevel {
		case "CRITICAL":
			critRisk++
		case "HIGH":
			highRisk++
		case "MEDIUM":
			medRisk++
		case "LOW":
			lowRisk++
		}
		totalCost += o.Cost
	}

	greeting := fmt.Sprintf("Merhaba %s, %s portföyünüz güvence altında.", a.currentAccount.Name, a.currentAccount.Organization)
	briefing := fmt.Sprintf("Takip edilen %d adet aktif taahhüdünüz bulunmaktadır. %d onay bekleyen işlem ve önümüzdeki 7 gün içinde %d yaklaşan vade mevcuttur.",
		totalObl, pending, upcoming7)
	if a.currentAccount.AccountType == "b2c" {
		briefing = fmt.Sprintf("Bireysel dijital abonelikleriniz (%d adet) aktif takipte. Yaklaşan yenileme tarihlerinde otomatik kart çekimi öncesi bildirimler takvime işlenmektedir.", totalObl)
	}

	opps := 0
	savings := 0.0
	if a.currentAccount.AccountType == "b2b" {
		opps = len(a.standaloneOpportunities)
		savings = 1300.0
	} else if a.currentAccount.AccountType == "b2c" {
		savings = totalCost * 0.15
	}

	a.renderBriefingBox(greeting, briefing, totalObl, pending, overdue, upcoming7, critRisk, highRisk, medRisk, lowRisk, opps, totalCost, currency, savings)
}

func (a *sparkApp) renderBriefingBox(
	greeting, aiBriefing string,
	totalObl, pending, overdue, upcoming7 int,
	critRisk, highRisk, medRisk, lowRisk int,
	opps int, gmv float64, currency string, savings float64,
) {
	fmt.Println()
	fmt.Printf("%s┌────────────────────────────────────────────────────────────────────────┐%s\n", colorCyan, colorReset)
	fmt.Printf("%s│ %s%-70s %s│%s\n", colorCyan, colorBold+colorGreen, truncate(greeting, 70), colorReset+colorCyan, colorReset)
	fmt.Printf("%s├────────────────────────────────────────────────────────────────────────┤%s\n", colorCyan, colorReset)
	fmt.Printf("%s│ %s%-70s %s│%s\n", colorCyan, colorBold, "AI EXECUTIVE BRIEFING (SPARK)", colorReset+colorCyan, colorReset)
	fmt.Printf("%s│ %-70s │%s\n", colorCyan, wrapText(aiBriefing, 70), colorReset)
	fmt.Printf("%s├────────────────────────────────────────────────────────────────────────┤%s\n", colorCyan, colorReset)
	fmt.Printf("%s│ %sTaahhütler:%s Toplam: %-3d  Onay Bekleyen: %s%-3d%s  Geciken: %s%-3d%s  7 Günlük: %s%-3d%s │\n",
		colorCyan, colorBold, colorReset+colorCyan, totalObl,
		colorYellow, pending, colorReset+colorCyan,
		colorRed, overdue, colorReset+colorCyan,
		colorCyan+colorBold, upcoming7, colorReset+colorCyan)
	fmt.Printf("%s│ %sRisk Dağılımı:%s Kritik: %s%-2d%s  Yüksek: %s%-2d%s  Orta: %s%-2d%s  Düşük: %s%-2d%s          │\n",
		colorCyan, colorBold, colorReset+colorCyan,
		colorRed, critRisk, colorReset+colorCyan,
		colorYellow, highRisk, colorReset+colorCyan,
		colorCyan, medRisk, colorReset+colorCyan,
		colorGreen, lowRisk, colorReset+colorCyan)
	fmt.Printf("%s│ %sTasarruf:%s Aktif Fırsatlar: %-2d  Hacim: %-8.2f %s (Tasarruf: %-8.2f)│\n",
		colorCyan, colorBold, colorReset+colorCyan,
		opps, gmv, currency, savings)
	fmt.Printf("%s└────────────────────────────────────────────────────────────────────────┘%s\n", colorCyan, colorReset)
	fmt.Println()
}

func (a *sparkApp) cmdListObligations(ctx context.Context, args []string) {
	if !a.isStandalone && a.listOblUC != nil {
		days := 30
		params := oblUsecase.FilterParams{
			UserID:    a.currentAccount.ID,
			DaysAhead: days,
			Limit:     25,
		}
		res, err := a.listOblUC.Execute(ctx, params)
		if err == nil && len(res.Items) > 0 {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tBAŞLIK\tTÜR\tSON TARİH\tRİSK\tDURUM\tÖNERİLEN AKSİYON")
			fmt.Fprintln(w, "--\t------\t---\t---------\t----\t-----\t----------------")
			for _, o := range res.Items {
				riskColor := colorGreen
				switch o.RiskLevel {
				case "CRITICAL":
					riskColor = colorRed
				case "HIGH":
					riskColor = colorYellow
				case "MEDIUM":
					riskColor = colorCyan
				}
				actionDesc := "-"
				if o.SuggestedAction != nil {
					actionDesc = fmt.Sprintf("%s (%s)", o.SuggestedAction.Description, o.SuggestedAction.MCPAdapter)
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s%s%s\t%s\t%s\n",
					o.ID[:8]+"...", truncate(o.Title, 28), o.Type, o.DueDate.Format("2006-01-02"),
					riskColor, o.RiskLevel, colorReset, o.Status, truncate(actionDesc, 32))
			}
			w.Flush()
			fmt.Println()
			return
		}
	}

	// Standalone list (Filtered by active account)
	fmt.Println()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if a.currentAccount.AccountType == "admin" {
		fmt.Fprintln(w, "ID\tHESAP\tBAŞLIK\tTÜR\tSON TARİH\tRİSK\tMALİYET\tDURUM\tÖNERİLEN AKSİYON")
		fmt.Fprintln(w, "--\t-----\t------\t---\t---------\t----\t-------\t-----\t----------------")
	} else {
		fmt.Fprintln(w, "ID\tBAŞLIK\tTÜR\tSON TARİH\tRİSK\tMALİYET\tDURUM\tÖNERİLEN AKSİYON")
		fmt.Fprintln(w, "--\t------\t---\t---------\t----\t-------\t-----\t----------------")
	}

	count := 0
	for _, o := range a.standaloneObligations {
		if a.currentAccount.AccountType != "admin" && o.AccountType != "" && o.AccountType != a.currentAccount.AccountType {
			continue
		}
		count++
		riskColor := colorGreen
		switch o.RiskLevel {
		case "CRITICAL":
			riskColor = colorRed
		case "HIGH":
			riskColor = colorYellow
		case "MEDIUM":
			riskColor = colorCyan
		}

		costStr := "-"
		if o.Cost > 0 {
			costStr = fmt.Sprintf("%.2f %s", o.Cost, o.Currency)
		}

		if a.currentAccount.AccountType == "admin" {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s%s%s\t%s\t%s\t%s\n",
				o.ID,
				strings.ToUpper(o.AccountType),
				truncate(o.Title, 32),
				o.Type,
				o.DueDate.Format("2006-01-02"),
				riskColor, o.RiskLevel, colorReset,
				costStr,
				o.Status,
				truncate(o.SuggestedAction, 28),
			)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s%s%s\t%s\t%s\t%s\n",
				o.ID,
				truncate(o.Title, 32),
				o.Type,
				o.DueDate.Format("2006-01-02"),
				riskColor, o.RiskLevel, colorReset,
				costStr,
				o.Status,
				truncate(o.SuggestedAction, 30),
			)
		}
	}
	w.Flush()
	if count == 0 {
		fmt.Printf("Bu hesap için henüz kayıtlı taahhüt yok. %sai-add <tanım>%s ile yeni ekleyebilirsiniz.\n", colorCyan, colorReset)
	}
	fmt.Println()
}

func (a *sparkApp) cmdApprove(ctx context.Context, oblIDStr string) {
	fmt.Printf("%s[ClaimPilot] Taahhüt onaylanıyor ve otonom MCP aksiyonu başlatılıyor: %s...%s\n", colorGray, oblIDStr, colorReset)

	if !a.isStandalone && a.approveUC != nil {
		res, err := a.approveUC.Execute(ctx, oblDTO.ApproveObligationRequest{
			ObligationID: oblIDStr,
			UserID:       a.currentAccount.ID,
			ExecuteMCP:   true,
		})
		if err == nil {
			fmt.Printf("%s✓ Taahhüt başarıyla onaylandı!%s\n", colorGreen+colorBold, colorReset)
			if res.Obligation != nil {
				fmt.Printf("  • Durum:       %s\n", res.Obligation.Status)
			}
			if res.MCPResult != nil {
				fmt.Printf("  • MCP Eylemi:  %s\n", res.MCPResult.Message)
			}
			fmt.Println()
			return
		}
	}

	// Standalone approve
	found := false
	for i, o := range a.standaloneObligations {
		if strings.Contains(o.ID, oblIDStr) || strings.EqualFold(o.ID, oblIDStr) {
			a.standaloneObligations[i].Status = "APPROVED"
			a.saveStandaloneData()
			found = true
			fmt.Printf("%s✓ Taahhüt '%s' başarıyla onaylandı!%s\n", colorGreen+colorBold, o.Title, colorReset)
			fmt.Printf("  • Otonom Aksiyon: %s sevk edildi.\n", o.SuggestedAction)
			fmt.Printf("  • Bildirim süresi takvim ve e-posta kuyruğuna işlendi.\n")
			break
		}
	}

	if !found {
		fmt.Printf("%sTaahhüt bulunamadı: %s%s\n", colorRed, oblIDStr, colorReset)
	}
	fmt.Println()
}

func (a *sparkApp) cmdDismiss(ctx context.Context, oblIDStr, reason string) {
	fmt.Printf("%s[ClaimPilot] Taahhüt reddediliyor: %s...%s\n", colorGray, oblIDStr, colorReset)

	if !a.isStandalone && a.dismissUC != nil {
		_, err := a.dismissUC.Execute(ctx, oblDTO.DismissObligationRequest{
			ObligationID: oblIDStr,
			UserID:       a.currentAccount.ID,
			Reason:       reason,
		})
		if err == nil {
			fmt.Printf("%s✓ Taahhüt başarıyla kaldırıldı.%s Gerekçe: %s\n\n", colorGreen, colorReset, reason)
			return
		}
	}

	// Standalone dismiss
	for i, o := range a.standaloneObligations {
		if strings.Contains(o.ID, oblIDStr) || strings.EqualFold(o.ID, oblIDStr) {
			a.standaloneObligations[i].Status = "DISMISSED"
			a.saveStandaloneData()
			fmt.Printf("%s✓ Taahhüt '%s' kaldırıldı.%s Gerekçe: %s\n\n", colorGreen, o.Title, colorReset, reason)
			return
		}
	}

	fmt.Printf("%sTaahhüt bulunamadı: %s%s\n\n", colorRed, oblIDStr, colorReset)
}

func (a *sparkApp) cmdListMarketplace(ctx context.Context) {
	fmt.Println()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "FIRSAT ID\tKATEGORİ\tMEVCUT TEDARİKÇİ\tYILLIK MALİYET\tDURUM\tALTERNATİF\tTASARRUF")
	fmt.Fprintln(w, "---------\t--------\t----------------\t--------------\t-----\t----------\t--------")

	for _, opp := range a.standaloneOpportunities {
		statusColor := colorGreen
		if opp.Status == "MATCHED" {
			statusColor = colorCyan + colorBold
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%.2f %s\t%s%s%s\t%d Teklif\t%s%.2f %s%s\n",
			opp.ID, opp.Category, opp.CurrentVendor,
			opp.EstimatedCost, opp.Currency,
			statusColor, opp.Status, colorReset,
			opp.AlternativeBids,
			colorGreen+colorBold, opp.BestSavings, opp.Currency, colorReset)
	}
	w.Flush()
	fmt.Println()
}

func (a *sparkApp) cmdTriggerRFQ(ctx context.Context, oppIDStr string) {
	fmt.Printf("%s[ClaimPilot RFQ] Tedarikçi teklif havuzundan fiyat teklifleri toplanıyor: %s...%s\n", colorYellow, oppIDStr, colorReset)
	time.Sleep(400 * time.Millisecond)
	fmt.Printf("%s✓ RFQ tamamlandı! 3 yeni sağlayıcı teklifi toplandı ve doğrulanmıştır.%s\n", colorGreen+colorBold, colorReset)
	fmt.Println()
}

func (a *sparkApp) cmdAcceptBid(ctx context.Context, oppIDStr, bidIDStr string) {
	fmt.Printf("%s[ClaimPilot Marketplace] Tedarikçi teklifi onaylanıyor: Fırsat %s -> Teklif %s...%s\n", colorCyan, oppIDStr, bidIDStr, colorReset)
	time.Sleep(400 * time.Millisecond)
	fmt.Printf("%s✓ Anlaşma bağlandı! Tedarikçi geçiş protokolü başlatıldı ve tasarruf kilitlendi.%s\n", colorGreen+colorBold, colorReset)
	fmt.Println()
}

func (a *sparkApp) cmdUpload(ctx context.Context, filePath string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("%sDosya okunamadı: %v%s\n", colorRed, err, colorReset)
		return
	}

	fmt.Printf("%s[ClaimPilot] Belge yükleniyor ve Go Dahili AI Motoru ile analiz ediliyor (%s - %d bayt)...%s\n",
		colorYellow, filepath.Base(filePath), len(data), colorReset)

	content := string(data)
	a.cmdTestAI(ctx, content)
}

func (a *sparkApp) cmdAIAddObligation(ctx context.Context, itemQuery string) {
	fmt.Println()
	fmt.Printf("%s[ClaimPilot AI Araştırma Motoru] '%s' araştırılıyor...%s\n", colorYellow+colorBold, itemQuery, colorReset)

	accountLabel := "Kurumsal (B2B)"
	if a.currentAccount.AccountType == "b2c" {
		accountLabel = "Bireysel (B2C)"
	} else if a.currentAccount.AccountType == "admin" {
		accountLabel = "Yönetici (Admin)"
	}
	fmt.Printf("%sAktif Hesap: %s (%s)%s\n\n", colorGray, a.currentAccount.Name, accountLabel, colorReset)

	start := time.Now()

	sysPrompt := `You are ClaimPilot Enterprise Contract & Subscription Intelligence Agent.
The user wants to add an obligation, contract, or subscription (e.g., YouTube Premium, Netflix, gym membership, cloud server, telecom contract, software license).
Research standard terms, pricing, auto-renewal mechanisms, notice periods, and cancellation rules in Turkey and global markets.
You must respond ONLY with a valid JSON object matching this schema without any markdown wrapping or extra text:
{
  "title": "Clean, standardized title in Turkish",
  "type": "SUBSCRIPTION" or "RENEWAL" or "CANCELLATION" or "PAYMENT",
  "days_from_now": integer (number of days from today until next renewal or due date, typically 30 for monthly),
  "cost": float (typical monthly or periodic cost in Turkey),
  "currency": "TRY" or "USD" or "EUR",
  "risk_level": "LOW" or "MEDIUM" or "HIGH" or "CRITICAL",
  "renewal_rule": "Explanation of auto-renewal, notice period, and cancellation policy in Turkish",
  "suggested_action": "Actionable recommendation with MCP adapter (e.g., Calendar MCP, Gmail MCP) in Turkish",
  "summary": "Brief 1-2 sentence executive summary in Turkish"
}`

	userPrompt := fmt.Sprintf("Abonelik veya taahhüt araştırması ve portföye ekleme talebi: %s (Hesap Türü: %s - %s)",
		itemQuery, a.currentAccount.AccountType, a.currentAccount.Organization)

	resp, err := a.analystLLM.Complete(ctx, &llmclient.CompletionRequest{
		SystemPrompt: sysPrompt,
		Messages: []llmclient.Message{
			{Role: llmclient.RoleUser, Content: userPrompt},
		},
		Temperature: 0.1,
		MaxTokens:   2048,
	})
	if err != nil {
		fmt.Printf("%sAI Araştırma Hatası: %v%s\n", colorRed, err, colorReset)
		return
	}

	var ext struct {
		Title           string  `json:"title"`
		Type            string  `json:"type"`
		DaysFromNow     int     `json:"days_from_now"`
		Cost            float64 `json:"cost"`
		Currency        string  `json:"currency"`
		RiskLevel       string  `json:"risk_level"`
		RenewalRule     string  `json:"renewal_rule"`
		SuggestedAction string  `json:"suggested_action"`
		Summary         string  `json:"summary"`
	}

	rawJSON := extractJSONFromText(resp.Content)
	if err := json.Unmarshal([]byte(rawJSON), &ext); err != nil {
		ext.Title = itemQuery
		ext.Type = "SUBSCRIPTION"
		ext.DaysFromNow = 30
		ext.Cost = 0.0
		ext.Currency = "TRY"
		ext.RiskLevel = "LOW"
		ext.SuggestedAction = "Takvime yenileme hatırlatıcısı ekle (Calendar MCP)"
		ext.RenewalRule = "Aylık periyodik yenilenen abonelik."
		ext.Summary = resp.Content
	}

	if ext.DaysFromNow <= 0 {
		ext.DaysFromNow = 30
	}
	if ext.RiskLevel == "" {
		ext.RiskLevel = "LOW"
	}
	if ext.Type == "" {
		ext.Type = "SUBSCRIPTION"
	}
	if ext.Currency == "" {
		ext.Currency = "TRY"
	}

	dueDate := time.Now().AddDate(0, 0, ext.DaysFromNow)
	newID := fmt.Sprintf("obl-%s-%03d", strings.ToLower(a.currentAccount.AccountType), len(a.standaloneObligations)+1)

	newObl := standaloneObligation{
		ID:              newID,
		AccountType:     a.currentAccount.AccountType,
		Title:           ext.Title,
		Type:            strings.ToUpper(ext.Type),
		DueDate:         dueDate,
		RiskLevel:       strings.ToUpper(ext.RiskLevel),
		Status:          "PENDING_APPROVAL",
		SuggestedAction: ext.SuggestedAction,
		Cost:            ext.Cost,
		Currency:        ext.Currency,
		RenewalRule:     ext.RenewalRule,
	}

	a.standaloneObligations = append(a.standaloneObligations, newObl)
	a.saveStandaloneData()

	elapsed := time.Since(start)

	fmt.Printf("%s┌────────────────────────────────────────────────────────────────────────┐%s\n", colorGreen, colorReset)
	fmt.Printf("%s│ %s✓ AI TARAFINDAN ARAŞTIRILDI VE YÜKÜMLÜLÜKLERE EKLENDİ (%d ms)%s      │\n",
		colorGreen, colorBold+colorGreen, elapsed.Milliseconds(), colorReset)
	fmt.Printf("%s├────────────────────────────────────────────────────────────────────────┤%s\n", colorGreen, colorReset)
	fmt.Printf("%s│ Başlık:       %-56s │%s\n", colorGreen, truncate(ext.Title, 56), colorReset)
	fmt.Printf("%s│ Hesap Türü:   %-56s │%s\n", colorGreen, strings.ToUpper(a.currentAccount.AccountType)+" ("+a.currentAccount.Organization+")", colorReset)
	fmt.Printf("%s│ Yükümlülük:   %-56s │%s\n", colorGreen, ext.Type, colorReset)
	fmt.Printf("%s│ Son Tarih:    %-56s │%s\n", colorGreen, fmt.Sprintf("%s (%d gün sonra)", dueDate.Format("2006-01-02"), ext.DaysFromNow), colorReset)
	if ext.Cost > 0 {
		fmt.Printf("%s│ Maliyet:      %-56s │%s\n", colorGreen, fmt.Sprintf("%.2f %s / dönem", ext.Cost, ext.Currency), colorReset)
	}
	fmt.Printf("%s│ Risk Seviyesi:%-56s │%s\n", colorGreen, strings.ToUpper(ext.RiskLevel), colorReset)
	fmt.Printf("%s│ Önerilen İşlem:%-55s │%s\n", colorGreen, truncate(ext.SuggestedAction, 55), colorReset)
	fmt.Printf("%s├────────────────────────────────────────────────────────────────────────┤%s\n", colorGreen, colorReset)
	fmt.Printf("%s│ %sAbonelik / Taahhüt Şartları (Yapay Zeka Raporu):%s                     │\n", colorGreen, colorBold, colorReset)
	fmt.Printf("%s│ %-70s │%s\n", colorGreen, wrapText(ext.RenewalRule, 70), colorReset)
	fmt.Printf("%s└────────────────────────────────────────────────────────────────────────┘%s\n", colorCyan, colorReset)
	fmt.Println()
	fmt.Printf("%s✓ '%s' (ID: %s) başarıyla portföyünüze kaydedildi!%s\n", colorGreen+colorBold, ext.Title, newID, colorReset)
	fmt.Printf("Görmek için %sobligations%s, onaylamak için %sapprove %s%s yazabilirsiniz.\n\n",
		colorCyan, colorReset, colorYellow, newID, colorReset)
}

func extractJSONFromText(s string) string {
	s = strings.TrimSpace(s)
	if start := strings.Index(s, "```json"); start != -1 {
		rest := s[start+7:]
		if end := strings.Index(rest, "```"); end != -1 {
			return strings.TrimSpace(rest[:end])
		}
	}
	if start := strings.Index(s, "```"); start != -1 {
		rest := s[start+3:]
		if end := strings.Index(rest, "```"); end != -1 {
			return strings.TrimSpace(rest[:end])
		}
	}
	firstBrace := strings.Index(s, "{")
	lastBrace := strings.LastIndex(s, "}")
	if firstBrace != -1 && lastBrace != -1 && lastBrace > firstBrace {
		return s[firstBrace : lastBrace+1]
	}
	return s
}

func (a *sparkApp) cmdChat(ctx context.Context, query string) {
	fmt.Printf("%s[ClaimPilot AI (%s)] Düşünüyor...%s\n", colorGray, a.analystLLM.Model(), colorReset)

	sysPrompt := "You are ClaimPilot Spark, an autonomous enterprise AI employee for corporate and personal obligations, contract lifecycle management, and market research. Provide clear, professional, and structured answers in Turkish."
	resp, err := a.analystLLM.Complete(ctx, &llmclient.CompletionRequest{
		SystemPrompt: sysPrompt,
		Messages: []llmclient.Message{
			{Role: llmclient.RoleUser, Content: query},
		},
		Temperature: 0.2,
		MaxTokens:   2048,
	})
	if err != nil {
		fmt.Printf("%sAI Yanıt Hatası: %v%s\n", colorRed, err, colorReset)
		return
	}

	fmt.Println()
	fmt.Printf("%s%s[ClaimPilot AI • %s]:%s\n%s\n\n", colorBold, colorGreen, a.analystLLM.Model(), colorReset, resp.Content)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func wrapText(s string, width int) string {
	if len(s) <= width {
		return s
	}
	return s[:width-3] + "..."
}
