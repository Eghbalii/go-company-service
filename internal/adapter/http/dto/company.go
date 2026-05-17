package dto

import (
	"time"

	"github.com/eghbalii/go-company-service/internal/domain/entity"
	"github.com/eghbalii/go-company-service/internal/usecase/company"
	"github.com/google/uuid"
)

// CreateCompanyRequest is the request body for POST /companies.
type CreateCompanyRequest struct {
	Name              string `json:"name"               validate:"required,max=15"`
	Description       string `json:"description"        validate:"max=3000"`
	AmountOfEmployees int    `json:"amount_of_employees" validate:"required,min=0"`
	Registered        bool   `json:"registered"`
	Type              string `json:"type"               validate:"required"`
}

// UpdateCompanyRequest is the request body for PATCH /companies/:id.
// All fields are optional — only non-nil fields are applied.
type UpdateCompanyRequest struct {
	Name              *string `json:"name"               validate:"omitempty,max=15"`
	Description       *string `json:"description"        validate:"omitempty,max=3000"`
	AmountOfEmployees *int    `json:"amount_of_employees" validate:"omitempty,min=0"`
	Registered        *bool   `json:"registered"`
	Type              *string `json:"type"               validate:"omitempty"`
}

// CompanyResponse is the outbound shape for a Company.
type CompanyResponse struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description,omitempty"`
	AmountOfEmployees int       `json:"amount_of_employees"`
	Registered        bool      `json:"registered"`
	Type              string    `json:"type"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ToCreateInput maps the request DTO to the use-case input type.
func (r *CreateCompanyRequest) ToCreateInput() company.CreateInput {
	return company.CreateInput{
		Name:              r.Name,
		Description:       r.Description,
		AmountOfEmployees: r.AmountOfEmployees,
		Registered:        r.Registered,
		Type:              entity.CompanyType(r.Type),
	}
}

// ToUpdateInput maps the request DTO to the use-case input type.
func (r *UpdateCompanyRequest) ToUpdateInput() company.UpdateInput {
	in := company.UpdateInput{
		Name:              r.Name,
		Description:       r.Description,
		AmountOfEmployees: r.AmountOfEmployees,
		Registered:        r.Registered,
	}
	if r.Type != nil {
		t := entity.CompanyType(*r.Type)
		in.Type = &t
	}
	return in
}

// ToCompanyResponse converts a domain Company to the response DTO.
func ToCompanyResponse(c *entity.Company) *CompanyResponse {
	return &CompanyResponse{
		ID:                c.ID,
		Name:              c.Name,
		Description:       c.Description,
		AmountOfEmployees: c.AmountOfEmployees,
		Registered:        c.Registered,
		Type:              string(c.Type),
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}
}
