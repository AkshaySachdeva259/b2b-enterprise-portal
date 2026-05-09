package services

import (
	"errors"

	"com.jetapcglobal.b2b.com/models"
	"com.jetapcglobal.b2b.com/repository"
)

type CatalogService interface {
	GetByPageName(pageName string) ([]models.Catalog, error)
}

type catalogService struct {
	repo repository.CatalogRepository
}

func NewCatalogService(repo repository.CatalogRepository) CatalogService {
	return &catalogService{repo: repo}
}

func (s *catalogService) GetByPageName(pageName string) ([]models.Catalog, error) {
	if pageName == "" {
		return nil, errors.New("page_name is required")
	}
	return s.repo.GetByPageName(pageName)
}
