package repository

import (
	"com.jetapcglobal.b2b.com/models"
	"gorm.io/gorm"
)

type DestinationRepository interface {
	ListDestinations() ([]models.DestinationOption, error)
}

type destinationRepository struct {
	db *gorm.DB
}

func NewDestinationRepository(db *gorm.DB) DestinationRepository {
	return &destinationRepository{db: db}
}

func (r *destinationRepository) ListDestinations() ([]models.DestinationOption, error) {
	results := make([]models.DestinationOption, 0)
	err := r.db.
		Model(&models.Destination{}).
		Select("name, display_name").
		Where("deleted_at IS NULL AND visibility = ?", true).
		Order("display_name ASC").
		Find(&results).Error
	return results, err
}
