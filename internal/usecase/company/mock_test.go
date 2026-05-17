package company_test

import (
	"context"

	"github.com/eghbalii/go-company-service/internal/domain/entity"
	"github.com/eghbalii/go-company-service/internal/domain/port"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// mockCompanyRepo is a testify mock for port.CompanyRepository.
type mockCompanyRepo struct{ mock.Mock }

func (m *mockCompanyRepo) Create(ctx context.Context, c *entity.Company) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *mockCompanyRepo) Update(ctx context.Context, id uuid.UUID, fields map[string]interface{}) (*entity.Company, error) {
	args := m.Called(ctx, id, fields)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Company), args.Error(1)
}

func (m *mockCompanyRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockCompanyRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Company, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Company), args.Error(1)
}

func (m *mockCompanyRepo) ExistsByName(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

// mockPublisher is a testify mock for port.EventPublisher.
type mockPublisher struct{ mock.Mock }

func (m *mockPublisher) Publish(ctx context.Context, e *port.Event) error {
	return m.Called(ctx, e).Error(0)
}

func (m *mockPublisher) Close() error { return m.Called().Error(0) }
