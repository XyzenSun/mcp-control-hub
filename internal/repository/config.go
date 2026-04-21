package repository

import (
	"strconv"

	"gorm.io/gorm"
	"mcp-control-hub/internal/models"
)

type ConfigRepository struct {
	db *gorm.DB
}

func NewConfigRepository(db *gorm.DB) *ConfigRepository {
	return &ConfigRepository{db: db}
}

// Get retrieves a config value by key
func (r *ConfigRepository) Get(key string) (string, error) {
	var config models.SystemConfig
	result := r.db.Where("key = ?", key).First(&config)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Return default value if not found
			if defaultValue, ok := models.DefaultConfigValues[key]; ok {
				return defaultValue, nil
			}
			return "", nil
		}
		return "", result.Error
	}
	return config.Value, nil
}

// GetInt retrieves a config value as integer
func (r *ConfigRepository) GetInt(key string) (int, error) {
	value, err := r.Get(key)
	if err != nil {
		return 0, err
	}

	// If empty, return default
	if value == "" {
		if defaultValue, ok := models.DefaultConfigValues[key]; ok {
			return strconv.Atoi(defaultValue)
		}
		return 0, nil
	}

	return strconv.Atoi(value)
}

// Set updates or creates a config value
func (r *ConfigRepository) Set(key, value string) error {
	var config models.SystemConfig
	result := r.db.Where("key = ?", key).First(&config)

	if result.Error == gorm.ErrRecordNotFound {
		// Create new
		config = models.SystemConfig{
			Key:   key,
			Value: value,
		}
		return r.db.Create(&config).Error
	}

	if result.Error != nil {
		return result.Error
	}

	// Update existing
	return r.db.Model(&config).Update("value", value).Error
}

// SetInt updates or creates a config value from integer
func (r *ConfigRepository) SetInt(key string, value int) error {
	return r.Set(key, strconv.Itoa(value))
}

// GetAll retrieves all config values
func (r *ConfigRepository) GetAll() (map[string]string, error) {
	var configs []models.SystemConfig
	if err := r.db.Find(&configs).Error; err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, config := range configs {
		result[config.Key] = config.Value
	}

	// Fill in defaults for missing keys
	for key, defaultValue := range models.DefaultConfigValues {
		if _, ok := result[key]; !ok {
			result[key] = defaultValue
		}
	}

	return result, nil
}

// GetAllWithDefinitions retrieves all configs with their definitions
func (r *ConfigRepository) GetAllWithDefinitions() ([]models.ConfigDefinition, error) {
	configs, err := r.GetAll()
	if err != nil {
		return nil, err
	}

	definitions := models.GetConfigDefinitions()
	for i := range definitions {
		if value, ok := configs[definitions[i].Key]; ok {
			definitions[i].Value = value
		} else {
			definitions[i].Value = definitions[i].DefaultValue
		}
	}

	return definitions, nil
}

// InitializeDefaults creates default config entries if they don't exist
func (r *ConfigRepository) InitializeDefaults() error {
	for key, value := range models.DefaultConfigValues {
		var config models.SystemConfig
		result := r.db.Where("key = ?", key).First(&config)
		if result.Error == gorm.ErrRecordNotFound {
			config = models.SystemConfig{
				Key:   key,
				Value: value,
			}
			if err := r.db.Create(&config).Error; err != nil {
				return err
			}
		}
	}
	return nil
}