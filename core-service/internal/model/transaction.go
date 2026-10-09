package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TransactionType enumerates the kinds of ledger entries.
type TransactionType string

const (
	TypeTopup       TransactionType = "topup"
	TypeTransferIn  TransactionType = "transfer_in"
	TypeTransferOut TransactionType = "transfer_out"
	TypePayment     TransactionType = "payment"
)

// TransactionStatus enumerates the lifecycle states of a transaction.
type TransactionStatus string

const (
	StatusCompleted TransactionStatus = "completed"
	StatusPending   TransactionStatus = "pending"
	StatusFailed    TransactionStatus = "failed"
)

// Transaction is a single ledger entry belonging to a user.
//
// For transfers two rows are created once settled: a transfer_out for the
// sender and a transfer_in for the receiver, linked by GroupID. Payment rows
// are written when the Payment Service debits a wallet via the internal API,
// so they still appear in a user's transaction history.
type Transaction struct {
	ID          uuid.UUID         `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID         `gorm:"type:uuid;index;not null"`
	Type        TransactionType   `gorm:"type:varchar(20);not null"`
	Amount      float64           `gorm:"type:numeric(20,2);not null"`
	From        string            `gorm:"type:varchar(255)"`
	To          string            `gorm:"type:varchar(255)"`
	Status      TransactionStatus `gorm:"type:varchar(20);not null;index"`
	Description string            `gorm:"type:text"`

	// GroupID links the two legs of a transfer together. Nil for other types.
	GroupID *uuid.UUID `gorm:"type:uuid;index"`

	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
}

// BeforeCreate assigns a UUID if one has not been set.
func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
