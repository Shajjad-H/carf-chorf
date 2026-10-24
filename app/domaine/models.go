package domaine

import (
	"time"

	"gorm.io/gorm"
)

type GormModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type User struct {
	GormModel
	Email    string `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Credits  uint   `gorm:"not null;default:0" json:"credits"`
	IsActive bool   `gorm:"not null;default:true" json:"is_active"`
}

func MigrationModels() []any {
	return []any{&User{}}
}
