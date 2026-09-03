package main

import (
	"bufio"
	"bytes"
	"context"
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
	docDTO "github.com/masterfabric-go/masterfabric/internal/application/document/dto"
	docUsecase "github.com/masterfabric-go/masterfabric/internal/application/document/usecase"
	mktDTO "github.com/masterfabric-go/masterfabric/internal/application/marketplace/dto"
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
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

type sparkApp struct {
	cfg         *config.Config
	userID      uuid.UUID
	dashboardUC *dashUsecase.GetDashboardSummaryUseCase
	listOblUC   *oblUsecase.ListObligationsUseCase
	approveUC   *oblUsecase.ApproveObligationUseCase
	dismissUC   *oblUsecase.DismissObligationUseCase
	rfqUC       *mktUsecase.TriggerRFQUseCase
	acceptBidUC *mktUsecase.AcceptBidUseCase
	metricsUC   *mktUsecase.GetMetricsUseCase
	listOppUC   *mktUsecase.ListOpportunitiesUseCase
	uploadUC    *docUsecase.UploadDocumentUseCase
	analystLLM  llmclient.Provider
}

func main() {
	briefingFlag := flag.Bool("briefing", false, "Print Good Morning briefing and exit")
	obligationsFlag := flag.Bool("obligations", false, "List upcoming obligations and exit")
	uploadFlag := flag.String("upload", "", "Upload and analyze a document file")
	flag.Parse()

	app, err := initSparkApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError initializing ClaimPilot Spark:%s %v\n", colorRed, colorReset, err)
		os.Exit(1)
	}

	ctx := context.Background()

	// Handle one-shot CLI flags
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

	// Interactive REPL mode
	app.runREPL()
}

func initSparkApp() (*sparkApp, error) {
	cfg := config.Load()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil)) // quiet CLI logger

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mongoClient, err := database.NewMongoClient(ctx, cfg.MongoDB)
	if err != nil {
		return nil, fmt.Errorf("connect to MongoDB (%s): %w", cfg.MongoDB.URI, err)
	}

	mongoDB := database.GetDatabase(mongoClient, cfg.MongoDB)

	// Repositories
	docRepo := mongoDoc.NewMongoRepository(mongoDB)
	oblRepo := mongoObl.NewMongoRepository(mongoDB)
	mktRepo := mongoMkt.NewMongoRepository(mongoDB)
	auditRepo := mongoAudit.NewMongoAgentAuditRepository(mongoDB)

	// MCP Registry
	mcpReg := mcp.NewRegistry()
	mcpReg.Register(mcpGmail.New(logger))
	mcpReg.Register(mcpCalendar.New(logger))
	mcpReg.Register(mcpSlack.New(logger))

	// Event bus
	eventBus := events.NewInProcessBus(logger, 64)

	// LLM Providers (read strictly from config)
	analystProvider, _ := llmclient.NewProvider(cfg.LLMAnalyst)
	verifierProvider, _ := llmclient.NewProvider(cfg.LLMVerifier)

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

	// Document storage & parser for upload use case
	storageService, _ := docStorage.NewLocalStorageService("uploads/documents")
	parserService := docParser.NewDefaultParser()

	uploadUC := docUsecase.NewUploadDocumentUseCase(docUsecase.UploadConfig{
		DocRepo:      docRepo,
		Storage:      storageService,
		Parser:       parserService,
		Orchestrator: orch,
		EventBus:     eventBus,
		Logger:       logger,
	})

	// Use Cases
	dashUC := dashUsecase.NewGetDashboardSummaryUseCase(dashUsecase.Config{
		OblRepo:  oblRepo,
		MktRepo:  mktRepo,
		MCPReg:   mcpReg,
		Analyst:  analystProvider,
		Verifier: verifierProvider,
		Logger:   logger,
	})

	listOblUC := oblUsecase.NewListObligationsUseCase(oblRepo)
	approveUC := oblUsecase.NewApproveObligationUseCase(oblUsecase.ApproveConfig{
		OblRepo:   oblRepo,
		AuditRepo: auditRepo,
		MCPReg:    mcpReg,
		EventBus:  eventBus,
		Logger:    logger,
	})
	dismissUC := oblUsecase.NewDismissObligationUseCase(oblRepo, auditRepo, logger)

	rfqEngine := mktAgent.NewRFQEngine(mktRepo, verifierAgent, logger)
	rfqUC := mktUsecase.NewTriggerRFQUseCase(rfqEngine, mktRepo, eventBus, logger)
	acceptBidUC := mktUsecase.NewAcceptBidUseCase(mktUsecase.AcceptBidConfig{
		MktRepo:   mktRepo,
		AuditRepo: auditRepo,
		MCPReg:    mcpReg,
		EventBus:  eventBus,
		Logger:    logger,
	})
	metricsUC := mktUsecase.NewGetMetricsUseCase(mktRepo)
	listOppUC := mktUsecase.NewListOpportunitiesUseCase(mktRepo)

	demoUserID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	return &sparkApp{
		cfg:         cfg,
		userID:      demoUserID,
		dashboardUC: dashUC,
		listOblUC:   listOblUC,
		approveUC:   approveUC,
		dismissUC:   dismissUC,
		rfqUC:       rfqUC,
		acceptBidUC: acceptBidUC,
		metricsUC:   metricsUC,
		listOppUC:   listOppUC,
		uploadUC:    uploadUC,
		analystLLM:  analystProvider,
	}, nil
}

func (a *sparkApp) runREPL() {
	printBanner()

	scanner := bufio.NewScanner(os.Stdin)
	ctx := context.Background()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Printf("\n%sClaimPilot Spark session ended. Goodbye!%s\n", colorGray, colorReset)
		os.Exit(0)
	}()

	for {
		fmt.Printf("%sclaimpilot%s> ", colorCyan+colorBold, colorReset)
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
		case "briefing", "morning", "summary":
			a.cmdBriefing(ctx)
		case "obligations", "list":
			a.cmdListObligations(ctx, args)
		case "approve":
			if len(args) == 0 {
				fmt.Printf("%sUsage: approve <obligation_id>%s\n", colorYellow, colorReset)
			} else {
				a.cmdApprove(ctx, args[0])
			}
		case "dismiss":
			if len(args) == 0 {
				fmt.Printf("%sUsage: dismiss <obligation_id> [reason]%s\n", colorYellow, colorReset)
			} else {
				reason := "Dismissed via Spark CLI"
				if len(args) > 1 {
					reason = strings.Join(args[1:], " ")
				}
				a.cmdDismiss(ctx, args[0], reason)
			}
		case "marketplace", "deals":
			a.cmdListMarketplace(ctx)
		case "rfq":
			if len(args) == 0 {
				fmt.Printf("%sUsage: rfq <opportunity_id>%s\n", colorYellow, colorReset)
			} else {
				a.cmdTriggerRFQ(ctx, args[0])
			}
		case "accept":
			if len(args) < 2 {
				fmt.Printf("%sUsage: accept <opportunity_id> <bid_id>%s\n", colorYellow, colorReset)
			} else {
				a.cmdAcceptBid(ctx, args[0], args[1])
			}
		case "upload":
			if len(args) == 0 {
				fmt.Printf("%sUsage: upload <file_path>%s\n", colorYellow, colorReset)
			} else {
				a.cmdUpload(ctx, args[0])
			}
		case "chat", "ask":
			if len(args) == 0 {
				fmt.Printf("%sUsage: chat <your question or obligation query>%s\n", colorYellow, colorReset)
			} else {
				a.cmdChat(ctx, strings.Join(args, " "))
			}
		case "help":
			printHelp()
		case "exit", "quit", "q":
			fmt.Printf("%sClaimPilot Spark session ended. Goodbye!%s\n", colorGray, colorReset)
			return
		default:
			// Treat arbitrary input as a chat query to the AI employee
			a.cmdChat(ctx, line)
		}
	}
}

func printBanner() {
	fmt.Println()
	fmt.Printf("%s%s", colorCyan, colorBold)
	fmt.Println(`   ____ _       _           ____  _ _       _   `)
	fmt.Println(`  / ___| | __ _(_)_ __ ___ |  _ \(_) | ___ | |_ `)
	fmt.Println(` | |   | |/ _`+"`"+` | | '_ `+"`"+` _ \| |_) | | |/ _ \| __|`)
	fmt.Println(` | |___| | (_| | | | | | | |  __/| | | (_) | |_ `)
	fmt.Println(`  \____|_|\__,_|_|_| |_| |_|_|   |_|_|\___/ \__|`)
	fmt.Printf("%s", colorReset)
	fmt.Printf("%s AI Employee for Corporate & Personal Obligations • Spark CLI v0.1.0%s\n", colorGray, colorReset)
	fmt.Printf(" Type %shelp%s for commands, or %sbriefing%s for your daily executive summary.\n\n",
		colorBold+colorYellow, colorReset, colorBold+colorGreen, colorReset)
}

func printHelp() {
	fmt.Println()
	fmt.Printf("%s%sAvailable Commands:%s\n", colorBold, colorYellow, colorReset)
	fmt.Printf("  %s%-28s%s %s\n", colorCyan, "briefing / morning", colorReset, "View daily executive briefing and risk metrics")
	fmt.Printf("  %s%-28s%s %s\n", colorCyan, "obligations / list", colorReset, "List all obligations and upcoming deadlines")
	fmt.Printf("  %s%-28s%s %s\n", colorCyan, "approve <id>", colorReset, "Approve obligation and dispatch autonomous MCP action")
	fmt.Printf("  %s%-28s%s %s\n", colorCyan, "dismiss <id> [reason]", colorReset, "Dismiss an obligation")
	fmt.Printf("  %s%-28s%s %s\n", colorCyan, "marketplace / deals", colorReset, "List opportunities and alternative vendor bids")
	fmt.Printf("  %s%-28s%s %s\n", colorCyan, "rfq <opportunity_id>", colorReset, "Trigger automated RFQ supplier bid collection")
	fmt.Printf("  %s%-28s%s %s\n", colorCyan, "accept <opp_id> <bid_id>", colorReset, "Accept vendor bid, execute transaction & take-rate")
	fmt.Printf("  %s%-28s%s %s\n", colorCyan, "upload <file_path>", colorReset, "Upload and autonomously analyze a document")
	fmt.Printf("  %s%-28s%s %s\n", colorCyan, "chat <prompt>", colorReset, "Ask the AI Agent anything about contracts & obligations")
	fmt.Printf("  %s%-28s%s %s\n", colorCyan, "exit / quit", colorReset, "Exit the terminal")
	fmt.Println()
}

func (a *sparkApp) cmdBriefing(ctx context.Context) {
	fmt.Printf("%s[ClaimPilot] Synthesizing executive morning briefing...%s\n", colorGray, colorReset)
	summary, err := a.dashboardUC.Execute(ctx, a.userID)
	if err != nil {
		fmt.Printf("%sError fetching briefing: %v%s\n", colorRed, err, colorReset)
		return
	}

	fmt.Println()
	fmt.Printf("%s┌────────────────────────────────────────────────────────────────────────┐%s\n", colorCyan, colorReset)
	fmt.Printf("%s│ %s%-70s %s│%s\n", colorCyan, colorBold+colorGreen, summary.Greeting, colorReset+colorCyan, colorReset)
	fmt.Printf("%s├────────────────────────────────────────────────────────────────────────┤%s\n", colorCyan, colorReset)
	fmt.Printf("%s│ %s%-70s %s│%s\n", colorCyan, colorBold, "AI EXECUTIVE BRIEFING", colorReset+colorCyan, colorReset)
	fmt.Printf("%s│ %-70s │%s\n", colorCyan, wrapText(summary.AIBriefing, 70), colorReset)
	fmt.Printf("%s├────────────────────────────────────────────────────────────────────────┤%s\n", colorCyan, colorReset)
	fmt.Printf("%s│ %sObligations:%s Total: %-3d  Pending: %s%-3d%s  Overdue: %s%-3d%s  Next 7d: %s%-3d%s        │\n",
		colorCyan, colorBold, colorReset+colorCyan, summary.TotalObligations,
		colorYellow, summary.PendingApprovals, colorReset+colorCyan,
		colorRed, summary.OverdueCount, colorReset+colorCyan,
		colorCyan+colorBold, summary.Upcoming7Days, colorReset+colorCyan)
	fmt.Printf("%s│ %sRisk Distribution:%s Critical: %s%-2d%s  High: %s%-2d%s  Medium: %s%-2d%s  Low: %s%-2d%s           │\n",
		colorCyan, colorBold, colorReset+colorCyan,
		colorRed, summary.RiskBreakdown.Critical, colorReset+colorCyan,
		colorYellow, summary.RiskBreakdown.High, colorReset+colorCyan,
		colorCyan, summary.RiskBreakdown.Medium, colorReset+colorCyan,
		colorGreen, summary.RiskBreakdown.Low, colorReset+colorCyan)
	fmt.Printf("%s│ %sFinancials:%s Active Opportunities: %-2d  GMV: %-10.2f %s (Saved: %-10.2f)│\n",
		colorCyan, colorBold, colorReset+colorCyan,
		summary.FinancialSummary.ActiveOpportunities,
		summary.FinancialSummary.TotalGMV.Amount, summary.FinancialSummary.TotalGMV.Currency,
		summary.FinancialSummary.TotalSavings.Amount)
	fmt.Printf("%s└────────────────────────────────────────────────────────────────────────┘%s\n", colorCyan, colorReset)
	fmt.Println()
}

func (a *sparkApp) cmdListObligations(ctx context.Context, args []string) {
	days := 30
	params := oblUsecase.FilterParams{
		UserID:    a.userID,
		DaysAhead: days,
		Limit:     25,
	}

	res, err := a.listOblUC.Execute(ctx, params)
	if err != nil {
		fmt.Printf("%sError listing obligations: %v%s\n", colorRed, err, colorReset)
		return
	}

	if len(res.Items) == 0 {
		fmt.Printf("%sNo obligations found for the selected filter.%s\n", colorGray, colorReset)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tTYPE\tDUE DATE\tRISK\tSTATUS\tSUGGESTED ACTION")
	fmt.Fprintln(w, "--\t-----\t----\t--------\t----\t------\t----------------")

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
			o.ID[:8]+"...",
			truncate(o.Title, 28),
			o.Type,
			o.DueDate.Format("2006-01-02"),
			riskColor, o.RiskLevel, colorReset,
			o.Status,
			truncate(actionDesc, 32),
		)
	}
	_ = w.Flush()
	fmt.Println()
}

func (a *sparkApp) cmdApprove(ctx context.Context, oblIDStr string) {
	fmt.Printf("%sApproving obligation %s and executing MCP action...%s\n", colorGray, oblIDStr, colorReset)
	req := oblDTO.ApproveObligationRequest{
		ObligationID: oblIDStr,
		UserID:       a.userID,
		ExecuteMCP:   true,
	}

	resp, err := a.approveUC.Execute(ctx, req)
	if err != nil {
		fmt.Printf("%sError: %v%s\n", colorRed, err, colorReset)
		return
	}

	fmt.Printf("%s✓ Obligation successfully approved! Status: %s%s\n", colorGreen, resp.Obligation.Status, colorReset)
	if resp.MCPResult != nil {
		fmt.Printf("  %sMCP Action Result:%s %s (External ID: %s)\n",
			colorBold, colorReset, resp.MCPResult.Message, resp.MCPResult.ExternalID)
	}
	fmt.Println()
}

func (a *sparkApp) cmdDismiss(ctx context.Context, oblIDStr, reason string) {
	req := oblDTO.DismissObligationRequest{
		ObligationID: oblIDStr,
		UserID:       a.userID,
		Reason:       reason,
	}

	resp, err := a.dismissUC.Execute(ctx, req)
	if err != nil {
		fmt.Printf("%sError: %v%s\n", colorRed, err, colorReset)
		return
	}

	fmt.Printf("%s✓ Obligation %s dismissed. (Reason: %s)%s\n\n",
		colorGreen, resp.ID, reason, colorReset)
}

func (a *sparkApp) cmdListMarketplace(ctx context.Context) {
	opps, err := a.listOppUC.List(ctx, a.userID, nil, 20, 0)
	if err != nil {
		fmt.Printf("%sError: %v%s\n", colorRed, err, colorReset)
		return
	}

	metrics, _ := a.metricsUC.Execute(ctx, a.userID)

	if metrics != nil && metrics.TotalGMV != nil {
		fmt.Printf("%sPlatform GMV: %.2f %s | Realized Savings: %.2f %s | Completed Deals: %d%s\n\n",
			colorBold+colorYellow, metrics.TotalGMV.Amount, metrics.TotalGMV.Currency,
			metrics.TotalSavings.Amount, metrics.TotalSavings.Currency,
			metrics.CompletedDeals, colorReset)
	}

	if len(opps) == 0 {
		fmt.Printf("%sNo marketplace opportunities available. Run an RFQ to collect bids.%s\n", colorGray, colorReset)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "OPP ID\tCATEGORY\tCURRENT VENDOR\tCOST\tSTATUS\tBIDS\tBEST SAVINGS")
	fmt.Fprintln(w, "------\t--------\t--------------\t----\t------\t----\t------------")

	for _, opp := range opps {
		costStr := "-"
		if opp.CurrentVendorCost != nil {
			costStr = fmt.Sprintf("%.2f %s", opp.CurrentVendorCost.Amount, opp.CurrentVendorCost.Currency)
		}

		bestSavings := "-"
		if len(opp.Bids) > 0 {
			bestBid := opp.Bids[0]
			for _, b := range opp.Bids {
				if b.SavingsRate > bestBid.SavingsRate {
					bestBid = b
				}
			}
			bestSavings = fmt.Sprintf("%.1f%% (%s)", bestBid.SavingsRate, bestBid.Vendor.Name)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\t%s%s%s\n",
			opp.ID[:8]+"...",
			opp.Category,
			truncate(opp.CurrentVendorName, 18),
			costStr,
			opp.Status,
			opp.BidCount,
			colorGreen, bestSavings, colorReset,
		)
	}
	_ = w.Flush()
	fmt.Println()
}

func (a *sparkApp) cmdTriggerRFQ(ctx context.Context, oppIDStr string) {
	fmt.Printf("%sTriggering RFQ supplier discovery for %s...%s\n", colorGray, oppIDStr, colorReset)
	resp, err := a.rfqUC.Execute(ctx, oppIDStr)
	if err != nil {
		fmt.Printf("%sError: %v%s\n", colorRed, err, colorReset)
		return
	}

	fmt.Printf("%s✓ RFQ Completed! %d competitive bids received for %s:%s\n\n",
		colorGreen+colorBold, len(resp.Bids), resp.Category, colorReset)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "BID ID\tVENDOR\tRATING\tPRICE\tSAVINGS\tRISK SCORE\tTERMS")
	fmt.Fprintln(w, "------\t------\t------\t-----\t-------\t----------\t-----")

	for _, b := range resp.Bids {
		riskStr := "0.10"
		if b.RiskScore != nil {
			riskStr = fmt.Sprintf("%.2f", *b.RiskScore)
		}
		fmt.Fprintf(w, "%s\t%s\t%.1f★\t%.2f %s\t%s%.1f%%%s\t%s\t%s\n",
			b.ID[:8]+"...",
			b.Vendor.Name,
			b.Vendor.Rating,
			b.Price.Amount, b.Price.Currency,
			colorGreen+colorBold, b.SavingsRate, colorReset,
			riskStr,
			truncate(b.Terms, 32),
		)
	}
	_ = w.Flush()
	fmt.Println()
}

func (a *sparkApp) cmdAcceptBid(ctx context.Context, oppIDStr, bidIDStr string) {
	fmt.Printf("%sAccepting bid and closing transaction...%s\n", colorGray, colorReset)
	req := mktDTO.AcceptBidRequest{
		OpportunityID: oppIDStr,
		BidID:         bidIDStr,
		UserID:        a.userID,
	}

	resp, err := a.acceptBidUC.Execute(ctx, req)
	if err != nil {
		fmt.Printf("%sError: %v%s\n", colorRed, err, colorReset)
		return
	}

	fmt.Printf("%s%s🎉 %s%s\n", colorBold, colorGreen, resp.Message, colorReset)
	if resp.Transaction != nil {
		fmt.Printf("  • Gross Deal Value (GMV): %.2f %s\n", resp.Transaction.Amount.Amount, resp.Transaction.Amount.Currency)
		fmt.Printf("  • Platform Fee (4%%):       %.2f %s\n", resp.Transaction.Commission.Amount, resp.Transaction.Commission.Currency)
		if resp.Transaction.SavingsRealized != nil {
			fmt.Printf("  • Annual Net Savings:     %.2f %s\n", resp.Transaction.SavingsRealized.Amount, resp.Transaction.SavingsRealized.Currency)
		}
	}
	fmt.Println()
}

func (a *sparkApp) cmdUpload(ctx context.Context, filePath string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("%sFailed to read file: %v%s\n", colorRed, err, colorReset)
		return
	}

	fmt.Printf("%sUploading and analyzing document %s (%d bytes)...%s\n", colorGray, filePath, len(data), colorReset)

	req := docDTO.UploadDocumentRequest{
		FileName:    filepath.Base(filePath),
		FileContent: bytes.NewReader(data),
		UserID:      a.userID,
	}

	docResp, err := a.uploadUC.Execute(ctx, req)
	if err != nil {
		fmt.Printf("%sDocument upload error: %v%s\n", colorRed, err, colorReset)
		return
	}

	fmt.Printf("%s✓ Document successfully uploaded and processed!%s\n", colorGreen+colorBold, colorReset)
	fmt.Printf("  • Document ID: %s\n", docResp.ID)
	fmt.Printf("  • Type:        %s\n", docResp.FileType)
	fmt.Printf("  • Status:      %s\n", docResp.Status)
	fmt.Printf("  • File Size:   %d bytes\n", docResp.FileSize)
	fmt.Println()
}

func (a *sparkApp) cmdChat(ctx context.Context, query string) {
	if a.analystLLM == nil {
		fmt.Printf("%sLLM provider is not configured. Set ANALYST_LLM_PROVIDER in .env%s\n", colorYellow, colorReset)
		return
	}

	sysPrompt := "You are ClaimPilot Spark, the enterprise AI employee for corporate and personal obligations. Answer the user's question concisely, helpfully, and with a business-oriented tone."

	fmt.Printf("%sClaimPilot AI is thinking...%s\n", colorGray, colorReset)
	resp, err := a.analystLLM.Complete(ctx, &llmclient.CompletionRequest{
		SystemPrompt: sysPrompt,
		Messages: []llmclient.Message{
			{Role: llmclient.RoleUser, Content: query},
		},
		Temperature: 0.3,
		MaxTokens:   500,
	})
	if err != nil {
		fmt.Printf("%sError from AI: %v%s\n", colorRed, err, colorReset)
		return
	}

	fmt.Println()
	fmt.Printf("%s%s[ClaimPilot AI]%s %s\n\n", colorBold, colorGreen, colorReset, resp.Content)
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
