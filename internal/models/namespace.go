package models

import (
	"time"
)

type Namespace struct {
	ID          string    `gorm:"primaryKey;size:50" json:"id" binding:"required,alphanum,min=1,max=50"` // 字母数字，用于 URL
	Name        string    `gorm:"not null;size:16" json:"name" binding:"required,min=1,max=16"`          // 显示名称，1-16 字符
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Tools       []Tool    `gorm:"many2many:namespace_tools;" json:"tools,omitempty"`
}
