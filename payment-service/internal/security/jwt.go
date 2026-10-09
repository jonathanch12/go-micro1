package security

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ErrInvalidToken is returned when a token cannot be validated.
var ErrInvalidToken = errors.New("invalid token")

// JWTVerifier validates JWTs minted by core-service using the shared secret.
// Payment-service only needs to verify tokens, never issue them.
type JWTVerifier struct {
	secret []byte
}

// Claims mirrors the core-service claim set.
type Claims struct {
	UserID string `json:"userID"`
	jwt.RegisteredClaims
}

// NewJWTVerifier constructs a JWTVerifier with the shared secret.
func NewJWTVerifier(secret string) *JWTVerifier {
	return &JWTVerifier{secret: []byte(secret)}
}

// Verify parses and validates a token string, returning the user ID.
func (v *JWTVerifier) Verify(tokenStr string) (uuid.UUID, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return v.secret, nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}
	return userID, nil
}
