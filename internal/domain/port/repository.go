package port

import (
	"context"

	"github.com/eghbalii/go-company-service/internal/domain/entity"
	"github.com/google/uuid"
)

// CompanyRepository is the persistence contract for Company aggregates.
type CompanyRepository interface {
	Create(ctx context.Context, company *entity.Company) error
	Update(ctx context.Context, id uuid.UUID, fields map[string]interface{}) (*entity.Company, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Company, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
}

// UserRepository is the persistence contract for User aggregates.
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
}
