package router

import (
	"core-service/internal/handler"
	"core-service/internal/middleware"
	"core-service/internal/security"

	"github.com/gin-gonic/gin"
)

// Handlers bundles the HTTP handlers wired into the router.
type Handlers struct {
	User        *handler.UserHandler
	Wallet      *handler.WalletHandler
	Transaction *handler.TransactionHandler
	Internal    *handler.InternalHandler
}

// New builds the Gin engine. Public user-facing routes live under /api/v1 and
// use the JWT auth middleware; service-to-service routes live under /internal
// and use the shared internal API key.
func New(h Handlers, jwt *security.JWTManager, internalKey string) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	v1 := engine.Group("/api/v1")

	// Public routes.
	users := v1.Group("/users")
	{
		users.POST("/register", h.User.Register)
		users.POST("/login", h.User.Login)
	}

	// JWT-protected routes.
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
		txs.GET("/history", h.Transaction.History)
	}

	// Service-to-service routes (Payment Service -> Core Service).
	internal := engine.Group("/internal")
	internal.Use(middleware.InternalAuth(internalKey))
	{
		internal.POST("/wallets/debit", h.Internal.DebitWallet)
	}

	return engine
}
