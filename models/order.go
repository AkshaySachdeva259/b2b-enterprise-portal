package models

import (
	"encoding/json"
	"time"
)

const (
	OrderProductPack     = "PACK"
	OrderStatusPending   = "pending"
	OrderStatusCompleted = "completed"
	OrderStatusFailed    = "failed"
)

type OrderRecord struct {
	ID            int64           `gorm:"column:id;primaryKey"                  json:"id"`
	TenantID      int64           `gorm:"column:tenant_id;type:bigint;index"    json:"tenant_id"`
	OrderID       string          `gorm:"column:order_id"                       json:"order_id"`
	TotalAmount   *float64        `gorm:"column:total_amount;type:numeric(18,2)" json:"total_amount,omitempty"`
	RequestObject json.RawMessage `gorm:"column:request_object;type:jsonb"      json:"request_object,omitempty"`
	Status        string          `gorm:"column:status"                         json:"status"`
	IsActive      bool            `gorm:"column:is_active"                      json:"is_active"`
	CreatedAt     *time.Time      `gorm:"column:created_at;autoCreateTime"      json:"created_at,omitempty"`
	CreatedBy     *string         `gorm:"column:created_by"                     json:"created_by,omitempty"`
	UpdatedAt     *time.Time      `gorm:"column:updated_at;autoUpdateTime"      json:"updated_at,omitempty"`
	UpdatedBy     *string         `gorm:"column:updated_by"                     json:"updated_by,omitempty"`
}

func (OrderRecord) TableName() string { return "tbl_order_record" }

type OrderRequestObject struct {
	Product              string `json:"product,omitempty"`
	TenantID             int64  `json:"tenant_id,omitempty"`
	CatalogID            int64  `json:"catalog_id,omitempty"`
	PackName             string `json:"pack_name,omitempty"`
	PageName             string `json:"page_name,omitempty"`
	ReceiverUserID       string `json:"receiver_user_id,omitempty"`
	OriginalPriceUSD     string `json:"original_price_usd,omitempty"`
	SoldPriceUSD         string `json:"sold_price_usd,omitempty"`
	Currency             string `json:"currency,omitempty"`
	PaymentTransactionID string `json:"payment_transaction_id,omitempty"`
	InvoiceID            string `json:"invoice_id,omitempty"`
	RequestID            string `json:"request_id,omitempty"`
	WalletBalanceBefore  string `json:"wallet_balance_before,omitempty"`
	WalletBalanceAfter   string `json:"wallet_balance_after,omitempty"`
	AssignmentStatus     string `json:"assignment_status,omitempty"`
	EsimICCID            string `json:"esim_iccid,omitempty"`
	EsimSource           string `json:"esim_source,omitempty"`
}

type OrderListItem struct {
	OrderID          string     `json:"order_id"`
	TransactionID    string     `json:"transaction_id,omitempty"`
	TenantID         int64      `json:"tenant_id"`
	CatalogID        int64      `json:"catalog_id,omitempty"`
	PackName         string     `json:"pack_name,omitempty"`
	PageName         string     `json:"page_name,omitempty"`
	ReceiverUserID   string     `json:"receiver_user_id,omitempty"`
	OriginalPriceUSD string     `json:"original_price_usd,omitempty"`
	SoldPriceUSD     string     `json:"sold_price_usd,omitempty"`
	Currency         string     `json:"currency,omitempty"`
	Status           string     `json:"status"`
	AssignmentStatus string     `json:"assignment_status,omitempty"`
	CreatedAt        *time.Time `json:"created_at,omitempty"`
}

type OrderTodaySummary struct {
	TenantID        int64  `json:"tenant_id"`
	Date            string `json:"date"`
	Timezone        string `json:"timezone"`
	Currency        string `json:"currency"`
	TodayRevenueUSD string `json:"today_revenue_usd"`
	TodayPacksSold  int64  `json:"today_packs_sold"`
}
