package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"mcp-control-hub/internal/mcp/registry"
	"mcp-control-hub/internal/mcp/server"
	"mcp-control-hub/internal/models"
	"mcp-control-hub/internal/repository"
	"mcp-control-hub/internal/service"
)

type ToolHandler struct {
	toolRepo      *repository.ToolRepository
	serverService *service.ServerService
	toolRegistry  *registry.ToolRegistry
	mcpServer     *server.MCPServer
}

func NewToolHandler(toolRepo *repository.ToolRepository, serverService *service.ServerService, toolRegistry *registry.ToolRegistry, mcpServer *server.MCPServer) *ToolHandler {
	return &ToolHandler{
		toolRepo:      toolRepo,
		serverService: serverService,
		toolRegistry:  toolRegistry,
		mcpServer:     mcpServer,
	}
}

func (h *ToolHandler) List(c *gin.Context) {
	var tools []models.Tool
	var err error

	serverID := c.Query("server_id")
	enabled := c.Query("enabled")

	if serverID != "" {
		id, err := strconv.ParseUint(serverID, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid server_id"})
			return
		}
		tools, err = h.toolRepo.FindByServerID(uint(id))
	} else if enabled == "true" {
		tools, err = h.toolRepo.FindAllEnabled()
	} else {
		tools, err = h.toolRepo.FindAll()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tools"})
		return
	}

	c.JSON(http.StatusOK, tools)
}

func (h *ToolHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tool ID"})
		return
	}

	tool, err := h.toolRepo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tool not found"})
		return
	}

	c.JSON(http.StatusOK, tool)
}

func (h *ToolHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tool ID"})
		return
	}

	tool, err := h.toolRepo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tool not found"})
		return
	}

	var updateData struct {
		Enabled             *bool   `json:"enabled"`
		OverrideDescription *string `json:"override_description"`
		UseOverride         *bool   `json:"use_override"` // 是否使用自定义描述
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if updateData.Enabled != nil {
		tool.Enabled = *updateData.Enabled
	}
	if updateData.OverrideDescription != nil {
		tool.OverrideDescription = updateData.OverrideDescription
	}
	if updateData.UseOverride != nil {
		tool.UseOverride = *updateData.UseOverride
	}

	if err := h.toolRepo.Update(tool); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tool"})
		return
	}

	// Refresh all MCP servers after tool update
	h.mcpServer.RefreshAll()

	c.JSON(http.StatusOK, tool)
}

func (h *ToolHandler) Enable(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tool ID"})
		return
	}

	tool, err := h.toolRepo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tool not found"})
		return
	}

	tool.Enabled = true
	if err := h.toolRepo.Update(tool); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enable tool"})
		return
	}

	// Refresh all MCP servers after tool update
	h.mcpServer.RefreshAll()

	c.JSON(http.StatusOK, tool)
}

func (h *ToolHandler) Disable(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tool ID"})
		return
	}

	tool, err := h.toolRepo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tool not found"})
		return
	}

	tool.Enabled = false
	if err := h.toolRepo.Update(tool); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to disable tool"})
		return
	}

	// Refresh all MCP servers after tool update
	h.mcpServer.RefreshAll()

	c.JSON(http.StatusOK, tool)
}

func (h *ToolHandler) Refresh(c *gin.Context) {
	if err := h.serverService.RefreshAllTools(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Refresh all MCP servers after tools refresh
	h.mcpServer.RefreshAll()

	c.JSON(http.StatusOK, gin.H{"message": "tools refreshed successfully"})
}
