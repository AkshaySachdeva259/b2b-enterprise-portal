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

		return nil
	})
	if err != nil {
		return nil, nil, false, err
	}

	return &selectedEsim, allocation, autoAllocatedEsim, nil
}

func (r *esimRepository) PurchaseAndAssignCatalog(tenantID int64, receiverUserID string, catalogID int64, amountUSD string, currency string, orderID string, invoiceID string, requestID string, transactionID string, orderRequest models.OrderRequestObject) (*models.PackAssignmentResult, error) {
	result := &models.PackAssignmentResult{
		OrderID:          orderID,
		TenantID:         tenantID,
		ReceiverUserID:   receiverUserID,
		CatalogID:        catalogID,
		InvoiceID:        invoiceID,
		RequestID:        requestID,
		TransactionID:    transactionID,
		AmountChargedUSD: amountUSD,
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

	err = r.db.Transaction(func(tx *gorm.DB) error {
		if err := lockReceiverAssignmentTx(tx, receiverUserID); err != nil {
			return err
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

		if availableBalanceRat.Cmp(amountRat) < 0 {
			return &InsufficientWalletBalanceError{
				AvailableBalance: result.WalletBalanceBefore,
				RequiredAmount:   formatMoney(amountRat),
				Currency:         walletCurrency,
			}
		}

		now := time.Now()
		systemUser := "SYSTEM"
		orderRequest.PaymentTransactionID = transactionID
		orderRequest.InvoiceID = invoiceID
		orderRequest.RequestID = requestID
		orderRequest.WalletBalanceBefore = result.WalletBalanceBefore
		orderRequest.AssignmentStatus = models.OrderStatusPending

		initialOrderRequest, err := json.Marshal(orderRequest)
		if err != nil {
			return err
		}

		order := &models.OrderRecord{
			TenantID:      tenantID,
			OrderID:       orderID,
			TotalAmount:   &amountFloat,
			RequestObject: initialOrderRequest,
			Status:        models.OrderStatusPending,
			IsActive:      true,
			CreatedAt:     &now,
			CreatedBy:     &systemUser,
			UpdatedAt:     &now,
			UpdatedBy:     &systemUser,
		}
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		selectedEsim, esimSource, err := r.selectAssignableEsimTx(tx, tenantIDString, receiverUserID)
		if err != nil {
			return err
		}

		transactionType := "PACK_ASSIGNMENT"
		transactionKind := "DEBIT"
		product := "PACK"

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

		walletTransaction := &models.CreditLedgerTransaction{
			TenantID:          tenantID,
			Currency:          stringPointer(walletCurrency),
			TransactionAmount: &amountFloat,
			Type:              &transactionKind,
			Product:           &product,
			OrderID:           &invoiceID,
			TransactionType:   &transactionType,
			TransactionID:     &transactionID,
			CreatedAt:         &now,
			CreatedBy:         &systemUser,
			UpdatedAt:         &now,
			UpdatedBy:         &systemUser,
		}
		if err := tx.Create(walletTransaction).Error; err != nil {
			return err
		}

		receiverID := selectedEsim.ICCID
		allocation := &models.B2BAllocation{
			CatalogID:     catalogID,
			OwnerID:       tenantIDString,
			ReceiverID:    &receiverID,
			InvoiceID:     invoiceID,
			Tenant:        &tenantIDString,
			RequestID:     requestID,
			TransactionID: &transactionID,
			CreatedBy:     systemUser,
			UpdatedBy:     systemUser,
		}
		if err := tx.Create(allocation).Error; err != nil {
			return err
		}

		result.WalletBalanceAfter = formatMoney(new(big.Rat).Sub(availableBalanceRat, amountRat))
		orderRequest.WalletBalanceAfter = result.WalletBalanceAfter
		orderRequest.AssignmentStatus = models.OrderStatusCompleted
		orderRequest.EsimICCID = selectedEsim.ICCID
		orderRequest.EsimSource = esimSource

		finalOrderRequest, err := json.Marshal(orderRequest)
		if err != nil {
			return err
		}
		if err := tx.
			Model(&models.OrderRecord{}).
			Where("id = ?", order.ID).
			Updates(map[string]interface{}{
				"request_object": finalOrderRequest,
				"status":         models.OrderStatusCompleted,
				"updated_at":     now,
				"updated_by":     systemUser,
				"is_active":      true,
			}).Error; err != nil {
			return err
		}

		result.OrderStatus = models.OrderStatusCompleted
		result.EsimSource = esimSource
		result.Esim = selectedEsim
		result.Allocation = allocation
		result.WalletTransaction = &models.WalletTransaction{
			ID:              walletTransaction.ID,
			Currency:        walletTransaction.Currency,
			Amount:          amountFloat,
			Type:            walletTransaction.Type,
			Product:         walletTransaction.Product,
			OrderID:         walletTransaction.OrderID,
			TransactionType: walletTransaction.TransactionType,
			TransactionID:   walletTransaction.TransactionID,
			CreatedAt:       walletTransaction.CreatedAt,
		}

		return nil
	})
	if err != nil {
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
				"user_id":    receiverUserID,
				"updated_at": now,
				"updated_by": "SYSTEM",
			}).Error; err != nil {
			return nil, "", err
		}

		tenantInventoryEsim.UserID = stringPointer(receiverUserID)
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
			"user_id":    receiverUserID,
			"updated_at": now,
			"updated_by": "SYSTEM",
		}).Error; err != nil {
		return nil, "", err
	}

	selectedEsim.TenantID = stringPointer(tenantID)
	selectedEsim.UserID = stringPointer(receiverUserID)
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
		Where("user_id = ?", receiverUserID).
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
		Where("user_id = ?", receiverUserID).
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
