package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionType string

const (
	TypeTopup       TransactionType = "topup"
	TypeTransferIn  TransactionType = "transfer_in"
	TypeTransferOut TransactionType = "transfer_out"
	TypePayment     TransactionType = "payment"
)

type TransactionStatus string

const (
	StatusCompleted TransactionStatus = "completed"
	StatusPending   TransactionStatus = "pending"
	StatusFailed    TransactionStatus = "failed"
	StatusExpired   TransactionStatus = "expired"
)

type Transaction struct {
	ID          uuid.UUID         `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID         `gorm:"type:uuid;index;not null"`
	Type        TransactionType   `gorm:"type:varchar(20);not null"`
	Amount      float64           `gorm:"type:numeric(20,2);not null"`
	From        string            `gorm:"type:varchar(255)"`
	To          string            `gorm:"type:varchar(255)"`
	Status      TransactionStatus `gorm:"type:varchar(20);not null;index"`
	Description string            `gorm:"type:text"`

	GroupID *uuid.UUID `gorm:"type:uuid;index"`

	ExpiresAt *time.Time `gorm:"index"`

	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
