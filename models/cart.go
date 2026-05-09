package models

import (
	"time"

	"github.com/google/uuid"
)

const DefaultCartCurrency = "usd"

type Cart struct {
	ID        uuid.UUID  `gorm:"column:id;primaryKey;type:uuid"   json:"id"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"                json:"deleted_at,omitempty"`
	UserID    string     `gorm:"column:user_id;index"             json:"user_id"`
	Currency  string     `gorm:"column:currency"                  json:"currency"`
	Items     []CartItem `gorm:"foreignKey:CartID;references:ID"  json:"-"`
}

func (Cart) TableName() string { return "tbl_cart" }

type CartItem struct {
	ID        uuid.UUID  `gorm:"column:id;primaryKey;type:uuid"   json:"id"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"                json:"deleted_at,omitempty"`
	CartID    uuid.UUID  `gorm:"column:cart_id;type:uuid;index"   json:"cart_id"`
	CatalogID int64      `gorm:"column:catalog_id;index"          json:"catalog_id"`
	Quantity  int        `gorm:"column:quantity"                  json:"quantity"`
}

func (CartItem) TableName() string { return "tbl_cart_items" }

type CartUpdateItemInput struct {
	CatalogID int64 `json:"catalog_id"`
	Quantity  int   `json:"quantity"`
}

type CartPriceBreakdown struct {
	Currency       string `json:"currency"`
	Symbol         string `json:"symbol,omitempty"`
	UnitAmount     string `json:"unit_amount"`
	UnitListAmount string `json:"unit_list_amount"`
	LineAmount     string `json:"line_amount"`
	LineListAmount string `json:"line_list_amount"`
	LineSavings    string `json:"line_savings"`
}

type CartLineItem struct {
	ID        uuid.UUID          `json:"id"`
	CatalogID int64              `json:"catalog_id"`
	Quantity  int                `json:"quantity"`
	Catalog   Catalog            `json:"catalog"`
	Pricing   CartPriceBreakdown `json:"pricing"`
}

type CartSummary struct {
	Currency      string `json:"currency"`
	Symbol        string `json:"symbol,omitempty"`
	ItemCount     int    `json:"item_count"`
	QuantityTotal int    `json:"quantity_total"`
	Subtotal      string `json:"subtotal"`
	ListTotal     string `json:"list_total"`
	Savings       string `json:"savings"`
}

type CartDetail struct {
	ID        uuid.UUID      `json:"id"`
	UserID    string         `json:"user_id"`
	Currency  string         `json:"currency"`
	Items     []CartLineItem `json:"items"`
	Summary   CartSummary    `json:"summary"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}
