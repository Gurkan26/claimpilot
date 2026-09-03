package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	auditModel "github.com/masterfabric-go/masterfabric/internal/domain/audit/model"
	oblModel "github.com/masterfabric-go/masterfabric/internal/domain/obligation/model"
	mongoAudit "github.com/masterfabric-go/masterfabric/internal/infrastructure/mongodb/audit"
	mongoDoc "github.com/masterfabric-go/masterfabric/internal/infrastructure/mongodb/document"
	mongoObl "github.com/masterfabric-go/masterfabric/internal/infrastructure/mongodb/obligation"
	"github.com/masterfabric-go/masterfabric/internal/mcp"
	mcpCalendar "github.com/masterfabric-go/masterfabric/internal/mcp/calendar"
	mcpGmail "github.com/masterfabric-go/masterfabric/internal/mcp/gmail"
	mcpSlack "github.com/masterfabric-go/masterfabric/internal/mcp/slack"
	"github.com/masterfabric-go/masterfabric/internal/shared/config"
	"github.com/masterfabric-go/masterfabric/internal/shared/database"
	"github.com/masterfabric-go/masterfabric/internal/shared/events"
	"github.com/masterfabric-go/masterfabric/internal/shared/logger"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "worker error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()
	log := logger.New(cfg.Log.Level, cfg.Log.Format).With("component", "claimpilot-worker")
	slog.SetDefault(log)

	log.Info("starting claimpilot-worker daemon")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := database.NewMongoClient(ctx, cfg.MongoDB)
	if err != nil {
		return fmt.Errorf("connect mongodb: %w", err)
	}
	defer func() { _ = mongoClient.Disconnect(context.Background()) }()

	mongoDB := database.GetDatabase(mongoClient, cfg.MongoDB)

	// Repositories
	oblRepo := mongoObl.NewMongoRepository(mongoDB)
	docRepo := mongoDoc.NewMongoRepository(mongoDB)
	auditRepo := mongoAudit.NewMongoAgentAuditRepository(mongoDB)
	_ = docRepo

	// Event bus
	eventBus := events.NewInProcessBus(log, 256)
	defer func() { _ = eventBus.Close() }()

	// MCP Registry
	mcpReg := mcp.NewRegistry()
	mcpReg.Register(mcpGmail.New(log))
	mcpReg.Register(mcpCalendar.New(log))
	mcpReg.Register(mcpSlack.New(log))
	log.Info("registered worker MCP adapters", "adapters", mcpReg.List())

	// Scan interval configuration
	intervalSec := 30
	if envInterval := os.Getenv("WORKER_SCAN_INTERVAL_SECONDS"); envInterval != "" {
		if val, err := strconv.Atoi(envInterval); err == nil && val > 0 {
			intervalSec = val
		}
	}
	scanTicker := time.NewTicker(time.Duration(intervalSec) * time.Second)
	defer scanTicker.Stop()

	log.Info("obligation deadline scheduler started", "interval_seconds", intervalSec)

	// Graceful shutdown handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-scanTicker.C:
			scanCtx, scanCancel := context.WithTimeout(context.Background(), 20*time.Second)
			scanAndDispatchDeadlines(scanCtx, oblRepo, auditRepo, mcpReg, log)
			scanCancel()

		case sig := <-stop:
			log.Info("shutdown signal received, stopping worker", "signal", sig)
			return nil
		}
	}
}

// scanAndDispatchDeadlines checks obligations due within 7 days and triggers reminders.
func scanAndDispatchDeadlines(
	ctx context.Context,
	oblRepo *mongoObl.MongoRepository,
	auditRepo *mongoAudit.MongoAgentAuditRepository,
	mcpReg *mcp.Registry,
	log *slog.Logger,
) {
	log.Debug("scanning upcoming obligations for proactive reminders")

	// Query upcoming obligations for demo user
	demoUserID := bson.NilObjectID // or query upcoming across all users if needed
	_ = demoUserID

	// Upcoming in next 7 days
	upcoming, err := oblRepo.FindUpcoming(ctx, [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}, 7)
	if err != nil {
		log.Warn("failed to scan upcoming obligations", "error", err)
		return
	}

	if len(upcoming) == 0 {
		return
	}

	log.Info("found upcoming critical obligations", "count", len(upcoming))

	for _, obl := range upcoming {
		// Only notify if in pending or in-progress state
		if obl.Status != oblModel.ObligationStatusPendingApproval && obl.Status != oblModel.ObligationStatusInProgress {
			continue
		}

		daysRemaining := int(time.Until(obl.DueDate).Hours() / 24)

		// Dispatch Slack reminder for high/critical risk
		if obl.RiskLevel == oblModel.RiskLevelHigh || obl.RiskLevel == oblModel.RiskLevelCritical {
			if slackAdapter, ok := mcpReg.Get("slack"); ok {
				msg := fmt.Sprintf("CRITICAL DEADLINE ALERT: %s is due in %d day(s) on %s! Immediate action required.",
					obl.Title, daysRemaining, obl.DueDate.Format("2006-01-02"))

				action := &mcp.Action{
					Type:      mcp.ActionTypeSendSlack,
					AdapterID: "slack",
					Parameters: map[string]any{
						"title":      fmt.Sprintf("Urgent Deadline: %s", obl.Title),
						"message":    msg,
						"risk_level": string(obl.RiskLevel),
					},
				}

				res, err := slackAdapter.Execute(ctx, action)
				if err != nil {
					log.Error("worker slack reminder dispatch failed", "error", err, "obligation_id", obl.ID.Hex())
				} else {
					log.Info("worker proactive reminder dispatched",
						"obligation_id", obl.ID.Hex(),
						"message", res.Message,
					)

					// Record audit trail entry
					_ = auditRepo.Create(ctx, &auditModel.AgentAuditEntry{
						UserID:         obl.UserID,
						OrganizationID: obl.OrganizationID,
						ActionType:     auditModel.AuditActionMCPActionExecuted,
						ObligationID:   &obl.ID,
						DocumentID:     &obl.SourceDocumentID,
						AdapterID:      "slack",
						ApprovalType:   auditModel.ApprovalTypeAutonomous,
						Details: bson.M{
							"reason":         "worker_proactive_deadline_scan",
							"days_remaining": daysRemaining,
							"mcp_result":     res,
						},
						Status:    "SUCCESS",
						CreatedAt: time.Now().UTC(),
					})
				}
			}
		}
	}
}
