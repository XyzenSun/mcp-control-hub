package repository

import (
	"errors"
	"fmt"
	"os"

	"mcp-control-hub/internal/models"
	"mcp-control-hub/pkg/utils"
	"gorm.io/gorm"
)

type APIKeyRepository struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) Create(apiKey *models.APIKey) error {
	return r.db.Create(apiKey).Error
}

func (r *APIKeyRepository) FindByHash(hash string) (*models.APIKey, error) {
	var apiKey models.APIKey
	if err := r.db.Where("key_hash = ?", hash).First(&apiKey).Error; err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (r *APIKeyRepository) FindAll() ([]models.APIKey, error) {
	var apiKeys []models.APIKey
	if err := r.db.Find(&apiKeys).Error; err != nil {
		return nil, err
	}
	return apiKeys, nil
}

func (r *APIKeyRepository) UpdateLastUsed(id uint) error {
	return r.db.Model(&models.APIKey{}).Where("id = ?", id).Update("last_used", gorm.Expr("NOW()")).Error
}

func (r *APIKeyRepository) Delete(id uint) error {
	return r.db.Delete(&models.APIKey{}, id).Error
}

// InitializeBootstrapKey initializes a bootstrap API key from environment variable
// if no API keys exist in the database.
// Returns error if no API keys exist and BOOTSTRAP_API_KEY env var is not set.
func (r *APIKeyRepository) InitializeBootstrapKey() error {
	// Check if any API key exists
	var count int64
	if err := r.db.Model(&models.APIKey{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count api keys: %w", err)
	}

	if count > 0 {
		return nil // Already have keys, skip initialization
	}

	// Get bootstrap key from environment
	apiKey := os.Getenv("BOOTSTRAP_API_KEY")
	if apiKey == "" {
		return errors.New("no API keys found in database and BOOTSTRAP_API_KEY environment variable is not set")
	}

	// Hash and store the key
	keyHash := utils.HashAPIKey(apiKey)
	bootstrapKey := &models.APIKey{
		Name:    "bootstrap",
		KeyHash: keyHash,
	}

	if err := r.Create(bootstrapKey); err != nil {
		return fmt.Errorf("failed to create bootstrap api key: %w", err)
	}

	fmt.Printf("\n==========================================\n")
	fmt.Printf("Bootstrap API Key initialized\n")
	fmt.Printf("==========================================\n\n")

	return nil
}
