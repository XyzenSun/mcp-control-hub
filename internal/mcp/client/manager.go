package client

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"mcp-control-hub/internal/mcp/transport"
	"mcp-control-hub/internal/models"
	"mcp-control-hub/internal/repository"
	"mcp-control-hub/pkg/logger"
)

// MCPClient represents a connection to an upstream MCP server
type MCPClient struct {
	serverID   uint
	serverName string
	protocol   string
	conn       *transport.ClientConnection
	connected  bool
	lastError  error
	mu         sync.RWMutex
}

// ClientManager manages connections to all upstream MCP servers
type ClientManager struct {
	clients    map[string]*MCPClient // key: server name
	mu         sync.RWMutex
	serverRepo *repository.ServerRepository
	configRepo *repository.ConfigRepository
	logger     *logger.Logger
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewClientManager creates a new client manager
func NewClientManager(serverRepo *repository.ServerRepository, configRepo *repository.ConfigRepository, log *logger.Logger) *ClientManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &ClientManager{
		clients:    make(map[string]*MCPClient),
		serverRepo: serverRepo,
		configRepo: configRepo,
		logger:     log,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// GetTimeout retrieves a timeout value from config (in seconds)
func (m *ClientManager) GetTimeout(key string) time.Duration {
	seconds, err := m.configRepo.GetInt(key)
	if err != nil || seconds <= 0 {
		// Use default value
		if defaultVal, ok := models.DefaultConfigValues[key]; ok {
			if sec, err := time.ParseDuration(defaultVal + "s"); err == nil {
				return sec
			}
		}
		return 30 * time.Second // fallback
	}
	return time.Duration(seconds) * time.Second
}

// Start initializes connections to all enabled servers
func (m *ClientManager) Start() error {
	servers, err := m.serverRepo.FindAllEnabled()
	if err != nil {
		return fmt.Errorf("failed to load servers: %w", err)
	}

	m.logger.Info().Int("count", len(servers)).Msg("Starting MCP client manager")

	for _, server := range servers {
		if err := m.connectServer(&server); err != nil {
			m.logger.Error().
				Err(err).
				Uint("server_id", server.ID).
				Str("server_name", server.Name).
				Msg("Failed to connect to server")
			continue
		}
	}

	// Start health check goroutine
	go m.healthCheckLoop()

	return nil
}

// Stop gracefully shuts down all connections
func (m *ClientManager) Stop() error {
	m.logger.Info().Msg("Stopping MCP client manager")
	m.cancel()

	m.mu.Lock()
	defer m.mu.Unlock()

	for name, client := range m.clients {
		if err := client.conn.Close(); err != nil {
			m.logger.Error().
				Err(err).
				Str("server_name", name).
				Msg("Failed to close connection")
		}
	}

	m.clients = make(map[string]*MCPClient)
	return nil
}

// GetClient returns the client for a specific server
func (m *ClientManager) GetClient(serverName string) (*MCPClient, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, exists := m.clients[serverName]
	if !exists {
		return nil, fmt.Errorf("server not found: %s", serverName)
	}

	if !client.connected {
		return nil, fmt.Errorf("server not connected: %s", serverName)
	}

	return client, nil
}

// ReloadServer reloads a single server configuration
func (m *ClientManager) ReloadServer(serverName string) error {
	// Disconnect existing client
	m.mu.Lock()
	if client, exists := m.clients[serverName]; exists {
		client.conn.Close()
		delete(m.clients, serverName)
	}
	m.mu.Unlock()

	// Load server from database
	server, err := m.serverRepo.FindByName(serverName)
	if err != nil {
		return fmt.Errorf("failed to load server: %w", err)
	}

	if !server.Enabled {
		return fmt.Errorf("server is disabled: %s", serverName)
	}

	// Reconnect
	return m.connectServer(server)
}

// ListTools returns all tools from a specific server
func (m *ClientManager) ListTools(serverName string) ([]map[string]interface{}, error) {
	mcpClient, err := m.GetClient(serverName)
	if err != nil {
		return nil, err
	}

	mcpClient.mu.RLock()
	defer mcpClient.mu.RUnlock()

	timeout := m.GetTimeout(models.ConfigKeyListToolsTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result, err := mcpClient.conn.Session.ListTools(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	// Convert to generic map format
	tools := make([]map[string]interface{}, len(result.Tools))
	for i, tool := range result.Tools {
		// Convert input schema to map
		var inputSchema map[string]interface{}
		if tool.InputSchema != nil {
			schemaBytes, _ := json.Marshal(tool.InputSchema)
			json.Unmarshal(schemaBytes, &inputSchema)
		}

		tools[i] = map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": inputSchema,
		}
	}

	return tools, nil
}

// CallTool executes a tool on a specific server
func (m *ClientManager) CallTool(serverName, toolName string, arguments map[string]interface{}) (interface{}, error) {
	mcpClient, err := m.GetClient(serverName)
	if err != nil {
		return nil, err
	}

	mcpClient.mu.RLock()
	defer mcpClient.mu.RUnlock()

	timeout := m.GetTimeout(models.ConfigKeyCallToolTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result, err := mcpClient.conn.Session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: arguments,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call tool: %w", err)
	}

	return result, nil
}

// connectServer establishes a connection to a server
func (m *ClientManager) connectServer(server *models.UpstreamServer) error {
	m.logger.Info().
		Uint("server_id", server.ID).
		Str("server_name", server.Name).
		Str("protocol", server.Protocol).
		Msg("Connecting to MCP server")

	// Parse config
	var config transport.Config
	if err := json.Unmarshal([]byte(server.Config), &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	timeout := m.GetTimeout(models.ConfigKeyConnectTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var conn *transport.ClientConnection
	var err error

	switch server.Protocol {
	case "streamable":
		if config.URL == "" {
			return fmt.Errorf("streamable protocol requires 'url' in config")
		}
		conn, err = transport.NewStreamableConnection(ctx, config.URL, config.Headers)
	case "sse":
		if config.BaseURL == "" {
			return fmt.Errorf("sse protocol requires 'base_url' in config")
		}
		conn, err = transport.NewSSEConnection(ctx, config.BaseURL, config.Headers)
	case "stdio":
		return fmt.Errorf("stdio protocol is experimental and not yet implemented")
	default:
		return fmt.Errorf("unsupported protocol: %s", server.Protocol)
	}

	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Store client
	client := &MCPClient{
		serverID:   server.ID,
		serverName: server.Name,
		protocol:   server.Protocol,
		conn:       conn,
		connected:  true,
	}

	m.mu.Lock()
	m.clients[server.Name] = client
	m.mu.Unlock()

	m.logger.Info().
		Str("server_name", server.Name).
		Msg("Successfully connected to MCP server")

	return nil
}

// healthCheckLoop periodically checks the health of all connections
func (m *ClientManager) healthCheckLoop() {
	// Get interval from config
	intervalSeconds, err := m.configRepo.GetInt(models.ConfigKeyHealthCheckInterval)
	if err != nil || intervalSeconds <= 0 {
		intervalSeconds = 30 // default
	}
	ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.performHealthCheck()
		}
	}
}

// performHealthCheck checks all connections and attempts to reconnect if needed
func (m *ClientManager) performHealthCheck() {
	m.mu.RLock()
	clients := make([]*MCPClient, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	m.mu.RUnlock()

	timeout := m.GetTimeout(models.ConfigKeyHealthCheckTimeout)
	for _, client := range clients {
		// Simple check: try to list tools
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		_, err := client.conn.Session.ListTools(ctx, nil)
		cancel()

		if err != nil {
			m.logger.Warn().
				Str("server_name", client.serverName).
				Err(err).
				Msg("Health check failed, attempting to reconnect")

			if err := m.ReloadServer(client.serverName); err != nil {
				m.logger.Error().
					Err(err).
					Str("server_name", client.serverName).
					Msg("Failed to reconnect")
			}
		}
	}
}