package repository

import (
	"errors"

	"ewallet/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WalletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) Create(tx *gorm.DB, wallet *model.Wallet) error {
	return r.conn(tx).Create(wallet).Error
}

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
