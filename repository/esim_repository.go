package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"math/big"
	"strconv"
	"strings"
	"time"

	"com.jetapcglobal.b2b.com/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrInsufficientEsimInventory = errors.New("insufficient released esim inventory")
var ErrTenantEsimNotFound = errors.New("esim not found in tenant inventory")
var ErrTenantHasNoEsims = errors.New("tenant has no esims in inventory")
var ErrPackAssignmentTenantWalletNotFound = errors.New("tenant wallet not found")
var ErrPackAssignmentTenantWalletInactive = errors.New("tenant wallet is not active")
var ErrPackAssignmentWalletCurrencyUnsupported = errors.New("tenant wallet currency is not supported")

type InsufficientWalletBalanceError struct {
	AvailableBalance string
	RequiredAmount   string
	Currency         string
}

func (e *InsufficientWalletBalanceError) Error() string {
	return "insufficient wallet balance"
}

type EsimRepository interface {
	GetInventoryByTenantID(tenantID string, filter models.EsimInventoryFilter) ([]models.Esim, error)
	AllocateReleasedInventory(tenantID string, quantity int) ([]models.Esim, error)
	AssignCatalogToEsim(tenantID string, receiverUserID string, catalogID int64, iccid string, autoAllocateEsim bool, invoiceID string, requestID string) (*models.Esim, *models.B2BAllocation, bool, error)
	PurchaseAndAssignCatalog(tenantID int64, receiverUserID string, catalogID int64, amountUSD string, currency string, orderID string, invoiceID string, requestID string, transactionID string, orderRequest models.OrderRequestObject) (*models.PackAssignmentResult, error)
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
		if err := r.createUserPlanEntryTx(tx, &selectedEsim, receiverUserID, catalogID, invoiceID, requestID, "SYSTEM"); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, nil, false, err
	}

	return &selectedEsim, allocation, autoAllocatedEsim, nil
}

func (r *esimRepository) PurchaseAndAssignCatalog(tenantID int64, receiverUserID string, catalogID int64, amountUSD string, currency string, orderID string, invoiceID string, requestID string, transactionID string, orderRequest models.OrderRequestObject) (*models.PackAssignmentResult, error) {
	receiverUserID = strings.TrimSpace(receiverUserID)

	result := &models.PackAssignmentResult{
		OrderID:          orderID,
		TenantID:         tenantID,
		ReceiverUserID:   receiverUserID,
		CatalogID:        catalogID,
		InvoiceID:        invoiceID,
		RequestID:        requestID,
		TransactionID:    transactionID,
		AmountChargedUSD: "0.00",
	}

	if receiverUserID != "" {
		reusedResult, reused, err := r.assignUnallocatedPackWithoutPurchase(tenantID, receiverUserID, catalogID)
		if err != nil {
			return nil, err
		}
		if reused {
			return reusedResult, nil
		}
	}

	normalizedCurrency := strings.ToUpper(strings.TrimSpace(currency))
	if normalizedCurrency == "" {
		normalizedCurrency = "USD"
	}

	amountRat, err := parseMoney(amountUSD)
	if err != nil {
		return nil, err
	}

	amountFloat, err := strconv.ParseFloat(formatMoney(amountRat), 64)
	if err != nil {
		return nil, err
	}

	tenantIDString := strconv.FormatInt(tenantID, 10)
	systemUser := "SYSTEM"
	ledgerInitiatedCreated := false

	orderRequest.PaymentTransactionID = transactionID
	orderRequest.InvoiceID = invoiceID
	orderRequest.RequestID = requestID
	orderRequest.AssignmentStatus = models.OrderStatusInitiated

	initiatedOrderRequest, err := json.Marshal(orderRequest)
	if err != nil {
		return nil, err
	}

	if err := r.createOrderRecord(&models.OrderRecord{
		TenantID:      tenantID,
		OrderID:       orderID,
		TotalAmount:   &amountFloat,
		RequestObject: initiatedOrderRequest,
		Status:        models.OrderStatusInitiated,
		IsActive:      true,
		CreatedBy:     &systemUser,
		UpdatedBy:     &systemUser,
	}); err != nil {
		return nil, err
	}

	err = r.db.Transaction(func(tx *gorm.DB) error {
		if receiverUserID != "" {
			if err := lockReceiverAssignmentTx(tx, receiverUserID); err != nil {
				return err
			}
		}

		wallet, err := lockTenantWalletTx(tx, tenantID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrPackAssignmentTenantWalletNotFound
			}
			return err
		}

		walletCurrency := normalizeWalletCurrency(wallet.Currency)
		result.WalletCurrency = walletCurrency
		result.WalletBalanceBefore = moneyStringFromFloat(wallet.AvailableCredit)

		if !strings.EqualFold(valueOrEmpty(wallet.Status), "ACTIVE") {
			return ErrPackAssignmentTenantWalletInactive
		}

		if walletCurrency != normalizedCurrency {
			return ErrPackAssignmentWalletCurrencyUnsupported
		}

		availableBalanceRat, err := parseMoney(result.WalletBalanceBefore)
		if err != nil {
			return err
		}

		now := time.Now()
		orderRequest.WalletBalanceBefore = result.WalletBalanceBefore

		if availableBalanceRat.Cmp(amountRat) < 0 {
			return &InsufficientWalletBalanceError{
				AvailableBalance: result.WalletBalanceBefore,
				RequiredAmount:   formatMoney(amountRat),
				Currency:         walletCurrency,
			}
		}

		if err := r.createCreditLedgerEntryTx(tx, models.CreditLedgerTransaction{
			TenantID:          tenantID,
			Currency:          stringPointer(walletCurrency),
			TransactionAmount: &amountFloat,
			Status:            stringPointer("INITIATED"),
			Product:           stringPointer("PACK"),
			OrderID:           &orderID,
			TransactionType:   stringPointer("PACK_ASSIGNMENT"),
			TransactionID:     &transactionID,
			CreatedBy:         &systemUser,
			UpdatedBy:         &systemUser,
		}); err != nil {
			return err
		}
		ledgerInitiatedCreated = true

		if err := tx.
			Model(&models.TenantWallet{}).
			Where("id = ?", wallet.ID).
			Updates(map[string]interface{}{
				"available_credit": gorm.Expr("COALESCE(available_credit, 0) - CAST(? AS numeric)", formatMoney(amountRat)),
				"used_credit":      gorm.Expr("COALESCE(used_credit, 0) + CAST(? AS numeric)", formatMoney(amountRat)),
				"updated_at":       now,
				"version":          gorm.Expr("COALESCE(version, 0) + 1"),
			}).Error; err != nil {
			return err
		}

		completedLedger := models.CreditLedgerTransaction{
			TenantID:          tenantID,
			Currency:          stringPointer(walletCurrency),
			TransactionAmount: &amountFloat,
			Status:            stringPointer("COMPLETED"),
			Product:           stringPointer("PACK"),
			OrderID:           &orderID,
			TransactionType:   stringPointer("PACK_ASSIGNMENT"),
			TransactionID:     &transactionID,
			CreatedBy:         &systemUser,
			UpdatedBy:         &systemUser,
		}
		if err := r.createCreditLedgerEntryTx(tx, completedLedger); err != nil {
			return err
		}

		result.WalletTransaction = &models.WalletTransaction{
			ID:              completedLedger.ID,
			Currency:        completedLedger.Currency,
			Amount:          amountFloat,
			Status:          completedLedger.Status,
			Product:         completedLedger.Product,
			OrderID:         completedLedger.OrderID,
			TransactionType: completedLedger.TransactionType,
			TransactionID:   completedLedger.TransactionID,
			CreatedAt:       completedLedger.CreatedAt,
		}

		result.AmountChargedUSD = formatMoney(amountRat)
		result.WalletBalanceAfter = formatMoney(new(big.Rat).Sub(availableBalanceRat, amountRat))

		allocationStatus := models.AllocationStatusUnallocated
		var receiverID *string
		if receiverUserID != "" {
			allocationStatus = models.AllocationStatusAssigned
			receiverID = stringPointer(receiverUserID)
			result.EsimSource = "new_purchase_assigned"
		} else {
			result.EsimSource = "new_purchase_unallocated"
		}

		allocation := &models.B2BAllocation{
			CatalogID:     catalogID,
			OwnerID:       tenantIDString,
			ReceiverID:    receiverID,
			InvoiceID:     invoiceID,
			Tenant:        &tenantIDString,
			RequestID:     requestID,
			TransactionID: &transactionID,
			CreatedBy:     systemUser,
			UpdatedBy:     systemUser,
			Status:        stringPointer(allocationStatus),
			OrderID:       &orderID,
		}
		if err := tx.Create(allocation).Error; err != nil {
			return err
		}

		if receiverUserID != "" {
			selectedEsim, esimSource, err := r.selectAssignableEsimTx(tx, tenantIDString, receiverUserID)
			if err != nil {
				return err
			}
			if err := r.createUserPlanEntryTx(tx, selectedEsim, receiverUserID, catalogID, orderID, transactionID, systemUser); err != nil {
				return err
			}
			result.EsimSource = esimSource
			result.Esim = selectedEsim
			orderRequest.EsimICCID = selectedEsim.ICCID
		}

		orderRequest.WalletBalanceAfter = result.WalletBalanceAfter
		orderRequest.AssignmentStatus = models.OrderStatusCompleted
		orderRequest.EsimSource = result.EsimSource

		finalOrderRequest, err := json.Marshal(orderRequest)
		if err != nil {
			return err
		}
		if err := r.updateOrderStatusTx(tx, orderID, models.OrderStatusCompleted, finalOrderRequest, systemUser); err != nil {
			return err
		}

		result.OrderStatus = models.OrderStatusCompleted
		result.Allocation = allocation

		return nil
	})
	if err != nil {
		failedOrderRequest := orderRequest
		failedOrderRequest.AssignmentStatus = models.OrderStatusFailed
		failedOrderRequest.FailureReason = err.Error()
		if result.WalletBalanceBefore != "" {
			failedOrderRequest.WalletBalanceBefore = result.WalletBalanceBefore
		}
		if result.WalletBalanceAfter != "" {
			failedOrderRequest.WalletBalanceAfter = result.WalletBalanceAfter
		}

		failedPayload, marshalErr := json.Marshal(failedOrderRequest)
		if marshalErr != nil {
			return nil, err
		}
		if statusErr := r.updateOrderStatus(orderID, models.OrderStatusFailed, failedPayload, systemUser); statusErr != nil {
			return nil, fmt.Errorf("pack assignment failed: %w; also failed to mark order failed: %v", err, statusErr)
		}

		if ledgerInitiatedCreated {
			_ = r.createCreditLedgerEntry(models.CreditLedgerTransaction{
				TenantID:          tenantID,
				Currency:          stringPointer(normalizedCurrency),
				TransactionAmount: &amountFloat,
				Status:            stringPointer("FAILED"),
				Product:           stringPointer("PACK"),
				OrderID:           &orderID,
				TransactionType:   stringPointer("PACK_ASSIGNMENT"),
				TransactionID:     &transactionID,
				CreatedBy:         &systemUser,
				UpdatedBy:         &systemUser,
			})
		}
		return nil, err
	}

	return result, nil
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

func lockReceiverAssignmentTx(tx *gorm.DB, receiverUserID string) error {
	return tx.Exec("SELECT pg_advisory_xact_lock(?)", advisoryLockKey(receiverUserID)).Error
}

func advisoryLockKey(receiverUserID string) int64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(receiverUserID))
	return int64(hasher.Sum64())
}

func lockTenantWalletTx(tx *gorm.DB, tenantID int64) (*models.TenantWallet, error) {
	var wallet models.TenantWallet
	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tenant_id = ?", tenantID).
		Order("CASE WHEN UPPER(COALESCE(status, '')) = 'ACTIVE' THEN 0 ELSE 1 END").
		Order("version DESC NULLS LAST").
		Order("updated_at DESC NULLS LAST").
		Order("created_at DESC NULLS LAST").
		Order("id DESC").
		First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *esimRepository) selectAssignableEsimTx(tx *gorm.DB, tenantID string, receiverUserID string) (*models.Esim, string, error) {
	existingEsim, err := findExistingUserEsimTx(tx, tenantID, receiverUserID)
	if err == nil {
		now := time.Now()
		if err := tx.Model(&models.Esim{}).
			Where("id = ?", existingEsim.ID).
			Updates(map[string]interface{}{
				"user_email": receiverUserID,
				"user_id":    receiverUserID,
				"status":     "ASSIGNED",
				"updated_at": now,
				"updated_by": "SYSTEM",
			}).Error; err != nil {
			return nil, "", err
		}
		existingEsim.UserEmail = stringPointer(receiverUserID)
		existingEsim.UserID = stringPointer(receiverUserID)
		existingEsim.Status = "ASSIGNED"
		existingEsim.UpdatedAt = &now
		existingEsim.UpdatedBy = "SYSTEM"
		return existingEsim, models.PackAssignmentEsimSourceExistingUser, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", err
	}

	tenantInventoryEsim, err := findTenantInventoryEsimTx(tx, tenantID)
	if err == nil {
		now := time.Now()
		if err := tx.
			Model(&models.Esim{}).
			Where("id = ?", tenantInventoryEsim.ID).
			Updates(map[string]interface{}{
				"user_email": receiverUserID,
				"user_id":    receiverUserID,
				"status":     "ASSIGNED",
				"updated_at": now,
				"updated_by": "SYSTEM",
			}).Error; err != nil {
			return nil, "", err
		}

		tenantInventoryEsim.UserEmail = stringPointer(receiverUserID)
		tenantInventoryEsim.UserID = stringPointer(receiverUserID)
		tenantInventoryEsim.Status = "ASSIGNED"
		tenantInventoryEsim.UpdatedAt = &now
		tenantInventoryEsim.UpdatedBy = "SYSTEM"
		return tenantInventoryEsim, models.PackAssignmentEsimSourceTenantInventory, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", err
	}

	globalInventory, err := r.allocateReleasedInventoryTx(tx, 1)
	if err != nil {
		return nil, "", err
	}

	selectedEsim := globalInventory[0]
	now := time.Now()
	if err := tx.
		Model(&models.Esim{}).
		Where("id = ?", selectedEsim.ID).
		Updates(map[string]interface{}{
			"tenant_id":  tenantID,
			"user_email": receiverUserID,
			"user_id":    receiverUserID,
			"status":     "ASSIGNED",
			"updated_at": now,
			"updated_by": "SYSTEM",
		}).Error; err != nil {
		return nil, "", err
	}

	selectedEsim.TenantID = stringPointer(tenantID)
	selectedEsim.UserEmail = stringPointer(receiverUserID)
	selectedEsim.UserID = stringPointer(receiverUserID)
	selectedEsim.Status = "ASSIGNED"
	selectedEsim.UpdatedAt = &now
	selectedEsim.UpdatedBy = "SYSTEM"
	return &selectedEsim, models.PackAssignmentEsimSourceAvailableInventory, nil
}

func findExistingUserEsimTx(tx *gorm.DB, tenantID string, receiverUserID string) (*models.Esim, error) {
	var esim models.Esim

	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Model(&models.Esim{}).
		Where("deleted_at IS NULL").
		Where("COALESCE(tenant_id, '') = ?", tenantID).
		Where("(user_id = ? OR user_email = ?)", receiverUserID, receiverUserID).
		Where("UPPER(COALESCE(status, '')) <> ?", "TERMINATED").
		Order("updated_at DESC NULLS LAST").
		Order("created_at DESC NULLS LAST").
		Order("id DESC").
		First(&esim).Error
	if err == nil {
		return &esim, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	err = tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Model(&models.Esim{}).
		Where("deleted_at IS NULL").
		Where("(user_id = ? OR user_email = ?)", receiverUserID, receiverUserID).
		Where("UPPER(COALESCE(status, '')) <> ?", "TERMINATED").
		Order("updated_at DESC NULLS LAST").
		Order("created_at DESC NULLS LAST").
		Order("id DESC").
		First(&esim).Error
	if err != nil {
		return nil, err
	}

	return &esim, nil
}

func findTenantInventoryEsimTx(tx *gorm.DB, tenantID string) (*models.Esim, error) {
	var esim models.Esim
	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Model(&models.Esim{}).
		Where("deleted_at IS NULL").
		Where("tenant_id = ?", tenantID).
		Where("COALESCE(user_id, '') = ''").
		Where("COALESCE(user_email, '') = ''").
		Where("UPPER(COALESCE(status, '')) = ?", "AVAILABLE").
		Where("UPPER(COALESCE(telna_status, '')) = ?", "RELEASED").
		Order("created_at ASC").
		First(&esim).Error
	if err != nil {
		return nil, err
	}
	return &esim, nil
}

func findUnallocatedAllocationTx(tx *gorm.DB, tenantID string, catalogID int64) (*models.B2BAllocation, error) {
	var allocation models.B2BAllocation
	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Model(&models.B2BAllocation{}).
		Where("deleted_at IS NULL").
		Where("COALESCE(tenant, '') = ?", tenantID).
		Where("catalog_id = ?", catalogID).
		Where("COALESCE(receiver_id, '') = ''").
		Where("(COALESCE(status, '') = '' OR UPPER(status) = ?)", models.AllocationStatusUnallocated).
		Order("created_at ASC").
		First(&allocation).Error
	if err != nil {
		return nil, err
	}
	return &allocation, nil
}

func (r *esimRepository) assignUnallocatedPackWithoutPurchase(tenantID int64, receiverUserID string, catalogID int64) (*models.PackAssignmentResult, bool, error) {
	tenantIDString := strconv.FormatInt(tenantID, 10)
	systemUser := "SYSTEM"
	result := &models.PackAssignmentResult{
		TenantID:         tenantID,
		ReceiverUserID:   receiverUserID,
		CatalogID:        catalogID,
		AmountChargedUSD: "0.00",
		OrderStatus:      models.OrderStatusCompleted,
		EsimSource:       "unallocated_inventory",
	}

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := lockReceiverAssignmentTx(tx, receiverUserID); err != nil {
			return err
		}

		unallocated, err := findUnallocatedAllocationTx(tx, tenantIDString, catalogID)
		if err != nil {
			return err
		}

		now := time.Now()
		assignedStatus := models.AllocationStatusAssigned
		if err := tx.Model(&models.B2BAllocation{}).
			Where("id = ?", unallocated.ID).
			Updates(map[string]interface{}{
				"receiver_id": receiverUserID,
				"status":      assignedStatus,
				"updated_at":  now,
				"updated_by":  systemUser,
			}).Error; err != nil {
			return err
		}

		unallocated.ReceiverID = stringPointer(receiverUserID)
		unallocated.Status = stringPointer(assignedStatus)
		result.Allocation = unallocated

		selectedEsim, esimSource, err := r.selectAssignableEsimTx(tx, tenantIDString, receiverUserID)
		if err != nil {
			return err
		}
		transactionID := ""
		if unallocated.TransactionID != nil {
			transactionID = strings.TrimSpace(*unallocated.TransactionID)
		}
		if err := r.createUserPlanEntryTx(tx, selectedEsim, receiverUserID, catalogID, result.OrderID, transactionID, systemUser); err != nil {
			return err
		}
		result.Esim = selectedEsim
		result.EsimSource = esimSource

		if unallocated.OrderID != nil {
			result.OrderID = strings.TrimSpace(*unallocated.OrderID)
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return result, true, nil
}

func parseMoney(value string) (*big.Rat, error) {
	decimal := new(big.Rat)
	if _, ok := decimal.SetString(strings.TrimSpace(value)); !ok {
		return nil, fmt.Errorf("invalid money value: %s", value)
	}
	return decimal, nil
}

func formatMoney(value *big.Rat) string {
	if value == nil {
		return "0.00"
	}
	return value.FloatString(2)
}

func moneyStringFromFloat(value *float64) string {
	if value == nil {
		return "0.00"
	}
	return strconv.FormatFloat(*value, 'f', 2, 64)
}

func normalizeWalletCurrency(currency *string) string {
	if currency == nil || strings.TrimSpace(*currency) == "" {
		return "USD"
	}
	return strings.ToUpper(strings.TrimSpace(*currency))
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func stringPointer(value string) *string {
	value = strings.TrimSpace(value)
	return &value
}

func (r *esimRepository) createOrderRecord(order *models.OrderRecord) error {
	return r.db.Create(order).Error
}

func (r *esimRepository) updateOrderStatus(orderID string, status string, requestObject []byte, updatedBy string) error {
	return r.db.Model(&models.OrderRecord{}).
		Where("order_id = ?", orderID).
		Updates(map[string]interface{}{
			"status":         status,
			"request_object": requestObject,
			"updated_at":     time.Now(),
			"updated_by":     updatedBy,
			"is_active":      true,
		}).Error
}

func (r *esimRepository) updateOrderStatusTx(tx *gorm.DB, orderID string, status string, requestObject []byte, updatedBy string) error {
	return tx.Model(&models.OrderRecord{}).
		Where("order_id = ?", orderID).
		Updates(map[string]interface{}{
			"status":         status,
			"request_object": requestObject,
			"updated_at":     time.Now(),
			"updated_by":     updatedBy,
			"is_active":      true,
		}).Error
}

func (r *esimRepository) createCreditLedgerEntry(ledger models.CreditLedgerTransaction) error {
	return r.db.Create(&ledger).Error
}

func (r *esimRepository) createCreditLedgerEntryTx(tx *gorm.DB, ledger models.CreditLedgerTransaction) error {
	return tx.Create(&ledger).Error
}

func (r *esimRepository) createUserPlanEntryTx(
	tx *gorm.DB,
	esim *models.Esim,
	receiverUserID string,
	catalogID int64,
	orderID string,
	transactionID string,
	systemUser string,
) error {
	receiverUserID = strings.TrimSpace(receiverUserID)
	if receiverUserID == "" || esim == nil || strings.TrimSpace(esim.ICCID) == "" {
		return nil
	}

	vendor := "UNKNOWN"
	if esim.Vendor != nil && strings.TrimSpace(*esim.Vendor) != "" {
		vendor = strings.TrimSpace(*esim.Vendor)
	}

	recordID := time.Now().UnixNano()
	appliedPlanID := strconv.FormatInt(catalogID, 10)
	paymentIntent := strings.TrimSpace(orderID)
	if paymentIntent == "" {
		paymentIntent = strings.TrimSpace(transactionID)
	}
	paymentIntentPtr := stringPointer(paymentIntent)
	if paymentIntent == "" {
		paymentIntentPtr = nil
	}

	userPlan := models.UserPlan{
		ID:            recordID,
		RowID:         recordID,
		CreatedBy:     systemUser,
		UpdatedBy:     systemUser,
		UserID:        receiverUserID,
		ICCID:         strings.TrimSpace(esim.ICCID),
		Vendor:        vendor,
		CatalogID:     catalogID,
		AppliedPlanID: appliedPlanID,
		PlanStatus:    models.UserPlanStatusActive,
		PaymentIntent: paymentIntentPtr,
	}

	return tx.Create(&userPlan).Error
}
