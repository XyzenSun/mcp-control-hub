package repository

import (
	"mcp-control-hub/internal/models"
	"gorm.io/gorm"
)

type ServerRepository struct {
	db *gorm.DB
}

func NewServerRepository(db *gorm.DB) *ServerRepository {
	return &ServerRepository{db: db}
}

func (r *ServerRepository) Create(server *models.UpstreamServer) error {
	return r.db.Create(server).Error
}

func (r *ServerRepository) FindByID(id uint) (*models.UpstreamServer, error) {
	var server models.UpstreamServer
	if err := r.db.Preload("Tools").First(&server, id).Error; err != nil {
		return nil, err
	}
	return &server, nil
}

func (r *ServerRepository) FindByName(name string) (*models.UpstreamServer, error) {
	var server models.UpstreamServer
	if err := r.db.Where("name = ?", name).First(&server).Error; err != nil {
		return nil, err
	}
	return &server, nil
}

func (r *ServerRepository) FindAll() ([]models.UpstreamServer, error) {
	var servers []models.UpstreamServer
	if err := r.db.Find(&servers).Error; err != nil {
		return nil, err
	}
	return servers, nil
}

func (r *ServerRepository) FindAllEnabled() ([]models.UpstreamServer, error) {
	var servers []models.UpstreamServer
	if err := r.db.Where("enabled = ?", true).Find(&servers).Error; err != nil {
		return nil, err
	}
	return servers, nil
}

func (r *ServerRepository) Update(server *models.UpstreamServer) error {
	return r.db.Save(server).Error
}

func (r *ServerRepository) Delete(id uint) error {
	return r.db.Delete(&models.UpstreamServer{}, id).Error
}
