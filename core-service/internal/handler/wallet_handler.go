package handler

import (
	"net/http"

	"core-service/internal/middleware"
	"core-service/internal/model"
	"core-service/internal/service"

	"github.com/gin-gonic/gin"
)

// WalletHandler serves the /wallet routes.
type WalletHandler struct {
	wallets *service.WalletService
}

func NewWalletHandler(wallets *service.WalletService) *WalletHandler {
	return &WalletHandler{wallets: wallets}
}

// Balance handles GET /wallet/balance.
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
