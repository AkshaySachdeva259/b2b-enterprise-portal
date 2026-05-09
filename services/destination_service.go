package services

import (
	"com.jetapcglobal.b2b.com/models"
	"com.jetapcglobal.b2b.com/repository"
)

type DestinationService interface {
	ListDestinations() ([]models.DestinationOption, error)
}

type destinationService struct {
	repo repository.DestinationRepository
}

func NewDestinationService(repo repository.DestinationRepository) DestinationService {
	return &destinationService{repo: repo}
}

func (s *destinationService) ListDestinations() ([]models.DestinationOption, error) {
	return s.repo.ListDestinations()
}
