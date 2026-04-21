package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"mcp-control-hub/internal/models"
	"mcp-control-hub/internal/repository"
	"mcp-control-hub/pkg/utils"
)

type APIKeyHandler struct {
	apiKeyRepo *repository.APIKeyRepository
	keyLength  int
}

func NewAPIKeyHandler(apiKeyRepo *repository.APIKeyRepository, keyLength int) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyRepo: apiKeyRepo,
		keyLength:  keyLength,
	}
}

func (h *APIKeyHandler) Create(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate API key
	plainKey, err := utils.GenerateAPIKey(h.keyLength)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate API key"})
		return
	}

	// Hash the key
	keyHash := utils.HashAPIKey(plainKey)

	apiKey := &models.APIKey{
		Name:    req.Name,
		KeyHash: keyHash,
	}

	if err := h.apiKeyRepo.Create(apiKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create API key"})
		return
	}

	// Return the plain key only once
	c.JSON(http.StatusCreated, gin.H{
		"id":      apiKey.ID,
		"name":    apiKey.Name,
		"api_key": plainKey,
		"message": "Save this API key - it will not be shown again",
	})
}

func (h *APIKeyHandler) List(c *gin.Context) {
	apiKeys, err := h.apiKeyRepo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list API keys"})
		return
	}

	c.JSON(http.StatusOK, apiKeys)
}

func (h *APIKeyHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid API key ID"})
		return
	}

	if err := h.apiKeyRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete API key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key deleted"})
}
