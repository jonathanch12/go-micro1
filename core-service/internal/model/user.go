package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User is the account entity. One user owns exactly one wallet.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name         string    `gorm:"type:varchar(255);not null"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string    `gorm:"type:varchar(255);not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time

	Wallet *Wallet `gorm:"constraint:OnDelete:CASCADE;"`
}

// BeforeCreate assigns a UUID if one has not been set.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
