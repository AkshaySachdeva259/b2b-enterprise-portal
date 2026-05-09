package services

import (
	"errors"
	"strconv"
	"strings"

	"com.jetapcglobal.b2b.com/models"
	"com.jetapcglobal.b2b.com/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrTenantIDMustBeInt64 = errors.New("tenant_id must be a valid int64")
var ErrTenantWalletInactive = errors.New("tenant wallet is not active")
var ErrUnsupportedTenantWalletCurrency = errors.New("tenant wallet currency is not supported")
var ErrCatalogUSDPriceUnavailable = errors.New("catalog usd price is not available")

type InsufficientWalletBalanceError = repository.InsufficientWalletBalanceError

type PackAssignmentService interface {
	AssignPack(tenantID string, receiverUserID string, catalogID int64) (*models.PackAssignmentResult, error)
}

type packAssignmentService struct {
	esimRepo    repository.EsimRepository
	catalogRepo repository.CatalogRepository
}

func NewPackAssignmentService(esimRepo repository.EsimRepository, catalogRepo repository.CatalogRepository) PackAssignmentService {
	return &packAssignmentService{
		esimRepo:    esimRepo,
		catalogRepo: catalogRepo,
	}
}

func (s *packAssignmentService) AssignPack(tenantID string, receiverUserID string, catalogID int64) (*models.PackAssignmentResult, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}

	receiverUserID = strings.TrimSpace(receiverUserID)
	if receiverUserID == "" {
		return nil, ErrReceiverUserIDRequired
	}

	if catalogID <= 0 {
		return nil, ErrCatalogIDRequired
	}

	tenantIDInt, err := strconv.ParseInt(tenantID, 10, 64)
	if err != nil {
		return nil, ErrTenantIDMustBeInt64
	}

	catalog, err := s.catalogRepo.GetByCatalogID(catalogID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCatalogNotFound
		}
		return nil, err
	}

	price, err := extractCatalogPrice(*catalog, "usd")
	if err != nil {
		return nil, ErrCatalogUSDPriceUnavailable
	}

	priceAmount, err := parseDecimal(price.Value)
	if err != nil {
		return nil, ErrCatalogUSDPriceUnavailable
	}

	orderID := uuid.NewString()
	orderRequest := models.OrderRequestObject{
		Product:          models.OrderProductPack,
		TenantID:         tenantIDInt,
		CatalogID:        catalogID,
		PackName:         strings.TrimSpace(catalog.Name),
		PageName:         firstNonEmptyOrderValue(valueOrEmptyString(catalog.PageName)),
		ReceiverUserID:   receiverUserID,
		OriginalPriceUSD: strings.TrimSpace(price.ListAmount),
		SoldPriceUSD:     formatDecimal(priceAmount),
		Currency:         "USD",
	}

	result, err := s.esimRepo.PurchaseAndAssignCatalog(
		tenantIDInt,
		receiverUserID,
		catalogID,
		formatDecimal(priceAmount),
		"USD",
		orderID,
		uuid.NewString(),
		uuid.NewString(),
		newPackAssignmentTransactionID(),
		orderRequest,
	)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrPackAssignmentTenantWalletNotFound):
			return nil, ErrTenantWalletNotFound
		case errors.Is(err, repository.ErrPackAssignmentTenantWalletInactive):
			return nil, ErrTenantWalletInactive
		case errors.Is(err, repository.ErrPackAssignmentWalletCurrencyUnsupported):
			return nil, ErrUnsupportedTenantWalletCurrency
		case errors.Is(err, repository.ErrInsufficientEsimInventory):
			return nil, ErrInsufficientEsimInventory
		}

		var balanceErr *repository.InsufficientWalletBalanceError
		if errors.As(err, &balanceErr) {
			return nil, err
		}

		return nil, err
	}

	return result, nil
}

func newPackAssignmentTransactionID() string {
	transactionID := strings.ReplaceAll(uuid.NewString(), "-", "")
	if len(transactionID) <= 10 {
		return transactionID
	}
	return transactionID[:10]
}

func valueOrEmptyString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
