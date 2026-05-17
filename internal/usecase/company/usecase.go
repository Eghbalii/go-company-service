package company

import (
	"context"
	"encoding/json"
	"time"

	"github.com/eghbalii/go-company-service/internal/domain/entity"
	"github.com/eghbalii/go-company-service/internal/domain/port"
	"github.com/eghbalii/go-company-service/pkg/apperr"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UseCase exposes the company business operations.
type UseCase interface {
	Create(ctx context.Context, in CreateInput) (*entity.Company, error)
	Update(ctx context.Context, id uuid.UUID, in UpdateInput) (*entity.Company, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Company, error)
}

// CreateInput carries the validated fields required to create a company.
type CreateInput struct {
	Name              string
	Description       string
	AmountOfEmployees int
	Registered        bool
	Type              entity.CompanyType
}

// UpdateInput carries the optional fields for a partial company update.
type UpdateInput struct {
	Name              *string
	Description       *string
	AmountOfEmployees *int
	Registered        *bool
	Type              *entity.CompanyType
}

type useCase struct {
	repo      port.CompanyRepository
	publisher port.EventPublisher
	log       *zap.Logger
}

// New builds a UseCase wired to the given repository and event publisher.
func New(repo port.CompanyRepository, publisher port.EventPublisher, log *zap.Logger) UseCase {
	return &useCase{repo: repo, publisher: publisher, log: log}
}

// Create validates, persists, and publishes a company.created event.
func (uc *useCase) Create(ctx context.Context, in CreateInput) (*entity.Company, error) {
	if !entity.IsValidType(in.Type) {
		return nil, apperr.New(apperr.ErrValidation, "invalid company type")
	}

	exists, err := uc.repo.ExistsByName(ctx, in.Name)
	if err != nil {
		return nil, apperr.New(apperr.ErrInternal, "check company name: "+err.Error())
	}
	if exists {
		return nil, apperr.New(apperr.ErrConflict, "company name already in use")
	}

	company := &entity.Company{
		ID:                uuid.New(),
		Name:              in.Name,
		Description:       in.Description,
		AmountOfEmployees: in.AmountOfEmployees,
		Registered:        in.Registered,
		Type:              in.Type,
	}

	if err := uc.repo.Create(ctx, company); err != nil {
		return nil, apperr.New(apperr.ErrInternal, "create company: "+err.Error())
	}

	uc.publishAsync(ctx, port.EventCompanyCreated, company)

	return company, nil
}

// Update applies a partial update and publishes a company.updated event.
func (uc *useCase) Update(ctx context.Context, id uuid.UUID, in UpdateInput) (*entity.Company, error) {
	fields := buildUpdateFields(in)
	if len(fields) == 0 {
		return uc.repo.GetByID(ctx, id)
	}

	if t, ok := fields["type"]; ok {
		if !entity.IsValidType(entity.CompanyType(t.(string))) {
			return nil, apperr.New(apperr.ErrValidation, "invalid company type")
		}
	}

	if name, ok := fields["name"]; ok {
		exists, err := uc.repo.ExistsByName(ctx, name.(string))
		if err != nil {
			return nil, apperr.New(apperr.ErrInternal, "check company name: "+err.Error())
		}
		if exists {
			return nil, apperr.New(apperr.ErrConflict, "company name already in use")
		}
	}

	company, err := uc.repo.Update(ctx, id, fields)
	if err != nil {
		return nil, err // already wrapped by the repo
	}

	uc.publishAsync(ctx, port.EventCompanyUpdated, company)

	return company, nil
}

// Delete removes a company and publishes a company.deleted event.
func (uc *useCase) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := uc.repo.GetByID(ctx, id); err != nil {
		return err // propagate ErrNotFound
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return apperr.New(apperr.ErrInternal, "delete company: "+err.Error())
	}

	uc.publishAsync(ctx, port.EventCompanyDeleted, map[string]string{"id": id.String()})

	return nil
}

// GetByID retrieves a single company.
func (uc *useCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Company, error) {
	return uc.repo.GetByID(ctx, id)
}

// buildUpdateFields converts UpdateInput into a map of non-nil fields for the repo.
func buildUpdateFields(in UpdateInput) map[string]interface{} {
	fields := make(map[string]interface{})
	if in.Name != nil {
		fields["name"] = *in.Name
	}
	if in.Description != nil {
		fields["description"] = *in.Description
	}
	if in.AmountOfEmployees != nil {
		fields["amount_of_employees"] = *in.AmountOfEmployees
	}
	if in.Registered != nil {
		fields["registered"] = *in.Registered
	}
	if in.Type != nil {
		fields["type"] = string(*in.Type)
	}
	return fields
}

// publishAsync fires an event in a goroutine so it never blocks the request path.
func (uc *useCase) publishAsync(ctx context.Context, eventType port.EventType, payload interface{}) {
	go func() {
		raw, err := json.Marshal(payload)
		if err != nil {
			uc.log.Error("marshal event payload", zap.Error(err))
			return
		}

		event := &port.Event{
			EventID:   uuid.New(),
			EventType: eventType,
			Timestamp: time.Now().UTC(),
			Payload:   raw,
		}

		if err := uc.publisher.Publish(ctx, event); err != nil {
			uc.log.Error("publish event", zap.String("type", string(eventType)), zap.Error(err))
		}
	}()
}
