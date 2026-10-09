package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Wallet struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"`
	Balance   float64   `gorm:"type:numeric(20,2);not null;default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (w *Wallet) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}
