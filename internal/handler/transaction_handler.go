package handler

import (
	"errors"
	"net/http"
	"strconv"

	"ewallet/internal/middleware"
	"ewallet/internal/model"
	"ewallet/internal/service"

	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	txs *service.TransactionService
}

func NewTransactionHandler(txs *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{txs: txs}
}

func (h *TransactionHandler) TopUp(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req model.TopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	t, newBalance, err := h.txs.TopUp(userID, req.Amount)
	if err != nil {
		h.respondTxError(c, err)
		return
	}

	c.JSON(http.StatusOK, model.TopUpResponse{
		TransactionID: t.ID.String(),
		Message:       "Top-up successful",
		NewBalance:    newBalance,
	})
}

func (h *TransactionHandler) Transfer(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req model.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	t, err := h.txs.Transfer(userID, req.ToUserID, req.Amount)
	if err != nil {
		h.respondTxError(c, err)
		return
	}

	c.JSON(http.StatusAccepted, model.TransferResponse{
		TransactionID: t.ID.String(),
		Status:        string(t.Status),
		Message:       "Transfer initiated successfully.",
	})
}

func (h *TransactionHandler) Pay(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req model.PayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	t, newBalance, err := h.txs.Pay(userID, req.MerchantID, req.Amount, req.Description)
	if err != nil {
		h.respondTxError(c, err)
		return
	}

	c.JSON(http.StatusOK, model.PayResponse{
		TransactionID: t.ID.String(),
		Message:       "Payment successful",
		NewBalance:    newBalance,
	})
}

func (h *TransactionHandler) History(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	page := parseIntDefault(c.Query("page"), 1)
	limit := parseIntDefault(c.Query("limit"), 10)

	resp, err := h.txs.History(userID, page, limit)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to fetch history")
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *TransactionHandler) respondTxError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInsufficientFunds),
		errors.Is(err, service.ErrInvalidAmount),
		errors.Is(err, service.ErrRecipientNotFound),
		errors.Is(err, service.ErrSelfTransfer):
		respondError(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrWalletNotFound):
		respondError(c, http.StatusNotFound, err.Error())
	default:
		respondError(c, http.StatusInternalServerError, "transaction failed")
	}
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 1 {
		return def
	}
	return v
}
