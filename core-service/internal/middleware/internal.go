package middleware

import (
	"crypto/subtle"
	"net/http"

	"core-service/internal/model"

	"github.com/gin-gonic/gin"
)

// InternalHeader is the header carrying the shared service-to-service key.
const InternalHeader = "X-Internal-Token"

// InternalAuth guards service-to-service routes with a shared secret key.
// It uses a constant-time comparison to avoid timing attacks.
func InternalAuth(key string) gin.HandlerFunc {
	expected := []byte(key)
	return func(c *gin.Context) {
		got := []byte(c.GetHeader(InternalHeader))
		if len(got) == 0 || subtle.ConstantTimeCompare(got, expected) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{Error: "unauthorized"})
			return
		}
		c.Next()
	}
}
