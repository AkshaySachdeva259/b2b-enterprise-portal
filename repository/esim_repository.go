package repository

import (
	"errors"
	"time"

	"com.jetapcglobal.b2b.com/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrInsufficientEsimInventory = errors.New("insufficient released esim inventory")
var ErrTenantEsimNotFound = errors.New("esim not found in tenant inventory")
var ErrTenantHasNoEsims = errors.New("tenant has no esims in inventory")

type EsimRepository interface {
	GetInventoryByTenantID(tenantID string, filter models.EsimInventoryFilter) ([]models.Esim, error)
	AllocateReleasedInventory(tenantID string, quantity int) ([]models.Esim, error)
	AssignCatalogToEsim(tenantID string, receiverUserID string, catalogID int64, iccid string, autoAllocateEsim bool, invoiceID string, requestID string) (*models.Esim, *models.B2BAllocation, bool, error)
}

type esimRepository struct {
	db *gorm.DB
}

func NewEsimRepository(db *gorm.DB) EsimRepository {
	return &esimRepository{db: db}
}

func (r *esimRepository) GetInventoryByTenantID(tenantID string, filter models.EsimInventoryFilter) ([]models.Esim, error) {
	var results []models.Esim

	query := r.db.
		Model(&models.Esim{}).
		Where("deleted_at IS NULL").
		Where("tenant_id = ?", tenantID)

	switch filter {
	case models.EsimInventoryFilterActive:
		query = query.Where("UPPER(COALESCE(telna_status, '')) <> ?", "RELEASED")
	case models.EsimInventoryFilterReleased:
		query = query.Where("UPPER(COALESCE(telna_status, '')) = ?", "RELEASED")
	case models.EsimInventoryFilterInstalled:
		query = query.Where("UPPER(COALESCE(telna_status, '')) = ?", "INSTALLED")
	}

	err := query.
		Order("created_at DESC").
		Find(&results).Error

	return results, err
}

func (r *esimRepository) AllocateReleasedInventory(tenantID string, quantity int) ([]models.Esim, error) {
	var allocated []models.Esim

	err := r.db.Transaction(func(tx *gorm.DB) error {
		inventory, err := r.allocateReleasedInventoryTx(tx, quantity)
		if err != nil {
			return err
		}

		ids := make([]int64, 0, len(inventory))
		for _, esim := range inventory {
			ids = append(ids, esim.ID)
		}

		if err := tx.
			Model(&models.Esim{}).
			Where("id IN ?", ids).
			Updates(map[string]interface{}{
				"tenant_id":  tenantID,
				"updated_at": time.Now(),
				"updated_by": "SYSTEM",
			}).Error; err != nil {
			return err
		}

		if err := tx.
			Where("id IN ?", ids).
			Order("created_at ASC").
			Find(&allocated).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return allocated, nil
}

func (r *esimRepository) AssignCatalogToEsim(tenantID string, receiverUserID string, catalogID int64, iccid string, autoAllocateEsim bool, invoiceID string, requestID string) (*models.Esim, *models.B2BAllocation, bool, error) {
	var selectedEsim models.Esim
	autoAllocatedEsim := false

	allocation := &models.B2BAllocation{
		CatalogID:  catalogID,
		OwnerID:    tenantID,
		ReceiverID: &receiverUserID,
		InvoiceID:  invoiceID,
		Tenant:     &tenantID,
		RequestID:  requestID,
		CreatedBy:  "SYSTEM",
		UpdatedBy:  "SYSTEM",
	}

	err := r.db.Transaction(func(tx *gorm.DB) error {
		query := tx.
			Model(&models.Esim{}).
			Where("deleted_at IS NULL").
			Where("tenant_id = ?", tenantID)

		if iccid != "" {
			query = query.Where("iccid = ?", iccid)
			if err := query.First(&selectedEsim).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrTenantEsimNotFound
				}
				return err
			}
		} else {
			if err := query.Order(clause.Expr{SQL: "RANDOM()"}).Limit(1).First(&selectedEsim).Error; err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}

				if !autoAllocateEsim {
					return ErrTenantHasNoEsims
				}

				inventory, err := r.allocateReleasedInventoryTx(tx, 1)
				if err != nil {
					return err
				}

				selectedEsim = inventory[0]
				autoAllocatedEsim = true

				if err := tx.
					Model(&models.Esim{}).
					Where("id = ?", selectedEsim.ID).
					Updates(map[string]interface{}{
						"tenant_id":  tenantID,
						"updated_at": time.Now(),
						"updated_by": "SYSTEM",
					}).Error; err != nil {
					return err
				}

				if err := tx.
					Where("id = ?", selectedEsim.ID).
					First(&selectedEsim).Error; err != nil {
					return err
				}
			}
		}

		if err := tx.Create(allocation).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, nil, false, err
	}

	return &selectedEsim, allocation, autoAllocatedEsim, nil
}

func (r *esimRepository) allocateReleasedInventoryTx(tx *gorm.DB, quantity int) ([]models.Esim, error) {
	var inventory []models.Esim

	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Model(&models.Esim{}).
		Where("deleted_at IS NULL").
		Where("user_email IS NULL").
		Where("COALESCE(tenant_id, '') = ''").
		Where("COALESCE(user_id, '') = ''").
		Where("UPPER(COALESCE(telna_status, '')) = ?", "RELEASED").
		Where("UPPER(COALESCE(status, '')) = ?", "AVAILABLE").
		Order("created_at ASC").
		Limit(quantity).
		Find(&inventory).Error
	if err != nil {
		return nil, err
	}

	if len(inventory) < quantity {
		return nil, ErrInsufficientEsimInventory
	}

	return inventory, nil
}
