package company_test

import (
	"context"
	"testing"

	"github.com/eghbalii/go-company-service/internal/domain/entity"
	"github.com/eghbalii/go-company-service/internal/usecase/company"
	"github.com/eghbalii/go-company-service/pkg/apperr"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newUC(repo *mockCompanyRepo, pub *mockPublisher) company.UseCase {
	return company.New(repo, pub, zap.NewNop())
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestCreate_Success(t *testing.T) {
	repo := new(mockCompanyRepo)
	pub := new(mockPublisher)
	uc := newUC(repo, pub)

	in := company.CreateInput{
		Name:              "Acme Corp",
		AmountOfEmployees: 100,
		Registered:        true,
		Type:              entity.CompanyTypeCorporations,
	}

	repo.On("ExistsByName", mock.Anything, "Acme Corp").Return(false, nil)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Company")).Return(nil)
	pub.On("Publish", mock.Anything, mock.Anything).Return(nil)

	result, err := uc.Create(context.Background(), in)

	require.NoError(t, err)
	assert.Equal(t, "Acme Corp", result.Name)
	assert.Equal(t, entity.CompanyTypeCorporations, result.Type)
	assert.NotEqual(t, uuid.Nil, result.ID)
}

func TestCreate_DuplicateName(t *testing.T) {
	repo := new(mockCompanyRepo)
	pub := new(mockPublisher)
	uc := newUC(repo, pub)

	repo.On("ExistsByName", mock.Anything, "Taken").Return(true, nil)

	_, err := uc.Create(context.Background(), company.CreateInput{
		Name:              "Taken",
		AmountOfEmployees: 10,
		Type:              entity.CompanyTypeCorporations,
	})

	require.Error(t, err)
	assert.True(t, apperr.Is(err, apperr.ErrConflict))
}

func TestCreate_InvalidType(t *testing.T) {
	repo := new(mockCompanyRepo)
	pub := new(mockPublisher)
	uc := newUC(repo, pub)

	_, err := uc.Create(context.Background(), company.CreateInput{
		Name:              "X Corp",
		AmountOfEmployees: 5,
		Type:              entity.CompanyType("Unknown"),
	})

	require.Error(t, err)
	assert.True(t, apperr.Is(err, apperr.ErrValidation))
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestGetByID_Success(t *testing.T) {
	repo := new(mockCompanyRepo)
	pub := new(mockPublisher)
	uc := newUC(repo, pub)

	id := uuid.New()
	expected := &entity.Company{ID: id, Name: "Acme"}
	repo.On("GetByID", mock.Anything, id).Return(expected, nil)

	result, err := uc.GetByID(context.Background(), id)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestGetByID_NotFound(t *testing.T) {
	repo := new(mockCompanyRepo)
	pub := new(mockPublisher)
	uc := newUC(repo, pub)

	id := uuid.New()
	repo.On("GetByID", mock.Anything, id).Return(nil, apperr.New(apperr.ErrNotFound, "company not found"))

	_, err := uc.GetByID(context.Background(), id)

	require.Error(t, err)
	assert.True(t, apperr.Is(err, apperr.ErrNotFound))
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestDelete_Success(t *testing.T) {
	repo := new(mockCompanyRepo)
	pub := new(mockPublisher)
	uc := newUC(repo, pub)

	id := uuid.New()
	existing := &entity.Company{ID: id}

	repo.On("GetByID", mock.Anything, id).Return(existing, nil)
	repo.On("Delete", mock.Anything, id).Return(nil)
	pub.On("Publish", mock.Anything, mock.Anything).Return(nil)

	err := uc.Delete(context.Background(), id)
	require.NoError(t, err)
}

func TestDelete_NotFound(t *testing.T) {
	repo := new(mockCompanyRepo)
	pub := new(mockPublisher)
	uc := newUC(repo, pub)

	id := uuid.New()
	repo.On("GetByID", mock.Anything, id).Return(nil, apperr.New(apperr.ErrNotFound, "company not found"))

	err := uc.Delete(context.Background(), id)

	require.Error(t, err)
	assert.True(t, apperr.Is(err, apperr.ErrNotFound))
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestUpdate_Success(t *testing.T) {
	repo := new(mockCompanyRepo)
	pub := new(mockPublisher)
	uc := newUC(repo, pub)

	id := uuid.New()
	newName := "New Name"
	updated := &entity.Company{ID: id, Name: newName}

	repo.On("ExistsByName", mock.Anything, newName).Return(false, nil)
	repo.On("Update", mock.Anything, id, mock.MatchedBy(func(f map[string]interface{}) bool {
		return f["name"] == newName
	})).Return(updated, nil)
	pub.On("Publish", mock.Anything, mock.Anything).Return(nil)

	result, err := uc.Update(context.Background(), id, company.UpdateInput{Name: &newName})

	require.NoError(t, err)
	assert.Equal(t, newName, result.Name)
}
