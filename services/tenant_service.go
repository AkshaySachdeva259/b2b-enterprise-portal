package services

import (
	"errors"

	"com.jetapcglobal.b2b.com/models"
	"com.jetapcglobal.b2b.com/repository"
	"gorm.io/gorm"
)

var ErrTenantNotFound = errors.New("no tenant registered")

type TenantService interface {
	FindByEmail(email string) (*models.Tenant, error)
}

type tenantService struct {
	repo repository.TenantRepository
}

func NewTenantService(repo repository.TenantRepository) TenantService {
	return &tenantService{repo: repo}
}

func (s *tenantService) FindByEmail(email string) (*models.Tenant, error) {
	tenant, err := s.repo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTenantNotFound
		}
		return nil, err
	}
	return tenant, nil
}
