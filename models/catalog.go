package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Catalog struct {
	ID              uuid.UUID       `gorm:"column:id;primaryKey;type:uuid"      json:"id"`
	CatalogID       *string         `gorm:"column:catalog_id"                   json:"catalog_id,omitempty"`
	CreatedAt       time.Time       `gorm:"column:created_at;autoCreateTime"    json:"created_at"`
	UpdatedAt       time.Time       `gorm:"column:updated_at;autoUpdateTime"    json:"updated_at"`
	DeletedAt       *time.Time      `gorm:"column:deleted_at"                   json:"deleted_at,omitempty"`
	Vendor          string          `gorm:"column:vendor"                       json:"vendor"`
	PackageType     string          `gorm:"column:package_type"                 json:"package_type"`
	PackageTypeID   *int64          `gorm:"column:package_type_id"              json:"package_type_id,omitempty"`
	Name            string          `gorm:"column:name"                         json:"name"`
	Description     pq.StringArray  `gorm:"column:description;type:text[]"      json:"description,omitempty"`
	Callout         *string         `gorm:"column:callout"                      json:"callout,omitempty"`
	DataInGB        float64         `gorm:"column:data_in_gb"                   json:"data_in_gb"`
	VoiceInMins     int64           `gorm:"column:voice_in_mins"                json:"voice_in_mins"`
	SMS             int64           `gorm:"column:sms"                          json:"sms"`
	ValidityInDays  int64           `gorm:"column:validity_in_days"             json:"validity_in_days"`
	Countries       pq.StringArray  `gorm:"column:countries;type:text[]"        json:"countries"`
	Visibility      bool            `gorm:"column:visibility"                   json:"visibility"`
	Supports5G      bool            `gorm:"column:supports_5g"                  json:"supports_5g"`
	RedirectURL     *string         `gorm:"column:redirect_url"                 json:"redirect_url,omitempty"`
	Tags            pq.StringArray  `gorm:"column:tags;type:text[]"             json:"tags"`
	Prices          json.RawMessage `gorm:"column:prices;type:jsonb"            json:"prices"`
	EsimConfigID    *int64          `gorm:"column:esim_config_id"               json:"esim_config_id,omitempty"`
	CatalogType     *string         `gorm:"column:catalog_type"                 json:"catalog_type,omitempty"`
	CatalogMetadata json.RawMessage `gorm:"column:catalog_metadata;type:jsonb"  json:"catalog_metadata,omitempty"`
	PageID          *uuid.UUID      `gorm:"column:page_id;type:uuid"            json:"page_id,omitempty"`
	PageName        *string         `gorm:"column:page_name"                    json:"page_name,omitempty"`
}

func (Catalog) TableName() string { return "tbl_catalog" }
