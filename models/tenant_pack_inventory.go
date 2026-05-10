package models

import "time"

type TenantPackInventoryItem struct {
	AllocationID     int64      `json:"allocation_id"`
	CatalogID        string     `json:"catalog_id"`
	PackName         string     `json:"pack_name"`
	PageName         string     `json:"page_name"`
	CountryName      string     `json:"country_name"`
	PriceUSD         string     `json:"price_usd"`
	ReceiverUserID   *string    `json:"receiver_user_id,omitempty"`
	AllocationStatus string     `json:"allocation_status"`
	OrderID          *string    `json:"order_id,omitempty"`
	InvoiceID        string     `json:"invoice_id"`
	RequestID        string     `json:"request_id"`
	TransactionID    *string    `json:"transaction_id,omitempty"`
	CreatedAt        *time.Time `json:"created_at,omitempty"`
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`
}
