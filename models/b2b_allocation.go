package models

import "time"

const (
	AllocationStatusUnallocated = "UNALLOCATED"
	AllocationStatusAssigned    = "ALLOCATED"
)

type B2BAllocation struct {
	ID            int64      `gorm:"column:id;primaryKey"               json:"id"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime"   json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;autoUpdateTime"   json:"updated_at"`
	DeletedAt     *time.Time `gorm:"column:deleted_at"                  json:"deleted_at,omitempty"`
	CreatedBy     string     `gorm:"column:created_by"                  json:"created_by"`
	UpdatedBy     string     `gorm:"column:updated_by"                  json:"updated_by"`
	CatalogID     string     `gorm:"column:catalog_id"                  json:"catalog_id"`
	OwnerID       string     `gorm:"column:owner_id"                    json:"owner_id"`
	ReceiverID    *string    `gorm:"column:receiver_id"                 json:"receiver_id,omitempty"`
	InvoiceID     string     `gorm:"column:invoice_id"                  json:"invoice_id"`
	Tenant        *string    `gorm:"column:tenant"                      json:"tenant,omitempty"`
	RequestID     string     `gorm:"column:request_id"                  json:"request_id"`
	TransactionID *string    `gorm:"column:transaction_id"              json:"transaction_id,omitempty"`
	Status        *string    `gorm:"column:status"                      json:"status,omitempty"`
	OrderID       *string    `gorm:"column:order_id"                    json:"order_id,omitempty"`
}

func (B2BAllocation) TableName() string { return "tbl_b2b_allocation" }
