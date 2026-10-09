package repository

import (
	"errors"

	"core-service/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// WalletRepository provides access to wallet records.
type WalletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

// Create inserts a wallet within the given tx or the base connection if nil.
func (r *WalletRepository) Create(tx *gorm.DB, wallet *model.Wallet) error {
	return r.conn(tx).Create(wallet).Error
}

// GetByUserID returns the wallet owned by the user or ErrNotFound.
func (r *WalletRepository) GetByUserID(userID uuid.UUID) (*model.Wallet, error) {
	var wallet model.Wallet
	err := r.db.Where("user_id = ?", userID).First(&wallet).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// GetByUserIDForUpdate returns the wallet with a row-level lock (SELECT ... FOR
// UPDATE). Must be called inside a transaction to avoid balance races.
func (r *WalletRepository) GetByUserIDForUpdate(tx *gorm.DB, userID uuid.UUID) (*model.Wallet, error) {
	var wallet model.Wallet
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).First(&wallet).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// UpdateBalance persists a wallet's balance within the given tx or base conn.
func (r *WalletRepository) UpdateBalance(tx *gorm.DB, walletID uuid.UUID, balance float64) error {
	return r.conn(tx).Model(&model.Wallet{}).
		Where("id = ?", walletID).
		Update("balance", balance).Error
}

func (r *WalletRepository) conn(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}
