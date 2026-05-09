package services

import (
	"errors"

	"com.jetapcglobal.b2b.com/models"
	"com.jetapcglobal.b2b.com/repository"
	"gorm.io/gorm"
)

var ErrTenantCreditLimitNotFound = errors.New("tenant credit limit not found")

type TenantCreditLimitService interface {
	GetCurrentByTenantID(tenantID int64) (*models.TenantCreditLimit, error)
}

type tenantCreditLimitService struct {
	repo repository.TenantCreditLimitRepository
}

func NewTenantCreditLimitService(repo repository.TenantCreditLimitRepository) TenantCreditLimitService {
	return &tenantCreditLimitService{repo: repo}
}

func (s *tenantCreditLimitService) GetCurrentByTenantID(tenantID int64) (*models.TenantCreditLimit, error) {
	result, err := s.repo.GetCurrentByTenantID(tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTenantCreditLimitNotFound
		}
		return nil, err
	}
	return result, nil
}
