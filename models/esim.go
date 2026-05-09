package models

import "time"

type EsimInventoryFilter string

const (
	EsimInventoryFilterAll       EsimInventoryFilter = "all"
	EsimInventoryFilterActive    EsimInventoryFilter = "active"
	EsimInventoryFilterReleased  EsimInventoryFilter = "released"
	EsimInventoryFilterInstalled EsimInventoryFilter = "installed"
)

type Esim struct {
	ID             int64      `gorm:"column:id;primaryKey"            json:"id"`
	CreatedAt      *time.Time `gorm:"column:created_at"               json:"created_at,omitempty"`
	UpdatedAt      *time.Time `gorm:"column:updated_at"               json:"updated_at,omitempty"`
	UpdatedBy      string     `gorm:"column:updated_by"               json:"updated_by"`
	ICCID          string     `gorm:"column:iccid"                    json:"iccid"`
	MatchingID     string     `gorm:"column:matching_id"              json:"matching_id"`
	SmDpAddress    string     `gorm:"column:sm_dp_address"            json:"sm_dp_address"`
	QRCode         string     `gorm:"column:qr_code"                  json:"qr_code"`
	Status         string     `gorm:"column:status"                   json:"status"`
	CreatedBy      string     `gorm:"column:created_by"               json:"created_by"`
	UserEmail      *string    `gorm:"column:user_email"               json:"user_email,omitempty"`
	UserName       *string    `gorm:"column:user_name"                json:"user_name,omitempty"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"               json:"deleted_at,omitempty"`
	TelnaStatus    *string    `gorm:"column:telna_status"             json:"telna_status,omitempty"`
	ProvisionedAt  *time.Time `gorm:"column:provisioned_at"           json:"provisioned_at,omitempty"`
	TenantID       *string    `gorm:"column:tenant_id"                json:"tenant_id,omitempty"`
	UserID         *string    `gorm:"column:user_id"                  json:"user_id,omitempty"`
	Vendor         *string    `gorm:"column:vendor"                   json:"vendor,omitempty"`
	IMSIProfileID  *string    `gorm:"column:imsi_profile_id"          json:"imsi_profile_id,omitempty"`
	VendorStatus   *string    `gorm:"column:vendor_status"            json:"vendor_status,omitempty"`
	WhitelistID    *string    `gorm:"column:whitelist_id"             json:"whitelist_id,omitempty"`
	EsimConfigID   int64      `gorm:"column:esim_config_id"           json:"esim_config_id"`
	BillingGroupID int64      `gorm:"column:billing_group_id"         json:"billing_group_id"`
}

func (Esim) TableName() string { return "tbl_esim" }
