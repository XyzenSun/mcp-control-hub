package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"mcp-control-hub/internal/models"
	"mcp-control-hub/internal/repository"
	"mcp-control-hub/internal/service"
)

type ServerHandler struct {
	serverRepo    *repository.ServerRepository
	serverService *service.ServerService
}

func NewServerHandler(serverRepo *repository.ServerRepository, serverService *service.ServerService) *ServerHandler {
	return &ServerHandler{
		serverRepo:    serverRepo,
		serverService: serverService,
	}
}

func (h *ServerHandler) Create(c *gin.Context) {
	var server models.UpstreamServer
	if err := c.ShouldBindJSON(&server); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.serverRepo.Create(&server); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create server"})
		return
	}

	c.JSON(http.StatusCreated, server)
}

func (h *ServerHandler) List(c *gin.Context) {
	servers, err := h.serverRepo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list servers"})
		return
	}

	c.JSON(http.StatusOK, servers)
}

func (h *ServerHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid server ID"})
		return
	}

	server, err := h.serverRepo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return
	}

	c.JSON(http.StatusOK, server)
}

func (h *ServerHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid server ID"})
		return
	}

	server, err := h.serverRepo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return
	}

	if err := c.ShouldBindJSON(server); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.serverRepo.Update(server); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update server"})
		return
	}

	c.JSON(http.StatusOK, server)
}

func (h *ServerHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid server ID"})
		return
	}

	if err := h.serverRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete server"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "server deleted"})
}

func (h *ServerHandler) Enable(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid server ID"})
		return
	}

	server, err := h.serverRepo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return
	}

	server.Enabled = true
	if err := h.serverRepo.Update(server); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enable server"})
		return
	}

	c.JSON(http.StatusOK, server)
}

func (h *ServerHandler) Disable(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid server ID"})
		return
	}

	server, err := h.serverRepo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return
	}

	server.Enabled = false
	if err := h.serverRepo.Update(server); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to disable server"})
		return
	}

	c.JSON(http.StatusOK, server)
}

func (h *ServerHandler) Sync(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid server ID"})
		return
	}

	if err := h.serverService.SyncTools(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tools synced successfully"})
}
