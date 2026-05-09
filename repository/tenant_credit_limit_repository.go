package repository

import (
	"errors"

	"com.jetapcglobal.b2b.com/models"
	"gorm.io/gorm"
)

type TenantCreditLimitRepository interface {
	GetCurrentByTenantID(tenantID int64) (*models.TenantCreditLimit, error)
}

type tenantCreditLimitRepository struct {
	db *gorm.DB
}

func NewTenantCreditLimitRepository(db *gorm.DB) TenantCreditLimitRepository {
	return &tenantCreditLimitRepository{db: db}
}

func (r *tenantCreditLimitRepository) GetCurrentByTenantID(tenantID int64) (*models.TenantCreditLimit, error) {
	var result models.TenantCreditLimit
	err := r.db.
		Where("tenant_id = ? AND is_active = ? AND is_deleted = ?", tenantID, true, false).
		First(&result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &result, nil
}
