package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"mcp-control-hub/internal/repository"
)

type ConfigHandler struct {
	configRepo *repository.ConfigRepository
}

func NewConfigHandler(configRepo *repository.ConfigRepository) *ConfigHandler {
	return &ConfigHandler{
		configRepo: configRepo,
	}
}

// List returns all configuration items with definitions
func (h *ConfigHandler) List(c *gin.Context) {
	configs, err := h.configRepo.GetAllWithDefinitions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, configs)
}

// Get returns a single config value
func (h *ConfigHandler) Get(c *gin.Context) {
	key := c.Param("key")

	var request struct {
		Value string `json:"value"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.configRepo.Set(key, request.Value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Config updated"})
}

// Update updates a config value
func (h *ConfigHandler) Update(c *gin.Context) {
	key := c.Param("key")

	var request struct {
		Value string `json:"value"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.configRepo.Set(key, request.Value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Config updated", "key": key, "value": request.Value})
}

// UpdateAll updates multiple config values
func (h *ConfigHandler) UpdateAll(c *gin.Context) {
	var request map[string]string

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for key, value := range request {
		if err := h.configRepo.Set(key, value); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Configs updated"})
}