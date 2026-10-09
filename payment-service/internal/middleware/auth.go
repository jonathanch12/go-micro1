package middleware

import (
	"net/http"
	"strings"

	"payment-service/internal/model"
	"payment-service/internal/security"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ContextUserIDKey is the Gin context key holding the authenticated user ID.
const ContextUserIDKey = "userID"

// Auth validates a Bearer JWT (shared secret) and injects the user ID.
func Auth(verifier *security.JWTVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			abortUnauthorized(c, "missing authorization header")
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			abortUnauthorized(c, "invalid authorization header")
			return
		}

		userID, err := verifier.Verify(parts[1])
		if err != nil {
			abortUnauthorized(c, "invalid or expired token")
			return
		}

		c.Set(ContextUserIDKey, userID)
		c.Next()
	}
}

// UserIDFromContext extracts the authenticated user ID set by Auth.
func UserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(ContextUserIDKey)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

func abortUnauthorized(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{Error: msg})
}
