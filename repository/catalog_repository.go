package repository

import (
	"errors"

	"com.jetapcglobal.b2b.com/models"
	"gorm.io/gorm"
)

type CatalogRepository interface {
	ListPacks(pageName string, limit int) ([]models.Catalog, error)
	GetByCatalogID(catalogID int64) (*models.Catalog, error)
	GetByCatalogIDs(catalogIDs []int64) ([]models.Catalog, error)
	ExistsByCatalogID(catalogID int64) (bool, error)
}

type catalogRepository struct {
	db *gorm.DB
}

func NewCatalogRepository(db *gorm.DB) CatalogRepository {
	return &catalogRepository{db: db}
}

func (r *catalogRepository) ListPacks(pageName string, limit int) ([]models.Catalog, error) {
	var results []models.Catalog

	query := r.db.
		Where("deleted_at IS NULL AND visibility = ?", true).
		Order("created_at DESC")

	if pageName != "" {
		query = query.Where("page_name = ?", pageName)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&results).Error
	return results, err
}

func (r *catalogRepository) GetByCatalogID(catalogID int64) (*models.Catalog, error) {
	var result models.Catalog
	err := r.db.
		Where("catalog_id = ? AND deleted_at IS NULL AND visibility = ?", catalogID, true).
		Order("created_at DESC").
		First(&result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &result, nil
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
