package worker

import (
	"context"
	"log"
	"time"

	"ewallet/internal/service"
)

type Manager struct {
	txService *service.TransactionService

	transferInterval time.Duration
	expiryInterval   time.Duration
	batchSize        int
}

func NewManager(txService *service.TransactionService) *Manager {
	return &Manager{
		txService:        txService,
		transferInterval: 2 * time.Second,
		expiryInterval:   30 * time.Second,
		batchSize:        100,
	}
}

func (m *Manager) Start(ctx context.Context) {
	go m.runTransferWorker(ctx)
	go m.runExpirySweeper(ctx)
	log.Println("background workers started")
}

func (m *Manager) runTransferWorker(ctx context.Context) {
	ticker := time.NewTicker(m.transferInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("transfer worker stopped")
			return
		case <-ticker.C:
			if n, err := m.txService.ProcessPendingTransfers(m.batchSize); err != nil {
				log.Printf("transfer worker: %v", err)
			} else if n > 0 {
				log.Printf("transfer worker: settled %d transfer(s)", n)
			}
		}
	}
}

func (m *Manager) runExpirySweeper(ctx context.Context) {
	ticker := time.NewTicker(m.expiryInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("payment expiry sweeper stopped")
			return
		case <-ticker.C:
			if n, err := m.txService.ExpireStalePayments(m.batchSize); err != nil {
				log.Printf("expiry sweeper: %v", err)
			} else if n > 0 {
				log.Printf("expiry sweeper: expired %d payment session(s)", n)
			}
		}
	}
}
