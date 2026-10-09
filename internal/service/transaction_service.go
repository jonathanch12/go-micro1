package service

import (
	"errors"
	"math"
	"time"

	"ewallet/internal/model"
	"ewallet/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const PaymentExpiry = 10 * time.Minute

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

func (s *TransactionService) Pay(userID uuid.UUID, merchantID string, amount float64, description string) (*model.Transaction, float64, error) {
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
		if wallet.Balance < amount {
			return ErrInsufficientFunds
		}

		expiresAt := time.Now().Add(PaymentExpiry)
		t := &model.Transaction{
			UserID:      userID,
			Type:        model.TypePayment,
			Amount:      amount,
			From:        userID.String(),
			To:          merchantID,
			Status:      model.StatusPending,
			Description: description,
			ExpiresAt:   &expiresAt,
		}
		if err := s.txs.Create(tx, t); err != nil {
			return err
		}

		newBalance = round2(wallet.Balance - amount)
		if err := s.wallets.UpdateBalance(tx, wallet.ID, newBalance); err != nil {
			return err
		}
		if err := s.txs.UpdateStatus(tx, t.ID, model.StatusCompleted); err != nil {
			return err
		}
		t.Status = model.StatusCompleted
		created = t
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	return created, newBalance, nil
}

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

func mapWalletErr(err error) error {
	if errors.Is(err, repository.ErrNotFound) {
		return ErrWalletNotFound
	}
	return err
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
