package repository

import (
	"mcp-control-hub/internal/models"
	"gorm.io/gorm"
)

type NamespaceRepository struct {
	db *gorm.DB
}

func NewNamespaceRepository(db *gorm.DB) *NamespaceRepository {
	return &NamespaceRepository{db: db}
}

func (r *NamespaceRepository) Create(namespace *models.Namespace) error {
	return r.db.Create(namespace).Error
}

func (r *NamespaceRepository) FindByID(id string) (*models.Namespace, error) {
	var namespace models.Namespace
	if err := r.db.Preload("Tools").Preload("Tools.Server").Where("id = ?", id).First(&namespace).Error; err != nil {
		return nil, err
	}
	return &namespace, nil
}

func (r *NamespaceRepository) FindByName(name string) (*models.Namespace, error) {
	var namespace models.Namespace
	if err := r.db.Preload("Tools").Preload("Tools.Server").Where("name = ?", name).First(&namespace).Error; err != nil {
		return nil, err
	}
	return &namespace, nil
}

func (r *NamespaceRepository) FindAll() ([]models.Namespace, error) {
	var namespaces []models.Namespace
	if err := r.db.Find(&namespaces).Error; err != nil {
		return nil, err
	}
	return namespaces, nil
}

func (r *NamespaceRepository) Update(namespace *models.Namespace) error {
	return r.db.Save(namespace).Error
}

func (r *NamespaceRepository) Delete(id string) error {
	return r.db.Select("Tools").Delete(&models.Namespace{ID: id}).Error
}

func (r *NamespaceRepository) AddTool(namespaceID string, toolID uint) error {
	return r.db.Exec("INSERT INTO namespace_tools (namespace_id, tool_id) VALUES (?, ?) ON CONFLICT DO NOTHING", namespaceID, toolID).Error
}

func (r *NamespaceRepository) RemoveTool(namespaceID string, toolID uint) error {
	return r.db.Exec("DELETE FROM namespace_tools WHERE namespace_id = ? AND tool_id = ?", namespaceID, toolID).Error
}
