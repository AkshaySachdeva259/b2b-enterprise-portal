package repository

import (
	"com.jetapcglobal.b2b.com/models"
	"gorm.io/gorm"
)

type CatalogRepository interface {
	GetByPageName(pageName string) ([]models.Catalog, error)
}

type catalogRepository struct {
	db *gorm.DB
}

func NewCatalogRepository(db *gorm.DB) CatalogRepository {
	return &catalogRepository{db: db}
}

func (r *catalogRepository) GetByPageName(pageName string) ([]models.Catalog, error) {
	var results []models.Catalog
	err := r.db.
		Where("page_name = ? AND deleted_at IS NULL", pageName).
		Order("created_at DESC").
		Find(&results).Error
	return results, err
}
