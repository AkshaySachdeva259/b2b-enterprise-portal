package repository

import (
	"errors"

	"com.jetapcglobal.b2b.com/models"
	"gorm.io/gorm"
)

type TenantWalletRepository interface {
	GetWalletByTenantID(tenantID int64) (*models.TenantWallet, error)
	GetRecentTransactionsByTenantID(tenantID int64, limit int) ([]models.CreditLedgerTransaction, error)
	TenantExists(tenantID int64) (bool, error)
}

type tenantWalletRepository struct {
	db *gorm.DB
}

func NewTenantWalletRepository(db *gorm.DB) TenantWalletRepository {
	return &tenantWalletRepository{db: db}
}

func (r *tenantWalletRepository) GetWalletByTenantID(tenantID int64) (*models.TenantWallet, error) {
	var result models.TenantWallet
	err := r.db.
		Where("tenant_id = ?", tenantID).
		Order("CASE WHEN UPPER(COALESCE(status, '')) = 'ACTIVE' THEN 0 ELSE 1 END").
		Order("version DESC NULLS LAST").
		Order("updated_at DESC NULLS LAST").
		Order("created_at DESC NULLS LAST").
		Order("id DESC").
		First(&result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &result, nil
}

func (r *tenantWalletRepository) TenantExists(tenantID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.TenantWallet{}).
		Where("tenant_id = ?", tenantID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *tenantWalletRepository) GetRecentTransactionsByTenantID(tenantID int64, limit int) ([]models.CreditLedgerTransaction, error) {
	var results []models.CreditLedgerTransaction
	err := r.db.
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC NULLS LAST").
		Order("id DESC").
		Limit(limit).
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}
