package worker

import (
	"context"
	"log"
	"time"

	"core-service/internal/service"
)

// Manager runs the async transfer settlement worker using a time.Ticker.
type Manager struct {
	txService *service.TransactionService

	transferInterval time.Duration
	batchSize        int
}

// NewManager builds a worker Manager with a sensible default interval.
func NewManager(txService *service.TransactionService) *Manager {
	return &Manager{
		txService:        txService,
		transferInterval: 2 * time.Second,
		batchSize:        100,
	}
}

// Start launches the transfer worker as a goroutine until ctx is cancelled.
func (m *Manager) Start(ctx context.Context) {
	go m.runTransferWorker(ctx)
	log.Println("transfer settlement worker started")
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
