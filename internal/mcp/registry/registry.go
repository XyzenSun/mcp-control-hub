package registry

import (
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"mcp-control-hub/internal/mcp/client"
	"mcp-control-hub/internal/repository"
	"mcp-control-hub/pkg/logger"
)

// ToolInfo represents aggregated tool information
type ToolInfo struct {
	Name        string
	Description string
	InputSchema interface{}
	ServerID    uint
	ServerName  string
	Enabled     bool
}

// ToolRegistry manages tool aggregation and routing
type ToolRegistry struct {
	clientManager *client.ClientManager
	toolRepo      *repository.ToolRepository
	namespaceRepo *repository.NamespaceRepository
	logger        *logger.Logger
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry(
	clientManager *client.ClientManager,
	toolRepo *repository.ToolRepository,
	namespaceRepo *repository.NamespaceRepository,
	logger *logger.Logger,
) *ToolRegistry {
	return &ToolRegistry{
		clientManager: clientManager,
		toolRepo:      toolRepo,
		namespaceRepo: namespaceRepo,
		logger:        logger,
	}
}

// GetToolsForNamespace returns all enabled tools for a namespace
func (r *ToolRegistry) GetToolsForNamespace(namespaceName string) ([]*ToolInfo, error) {
	r.logger.Debug().Str("namespace", namespaceName).Msg("Fetching tools from database")

	// Get namespace by ID (namespaceName is actually the namespace ID from URL)
	namespace, err := r.namespaceRepo.FindByID(namespaceName)
	if err != nil {
		return nil, fmt.Errorf("namespace not found: %w", err)
	}

	// Get tools for namespace
	tools, err := r.toolRepo.FindByNamespaceID(namespace.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tools: %w", err)
	}

	// Convert to ToolInfo
	toolInfos := make([]*ToolInfo, 0, len(tools))
	for _, tool := range tools {
		if !tool.Enabled {
			continue
		}

		// Parse input schema
		var inputSchema interface{}
		if tool.InputSchema != "" {
			json.Unmarshal([]byte(tool.InputSchema), &inputSchema)
		}

		// Use override description if UseOverride is true and override is set
		description := tool.OriginalDescription
		if tool.UseOverride && tool.OverrideDescription != nil && *tool.OverrideDescription != "" {
			description = *tool.OverrideDescription
		}

		toolInfos = append(toolInfos, &ToolInfo{
			Name:        tool.Name,
			Description: description,
			InputSchema: inputSchema,
			ServerID:    tool.ServerID,
			ServerName:  tool.Server.Name,
			Enabled:     tool.Enabled,
		})
	}

	return toolInfos, nil
}

// GetAllTools returns all enabled tools across all servers
func (r *ToolRegistry) GetAllTools() ([]*ToolInfo, error) {
	tools, err := r.toolRepo.FindAllEnabled()
	if err != nil {
		return nil, fmt.Errorf("failed to get tools: %w", err)
	}

	toolInfos := make([]*ToolInfo, 0, len(tools))
	for _, tool := range tools {
		// Parse input schema
		var inputSchema interface{}
		if tool.InputSchema != "" {
			json.Unmarshal([]byte(tool.InputSchema), &inputSchema)
		}

		// Use override description if UseOverride is true and override is set
		description := tool.OriginalDescription
		if tool.UseOverride && tool.OverrideDescription != nil && *tool.OverrideDescription != "" {
			description = *tool.OverrideDescription
		}

		toolInfos = append(toolInfos, &ToolInfo{
			Name:        tool.Name,
			Description: description,
			InputSchema: inputSchema,
			ServerID:    tool.ServerID,
			ServerName:  tool.Server.Name,
			Enabled:     tool.Enabled,
		})
	}

	return toolInfos, nil
}

// RouteToolCall routes a tool call to the appropriate upstream server
func (r *ToolRegistry) RouteToolCall(namespaceName, toolName string, arguments map[string]interface{}) (interface{}, error) {
	// Get namespace tools
	tools, err := r.GetToolsForNamespace(namespaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace tools: %w", err)
	}

	// Find the tool
	var targetTool *ToolInfo
	for _, tool := range tools {
		if tool.Name == toolName {
			targetTool = tool
			break
		}
	}

	if targetTool == nil {
		return nil, fmt.Errorf("tool not found in namespace: %s", toolName)
	}

	// Route to upstream server
	r.logger.Info().
		Str("namespace", namespaceName).
		Str("tool", toolName).
		Str("server", targetTool.ServerName).
		Msg("Routing tool call")

	result, err := r.clientManager.CallTool(targetTool.ServerName, toolName, arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to call tool: %w", err)
	}

	return result, nil
}

// ConvertToMCPTools converts ToolInfo to MCP Tool format
func (r *ToolRegistry) ConvertToMCPTools(tools []*ToolInfo) []*mcp.Tool {
	mcpTools := make([]*mcp.Tool, len(tools))
	for i, tool := range tools {
		mcpTools[i] = &mcp.Tool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.InputSchema,
		}
	}
	return mcpTools
}

// ResolveConflicts handles duplicate tool names across servers
// Returns a map of resolved tool names to original tool info
func (r *ToolRegistry) ResolveConflicts(tools []*ToolInfo) map[string]*ToolInfo {
	resolved := make(map[string]*ToolInfo)
	nameCounts := make(map[string]int)

	// Count occurrences of each tool name
	for _, tool := range tools {
		nameCounts[tool.Name]++
	}

	// Resolve conflicts by prefixing with server name
	for _, tool := range tools {
		resolvedName := tool.Name
		if nameCounts[tool.Name] > 1 {
			// Conflict detected, prefix with server name
			resolvedName = fmt.Sprintf("%s_%s", tool.ServerName, tool.Name)
			r.logger.Warn().
				Str("original_name", tool.Name).
				Str("resolved_name", resolvedName).
				Str("server", tool.ServerName).
				Msg("Tool name conflict resolved")
		}
		resolved[resolvedName] = tool
	}

	return resolved
}
