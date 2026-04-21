package models

import (
	"time"
)

type APIKey struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Name      string     `gorm:"not null" json:"name" binding:"required"`
	KeyHash   string     `gorm:"uniqueIndex;not null" json:"-"`
	LastUsed  *time.Time `json:"last_used"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
