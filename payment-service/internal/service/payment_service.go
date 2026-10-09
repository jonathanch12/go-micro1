package service

import (
	"errors"
	"time"

	"payment-service/internal/coreclient"
	"payment-service/internal/model"
	"payment-service/internal/repository"

	"github.com/google/uuid"
)

// PaymentExpiry is the lifetime of a payment session before it expires.
const PaymentExpiry = 10 * time.Minute

// PaymentService implements the /transactions/pay flow and payment-session
// expiry. The wallet itself lives in core-service, which this service calls
// to perform the actual debit.
type PaymentService struct {
	payments *repository.PaymentRepository
	core     *coreclient.Client
}

func NewPaymentService(payments *repository.PaymentRepository, core *coreclient.Client) *PaymentService {
	return &PaymentService{payments: payments, core: core}
}

// Pay creates a payment session (expires in 10m) and, per option (a),
// synchronously debits the wallet via core-service and marks it completed.
// If core-service reports insufficient funds, the session is marked failed and
// no balance is deducted.
func (s *PaymentService) Pay(userID uuid.UUID, merchantID string, amount float64, description string) (*model.PaymentSession, float64, error) {
	if amount <= 0 {
		return nil, 0, ErrInvalidAmount
	}

	now := time.Now()
	session := &model.PaymentSession{
		UserID:      userID,
		MerchantID:  merchantID,
		Amount:      amount,
		Description: description,
		Status:      model.StatusPending,
		ExpiresAt:   now.Add(PaymentExpiry),
	}
	if err := s.payments.Create(session); err != nil {
		return nil, 0, err
	}

	// Guard against a session that somehow already lapsed (clock skew / retry).
	if time.Now().After(session.ExpiresAt) {
		session.Status = model.StatusExpired
		_ = s.payments.Update(session)
		return nil, 0, ErrPaymentExpired
	}

	// Debit the wallet via core-service.
	resp, err := s.core.DebitWallet(model.WalletDebitRequest{
		UserID:        userID.String(),
		Amount:        amount,
		MerchantID:    merchantID,
		Description:   description,
		TransactionID: session.ID.String(),
	})
	if err != nil {
		// Mark the session failed; no balance was deducted.
		session.Status = model.StatusFailed
		_ = s.payments.Update(session)

		switch {
		case errors.Is(err, coreclient.ErrInsufficientFunds):
			return nil, 0, ErrInsufficientFunds
		case errors.Is(err, coreclient.ErrWalletNotFound):
			return nil, 0, ErrWalletNotFound
		default:
			return nil, 0, err
		}
	}

	// Success: record the ledger ID and mark completed.
	session.Status = model.StatusCompleted
	session.LedgerTransactionID = resp.TransactionID
	if err := s.payments.Update(session); err != nil {
		return nil, 0, err
	}

	return session, resp.NewBalance, nil
}

// ExpireStaleSessions marks pending sessions past their expiry as expired.
// Returns the number expired. No balance is involved.
func (s *PaymentService) ExpireStaleSessions() (int64, error) {
	return s.payments.MarkExpiredBatch(time.Now())
}
