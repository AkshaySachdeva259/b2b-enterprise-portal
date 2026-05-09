package services

import (
	"errors"
	"strings"

	"com.jetapcglobal.b2b.com/models"
	"com.jetapcglobal.b2b.com/repository"
	"github.com/google/uuid"
)

var (
	ErrTenantIDRequired           = errors.New("tenant_id is required")
	ErrReceiverUserIDRequired     = errors.New("receiver_user_id is required")
	ErrCatalogIDRequired          = errors.New("catalog_id is required")
	ErrInvalidEsimInventoryFilter = errors.New("status must be one of: active, all, released, installed")
	ErrInvalidEsimQuantity        = errors.New("quantity must be greater than zero")
	ErrInsufficientEsimInventory  = repository.ErrInsufficientEsimInventory
	ErrCatalogNotFound            = errors.New("catalog not found")
	ErrTenantEsimNotFound         = repository.ErrTenantEsimNotFound
	ErrTenantHasNoEsims           = repository.ErrTenantHasNoEsims
)

type EsimService interface {
	GetInventoryByTenantID(tenantID, filter string) ([]models.Esim, error)
	OrderEsims(tenantID string, quantity int) ([]models.Esim, error)
	AssignCatalog(tenantID string, receiverUserID string, catalogID int64, iccid string, autoAllocateEsim bool) (*models.Esim, *models.B2BAllocation, bool, error)
}

type esimService struct {
	repo        repository.EsimRepository
	catalogRepo repository.CatalogRepository
}

func NewEsimService(repo repository.EsimRepository, catalogRepo repository.CatalogRepository) EsimService {
	return &esimService{repo: repo, catalogRepo: catalogRepo}
}

func (s *esimService) GetInventoryByTenantID(tenantID, filter string) ([]models.Esim, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}

	normalizedFilter, err := normalizeEsimInventoryFilter(filter)
	if err != nil {
		return nil, err
	}

	esims, err := s.repo.GetInventoryByTenantID(tenantID, normalizedFilter)
	if err != nil {
		return nil, err
	}

	return esims, nil
}

func (s *esimService) OrderEsims(tenantID string, quantity int) ([]models.Esim, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}

	if quantity <= 0 {
		return nil, ErrInvalidEsimQuantity
	}

	esims, err := s.repo.AllocateReleasedInventory(tenantID, quantity)
	if err != nil {
		if errors.Is(err, repository.ErrInsufficientEsimInventory) {
			return nil, ErrInsufficientEsimInventory
		}
		return nil, err
	}

	return esims, nil
}

func (s *esimService) AssignCatalog(tenantID string, receiverUserID string, catalogID int64, iccid string, autoAllocateEsim bool) (*models.Esim, *models.B2BAllocation, bool, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, nil, false, ErrTenantIDRequired
	}

	receiverUserID = strings.TrimSpace(receiverUserID)
	if receiverUserID == "" {
		return nil, nil, false, ErrReceiverUserIDRequired
	}

	if catalogID <= 0 {
		return nil, nil, false, ErrCatalogIDRequired
	}

	iccid = strings.TrimSpace(iccid)

	exists, err := s.catalogRepo.ExistsByCatalogID(catalogID)
	if err != nil {
		return nil, nil, false, err
	}
	if !exists {
		return nil, nil, false, ErrCatalogNotFound
	}

	invoiceID := uuid.NewString()
	requestID := uuid.NewString()

	esim, allocation, autoAllocatedEsim, err := s.repo.AssignCatalogToEsim(tenantID, receiverUserID, catalogID, iccid, autoAllocateEsim, invoiceID, requestID)
	if err != nil {
		if errors.Is(err, repository.ErrTenantEsimNotFound) {
			return nil, nil, false, ErrTenantEsimNotFound
		}
		if errors.Is(err, repository.ErrTenantHasNoEsims) {
			return nil, nil, false, ErrTenantHasNoEsims
		}
		return nil, nil, false, err
	}

	return esim, allocation, autoAllocatedEsim, nil
}

func normalizeEsimInventoryFilter(filter string) (models.EsimInventoryFilter, error) {
	switch strings.ToLower(strings.TrimSpace(filter)) {
	case "", string(models.EsimInventoryFilterAll):
		return models.EsimInventoryFilterAll, nil
	case string(models.EsimInventoryFilterActive):
		return models.EsimInventoryFilterActive, nil
	case string(models.EsimInventoryFilterReleased):
		return models.EsimInventoryFilterReleased, nil
	case string(models.EsimInventoryFilterInstalled):
		return models.EsimInventoryFilterInstalled, nil
	default:
		return "", ErrInvalidEsimInventoryFilter
	}
}
