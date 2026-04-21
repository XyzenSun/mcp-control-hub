package main

import (
	"fmt"
	"log"
	"os"

	"mcp-control-hub/internal/api"
	"mcp-control-hub/internal/api/handlers"
	"mcp-control-hub/internal/api/middleware"
	"mcp-control-hub/internal/config"
	"mcp-control-hub/internal/database"
	"mcp-control-hub/internal/mcp/client"
	"mcp-control-hub/internal/mcp/registry"
	"mcp-control-hub/internal/mcp/server"
	"mcp-control-hub/internal/repository"
	"mcp-control-hub/internal/service"
	"mcp-control-hub/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	appLogger, err := logger.New(cfg.Logging.Level, cfg.Logging.Format, cfg.Logging.Output)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	appLogger.Info().Msg("Starting MCP Gateway")

	// Initialize database
	db, err := database.Initialize(
		cfg.Database.Driver,
		cfg.Database.DSN,
		cfg.Database.MaxOpenConns,
		cfg.Database.MaxIdleConns,
		cfg.Database.ConnMaxLifetime,
	)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to initialize database")
	}

	appLogger.Info().Str("driver", cfg.Database.Driver).Msg("Database initialized")

	// Initialize repositories
	serverRepo := repository.NewServerRepository(db)
	toolRepo := repository.NewToolRepository(db)
	namespaceRepo := repository.NewNamespaceRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	configRepo := repository.NewConfigRepository(db)

	// Initialize default config values
	if err := configRepo.InitializeDefaults(); err != nil {
		appLogger.Error().Err(err).Msg("Failed to initialize default config values")
	}

	// Initialize bootstrap API key if none exists
	if err := apiKeyRepo.InitializeBootstrapKey(); err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to initialize bootstrap api key")
	}

	// Initialize MCP client manager
	clientManager := client.NewClientManager(serverRepo, configRepo, appLogger)
	if err := clientManager.Start(); err != nil {
		appLogger.Error().Err(err).Msg("Failed to start client manager")
	}
	defer clientManager.Stop()

	// Initialize tool registry
	toolRegistry := registry.NewToolRegistry(clientManager, toolRepo, namespaceRepo, appLogger)

	// Initialize MCP server
	mcpServer := server.NewMCPServer(toolRegistry, apiKeyRepo, appLogger)

	// Initialize services
	serverService := service.NewServerService(serverRepo, toolRepo, clientManager, appLogger)

	// Initialize handlers
	serverHandler := handlers.NewServerHandler(serverRepo, serverService)
	toolHandler := handlers.NewToolHandler(toolRepo, serverService, toolRegistry, mcpServer)
	namespaceHandler := handlers.NewNamespaceHandler(namespaceRepo, toolRegistry, mcpServer)
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyRepo, cfg.Security.APIKeyLength)
	configHandler := handlers.NewConfigHandler(configRepo)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(apiKeyRepo)

	// Setup router
	r := api.Setup(
		serverHandler,
		toolHandler,
		namespaceHandler,
		apiKeyHandler,
		configHandler,
		authMiddleware,
	)

	// Add MCP endpoint with apikey in path
	// Format: /mcp/{apikey}/{namespace}
	r.POST("/mcp/:apikey/:namespace", gin.WrapH(mcpServer.Handler()))
	r.GET("/mcp/:apikey/:namespace", gin.WrapH(mcpServer.Handler()))

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	appLogger.Info().Str("address", addr).Msg("Starting HTTP server")

	if err := r.Run(addr); err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to start server")
	}
}
