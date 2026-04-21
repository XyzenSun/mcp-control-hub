package models

import (
	"time"
)

type UpstreamServer struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name" binding:"required"`
	Protocol    string    `gorm:"not null" json:"protocol" binding:"required,oneof=stdio sse streamable"`
	Config      string    `gorm:"type:text" json:"config" binding:"required"`
	Enabled     bool      `gorm:"default:true" json:"enabled"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Tools       []Tool    `gorm:"foreignKey:ServerID" json:"tools,omitempty"`
}
