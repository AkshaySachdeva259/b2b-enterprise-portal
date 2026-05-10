package models

import "time"

type TenantWallet struct {
	ID              int64      `gorm:"column:id;primaryKey"                    json:"id"`
	TenantID        int64      `gorm:"column:tenant_id;type:bigint;index"      json:"tenant_id"`
	Currency        *string    `gorm:"column:currency"                         json:"currency,omitempty"`
	CreditLimit     *float64   `gorm:"column:credit_limit;type:numeric(18,2)"  json:"credit_limit,omitempty"`
	AvailableCredit *float64   `gorm:"column:available_credit;type:numeric(18,2)" json:"available_credit,omitempty"`
	ReservedCredit  *float64   `gorm:"column:reserved_credit;type:numeric(18,2)" json:"reserved_credit,omitempty"`
	UsedCredit      *float64   `gorm:"column:used_credit;type:numeric(18,2)"   json:"used_credit,omitempty"`
	Status          *string    `gorm:"column:status"                           json:"status,omitempty"`
	Version         *int       `gorm:"column:version"                          json:"version,omitempty"`
	CreatedAt       *time.Time `gorm:"column:created_at"                       json:"created_at,omitempty"`
	UpdatedAt       *time.Time `gorm:"column:updated_at"                       json:"updated_at,omitempty"`
}

func (TenantWallet) TableName() string { return "tbl_tenant_wallet" }

type CreditLedgerTransaction struct {
	ID                int64      `gorm:"column:id;primaryKey"                     json:"id"`
	TenantID          int64      `gorm:"column:tenant_id;type:bigint;index"       json:"tenant_id"`
	Currency          *string    `gorm:"column:currency"                          json:"currency,omitempty"`
	TransactionAmount *float64   `gorm:"column:transaction_amount;type:numeric(18,2)" json:"transaction_amount,omitempty"`
	Status            *string    `gorm:"column:status"                            json:"status,omitempty"`
	Product           *string    `gorm:"column:product"                           json:"product,omitempty"`
	OrderID           *string    `gorm:"column:order_id"                          json:"order_id,omitempty"`
	TransactionType   *string    `gorm:"column:transaction_type"                  json:"transaction_type,omitempty"`
	TransactionID     *string    `gorm:"column:transaction_id"                    json:"transaction_id,omitempty"`
	CreatedAt         *time.Time `gorm:"column:created_at"                        json:"created_at,omitempty"`
	CreatedBy         *string    `gorm:"column:created_by"                        json:"created_by,omitempty"`
	UpdatedAt         *time.Time `gorm:"column:updated_at"                        json:"updated_at,omitempty"`
	UpdatedBy         *string    `gorm:"column:updated_by"                        json:"updated_by,omitempty"`
}

func (CreditLedgerTransaction) TableName() string { return "tbl_credit_ledger" }

type WalletTransaction struct {
	ID              int64      `json:"id"`
	Currency        *string    `json:"currency,omitempty"`
	Amount          float64    `json:"amount"`
	Status          *string    `json:"status,omitempty"`
	Product         *string    `json:"product,omitempty"`
	OrderID         *string    `json:"order_id,omitempty"`
	TransactionType *string    `json:"transaction_type,omitempty"`
	TransactionID   *string    `json:"transaction_id,omitempty"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
}

type TenantWalletSummary struct {
	TenantID         int64               `json:"tenant_id"`
	Currency         *string             `json:"currency,omitempty"`
	AvailableBalance float64             `json:"available_balance"`
	Status           *string             `json:"status,omitempty"`
	LastTransactions []WalletTransaction `json:"last_transactions"`
}
