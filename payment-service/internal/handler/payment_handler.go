package handler

import (
	"errors"
	"net/http"

	"payment-service/internal/middleware"
	"payment-service/internal/model"
	"payment-service/internal/service"

	"github.com/gin-gonic/gin"
)

// PaymentHandler serves the /transactions/pay route.
type PaymentHandler struct {
	payments *service.PaymentService
}

func NewPaymentHandler(payments *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{payments: payments}
}

// Pay handles POST /transactions/pay.
func (h *PaymentHandler) Pay(c *gin.Context) {
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

	session, newBalance, err := h.payments.Pay(userID, req.MerchantID, req.Amount, req.Description)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidAmount),
			errors.Is(err, service.ErrInsufficientFunds),
			errors.Is(err, service.ErrPaymentExpired):
			respondError(c, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrWalletNotFound):
			respondError(c, http.StatusNotFound, err.Error())
		default:
			respondError(c, http.StatusInternalServerError, "payment failed")
		}
		return
	}

	c.JSON(http.StatusOK, model.PayResponse{
		TransactionID: session.ID.String(),
		Message:       "Payment successful",
		NewBalance:    newBalance,
	})
}
