package repository

import (
	"context"
	"errors"
	"time"

	"github.com/eghbalii/go-company-service/internal/domain/entity"
	"github.com/eghbalii/go-company-service/pkg/apperr"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// userModel is the GORM persistence model for the users table.
type userModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (userModel) TableName() string { return "users" }

func toUserEntity(m *userModel) *entity.User {
	return &entity.User{
		ID:           m.ID,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// UserRepository is the GORM implementation of port.UserRepository.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a UserRepository backed by the given DB connection.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *entity.User) error {
	m := &userModel{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	u.CreatedAt = m.CreatedAt
	u.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var m userModel
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.New(apperr.ErrNotFound, "user not found")
		}
		return nil, apperr.New(apperr.ErrInternal, err.Error())
	}
	return toUserEntity(&m), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var m userModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.New(apperr.ErrNotFound, "user not found")
		}
		return nil, apperr.New(apperr.ErrInternal, err.Error())
	}
	return toUserEntity(&m), nil
}
