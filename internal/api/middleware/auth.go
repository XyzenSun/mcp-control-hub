package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"mcp-control-hub/internal/repository"
	"mcp-control-hub/pkg/utils"
)

type AuthMiddleware struct {
	apiKeyRepo *repository.APIKeyRepository
}

func NewAuthMiddleware(apiKeyRepo *repository.APIKeyRepository) *AuthMiddleware {
	return &AuthMiddleware{apiKeyRepo: apiKeyRepo}
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing API key"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		apiKey = strings.TrimPrefix(apiKey, "Bearer ")

		// Hash the key and look it up
		keyHash := utils.HashAPIKey(apiKey)
		key, err := m.apiKeyRepo.FindByHash(keyHash)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			c.Abort()
			return
		}

		// Update last used timestamp (async)
		go m.apiKeyRepo.UpdateLastUsed(key.ID)

		c.Set("api_key_id", key.ID)
		c.Next()
	}
}
