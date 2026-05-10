package services

import (
	"com.jetapcglobal.b2b.com/models"
	"com.jetapcglobal.b2b.com/repository"
)

type PackInventoryService interface {
	ListByTenantID(tenantID int64) ([]models.TenantPackInventoryItem, error)
}

type packInventoryService struct {
	repo repository.PackInventoryRepository
}

func NewPackInventoryService(repo repository.PackInventoryRepository) PackInventoryService {
	return &packInventoryService{repo: repo}
}

func (s *packInventoryService) ListByTenantID(tenantID int64) ([]models.TenantPackInventoryItem, error) {
	return s.repo.ListByTenantID(tenantID)
}
