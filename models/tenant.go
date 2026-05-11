package models

import "time"

type Tenant struct {
	ID                         int64      `gorm:"column:id;primaryKey"                          json:"id"`
	TenantCode                 *string    `gorm:"column:tenant_code"                            json:"tenant_code,omitempty"`
	CompanyName                *string    `gorm:"column:company_name"                           json:"company_name,omitempty"`
	TenantType                 *string    `gorm:"column:tenant_type"                            json:"tenant_type,omitempty"`
	LegalBusinessName          *string    `gorm:"column:legal_business_name"                    json:"legal_business_name,omitempty"`
	BusinessRegistrationNumber *string    `gorm:"column:business_registration_number"           json:"business_registration_number,omitempty"`
	BusinessType               *string    `gorm:"column:business_type"                          json:"business_type,omitempty"`
	Country                    *string    `gorm:"column:country"                                json:"country,omitempty"`
	City                       *string    `gorm:"column:city"                                   json:"city,omitempty"`
	RegisteredBusinessAddress  *string    `gorm:"column:registered_business_address"            json:"registered_business_address,omitempty"`
	WebsiteURL                 *string    `gorm:"column:website_url"                            json:"website_url,omitempty"`
	TenantCurrency             *string    `gorm:"column:tenant_currency"                        json:"tenant_currency,omitempty"`
	Status                     *string    `gorm:"column:status"                                 json:"status,omitempty"`
	IsActive                   bool       `gorm:"column:is_active"                              json:"is_active"`
	IsDeleted                  bool       `gorm:"column:is_deleted"                             json:"is_deleted"`
	ContactPerson              *string    `gorm:"column:contact_person"                         json:"contact_person,omitempty"`
	CommunicationLang          *string    `gorm:"column:communication_lang"                     json:"communication_lang,omitempty"`
	CreatedAt                  *time.Time `gorm:"column:created_at;autoCreateTime"              json:"created_at,omitempty"`
	CreatedBy                  *string    `gorm:"column:created_by"                             json:"created_by,omitempty"`
	UpdatedAt                  *time.Time `gorm:"column:updated_at;autoUpdateTime"              json:"updated_at,omitempty"`
	UpdatedBy                  *string    `gorm:"column:updated_by"                             json:"updated_by,omitempty"`
}

func (Tenant) TableName() string { return "tbl_tenant" }
