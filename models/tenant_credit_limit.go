package models

import (
	"time"
)

type TenantCreditLimit struct {
	// RowID             int64      `gorm:"column:rowid;primaryKey"               json:"rowid"`
	ID                int64      `gorm:"column:id;primaryKey"                             json:"id"`
	TenantID          int64      `gorm:"column:tenant_id;type:bigint;index"    json:"tenant_id"`
	CommercialType    *string    `gorm:"column:commercial_type"                 json:"commercial_type,omitempty"`
	CreditTermsPeriod *string    `gorm:"column:credit_terms_period"             json:"credit_terms_period,omitempty"`
	CreditLimit       *float64   `gorm:"column:credit_limit;type:numeric(18,2)" json:"credit_limit,omitempty"`
	IsActive          bool       `gorm:"column:is_active"                       json:"is_active"`
	CreatedAt         *time.Time `gorm:"column:created_at;autoCreateTime"       json:"created_at,omitempty"`
	CreatedBy         *string    `gorm:"column:created_by"                      json:"created_by,omitempty"`
	UpdatedAt         *time.Time `gorm:"column:updated_at;autoUpdateTime"       json:"updated_at,omitempty"`
	UpdatedBy         *string    `gorm:"column:updated_by"                      json:"updated_by,omitempty"`
	IsDeleted         bool       `gorm:"column:is_deleted"                      json:"is_deleted"`
}

func (TenantCreditLimit) TableName() string { return "tbl_tenant_commercials" }
