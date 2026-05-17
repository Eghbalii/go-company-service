package entity

import (
	"time"

	"github.com/google/uuid"
)

// CompanyType is the allowed set of company classifications.
type CompanyType string

const (
	CompanyTypeCorporations       CompanyType = "Corporations"
	CompanyTypeNonProfit          CompanyType = "NonProfit"
	CompanyTypeCooperative        CompanyType = "Cooperative"
	CompanyTypeSoleProprietorship CompanyType = "Sole Proprietorship"
)

// ValidCompanyTypes is the authoritative list used for validation.
var ValidCompanyTypes = map[CompanyType]struct{}{
	CompanyTypeCorporations:       {},
	CompanyTypeNonProfit:          {},
	CompanyTypeCooperative:        {},
	CompanyTypeSoleProprietorship: {},
}

// Company is the core business entity.
type Company struct {
	ID                uuid.UUID   `json:"id"`
	Name              string      `json:"name"`
	Description       string      `json:"description,omitempty"`
	AmountOfEmployees int         `json:"amount_of_employees"`
	Registered        bool        `json:"registered"`
	Type              CompanyType `json:"type"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
}

// IsValidType reports whether t is a recognised company type.
func IsValidType(t CompanyType) bool {
	_, ok := ValidCompanyTypes[t]
	return ok
}
