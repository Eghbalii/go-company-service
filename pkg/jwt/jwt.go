package jwt

import (
	"fmt"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims are the custom JWT payload fields.
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	gojwt.RegisteredClaims
}

// Manager signs and verifies JWTs.
type Manager struct {
	secret []byte
	expiry time.Duration
}

// NewManager creates a Manager with the given HMAC secret and token lifetime.
func NewManager(secret string, expiry time.Duration) *Manager {
	return &Manager{secret: []byte(secret), expiry: expiry}
}

// Generate mints a signed JWT for the given user.
func (m *Manager) Generate(userID uuid.UUID, email string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: gojwt.RegisteredClaims{
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(now.Add(m.expiry)),
		},
	}

	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

// Verify parses and validates a token string, returning its claims.
func (m *Manager) Verify(tokenStr string) (*Claims, error) {
	token, err := gojwt.ParseWithClaims(tokenStr, &Claims{}, func(t *gojwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
