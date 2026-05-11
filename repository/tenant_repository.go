package repository

import (
	"errors"

	"com.jetapcglobal.b2b.com/models"
	"gorm.io/gorm"
)

type TenantRepository interface {
	FindByEmail(email string) (*models.Tenant, error)
}

type tenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) TenantRepository {
	return &tenantRepository{db: db}
}

func (r *tenantRepository) FindByEmail(email string) (*models.Tenant, error) {
	var tenant models.Tenant
	err := r.db.
		Where("contact_person = ?", email).
		Where("is_deleted = false").
		First(&tenant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &tenant, nil
}
