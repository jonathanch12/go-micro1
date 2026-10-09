package worker

import (
	"context"
	"log"
	"time"

	"payment-service/internal/service"
)

// Manager runs the payment-session expiry sweeper using a time.Ticker.
type Manager struct {
	paymentService *service.PaymentService
	expiryInterval time.Duration
}

// NewManager builds a worker Manager with a sensible default interval.
func NewManager(paymentService *service.PaymentService) *Manager {
	return &Manager{
		paymentService: paymentService,
		expiryInterval: 30 * time.Second,
	}
}

// Start launches the expiry sweeper as a goroutine until ctx is cancelled.
func (m *Manager) Start(ctx context.Context) {
	go m.runExpirySweeper(ctx)
	log.Println("payment expiry sweeper started")
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
			if n, err := m.paymentService.ExpireStaleSessions(); err != nil {
				log.Printf("expiry sweeper: %v", err)
			} else if n > 0 {
				log.Printf("expiry sweeper: expired %d payment session(s)", n)
			}
		}
	}
}
