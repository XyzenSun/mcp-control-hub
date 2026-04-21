package repository

import (
	"mcp-control-hub/internal/models"
	"gorm.io/gorm"
)

type ToolRepository struct {
	db *gorm.DB
}

func NewToolRepository(db *gorm.DB) *ToolRepository {
	return &ToolRepository{db: db}
}

func (r *ToolRepository) Create(tool *models.Tool) error {
	return r.db.Create(tool).Error
}

func (r *ToolRepository) FindByID(id uint) (*models.Tool, error) {
	var tool models.Tool
	if err := r.db.Preload("Server").Preload("Namespaces").First(&tool, id).Error; err != nil {
		return nil, err
	}
	return &tool, nil
}

func (r *ToolRepository) FindByServerAndName(serverID uint, name string) (*models.Tool, error) {
	var tool models.Tool
	if err := r.db.Where("server_id = ? AND name = ?", serverID, name).First(&tool).Error; err != nil {
		return nil, err
	}
	return &tool, nil
}

func (r *ToolRepository) FindAll() ([]models.Tool, error) {
	var tools []models.Tool
	if err := r.db.Preload("Server").Find(&tools).Error; err != nil {
		return nil, err
	}
	return tools, nil
}

func (r *ToolRepository) FindAllEnabled() ([]models.Tool, error) {
	var tools []models.Tool
	if err := r.db.Preload("Server").Where("enabled = ?", true).Find(&tools).Error; err != nil {
		return nil, err
	}
	return tools, nil
}

func (r *ToolRepository) FindByNamespaceID(namespaceID string) ([]models.Tool, error) {
	var tools []models.Tool
	if err := r.db.Preload("Server").
		Joins("JOIN namespace_tools ON namespace_tools.tool_id = tools.id").
		Where("namespace_tools.namespace_id = ?", namespaceID).
		Find(&tools).Error; err != nil {
		return nil, err
	}
	return tools, nil
}

func (r *ToolRepository) FindByServerID(serverID uint) ([]models.Tool, error) {
	var tools []models.Tool
	if err := r.db.Where("server_id = ?", serverID).Find(&tools).Error; err != nil {
		return nil, err
	}
	return tools, nil
}

func (r *ToolRepository) Update(tool *models.Tool) error {
	return r.db.Save(tool).Error
}

func (r *ToolRepository) Delete(id uint) error {
	return r.db.Delete(&models.Tool{}, id).Error
}

func (r *ToolRepository) DeleteByServerID(serverID uint) error {
	return r.db.Where("server_id = ?", serverID).Delete(&models.Tool{}).Error
}
