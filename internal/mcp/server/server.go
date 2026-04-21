package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"mcp-control-hub/internal/mcp/registry"
	"mcp-control-hub/internal/repository"
	"mcp-control-hub/pkg/logger"
)

// MCPServer implements the MCP server for clients
type MCPServer struct {
	registry   *registry.ToolRegistry
	apiKeyRepo *repository.APIKeyRepository
	logger     *logger.Logger
	serversMu  sync.RWMutex
	servers    map[string]*mcp.Server // key: namespace
}

// NewMCPServer creates a new MCP server
func NewMCPServer(
	registry *registry.ToolRegistry,
	apiKeyRepo *repository.APIKeyRepository,
	logger *logger.Logger,
) *MCPServer {
	return &MCPServer{
		registry:   registry,
		apiKeyRepo: apiKeyRepo,
		logger:     logger,
		servers:    make(map[string]*mcp.Server),
	}
}

// Handler returns an HTTP handler for the MCP server
func (s *MCPServer) Handler() http.Handler {
	// Create streamable HTTP handler with stateless mode
	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		// Extract apikey and namespace from path
		// Format: /mcp/{apikey}/{namespace}
		path := r.URL.Path

		// Default values
		apiKey := ""
		namespace := "default"

		// Parse path: /mcp/{apikey}/{namespace}
		if len(path) > 5 && path[:5] == "/mcp/" {
			parts := strings.Split(path[5:], "/")
			if len(parts) >= 2 {
				apiKey = parts[0]
				namespace = parts[1]
			} else if len(parts) == 1 {
				// Only apikey provided, use default namespace
				apiKey = parts[0]
			}
		}

		// Validate API key
		if apiKey == "" || !strings.HasPrefix(apiKey, "sk_") {
			s.logger.Warn().Str("path", path).Msg("Invalid or missing API key in path")
			// Return empty server (will fail gracefully)
			return mcp.NewServer(&mcp.Implementation{
				Name:    "mcp-gateway",
				Version: "1.0.0",
			}, nil)
		}

		// Verify API key
		keyHash := hashAPIKey(apiKey)
		_, err := s.apiKeyRepo.FindByHash(keyHash)
		if err != nil {
			s.logger.Warn().Str("apikey", apiKey[:10]+"...").Msg("Invalid API key")
			// Return empty server
			return mcp.NewServer(&mcp.Implementation{
				Name:    "mcp-gateway",
				Version: "1.0.0",
			}, nil)
		}

		// Validate namespace (alphanumeric only)
		if !isValidNamespace(namespace) {
			s.logger.Warn().Str("namespace", namespace).Msg("Invalid namespace format")
			namespace = "default"
		}

		s.logger.Debug().
			Str("namespace", namespace).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("MCP request received")

		// Get or create MCP server for this namespace
		return s.getOrCreateServer(namespace)
	}, &mcp.StreamableHTTPOptions{
		Stateless: true,
	})

	return handler
}

// getOrCreateServer gets or creates an MCP server for a namespace
func (s *MCPServer) getOrCreateServer(namespace string) *mcp.Server {
	s.serversMu.RLock()
	server, exists := s.servers[namespace]
	s.serversMu.RUnlock()

	if exists {
		return server
	}

	// Create new server
	s.serversMu.Lock()
	defer s.serversMu.Unlock()

	// Double-check after acquiring write lock
	if server, exists := s.servers[namespace]; exists {
		return server
	}

	server = s.createServerForNamespace(namespace)
	s.servers[namespace] = server

	s.logger.Info().
		Str("namespace", namespace).
		Msg("Created new MCP server for namespace")

	return server
}

// createServerForNamespace creates an MCP server with tools for a specific namespace
func (s *MCPServer) createServerForNamespace(namespace string) *mcp.Server {
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "mcp-gateway",
		Version: "1.0.0",
	}, nil)

	// Get tools for namespace
	tools, err := s.registry.GetToolsForNamespace(namespace)
	if err != nil {
		s.logger.Error().Err(err).Str("namespace", namespace).Msg("Failed to get tools")
		return mcpServer
	}

	s.logger.Info().
		Int("tool_count", len(tools)).
		Str("namespace", namespace).
		Msg("Registering tools for namespace")

	// Register each tool
	for _, tool := range tools {
		s.registerTool(mcpServer, namespace, tool)
	}

	return mcpServer
}

// registerTool registers a single tool with the MCP server
func (s *MCPServer) registerTool(mcpServer *mcp.Server, namespace string, tool *registry.ToolInfo) {
	toolName := tool.Name
	serverName := tool.ServerName

	// Parse input schema
	var inputSchema interface{}
	if tool.InputSchema != nil {
		inputSchema = tool.InputSchema
	} else {
		// Default empty object schema
		inputSchema = map[string]interface{}{"type": "object"}
	}

	// Create MCP tool
	mcpTool := &mcp.Tool{
		Name:        tool.Name,
		Description: tool.Description,
		InputSchema: inputSchema,
	}

	// Create tool handler that routes to upstream server
	handler := func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		s.logger.Info().
			Str("tool", toolName).
			Str("server", serverName).
			Str("namespace", namespace).
			Msg("Tool called via MCP")

		// Convert arguments to map
		arguments := make(map[string]interface{})
		if req.Params.Arguments != nil {
			// Arguments is json.RawMessage, unmarshal it
			if err := json.Unmarshal(req.Params.Arguments, &arguments); err != nil {
				s.logger.Error().Err(err).Msg("Failed to unmarshal arguments")
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}
		}

		result, err := s.registry.RouteToolCall(namespace, toolName, arguments)
		if err != nil {
			s.logger.Error().
				Err(err).
				Str("tool", toolName).
				Msg("Tool call failed")
			return nil, err
		}

		// Convert result to CallToolResult
		if callResult, ok := result.(*mcp.CallToolResult); ok {
			return callResult, nil
		}

		// Fallback: wrap in text content
		resultJSON, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: string(resultJSON),
				},
			},
		}, nil
	}

	// Add tool to server
	mcpServer.AddTool(mcpTool, handler)

	s.logger.Debug().
		Str("tool", toolName).
		Str("namespace", namespace).
		Msg("Tool registered")
}

// RefreshNamespace refreshes tools for a specific namespace
func (s *MCPServer) RefreshNamespace(namespace string) error {
	s.serversMu.Lock()
	defer s.serversMu.Unlock()

	// Recreate server for this namespace (tools will be fetched directly from database)
	s.servers[namespace] = s.createServerForNamespace(namespace)

	s.logger.Info().
		Str("namespace", namespace).
		Msg("Namespace server refreshed")

	return nil
}

// RefreshAll refreshes all namespace servers
func (s *MCPServer) RefreshAll() error {
	s.serversMu.Lock()
	defer s.serversMu.Unlock()

	// Recreate all servers (tools will be fetched directly from database)
	for namespace := range s.servers {
		s.servers[namespace] = s.createServerForNamespace(namespace)
		s.logger.Info().
			Str("namespace", namespace).
			Msg("Namespace server refreshed")
	}

	return nil
}

// CleanupStaleServers removes servers that haven't been used recently
func (s *MCPServer) CleanupStaleServers() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.serversMu.Lock()
		// In a real implementation, we would track last access time
		// For now, we just log
		s.logger.Debug().
			Int("server_count", len(s.servers)).
			Msg("Namespace servers active")
		s.serversMu.Unlock()
	}
}

// hashAPIKey hashes an API key using SHA256
func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// isValidNamespace checks if namespace contains only alphanumeric characters
func isValidNamespace(namespace string) bool {
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_-]+$", namespace)
	return matched
}
