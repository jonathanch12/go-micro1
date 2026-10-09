package service

import (
	"errors"
	"math"

	"core-service/internal/model"
	"core-service/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TransactionService implements top-up, transfer, history and the internal
// payment-debit logic. Payment sessions themselves live in the Payment
// Service; this service only moves money and records ledger entries.
type TransactionService struct {
	db      *gorm.DB
	users   *repository.UserRepository
	wallets *repository.WalletRepository
	txs     *repository.TransactionRepository
}

func NewTransactionService(
	db *gorm.DB,
	users *repository.UserRepository,
	wallets *repository.WalletRepository,
	txs *repository.TransactionRepository,
) *TransactionService {
	return &TransactionService{db: db, users: users, wallets: wallets, txs: txs}
}

// TopUp credits the user's wallet and records a completed topup transaction.
func (s *TransactionService) TopUp(userID uuid.UUID, amount float64) (*model.Transaction, float64, error) {
	if amount <= 0 {
		return nil, 0, ErrInvalidAmount
	}

	var created *model.Transaction
	var newBalance float64

	err := s.db.Transaction(func(tx *gorm.DB) error {
		wallet, err := s.wallets.GetByUserIDForUpdate(tx, userID)
		if err != nil {
			return mapWalletErr(err)
		}

		newBalance = round2(wallet.Balance + amount)
		if err := s.wallets.UpdateBalance(tx, wallet.ID, newBalance); err != nil {
			return err
		}

		t := &model.Transaction{
			UserID: userID,
			Type:   model.TypeTopup,
			Amount: amount,
			To:     userID.String(),
			Status: model.StatusCompleted,
		}
		if err := s.txs.Create(tx, t); err != nil {
			return err
		}
		created = t
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	return created, newBalance, nil
}

// Transfer validates and records a PENDING transfer_out leg, then returns.
// Settlement happens asynchronously via SettleTransfer.
func (s *TransactionService) Transfer(fromUserID uuid.UUID, toUserIDStr string, amount float64) (*model.Transaction, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	toUserID, err := uuid.Parse(toUserIDStr)
	if err != nil {
		return nil, ErrRecipientNotFound
	}
	if toUserID == fromUserID {
		return nil, ErrSelfTransfer
	}

	if _, err := s.users.FindByID(toUserID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrRecipientNotFound
		}
		return nil, err
	}

	var pending *model.Transaction
	err = s.db.Transaction(func(tx *gorm.DB) error {
		wallet, err := s.wallets.GetByUserIDForUpdate(tx, fromUserID)
		if err != nil {
			return mapWalletErr(err)
		}
		if wallet.Balance < amount {
			return ErrInsufficientFunds
		}

		groupID := uuid.New()
		t := &model.Transaction{
			UserID:  fromUserID,
			Type:    model.TypeTransferOut,
			Amount:  amount,
			From:    fromUserID.String(),
			To:      toUserID.String(),
			Status:  model.StatusPending,
			GroupID: &groupID,
		}
		if err := s.txs.Create(tx, t); err != nil {
			return err
		}
		pending = t
		return nil
	})
	if err != nil {
		return nil, err
	}
	return pending, nil
}

// SettleTransfer completes a pending transfer_out atomically, debiting the
// sender, crediting the receiver and writing the transfer_in leg.
func (s *TransactionService) SettleTransfer(transferOutID uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		out, err := s.txs.FindByID(tx, transferOutID)
		if err != nil {
			return err
		}
		if out.Status != model.StatusPending {
			return nil
		}

		fromID, err := uuid.Parse(out.From)
		if err != nil {
			return s.txs.UpdateStatus(tx, out.ID, model.StatusFailed)
		}
		toID, err := uuid.Parse(out.To)
		if err != nil {
			return s.txs.UpdateStatus(tx, out.ID, model.StatusFailed)
		}

		senderWallet, err := s.wallets.GetByUserIDForUpdate(tx, fromID)
		if err != nil {
			return s.txs.UpdateStatus(tx, out.ID, model.StatusFailed)
		}
		receiverWallet, err := s.wallets.GetByUserIDForUpdate(tx, toID)
		if err != nil {
			return s.txs.UpdateStatus(tx, out.ID, model.StatusFailed)
		}

		if senderWallet.Balance < out.Amount {
			return s.txs.UpdateStatus(tx, out.ID, model.StatusFailed)
		}

		newSender := round2(senderWallet.Balance - out.Amount)
		newReceiver := round2(receiverWallet.Balance + out.Amount)
		if err := s.wallets.UpdateBalance(tx, senderWallet.ID, newSender); err != nil {
			return err
		}
		if err := s.wallets.UpdateBalance(tx, receiverWallet.ID, newReceiver); err != nil {
			return err
		}

		in := &model.Transaction{
			UserID:  toID,
			Type:    model.TypeTransferIn,
			Amount:  out.Amount,
			From:    out.From,
			To:      out.To,
			Status:  model.StatusCompleted,
			GroupID: out.GroupID,
		}
		if err := s.txs.Create(tx, in); err != nil {
			return err
		}

		return s.txs.UpdateStatus(tx, out.ID, model.StatusCompleted)
	})
}

// DebitForPayment is invoked by the Payment Service (via the internal API) to
// debit a user's wallet for a completed payment and record the payment ledger
// entry, all atomically. It returns the ledger transaction ID and new balance.
func (s *TransactionService) DebitForPayment(req model.WalletDebitRequest) (string, float64, error) {
	if req.Amount <= 0 {
		return "", 0, ErrInvalidAmount
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return "", 0, ErrWalletNotFound
	}

	var ledgerID string
	var newBalance float64

	err = s.db.Transaction(func(tx *gorm.DB) error {
		wallet, err := s.wallets.GetByUserIDForUpdate(tx, userID)
		if err != nil {
			return mapWalletErr(err)
		}
		if wallet.Balance < req.Amount {
			return ErrInsufficientFunds
		}

		newBalance = round2(wallet.Balance - req.Amount)
		if err := s.wallets.UpdateBalance(tx, wallet.ID, newBalance); err != nil {
			return err
		}

		t := &model.Transaction{
			UserID:      userID,
			Type:        model.TypePayment,
			Amount:      req.Amount,
			From:        userID.String(),
			To:          req.MerchantID,
			Status:      model.StatusCompleted,
			Description: req.Description,
		}
		if err := s.txs.Create(tx, t); err != nil {
			return err
		}
		ledgerID = t.ID.String()
		return nil
	})
	if err != nil {
		return "", 0, err
	}
	return ledgerID, newBalance, nil
}

// History returns a paginated view of the user's transactions.
func (s *TransactionService) History(userID uuid.UUID, page, limit int) (*model.TransactionHistoryResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	txs, total, err := s.txs.ListByUser(userID, page, limit)
	if err != nil {
		return nil, err
	}

	items := make([]model.TransactionHistoryItem, 0, len(txs))
	for _, t := range txs {
		items = append(items, model.TransactionHistoryItem{
			TransactionID: t.ID.String(),
			Type:          string(t.Type),
			Amount:        t.Amount,
			From:          t.From,
			To:            t.To,
			Timestamp:     t.CreatedAt,
			Status:        string(t.Status),
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	return &model.TransactionHistoryResponse{
		Transactions: items,
		Pagination: model.Pagination{
			CurrentPage: page,
			TotalPages:  totalPages,
			TotalItems:  total,
		},
	}, nil
}

// ProcessPendingTransfers settles a batch of pending transfers for the worker.
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

func mapWalletErr(err error) error {
	if errors.Is(err, repository.ErrNotFound) {
		return ErrWalletNotFound
	}
	return err
}

// round2 rounds a monetary value to 2 decimal places.
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
