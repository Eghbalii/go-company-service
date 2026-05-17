package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eghbalii/go-company-service/internal/adapter/http/handler"
	"github.com/eghbalii/go-company-service/internal/domain/entity"
	"github.com/eghbalii/go-company-service/internal/usecase/company"
	"github.com/eghbalii/go-company-service/pkg/apperr"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockCompanyUseCase is a testify mock for company.UseCase.
type mockCompanyUseCase struct{ mock.Mock }

func (m *mockCompanyUseCase) Create(ctx context.Context, in company.CreateInput) (*entity.Company, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Company), args.Error(1)
}

func (m *mockCompanyUseCase) Update(ctx context.Context, id uuid.UUID, in company.UpdateInput) (*entity.Company, error) {
	args := m.Called(ctx, id, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Company), args.Error(1)
}

func (m *mockCompanyUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockCompanyUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Company, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Company), args.Error(1)
}

type customValidator struct{ v *validator.Validate }

func (cv *customValidator) Validate(i interface{}) error { return cv.v.Struct(i) }

func newEcho() *echo.Echo {
	e := echo.New()
	e.Validator = &customValidator{v: validator.New()}
	return e
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestGetByID_Handler_Success(t *testing.T) {
	uc := new(mockCompanyUseCase)
	h := handler.NewCompanyHandler(uc)
	e := newEcho()

	id := uuid.New()
	c := &entity.Company{ID: id, Name: "Acme", Type: entity.CompanyTypeCorporations}
	uc.On("GetByID", mock.Anything, id).Return(c, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues(id.String())

	require.NoError(t, h.GetByID(ctx))
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	data := body["data"].(map[string]interface{})
	assert.Equal(t, id.String(), data["id"])
}

func TestGetByID_Handler_NotFound(t *testing.T) {
	uc := new(mockCompanyUseCase)
	h := handler.NewCompanyHandler(uc)
	e := newEcho()

	id := uuid.New()
	uc.On("GetByID", mock.Anything, id).Return(nil, apperr.New(apperr.ErrNotFound, "company not found"))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues(id.String())

	// The handler returns the apperr; the global error handler maps it to 404.
	// In unit tests we just confirm the returned error wraps ErrNotFound.
	err := h.GetByID(ctx)
	require.Error(t, err)
	assert.True(t, apperr.Is(err, apperr.ErrNotFound))
}

func TestGetByID_Handler_InvalidID(t *testing.T) {
	uc := new(mockCompanyUseCase)
	h := handler.NewCompanyHandler(uc)
	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("not-a-uuid")

	err := h.GetByID(ctx)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestCreate_Handler_Success(t *testing.T) {
	uc := new(mockCompanyUseCase)
	h := handler.NewCompanyHandler(uc)
	e := newEcho()

	body := `{"name":"Acme","amount_of_employees":50,"registered":true,"type":"Corporations"}`
	created := &entity.Company{
		ID:                uuid.New(),
		Name:              "Acme",
		AmountOfEmployees: 50,
		Registered:        true,
		Type:              entity.CompanyTypeCorporations,
	}

	uc.On("Create", mock.Anything, mock.AnythingOfType("company.CreateInput")).Return(created, nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	require.NoError(t, h.Create(ctx))
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestCreate_Handler_ValidationError(t *testing.T) {
	uc := new(mockCompanyUseCase)
	h := handler.NewCompanyHandler(uc)
	e := newEcho()

	// name exceeds 15 chars
	body := `{"name":"This Name Is Way Too Long","amount_of_employees":10,"type":"Corporations"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := h.Create(ctx)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnprocessableEntity, he.Code)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestDelete_Handler_Success(t *testing.T) {
	uc := new(mockCompanyUseCase)
	h := handler.NewCompanyHandler(uc)
	e := newEcho()

	id := uuid.New()
	uc.On("Delete", mock.Anything, id).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues(id.String())

	require.NoError(t, h.Delete(ctx))
	assert.Equal(t, http.StatusNoContent, rec.Code)
}
