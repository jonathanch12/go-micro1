package router

import (
	"ewallet/internal/handler"
	"ewallet/internal/middleware"
	"ewallet/internal/security"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	User        *handler.UserHandler
	Wallet      *handler.WalletHandler
	Transaction *handler.TransactionHandler
}

func New(h Handlers, jwt *security.JWTManager) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	v1 := engine.Group("/api/v1")

	users := v1.Group("/users")
	{
		users.POST("/register", h.User.Register)
		users.POST("/login", h.User.Login)
	}

	auth := middleware.Auth(jwt)

	wallet := v1.Group("/wallet")
	wallet.Use(auth)
	{
		wallet.GET("/balance", h.Wallet.Balance)
	}

	txs := v1.Group("/transactions")
	txs.Use(auth)
	{
		txs.POST("/topup", h.Transaction.TopUp)
		txs.POST("/transfer", h.Transaction.Transfer)
		txs.POST("/pay", h.Transaction.Pay)
		txs.GET("/history", h.Transaction.History)
	}

	return engine
}
