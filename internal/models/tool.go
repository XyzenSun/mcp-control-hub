package models

import (
	"time"
)

type Tool struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	ServerID            uint           `gorm:"not null;index" json:"server_id"`
	Name                string         `gorm:"not null;index" json:"name"`
	OriginalDescription string         `gorm:"type:text" json:"original_description"`
	OverrideDescription *string        `gorm:"type:text" json:"override_description"`
	UseOverride         bool           `gorm:"default:false" json:"use_override"` // 是否使用自定义描述
	InputSchema         string         `gorm:"type:text" json:"input_schema"`
	Enabled             bool           `gorm:"default:true" json:"enabled"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	Server              UpstreamServer `gorm:"foreignKey:ServerID" json:"server,omitempty"`
	Namespaces          []Namespace    `gorm:"many2many:namespace_tools;" json:"namespaces,omitempty"`
}

// TableName specifies the table name
func (Tool) TableName() string {
	return "tools"
}
