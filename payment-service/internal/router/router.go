package router

import (
	"payment-service/internal/handler"
	"payment-service/internal/middleware"
	"payment-service/internal/security"

	"github.com/gin-gonic/gin"
)

// Handlers bundles the HTTP handlers wired into the router.
type Handlers struct {
	Payment *handler.PaymentHandler
}

// New builds the Gin engine. The /transactions/pay route is JWT-protected
// using the shared secret.
func New(h Handlers, verifier *security.JWTVerifier) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	v1 := engine.Group("/api/v1")

	txs := v1.Group("/transactions")
	txs.Use(middleware.Auth(verifier))
	{
		txs.POST("/pay", h.Payment.Pay)
	}

	return engine
}
