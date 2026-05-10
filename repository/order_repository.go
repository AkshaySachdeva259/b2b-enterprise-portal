package repository

import (
	"com.jetapcglobal.b2b.com/models"
	"gorm.io/gorm"
	"strings"
	"time"
)

type OrderRepository interface {
	ListOrdersByTenantID(tenantID int64, limit int, productType string) ([]models.OrderRecord, error)
	GetCompletedPackOrderSummaryByTenantID(tenantID int64, start time.Time, end time.Time) (float64, int64, error)
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) ListOrdersByTenantID(tenantID int64, limit int, productType string) ([]models.OrderRecord, error) {
	results := make([]models.OrderRecord, 0)

	query := r.db.
		Model(&models.OrderRecord{}).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC NULLS LAST").
		Order("id DESC")

	switch strings.ToLower(strings.TrimSpace(productType)) {
	case models.OrderProductTypeCatalog:
		query = query.Where("LOWER(COALESCE(product_type, '')) = ?", models.OrderProductTypeCatalog)
	case models.OrderProductTypeEsim:
		query = query.Where("LOWER(COALESCE(product_type, '')) = ?", models.OrderProductTypeEsim)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&results).Error
	return results, err
}

func (r *orderRepository) GetCompletedPackOrderSummaryByTenantID(tenantID int64, start time.Time, end time.Time) (float64, int64, error) {
	var result struct {
		TodayRevenueUSD float64 `gorm:"column:today_revenue_usd"`
		TodayPacksSold  int64   `gorm:"column:today_packs_sold"`
	}

	err := r.db.
		Model(&models.OrderRecord{}).
		Select("COALESCE(SUM(total_amount), 0) AS today_revenue_usd, COUNT(*) AS today_packs_sold").
		Where("tenant_id = ?", tenantID).
		Where("LOWER(COALESCE(product_type, '')) = ?", models.OrderProductTypeCatalog).
		Where("LOWER(COALESCE(status, '')) = ?", models.OrderStatusCompleted).
		Where("is_active = ?", true).
		Where("created_at >= ? AND created_at < ?", start, end).
		Scan(&result).Error
	if err != nil {
		return 0, 0, err
	}

	return result.TodayRevenueUSD, result.TodayPacksSold, nil
}
