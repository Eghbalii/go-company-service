package auth

import (
	"context"
	"fmt"

	"github.com/eghbalii/go-company-service/internal/domain/port"
	"github.com/eghbalii/go-company-service/pkg/apperr"
	"github.com/eghbalii/go-company-service/pkg/hash"
	"github.com/eghbalii/go-company-service/pkg/jwt"
	"go.uber.org/zap"
)

// UseCase exposes authentication operations.
type UseCase interface {
	Login(ctx context.Context, email, password string) (token string, err error)
}

type useCase struct {
	users  port.UserRepository
	jwtMgr *jwt.Manager
	log    *zap.Logger
}

// New builds an auth UseCase.
func New(users port.UserRepository, jwtMgr *jwt.Manager, log *zap.Logger) UseCase {
	return &useCase{users: users, jwtMgr: jwtMgr, log: log}
}

// Login verifies credentials and returns a signed JWT on success.
func (uc *useCase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := uc.users.GetByEmail(ctx, email)
	if err != nil {
		// Intentionally return the same error regardless of whether the user
		// exists to avoid user-enumeration via timing differences.
		uc.log.Debug("login: user lookup failed", zap.String("email", email), zap.Error(err))
		return "", apperr.New(apperr.ErrInvalidCredential, "invalid email or password")
	}

	if !hash.CheckPassword(user.PasswordHash, password) {
		return "", apperr.New(apperr.ErrInvalidCredential, "invalid email or password")
	}

	token, err := uc.jwtMgr.Generate(user.ID, user.Email)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return token, nil
}
