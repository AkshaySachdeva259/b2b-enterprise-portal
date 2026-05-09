package repository

import (
	"com.jetapcglobal.b2b.com/models"
	"gorm.io/gorm"
)

type DestinationRepository interface {
	GetAll() ([]models.Destination, error)
}

type destinationRepository struct {
	db *gorm.DB
}

func NewDestinationRepository(db *gorm.DB) DestinationRepository {
	return &destinationRepository{db: db}
}

func (r *destinationRepository) GetAll() ([]models.Destination, error) {
	var results []models.Destination
	err := r.db.
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Find(&results).Error
	return results, err
}
