package service

import (
	"errors"

	"core-service/internal/repository"

	"github.com/google/uuid"
)

// WalletService handles wallet balance queries.
type WalletService struct {
	wallets *repository.WalletRepository
}

func NewWalletService(wallets *repository.WalletRepository) *WalletService {
	return &WalletService{wallets: wallets}
}

// GetBalance returns the balance for the given user's wallet.
func (s *WalletService) GetBalance(userID uuid.UUID) (float64, error) {
	wallet, err := s.wallets.GetByUserID(userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return 0, ErrWalletNotFound
		}
		return 0, err
	}
	return wallet.Balance, nil
}
