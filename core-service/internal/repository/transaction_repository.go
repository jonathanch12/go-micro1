package repository

import (
	"errors"

	"core-service/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TransactionRepository provides access to transaction records.
type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Create inserts a transaction within the given tx or the base connection.
func (r *TransactionRepository) Create(tx *gorm.DB, t *model.Transaction) error {
	return r.conn(tx).Create(t).Error
}

// FindByID returns a transaction by ID or ErrNotFound.
func (r *TransactionRepository) FindByID(tx *gorm.DB, id uuid.UUID) (*model.Transaction, error) {
	var t model.Transaction
	err := r.conn(tx).First(&t, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// UpdateStatus sets the status of a transaction.
func (r *TransactionRepository) UpdateStatus(tx *gorm.DB, id uuid.UUID, status model.TransactionStatus) error {
	return r.conn(tx).Model(&model.Transaction{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// ListByUser returns a page of a user's transactions (newest first) plus the
// total count for pagination metadata.
func (r *TransactionRepository) ListByUser(userID uuid.UUID, page, limit int) ([]model.Transaction, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	var total int64
	if err := r.db.Model(&model.Transaction{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var txs []model.Transaction
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&txs).Error
	if err != nil {
		return nil, 0, err
	}
	return txs, total, nil
}

// FindPendingTransfers returns pending transfer_out transactions awaiting
// async settlement.
func (r *TransactionRepository) FindPendingTransfers(limit int) ([]model.Transaction, error) {
	var txs []model.Transaction
	err := r.db.Where("type = ? AND status = ?", model.TypeTransferOut, model.StatusPending).
		Order("created_at ASC").
		Limit(limit).
		Find(&txs).Error
	if err != nil {
		return nil, err
	}
	return txs, nil
}

func (r *TransactionRepository) conn(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}
