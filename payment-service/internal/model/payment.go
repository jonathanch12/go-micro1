package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentStatus enumerates the lifecycle states of a payment session.
type PaymentStatus string

const (
	StatusPending   PaymentStatus = "pending"
	StatusCompleted PaymentStatus = "completed"
	StatusFailed    PaymentStatus = "failed"
	StatusExpired   PaymentStatus = "expired"
)

// PaymentSession represents a payment initiated at /transactions/pay. It is
// owned entirely by the Payment Service. A session expires 10 minutes after
// creation; expired sessions never result in a wallet debit.
type PaymentSession struct {
	ID          uuid.UUID     `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID     `gorm:"type:uuid;index;not null"`
	MerchantID  string        `gorm:"type:varchar(255);not null"`
	Amount      float64       `gorm:"type:numeric(20,2);not null"`
	Description string        `gorm:"type:text"`
	Status      PaymentStatus `gorm:"type:varchar(20);not null;index"`

	// LedgerTransactionID is the ID of the ledger entry created in core-service
	// once the wallet is successfully debited.
	LedgerTransactionID string `gorm:"type:varchar(64)"`

	ExpiresAt time.Time `gorm:"index;not null"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
}

// BeforeCreate assigns a UUID if one has not been set.
func (p *PaymentSession) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
