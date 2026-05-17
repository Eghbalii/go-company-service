package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/eghbalii/go-company-service/internal/domain/entity"
	"github.com/eghbalii/go-company-service/internal/domain/port"
	"github.com/eghbalii/go-company-service/internal/usecase/auth"
	"github.com/eghbalii/go-company-service/pkg/apperr"
	"github.com/eghbalii/go-company-service/pkg/hash"
	"github.com/eghbalii/go-company-service/pkg/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockUserRepo is a testify mock for port.UserRepository.
type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Create(ctx context.Context, u *entity.User) error {
	return m.Called(ctx, u).Error(0)
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

// ensure mockUserRepo satisfies the interface at compile time
var _ port.UserRepository = (*mockUserRepo)(nil)

func newJWTMgr() *jwt.Manager {
	return jwt.NewManager("test-secret-32-chars-minimum-len!", time.Hour)
}

func TestLogin_Success(t *testing.T) {
	repo := new(mockUserRepo)
	mgr := newJWTMgr()
	uc := auth.New(repo, mgr, zap.NewNop())

	pw, err := hash.Password("password123")
	require.NoError(t, err)

	user := &entity.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		PasswordHash: pw,
	}

	repo.On("GetByEmail", mock.Anything, "user@example.com").Return(user, nil)

	token, err := uc.Login(context.Background(), "user@example.com", "password123")

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify the token is valid
	claims, err := mgr.Verify(token)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := new(mockUserRepo)
	uc := auth.New(repo, newJWTMgr(), zap.NewNop())

	repo.On("GetByEmail", mock.Anything, "nobody@example.com").
		Return(nil, apperr.New(apperr.ErrNotFound, "user not found"))

	_, err := uc.Login(context.Background(), "nobody@example.com", "password123")

	require.Error(t, err)
	assert.True(t, apperr.Is(err, apperr.ErrInvalidCredential))
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := new(mockUserRepo)
	uc := auth.New(repo, newJWTMgr(), zap.NewNop())

	pw, _ := hash.Password("correct-password")
	user := &entity.User{ID: uuid.New(), Email: "user@example.com", PasswordHash: pw}

	repo.On("GetByEmail", mock.Anything, "user@example.com").Return(user, nil)

	_, err := uc.Login(context.Background(), "user@example.com", "wrong-password")

	require.Error(t, err)
	assert.True(t, apperr.Is(err, apperr.ErrInvalidCredential))
}
