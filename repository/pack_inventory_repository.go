package repository

import (
	"com.jetapcglobal.b2b.com/models"
	"gorm.io/gorm"
	"strconv"
)

type PackInventoryRepository interface {
	ListByTenantID(tenantID int64) ([]models.TenantPackInventoryItem, error)
}

type packInventoryRepository struct {
	db *gorm.DB
}

func NewPackInventoryRepository(db *gorm.DB) PackInventoryRepository {
	return &packInventoryRepository{db: db}
}

func (r *packInventoryRepository) ListByTenantID(tenantID int64) ([]models.TenantPackInventoryItem, error) {
	const query = `
WITH latest_catalog AS (
	SELECT DISTINCT ON (catalog_id)
		catalog_id,
		name,
		page_name,
		prices
	FROM tbl_catalog
	WHERE deleted_at IS NULL
	ORDER BY catalog_id, created_at DESC, id DESC
)
SELECT
	a.id AS allocation_id,
	a.catalog_id,
	COALESCE(lc.name, '') AS pack_name,
	COALESCE(lc.page_name, '') AS page_name,
	COALESCE(d.display_name, lc.page_name, lc.name, '') AS country_name,
	COALESCE(lc.prices -> 'usd' ->> 'value', '') AS price_usd,
	NULLIF(a.receiver_id, '') AS receiver_user_id,
	COALESCE(NULLIF(a.status, ''), CASE WHEN COALESCE(a.receiver_id, '') = '' THEN 'UNALLOCATED' ELSE 'ASSIGNED' END) AS allocation_status,
	a.order_id,
	a.invoice_id,
	a.request_id,
	a.transaction_id,
	a.created_at,
	a.updated_at
FROM tbl_b2b_allocation a
LEFT JOIN latest_catalog lc ON lc.catalog_id = a.catalog_id
LEFT JOIN tbl_destination d ON d.name = lc.page_name AND d.deleted_at IS NULL
WHERE a.deleted_at IS NULL
	AND a.tenant = ?
ORDER BY a.created_at DESC, a.id DESC;`

	results := make([]models.TenantPackInventoryItem, 0)
	tenantIDText := strconv.FormatInt(tenantID, 10)
	err := r.db.Raw(query, tenantIDText).Scan(&results).Error
	return results, err
}
