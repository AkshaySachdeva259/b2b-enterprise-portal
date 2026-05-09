package services

import (
	"com.jetapcglobal.b2b.com/models"
	"com.jetapcglobal.b2b.com/repository"
)

type DestinationService interface {
	GetAllPages() ([]models.Destination, error)
}

type destinationService struct {
	repo repository.DestinationRepository
}

func NewDestinationService(repo repository.DestinationRepository) DestinationService {
	return &destinationService{repo: repo}
}

func (s *destinationService) GetAllPages() ([]models.Destination, error) {
	return s.repo.GetAll()
}
