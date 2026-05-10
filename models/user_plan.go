package models

import "time"

const (
	UserPlanStatusActive = "ACTIVE"
)

type UserPlan struct {
	ID            int64      `gorm:"column:id;primaryKey"           json:"id"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt     *time.Time `gorm:"column:deleted_at"              json:"deleted_at,omitempty"`
	CreatedBy     string     `gorm:"column:created_by"              json:"created_by"`
	UpdatedBy     string     `gorm:"column:updated_by"              json:"updated_by"`
	UserID        string     `gorm:"column:user_id"                 json:"user_id"`
	ICCID         string     `gorm:"column:iccid"                   json:"iccid"`
	Vendor        string     `gorm:"column:vendor"                  json:"vendor"`
	CatalogID     string     `gorm:"column:catalog_id"              json:"catalog_id"`
	AppliedPlanID string     `gorm:"column:applied_plan_id"         json:"applied_plan_id"`
	PlanStatus    string     `gorm:"column:plan_status"             json:"plan_status"`
	RowID         int64      `gorm:"column:rowid"                   json:"rowid"`
	PaymentIntent *string    `gorm:"column:payment_intent"          json:"payment_intent,omitempty"`
	TimeboundID   *string    `gorm:"column:timebound_id"            json:"timebound_id,omitempty"`
}

func (UserPlan) TableName() string { return "tbl_user_plan" }
