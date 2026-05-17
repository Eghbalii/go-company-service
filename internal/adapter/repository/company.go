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

// companyModel is the GORM persistence model for the companies table.
// Kept separate from the domain entity to isolate ORM tag concerns.
type companyModel struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name              string    `gorm:"size:15;uniqueIndex;not null"`
	Description       string    `gorm:"size:3000"`
	AmountOfEmployees int       `gorm:"not null"`
	Registered        bool      `gorm:"not null"`
	Type              string    `gorm:"not null"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (companyModel) TableName() string { return "companies" }

func toCompanyEntity(m *companyModel) *entity.Company {
	return &entity.Company{
		ID:                m.ID,
		Name:              m.Name,
		Description:       m.Description,
		AmountOfEmployees: m.AmountOfEmployees,
		Registered:        m.Registered,
		Type:              entity.CompanyType(m.Type),
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
}

// CompanyRepository is the GORM implementation of port.CompanyRepository.
type CompanyRepository struct {
	db *gorm.DB
}

// NewCompanyRepository creates a CompanyRepository backed by the given DB connection.
func NewCompanyRepository(db *gorm.DB) *CompanyRepository {
	return &CompanyRepository{db: db}
}

func (r *CompanyRepository) Create(ctx context.Context, c *entity.Company) error {
	m := &companyModel{
		ID:                c.ID,
		Name:              c.Name,
		Description:       c.Description,
		AmountOfEmployees: c.AmountOfEmployees,
		Registered:        c.Registered,
		Type:              string(c.Type),
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	c.CreatedAt = m.CreatedAt
	c.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *CompanyRepository) Update(ctx context.Context, id uuid.UUID, fields map[string]interface{}) (*entity.Company, error) {
	res := r.db.WithContext(ctx).Model(&companyModel{}).Where("id = ?", id).Updates(fields)
	if res.Error != nil {
		return nil, apperr.New(apperr.ErrInternal, "update company: "+res.Error.Error())
	}
	if res.RowsAffected == 0 {
		return nil, apperr.New(apperr.ErrNotFound, "company not found")
	}
	return r.GetByID(ctx, id)
}

func (r *CompanyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ?", id).Delete(&companyModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperr.New(apperr.ErrNotFound, "company not found")
	}
	return nil
}

func (r *CompanyRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Company, error) {
	var m companyModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.New(apperr.ErrNotFound, "company not found")
		}
		return nil, apperr.New(apperr.ErrInternal, err.Error())
	}
	return toCompanyEntity(&m), nil
}

func (r *CompanyRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&companyModel{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}
