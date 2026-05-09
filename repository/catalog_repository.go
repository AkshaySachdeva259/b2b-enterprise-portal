package repository

import (
	"com.jetapcglobal.b2b.com/models"
	"gorm.io/gorm"
)

type CatalogRepository interface {
	GetByPageName(pageName string) ([]models.Catalog, error)
	GetByCatalogIDs(catalogIDs []int64) ([]models.Catalog, error)
	ExistsByCatalogID(catalogID int64) (bool, error)
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

func (r *catalogRepository) GetByCatalogIDs(catalogIDs []int64) ([]models.Catalog, error) {
	results := make([]models.Catalog, 0)
	if len(catalogIDs) == 0 {
		return results, nil
	}

	err := r.db.
		Where("catalog_id IN ? AND deleted_at IS NULL", catalogIDs).
		Order("created_at DESC").
		Find(&results).Error
	return results, err
}

func (r *catalogRepository) ExistsByCatalogID(catalogID int64) (bool, error) {
	var count int64
	err := r.db.
		Model(&models.Catalog{}).
		Where("catalog_id = ? AND deleted_at IS NULL", catalogID).
		Count(&count).Error
	return count > 0, err
}
