package services

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"com.jetapcglobal.b2b.com/models"
	"com.jetapcglobal.b2b.com/repository"
)

const defaultRecentOrderLimit = 5
const defaultOrderHistoryLimit = 20
const maxOrderHistoryLimit = 20

type OrderService interface {
	GetRecentOrdersByTenantID(tenantID int64, limit int, productType string) ([]models.OrderListItem, error)
	GetOrderHistoryByTenantID(tenantID int64, limit int, productType string) ([]models.OrderListItem, error)
	GetEsimOrderHistoryByTenantID(tenantID int64, limit int) ([]models.OrderListItem, error)
	GetTodayOrderSummaryByTenantID(tenantID int64, location *time.Location) (*models.OrderTodaySummary, error)
}

type orderService struct {
	repo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) OrderService {
	return &orderService{repo: repo}
}

func (s *orderService) GetRecentOrdersByTenantID(tenantID int64, limit int, productType string) ([]models.OrderListItem, error) {
	return s.listOrdersByTenantID(tenantID, resolveOrderLimit(limit, defaultRecentOrderLimit), normalizeOrderProductTypeFilter(productType))
}

func (s *orderService) GetOrderHistoryByTenantID(tenantID int64, limit int, productType string) ([]models.OrderListItem, error) {
	return s.listOrdersByTenantID(tenantID, resolveOrderLimit(limit, defaultOrderHistoryLimit), normalizeOrderProductTypeFilter(productType))
}

func (s *orderService) GetEsimOrderHistoryByTenantID(tenantID int64, limit int) ([]models.OrderListItem, error) {
	return s.listOrdersByTenantID(tenantID, resolveOrderLimit(limit, defaultOrderHistoryLimit), models.OrderProductTypeEsim)
}

func (s *orderService) GetTodayOrderSummaryByTenantID(tenantID int64, location *time.Location) (*models.OrderTodaySummary, error) {
	if location == nil {
		location = time.Local
	}

	now := time.Now().In(location)
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	dayEnd := dayStart.AddDate(0, 0, 1)

	todayRevenueUSD, todayPacksSold, err := s.repo.GetCompletedPackOrderSummaryByTenantID(tenantID, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}

	return &models.OrderTodaySummary{
		TenantID:        tenantID,
		Date:            dayStart.Format("2006-01-02"),
		Timezone:        location.String(),
		Currency:        "USD",
		TodayRevenueUSD: strconv.FormatFloat(todayRevenueUSD, 'f', 2, 64),
		TodayPacksSold:  todayPacksSold,
	}, nil
}

func (s *orderService) listOrdersByTenantID(tenantID int64, limit int, productType string) ([]models.OrderListItem, error) {
	orders, err := s.repo.ListOrdersByTenantID(tenantID, limit, productType)
	if err != nil {
		return nil, err
	}

	result := make([]models.OrderListItem, 0, len(orders))
	for _, order := range orders {
		requestObject := models.OrderRequestObject{}
		if len(order.RequestObject) > 0 {
			_ = json.Unmarshal(order.RequestObject, &requestObject)
		}

		result = append(result, models.OrderListItem{
			OrderID:          order.OrderID,
			TransactionID:    strings.TrimSpace(requestObject.PaymentTransactionID),
			TenantID:         order.TenantID,
			ProductType:      resolveOrderProductType(order.ProductType),
			CatalogID:        requestObject.CatalogID,
			PackName:         strings.TrimSpace(requestObject.PackName),
			PageName:         strings.TrimSpace(requestObject.PageName),
			ReceiverUserID:   strings.TrimSpace(requestObject.ReceiverUserID),
			OriginalPriceUSD: fallbackOrderAmount(strings.TrimSpace(requestObject.OriginalPriceUSD), order.TotalAmount),
			SoldPriceUSD:     fallbackOrderAmount(strings.TrimSpace(requestObject.SoldPriceUSD), order.TotalAmount),
			Currency:         strings.TrimSpace(requestObject.Currency),
			Status:           strings.TrimSpace(order.Status),
			AssignmentStatus: firstNonEmptyOrderValue(strings.TrimSpace(requestObject.AssignmentStatus), strings.TrimSpace(order.Status)),
			CreatedAt:        order.CreatedAt,
		})
	}

	return result, nil
}

func resolveOrderLimit(limit int, defaultLimit int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxOrderHistoryLimit {
		return maxOrderHistoryLimit
	}
	return limit
}

func fallbackOrderAmount(value string, amount *float64) string {
	if value != "" {
		return value
	}
	if amount == nil {
		return ""
	}
	return strconv.FormatFloat(*amount, 'f', 2, 64)
}

func firstNonEmptyOrderValue(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func normalizeOrderProductTypeFilter(productType string) string {
	switch strings.ToLower(strings.TrimSpace(productType)) {
	case models.OrderProductTypeCatalog:
		return models.OrderProductTypeCatalog
	case models.OrderProductTypeEsim:
		return models.OrderProductTypeEsim
	default:
		return ""
	}
}

func resolveOrderProductType(productType string) string {
	productType = normalizeOrderProductTypeFilter(productType)
	return productType
}
