package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"mcp-control-hub/internal/mcp/registry"
	"mcp-control-hub/internal/mcp/server"
	"mcp-control-hub/internal/models"
	"mcp-control-hub/internal/repository"
)

type NamespaceHandler struct {
	namespaceRepo *repository.NamespaceRepository
	toolRegistry  *registry.ToolRegistry
	mcpServer     *server.MCPServer
}

func NewNamespaceHandler(namespaceRepo *repository.NamespaceRepository, toolRegistry *registry.ToolRegistry, mcpServer *server.MCPServer) *NamespaceHandler {
	return &NamespaceHandler{
		namespaceRepo: namespaceRepo,
		toolRegistry:  toolRegistry,
		mcpServer:     mcpServer,
	}
}

func (h *NamespaceHandler) Create(c *gin.Context) {
	var namespace models.Namespace
	if err := c.ShouldBindJSON(&namespace); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.namespaceRepo.Create(&namespace); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create namespace"})
		return
	}

	c.JSON(http.StatusCreated, namespace)
}

func (h *NamespaceHandler) List(c *gin.Context) {
	namespaces, err := h.namespaceRepo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list namespaces"})
		return
	}

	c.JSON(http.StatusOK, namespaces)
}

func (h *NamespaceHandler) Get(c *gin.Context) {
	id := c.Param("id")

	namespace, err := h.namespaceRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "namespace not found"})
		return
	}

	c.JSON(http.StatusOK, namespace)
}

func (h *NamespaceHandler) Update(c *gin.Context) {
	id := c.Param("id")

	namespace, err := h.namespaceRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "namespace not found"})
		return
	}

	var updateData struct {
		Name        string `json:"name" binding:"required,min=1,max=16"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	namespace.Name = updateData.Name
	namespace.Description = updateData.Description

	if err := h.namespaceRepo.Update(namespace); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update namespace"})
		return
	}

	c.JSON(http.StatusOK, namespace)
}

func (h *NamespaceHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.namespaceRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete namespace"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "namespace deleted"})
}

func (h *NamespaceHandler) AddTool(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		ToolID uint `json:"tool_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.namespaceRepo.AddTool(id, req.ToolID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add tool to namespace"})
		return
	}

	// Refresh MCP server for this namespace
	h.mcpServer.RefreshNamespace(id)

	c.JSON(http.StatusOK, gin.H{"message": "tool added to namespace"})
}

func (h *NamespaceHandler) RemoveTool(c *gin.Context) {
	id := c.Param("id")

	toolID, err := strconv.ParseUint(c.Param("tool_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tool ID"})
		return
	}

	if err := h.namespaceRepo.RemoveTool(id, uint(toolID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove tool from namespace"})
		return
	}

	// Refresh MCP server for this namespace
	h.mcpServer.RefreshNamespace(id)

	c.JSON(http.StatusOK, gin.H{"message": "tool removed from namespace"})
}

func (h *NamespaceHandler) ListTools(c *gin.Context) {
	id := c.Param("id")

	namespace, err := h.namespaceRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "namespace not found"})
		return
	}

	c.JSON(http.StatusOK, namespace.Tools)
}
