package handler

import (
	"net/http"

	"ewallet/internal/middleware"
	"ewallet/internal/model"
	"ewallet/internal/service"

	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	wallets *service.WalletService
}

func NewWalletHandler(wallets *service.WalletService) *WalletHandler {
	return &WalletHandler{wallets: wallets}
}

func (h *WalletHandler) Balance(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	balance, err := h.wallets.GetBalance(userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to fetch balance")
		return
	}

	c.JSON(http.StatusOK, model.WalletBalanceResponse{
		UserID:  userID.String(),
		Balance: balance,
	})
}
