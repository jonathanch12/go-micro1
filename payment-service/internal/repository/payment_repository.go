package repository

import (
	"errors"
	"time"

	"payment-service/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrNotFound is returned when a payment session does not exist.
var ErrNotFound = errors.New("record not found")

// PaymentRepository provides access to payment session records.
type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// Create inserts a payment session.
func (r *PaymentRepository) Create(session *model.PaymentSession) error {
	return r.db.Create(session).Error
}

// Update persists changes to a payment session.
func (r *PaymentRepository) Update(session *model.PaymentSession) error {
	return r.db.Save(session).Error
}

// FindByID returns a payment session by ID or ErrNotFound.
func (r *PaymentRepository) FindByID(id uuid.UUID) (*model.PaymentSession, error) {
	var s model.PaymentSession
	err := r.db.First(&s, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// MarkExpiredBatch marks all pending sessions past their expiry as expired in
// a single UPDATE and returns the number affected. Balances are never touched.
func (r *PaymentRepository) MarkExpiredBatch(now time.Time) (int64, error) {
	res := r.db.Model(&model.PaymentSession{}).
		Where("status = ? AND expires_at < ?", model.StatusPending, now).
		Update("status", model.StatusExpired)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}
