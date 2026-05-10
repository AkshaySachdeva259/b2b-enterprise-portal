package services

import (
	"com.jetapcglobal.b2b.com/models"
	"com.jetapcglobal.b2b.com/repository"
	"errors"
	"strings"
)

type PackInventoryService interface {
	ListByTenantID(tenantID int64, statusFilter string) ([]models.TenantPackInventoryItem, error)
}

type packInventoryService struct {
	repo repository.PackInventoryRepository
}

var ErrInvalidPackInventoryStatusFilter = errors.New("status must be one of: allocated, unallocated")

func NewPackInventoryService(repo repository.PackInventoryRepository) PackInventoryService {
	return &packInventoryService{repo: repo}
}

func (s *packInventoryService) ListByTenantID(tenantID int64, statusFilter string) ([]models.TenantPackInventoryItem, error) {
	normalizedFilter := strings.ToLower(strings.TrimSpace(statusFilter))
	switch normalizedFilter {
	case "", "allocated", "unallocated":
	default:
		return nil, ErrInvalidPackInventoryStatusFilter
	}

	return s.repo.ListByTenantID(tenantID, normalizedFilter)
}
