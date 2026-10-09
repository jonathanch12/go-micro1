package service

import (
	"time"

	"ewallet/internal/model"
)

func (s *TransactionService) ProcessPendingTransfers(batch int) (int, error) {
	pending, err := s.txs.FindPendingTransfers(batch)
	if err != nil {
		return 0, err
	}
	processed := 0
	for _, t := range pending {
		if err := s.SettleTransfer(t.ID); err != nil {
			return processed, err
		}
		processed++
	}
	return processed, nil
}

func (s *TransactionService) ExpireStalePayments(batch int) (int, error) {
	stale, err := s.txs.FindExpiredPendingPayments(time.Now(), batch)
	if err != nil {
		return 0, err
	}
	expired := 0
	for _, t := range stale {
		if err := s.txs.UpdateStatus(nil, t.ID, model.StatusExpired); err != nil {
			return expired, err
		}
		expired++
	}
	return expired, nil
}
