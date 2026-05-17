//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/eghbalii/go-company-service/internal/adapter/http/handler"
	"github.com/eghbalii/go-company-service/internal/adapter/http/router"
	"github.com/eghbalii/go-company-service/internal/adapter/repository"
	"github.com/eghbalii/go-company-service/internal/domain/entity"
	"github.com/eghbalii/go-company-service/internal/domain/port"
	"github.com/eghbalii/go-company-service/internal/infrastructure/database"
	"github.com/eghbalii/go-company-service/internal/infrastructure/server"
	"github.com/eghbalii/go-company-service/internal/usecase/auth"
	"github.com/eghbalii/go-company-service/internal/usecase/company"
	"github.com/eghbalii/go-company-service/pkg/config"
	"github.com/eghbalii/go-company-service/pkg/hash"
	"github.com/eghbalii/go-company-service/pkg/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/labstack/echo/v4"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// noopPublisher discards events so integration tests don't need a Kafka broker.
type noopPublisher struct{}

func (n *noopPublisher) Publish(_ context.Context, _ *port.Event) error { return nil }
func (n *noopPublisher) Close() error                                    { return nil }

// CompanyIntegrationSuite spins up a real PostgreSQL container and tests the
// full HTTP stack end-to-end (handler → use case → GORM → postgres).
type CompanyIntegrationSuite struct {
	suite.Suite
	pgContainer testcontainers.Container
	echoServer  *echo.Echo
	jwtMgr      *jwt.Manager
	userRepo    *repository.UserRepository
	token       string
}

func (s *CompanyIntegrationSuite) SetupSuite() {
	ctx := context.Background()

	pgC, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(s.T(), err)
	s.pgContainer = pgC

	host, _ := pgC.Host(ctx)
	mappedPort, _ := pgC.MappedPort(ctx, "5432")

	dbCfg := config.Database{
		Host:            host,
		Port:            mappedPort.Port(),
		User:            "postgres",
		Password:        "postgres",
		Name:            "testdb",
		SSLMode:         "disable",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
	}

	db, err := database.NewPostgres(dbCfg)
	require.NoError(s.T(), err)

	// Auto-migrate models for the test database
	err = db.AutoMigrate(&struct {
		ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
		Email        string    `gorm:"uniqueIndex;not null"`
		PasswordHash string    `gorm:"not null"`
		CreatedAt    time.Time
		UpdatedAt    time.Time
	}{})
	require.NoError(s.T(), err)

	// Run raw SQL to create tables matching the migration files
	db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`)
	db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		email VARCHAR(255) NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`)
	db.Exec(`DO $$ BEGIN
		CREATE TYPE company_type AS ENUM (
			'Corporations','NonProfit','Cooperative','Sole Proprietorship'
		);
	EXCEPTION WHEN duplicate_object THEN NULL; END $$`)
	db.Exec(`CREATE TABLE IF NOT EXISTS companies (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name VARCHAR(15) NOT NULL UNIQUE,
		description VARCHAR(3000) NOT NULL DEFAULT '',
		amount_of_employees INTEGER NOT NULL CHECK (amount_of_employees >= 0),
		registered BOOLEAN NOT NULL DEFAULT FALSE,
		type company_type NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`)

	s.jwtMgr = jwt.NewManager("integration-test-secret-32chars!", time.Hour)
	log := zap.NewNop()

	companyRepo := repository.NewCompanyRepository(db)
	s.userRepo = repository.NewUserRepository(db)

	pub := &noopPublisher{}
	companyUC := company.New(companyRepo, pub, log)
	authUC := auth.New(s.userRepo, s.jwtMgr, log)

	e := server.New(log, false)
	companyH := handler.NewCompanyHandler(companyUC)
	authH := handler.NewAuthHandler(authUC)
	router.Register(e, s.jwtMgr, authH, companyH)
	s.echoServer = e

	// Seed a test user
	pw, _ := hash.Password("testpassword")
	testUser := &entity.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: pw,
	}
	require.NoError(s.T(), s.userRepo.Create(ctx, testUser))

	// Obtain a JWT for protected endpoints
	tok, err := s.jwtMgr.Generate(testUser.ID, testUser.Email)
	require.NoError(s.T(), err)
	s.token = tok
}

func (s *CompanyIntegrationSuite) TearDownSuite() {
	_ = s.pgContainer.Terminate(context.Background())
}

func (s *CompanyIntegrationSuite) doRequest(method, path, body string, auth bool) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != "" {
		reqBody = bytes.NewBufferString(body)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if auth {
		req.Header.Set("Authorization", "Bearer "+s.token)
	}

	rec := httptest.NewRecorder()
	s.echoServer.ServeHTTP(rec, req)
	return rec
}

// ── Auth flow ─────────────────────────────────────────────────────────────────

func (s *CompanyIntegrationSuite) TestLogin_Success() {
	body := `{"email":"test@example.com","password":"testpassword"}`
	rec := s.doRequest(http.MethodPost, "/api/v1/auth/login", body, false)

	assert.Equal(s.T(), http.StatusOK, rec.Code)

	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(rec.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.NotEmpty(s.T(), data["token"])
}

func (s *CompanyIntegrationSuite) TestLogin_WrongPassword() {
	body := `{"email":"test@example.com","password":"wrongpassword"}`
	rec := s.doRequest(http.MethodPost, "/api/v1/auth/login", body, false)
	assert.Equal(s.T(), http.StatusUnauthorized, rec.Code)
}

// ── Company CRUD ──────────────────────────────────────────────────────────────

func (s *CompanyIntegrationSuite) TestCreateCompany_Success() {
	body := fmt.Sprintf(`{
		"name":"%s",
		"description":"Integration test company",
		"amount_of_employees":42,
		"registered":true,
		"type":"Corporations"
	}`, "IntTest"+uuid.New().String()[:5])

	rec := s.doRequest(http.MethodPost, "/api/v1/companies", body, true)
	assert.Equal(s.T(), http.StatusCreated, rec.Code)
}

func (s *CompanyIntegrationSuite) TestCreateCompany_Unauthorized() {
	body := `{"name":"NoAuth","amount_of_employees":1,"type":"NonProfit"}`
	rec := s.doRequest(http.MethodPost, "/api/v1/companies", body, false)
	assert.Equal(s.T(), http.StatusUnauthorized, rec.Code)
}

func (s *CompanyIntegrationSuite) TestGetCompany_Success() {
	// Create a company first
	name := "GetTest" + uuid.New().String()[:3]
	createBody := fmt.Sprintf(`{"name":"%s","amount_of_employees":5,"registered":false,"type":"NonProfit"}`, name)
	createRec := s.doRequest(http.MethodPost, "/api/v1/companies", createBody, true)
	require.Equal(s.T(), http.StatusCreated, createRec.Code)

	var createResp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(createRec.Body.Bytes(), &createResp))
	id := createResp["data"].(map[string]interface{})["id"].(string)

	// Retrieve it
	rec := s.doRequest(http.MethodGet, "/api/v1/companies/"+id, "", false)
	assert.Equal(s.T(), http.StatusOK, rec.Code)

	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(rec.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(s.T(), id, data["id"])
	assert.Equal(s.T(), name, data["name"])
}

func (s *CompanyIntegrationSuite) TestPatchCompany_Success() {
	// Create
	name := "PatchMe" + uuid.New().String()[:2]
	createBody := fmt.Sprintf(`{"name":"%s","amount_of_employees":10,"registered":false,"type":"Cooperative"}`, name)
	createRec := s.doRequest(http.MethodPost, "/api/v1/companies", createBody, true)
	require.Equal(s.T(), http.StatusCreated, createRec.Code)

	var createResp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(createRec.Body.Bytes(), &createResp))
	id := createResp["data"].(map[string]interface{})["id"].(string)

	// Patch
	patchBody := `{"amount_of_employees":999,"registered":true}`
	rec := s.doRequest(http.MethodPatch, "/api/v1/companies/"+id, patchBody, true)
	assert.Equal(s.T(), http.StatusOK, rec.Code)

	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(rec.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(s.T(), float64(999), data["amount_of_employees"])
	assert.Equal(s.T(), true, data["registered"])
}

func (s *CompanyIntegrationSuite) TestDeleteCompany_Success() {
	// Create
	name := "DelTest" + uuid.New().String()[:2]
	createBody := fmt.Sprintf(`{"name":"%s","amount_of_employees":1,"registered":false,"type":"NonProfit"}`, name)
	createRec := s.doRequest(http.MethodPost, "/api/v1/companies", createBody, true)
	require.Equal(s.T(), http.StatusCreated, createRec.Code)

	var createResp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(createRec.Body.Bytes(), &createResp))
	id := createResp["data"].(map[string]interface{})["id"].(string)

	// Delete
	rec := s.doRequest(http.MethodDelete, "/api/v1/companies/"+id, "", true)
	assert.Equal(s.T(), http.StatusNoContent, rec.Code)

	// Verify gone
	getRec := s.doRequest(http.MethodGet, "/api/v1/companies/"+id, "", false)
	assert.Equal(s.T(), http.StatusNotFound, getRec.Code)
}

func TestCompanyIntegrationSuite(t *testing.T) {
	suite.Run(t, new(CompanyIntegrationSuite))
}
