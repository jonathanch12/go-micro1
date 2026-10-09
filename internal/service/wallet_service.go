package service

import (
	"errors"

	"ewallet/internal/repository"

	"github.com/google/uuid"
)

type WalletService struct {
	wallets *repository.WalletRepository
}

func NewWalletService(wallets *repository.WalletRepository) *WalletService {
	return &WalletService{wallets: wallets}
}

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
