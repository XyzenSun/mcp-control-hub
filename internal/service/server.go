package service

import (
	"encoding/json"
	"fmt"

	"mcp-control-hub/internal/mcp/client"
	"mcp-control-hub/internal/models"
	"mcp-control-hub/internal/repository"
	"mcp-control-hub/pkg/logger"
)

// ServerService handles business logic for servers
type ServerService struct {
	serverRepo    *repository.ServerRepository
	toolRepo      *repository.ToolRepository
	clientManager *client.ClientManager
	logger        *logger.Logger
}

// NewServerService creates a new server service
func NewServerService(
	serverRepo *repository.ServerRepository,
	toolRepo *repository.ToolRepository,
	clientManager *client.ClientManager,
	logger *logger.Logger,
) *ServerService {
	return &ServerService{
		serverRepo:    serverRepo,
		toolRepo:      toolRepo,
		clientManager: clientManager,
		logger:        logger,
	}
}

// SyncTools syncs tools from an upstream MCP server
func (s *ServerService) SyncTools(serverID uint) error {
	// Get server
	server, err := s.serverRepo.FindByID(serverID)
	if err != nil {
		return fmt.Errorf("failed to find server: %w", err)
	}

	if !server.Enabled {
		return fmt.Errorf("server is disabled")
	}

	s.logger.Info().
		Uint("server_id", serverID).
		Str("server_name", server.Name).
		Msg("Syncing tools from MCP server")

	// Get or create client connection
	_, err = s.clientManager.GetClient(server.Name)
	if err != nil {
		// Try to connect
		if err := s.clientManager.ReloadServer(server.Name); err != nil {
			return fmt.Errorf("failed to connect to server: %w", err)
		}
	}

	// List tools from upstream server
	tools, err := s.clientManager.ListTools(server.Name)
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}

	s.logger.Info().
		Int("tool_count", len(tools)).
		Str("server_name", server.Name).
		Msg("Retrieved tools from server")

	// Get existing tools for this server
	existingTools, err := s.toolRepo.FindByServerID(serverID)
	if err != nil {
		return fmt.Errorf("failed to get existing tools: %w", err)
	}

	existingToolMap := make(map[string]*models.Tool)
	for i := range existingTools {
		existingToolMap[existingTools[i].Name] = &existingTools[i]
	}

	// Sync tools
	syncedTools := make(map[string]bool)
	for _, toolData := range tools {
		toolName, ok := toolData["name"].(string)
		if !ok {
			s.logger.Warn().Interface("tool", toolData).Msg("Tool missing name field")
			continue
		}

		description, _ := toolData["description"].(string)

		// Convert input schema to JSON string
		inputSchemaJSON := "{}"
		if inputSchema, ok := toolData["inputSchema"]; ok {
			if schemaBytes, err := json.Marshal(inputSchema); err == nil {
				inputSchemaJSON = string(schemaBytes)
			}
		}

		syncedTools[toolName] = true

		if existingTool, exists := existingToolMap[toolName]; exists {
			// Update existing tool
			existingTool.OriginalDescription = description
			existingTool.InputSchema = inputSchemaJSON
			// Preserve user's override description and enabled status
			if err := s.toolRepo.Update(existingTool); err != nil {
				s.logger.Error().
					Err(err).
					Str("tool_name", toolName).
					Msg("Failed to update tool")
			}
		} else {
			// Create new tool
			newTool := &models.Tool{
				ServerID:            serverID,
				Name:                toolName,
				OriginalDescription: description,
				InputSchema:         inputSchemaJSON,
				Enabled:             true,
			}
			if err := s.toolRepo.Create(newTool); err != nil {
				s.logger.Error().
					Err(err).
					Str("tool_name", toolName).
					Msg("Failed to create tool")
			}
		}
	}

	// Mark tools that no longer exist as disabled (optional: could delete them)
	for toolName, existingTool := range existingToolMap {
		if !syncedTools[toolName] {
			s.logger.Info().
				Str("tool_name", toolName).
				Msg("Tool no longer exists on server, disabling")
			existingTool.Enabled = false
			s.toolRepo.Update(existingTool)
		}
	}

	s.logger.Info().
		Uint("server_id", serverID).
		Int("synced_count", len(syncedTools)).
		Msg("Tool sync completed")

	return nil
}

// RefreshAllTools syncs tools from all enabled servers
func (s *ServerService) RefreshAllTools() error {
	servers, err := s.serverRepo.FindAllEnabled()
	if err != nil {
		return fmt.Errorf("failed to get enabled servers: %w", err)
	}

	s.logger.Info().
		Int("server_count", len(servers)).
		Msg("Refreshing tools from all servers")

	errorCount := 0
	for _, server := range servers {
		if err := s.SyncTools(server.ID); err != nil {
			s.logger.Error().
				Err(err).
				Uint("server_id", server.ID).
				Str("server_name", server.Name).
				Msg("Failed to sync tools")
			errorCount++
		}
	}

	if errorCount > 0 {
		return fmt.Errorf("failed to sync %d servers", errorCount)
	}

	return nil
}
