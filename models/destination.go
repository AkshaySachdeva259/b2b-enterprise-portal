package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Destination struct {
	ID              uuid.UUID       `gorm:"column:id;primaryKey;type:uuid" json:"id"`
	CreatedAt       time.Time       `gorm:"column:created_at;autoCreateTime"  json:"created_at"`
	UpdatedAt       time.Time       `gorm:"column:updated_at;autoUpdateTime"  json:"updated_at"`
	DeletedAt       *time.Time      `gorm:"column:deleted_at"                 json:"deleted_at,omitempty"`
	Name            string          `gorm:"column:name;uniqueIndex"           json:"name"`
	DisplayName     string          `gorm:"column:display_name"               json:"display_name"`
	DestinationType string          `gorm:"column:destination_type"           json:"destination_type"`
	SeoMetadata     *string         `gorm:"column:seo_metadata"               json:"seo_metadata,omitempty"`
	Visibility      bool            `gorm:"column:visibility"                 json:"visibility"`
	Aliases         pq.StringArray  `gorm:"column:aliases;type:text[]"        json:"aliases,omitempty"`
	Flag            *string         `gorm:"column:flag"                       json:"flag,omitempty"`
	Tags            pq.StringArray  `gorm:"column:tags;type:text[]"           json:"tags,omitempty"`
	PageMetadata    json.RawMessage `gorm:"column:page_metadata;type:jsonb"   json:"page_metadata,omitempty"`
	Region          *string         `gorm:"column:region"                     json:"region,omitempty"`
}

func (Destination) TableName() string { return "tbl_destination" }
