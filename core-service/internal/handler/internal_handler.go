package handler

import (
	"errors"
	"net/http"

	"core-service/internal/model"
	"core-service/internal/service"

	"github.com/gin-gonic/gin"
)

// InternalHandler serves service-to-service routes consumed by the Payment
// Service. These are protected by the internal API key, not the user JWT.
type InternalHandler struct {
	txs *service.TransactionService
}

func NewInternalHandler(txs *service.TransactionService) *InternalHandler {
	return &InternalHandler{txs: txs}
}

// DebitWallet handles POST /internal/wallets/debit. It atomically debits the
// user's wallet and records a completed payment ledger entry so the payment
// still appears in the user's transaction history.
func (h *InternalHandler) DebitWallet(c *gin.Context) {
	var req model.WalletDebitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	ledgerID, newBalance, err := h.txs.DebitForPayment(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInsufficientFunds),
			errors.Is(err, service.ErrInvalidAmount):
			respondError(c, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrWalletNotFound):
			respondError(c, http.StatusNotFound, err.Error())
		default:
			respondError(c, http.StatusInternalServerError, "debit failed")
		}
		return
	}

	c.JSON(http.StatusOK, model.WalletDebitResponse{
		TransactionID: ledgerID,
		NewBalance:    newBalance,
	})
}
