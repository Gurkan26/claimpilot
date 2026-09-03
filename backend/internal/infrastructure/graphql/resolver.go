package graphql

import (
	"log/slog"

	docRepo "github.com/masterfabric-go/masterfabric/internal/domain/document/repository"
	mktRepo "github.com/masterfabric-go/masterfabric/internal/domain/marketplace/repository"
	oblRepo "github.com/masterfabric-go/masterfabric/internal/domain/obligation/repository"
	"github.com/masterfabric-go/masterfabric/internal/agent/llmclient"
	"github.com/masterfabric-go/masterfabric/internal/agent/orchestrator"
	"github.com/masterfabric-go/masterfabric/internal/mcp"
)

// Resolver is the root GraphQL resolver.
// It holds all dependencies needed to resolve queries, mutations, and subscriptions.
// Business logic lives in domain/application packages, not here — resolvers are thin.
type Resolver struct {
	DocRepo      docRepo.DocumentRepository
	OblRepo      oblRepo.ObligationRepository
	MktRepo      mktRepo.MarketplaceRepository
	Orchestrator *orchestrator.Orchestrator
	MCPRegistry  *mcp.Registry
	Analyst      llmclient.Provider
	Verifier     llmclient.Provider
	Logger       *slog.Logger
}

// NewResolver creates the root resolver with all dependencies.
func NewResolver(
	docRepo docRepo.DocumentRepository,
	oblRepo oblRepo.ObligationRepository,
	mktRepo mktRepo.MarketplaceRepository,
	orch *orchestrator.Orchestrator,
	mcpReg *mcp.Registry,
	analyst llmclient.Provider,
	verifier llmclient.Provider,
	logger *slog.Logger,
) *Resolver {
	return &Resolver{
		DocRepo:      docRepo,
		OblRepo:      oblRepo,
		MktRepo:      mktRepo,
		Orchestrator: orch,
		MCPRegistry:  mcpReg,
		Analyst:      analyst,
		Verifier:     verifier,
		Logger:       logger.With("component", "graphql"),
	}
}
