package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/masterfabric-go/masterfabric/internal/agent/analyst"
	"github.com/masterfabric-go/masterfabric/internal/agent/llmclient"
	mktAgent "github.com/masterfabric-go/masterfabric/internal/agent/marketplace"
	"github.com/masterfabric-go/masterfabric/internal/agent/orchestrator"
	"github.com/masterfabric-go/masterfabric/internal/agent/verifier"
	dashboardUsecase "github.com/masterfabric-go/masterfabric/internal/application/dashboard/usecase"
	docUsecase "github.com/masterfabric-go/masterfabric/internal/application/document/usecase"
	mktUsecase "github.com/masterfabric-go/masterfabric/internal/application/marketplace/usecase"
	oblUsecase "github.com/masterfabric-go/masterfabric/internal/application/obligation/usecase"
	docParser "github.com/masterfabric-go/masterfabric/internal/domain/document/parser"
	docStorage "github.com/masterfabric-go/masterfabric/internal/domain/document/storage"
	graphqlServer "github.com/masterfabric-go/masterfabric/internal/infrastructure/graphql"
	dashboardHttpHandler "github.com/masterfabric-go/masterfabric/internal/infrastructure/http/handler/dashboard"
	docHttpHandler "github.com/masterfabric-go/masterfabric/internal/infrastructure/http/handler/document"
	mcpHttpHandler "github.com/masterfabric-go/masterfabric/internal/infrastructure/http/handler/mcp"
	mktHttpHandler "github.com/masterfabric-go/masterfabric/internal/infrastructure/http/handler/marketplace"
	oblHttpHandler "github.com/masterfabric-go/masterfabric/internal/infrastructure/http/handler/obligation"
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
	"github.com/masterfabric-go/masterfabric/internal/shared/logger"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration (all values from env, nothing hardcoded)
	cfg := config.Load()

	// Initialize logger
	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	slog.SetDefault(log)

	log.Info("starting claimpilot-api",
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize MongoDB
	mongoClient, err := database.NewMongoClient(ctx, cfg.MongoDB)
	if err != nil {
		log.Warn("mongodb unavailable, ClaimPilot features disabled", "error", err)
	} else {
		defer func() { _ = mongoClient.Disconnect(context.Background()) }()
		log.Info("connected to mongodb",
			"uri", cfg.MongoDB.URI,
			"database", cfg.MongoDB.Database,
		)
	}

	// Initialize event bus
	eventBus := events.NewInProcessBus(log, 256)
	defer func() { _ = eventBus.Close() }()

	// Initialize LLM providers (config-driven, no hardcoded endpoints)
	analystProvider, err := llmclient.NewProvider(cfg.LLMAnalyst)
	if err != nil {
		log.Warn("analyst LLM provider not configured", "error", err)
	} else {
		log.Info("analyst LLM provider initialized",
			"provider", cfg.LLMAnalyst.Provider,
			"endpoint", cfg.LLMAnalyst.Endpoint,
			"model", cfg.LLMAnalyst.Model,
		)
	}

	verifierProvider, err := llmclient.NewProvider(cfg.LLMVerifier)
	if err != nil {
		log.Warn("verifier LLM provider not configured", "error", err)
	} else {
		log.Info("verifier LLM provider initialized",
			"provider", cfg.LLMVerifier.Provider,
			"endpoint", cfg.LLMVerifier.Endpoint,
			"model", cfg.LLMVerifier.Model,
		)
	}

	// Initialize MCP adapters
	mcpRegistry := mcp.NewRegistry()
	mcpRegistry.Register(mcpGmail.New(log))
	mcpRegistry.Register(mcpCalendar.New(log))
	mcpRegistry.Register(mcpSlack.New(log))
	log.Info("MCP adapters registered", "adapters", mcpRegistry.List())

	// Initialize PII redactor
	redactor := pii.NewPatternRedactor()

	// Initialize document storage & parser
	storageService, err := docStorage.NewLocalStorageService("uploads/documents")
	if err != nil {
		log.Warn("failed to initialize local storage service, falling back to default", "error", err)
	}
	parserService := docParser.NewDefaultParser()

	// Build GraphQL, Document, Obligation, Dashboard, MCP, and Marketplace handlers (if MongoDB is available)
	var graphqlHandler http.Handler
	var documentHandler *docHttpHandler.Handler
	var obligationHandler *oblHttpHandler.Handler
	var dashboardHandler *dashboardHttpHandler.Handler
	var mcpHandler *mcpHttpHandler.Handler
	var marketplaceHandler *mktHttpHandler.Handler

	if mongoClient != nil {
		mongoDB := database.GetDatabase(mongoClient, cfg.MongoDB)

		// Repositories
		docRepo := mongoDoc.NewMongoRepository(mongoDB)
		oblRepo := mongoObl.NewMongoRepository(mongoDB)
		mktRepo := mongoMkt.NewMongoRepository(mongoDB)

		// Agents
		var analystAgent *analyst.Analyst
		var verifierAgent *verifier.Verifier

		if analystProvider != nil {
			analystAgent = analyst.New(analystProvider, log)
		}
		if verifierProvider != nil {
			verifierAgent = verifier.New(verifierProvider, log)
		}

		// Orchestrator
		var orch *orchestrator.Orchestrator
		if analystAgent != nil && verifierAgent != nil {
			orch = orchestrator.New(orchestrator.Config{
				Analyst:  analystAgent,
				Verifier: verifierAgent,
				Redactor: redactor,
				DocRepo:  docRepo,
				OblRepo:  oblRepo,
				MktRepo:  mktRepo,
				EventBus: eventBus,
				Logger:   log,
			})
		}

		// Document Use Cases
		uploadUC := docUsecase.NewUploadDocumentUseCase(docUsecase.UploadConfig{
			DocRepo:      docRepo,
			Storage:      storageService,
			Parser:       parserService,
			Orchestrator: orch,
			EventBus:     eventBus,
			Logger:       log,
		})
		getUC := docUsecase.NewGetDocumentUseCase(docRepo)
		listUC := docUsecase.NewListDocumentsUseCase(docRepo)

		documentHandler = docHttpHandler.NewHandler(docHttpHandler.Config{
			UploadUC:     uploadUC,
			GetUC:        getUC,
			ListUC:       listUC,
			Orchestrator: orch,
			OblRepo:      oblRepo,
		})

		// Audit Repository
		auditRepo := mongoAudit.NewMongoAgentAuditRepository(mongoDB)

		// Obligation Use Cases
		approveOblUC := oblUsecase.NewApproveObligationUseCase(oblUsecase.ApproveConfig{
			OblRepo:   oblRepo,
			AuditRepo: auditRepo,
			MCPReg:    mcpRegistry,
			EventBus:  eventBus,
			Logger:    log,
		})
		dismissOblUC := oblUsecase.NewDismissObligationUseCase(oblRepo, auditRepo, log)
		listOblUC := oblUsecase.NewListObligationsUseCase(oblRepo)

		obligationHandler = oblHttpHandler.NewHandler(oblHttpHandler.Config{
			ApproveUC: approveOblUC,
			DismissUC: dismissOblUC,
			ListUC:    listOblUC,
		})

		// Dashboard Use Case
		dashboardUC := dashboardUsecase.NewGetDashboardSummaryUseCase(dashboardUsecase.Config{
			OblRepo:  oblRepo,
			MktRepo:  mktRepo,
			MCPReg:   mcpRegistry,
			Analyst:  analystProvider,
			Verifier: verifierProvider,
			Logger:   log,
		})
		dashboardHandler = dashboardHttpHandler.NewHandler(dashboardUC)

		// MCP Handler
		mcpHandler = mcpHttpHandler.NewHandler(mcpRegistry)

		// Marketplace Engine & Use Cases
		rfqEngine := mktAgent.NewRFQEngine(mktRepo, verifierAgent, log)
		rfqUC := mktUsecase.NewTriggerRFQUseCase(rfqEngine, mktRepo, eventBus, log)
		acceptBidUC := mktUsecase.NewAcceptBidUseCase(mktUsecase.AcceptBidConfig{
			MktRepo:   mktRepo,
			AuditRepo: auditRepo,
			MCPReg:    mcpRegistry,
			EventBus:  eventBus,
			Logger:    log,
		})
		metricsUC := mktUsecase.NewGetMetricsUseCase(mktRepo)
		listOppUC := mktUsecase.NewListOpportunitiesUseCase(mktRepo)

		marketplaceHandler = mktHttpHandler.NewHandler(mktHttpHandler.Config{
			RFQUC:     rfqUC,
			AcceptUC:  acceptBidUC,
			MetricsUC: metricsUC,
			ListUC:    listOppUC,
		})

		// GraphQL resolver
		resolver := graphqlServer.NewResolver(
			docRepo, oblRepo, mktRepo,
			orch, mcpRegistry,
			analystProvider, verifierProvider,
			log,
		)
		graphqlHandler = graphqlServer.NewServer(resolver, log)
	}

	// Build Chi HTTP router
	r := chi.NewRouter()

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"claimpilot-api"}`))
	})

	// Document REST API routes
	r.Route("/api/v1/documents", func(sub chi.Router) {
		if documentHandler != nil {
			documentHandler.Routes(sub)
		} else {
			sub.HandleFunc("/*", unavailableHandler)
		}
	})

	// Obligation REST API routes
	r.Route("/api/v1/obligations", func(sub chi.Router) {
		if obligationHandler != nil {
			obligationHandler.Routes(sub)
		} else {
			sub.HandleFunc("/*", unavailableHandler)
		}
	})

	// Dashboard REST API routes
	r.Route("/api/v1/dashboard", func(sub chi.Router) {
		if dashboardHandler != nil {
			dashboardHandler.Routes(sub)
		} else {
			sub.HandleFunc("/*", unavailableHandler)
		}
	})

	// MCP Tool API routes
	r.Route("/api/v1/mcp", func(sub chi.Router) {
		if mcpHandler != nil {
			mcpHandler.Routes(sub)
		} else {
			sub.HandleFunc("/*", unavailableHandler)
		}
	})

	// Marketplace REST API routes
	r.Route("/api/v1/marketplace", func(sub chi.Router) {
		if marketplaceHandler != nil {
			marketplaceHandler.Routes(sub)
		} else {
			sub.HandleFunc("/*", unavailableHandler)
		}
	})

	// GraphQL endpoints
	if graphqlHandler != nil {
		r.Handle("/graphql", graphqlHandler)
		r.Handle("/playground", graphqlHandler)
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		log.Info("claimpilot-api listening", "addr", addr)
		serverErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	case sig := <-shutdown:
		log.Info("shutdown signal received", "signal", sig)
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			_ = srv.Close()
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}
		log.Info("claimpilot-api stopped gracefully")
	}

	return nil
}

func unavailableHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte(`{"error":"MongoDB is unavailable; ClaimPilot service is inactive"}`))
}
