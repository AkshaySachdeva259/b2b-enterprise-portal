package models

import (
	"encoding/json"
	"time"
)

type DestinationOption struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

type PackSummary struct {
	ID             string  `json:"id"`
	CountryName    string  `json:"country_name"`
	DataInGB       float64 `json:"data_in_gb"`
	ValidityInDays int64   `json:"validity_in_days"`
	PriceUSD       string  `json:"price_usd"`
}

type PackDetail struct {
	ID                 string          `json:"id"`
	CountryName        string          `json:"country_name"`
	Vendor             string          `json:"vendor"`
	PackageType        string          `json:"package_type"`
	PackageTypeID      *int64          `json:"package_type_id,omitempty"`
	Name               string          `json:"name"`
	Description        []string        `json:"description,omitempty"`
	Callout            *string         `json:"callout,omitempty"`
	DataInGB           float64         `json:"data_in_gb"`
	VoiceInMins        int64           `json:"voice_in_mins"`
	SMS                int64           `json:"sms"`
	ValidityInDays     int64           `json:"validity_in_days"`
	SupportedCountries []string        `json:"supported_countries,omitempty"`
	Visibility         bool            `json:"visibility"`
	Supports5G         bool            `json:"supports_5g"`
	RedirectURL        *string         `json:"redirect_url,omitempty"`
	Tags               []string        `json:"tags,omitempty"`
	Prices             json.RawMessage `json:"prices"`
	PriceUSD           string          `json:"price_usd"`
	EsimConfigID       *int64          `json:"esim_config_id,omitempty"`
	CatalogType        *string         `json:"catalog_type,omitempty"`
	CatalogMetadata    json.RawMessage `json:"catalog_metadata,omitempty"`
	PageName           *string         `json:"page_name,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}
