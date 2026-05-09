package services

import (
	"errors"

	"com.jetapcglobal.b2b.com/models"
	"com.jetapcglobal.b2b.com/repository"
	"gorm.io/gorm"
)

var ErrTenantWalletNotFound = errors.New("tenant wallet not found")

type TenantWalletService interface {
	GetWalletSummaryByTenantID(tenantID int64) (*models.TenantWalletSummary, error)
}

type tenantWalletService struct {
	repo repository.TenantWalletRepository
}

func NewTenantWalletService(repo repository.TenantWalletRepository) TenantWalletService {
	return &tenantWalletService{repo: repo}
}

func (s *tenantWalletService) GetWalletSummaryByTenantID(tenantID int64) (*models.TenantWalletSummary, error) {
	wallet, err := s.repo.GetWalletByTenantID(tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTenantWalletNotFound
		}
		return nil, err
	}

	transactions, err := s.repo.GetRecentTransactionsByTenantID(tenantID, 5)
	if err != nil {
		return nil, err
	}

	result := &models.TenantWalletSummary{
		TenantID:         wallet.TenantID,
		Currency:         wallet.Currency,
		AvailableBalance: floatValue(wallet.AvailableCredit),
		Status:           wallet.Status,
		LastTransactions: make([]models.WalletTransaction, 0, len(transactions)),
	}

	for _, transaction := range transactions {
		result.LastTransactions = append(result.LastTransactions, models.WalletTransaction{
			ID:              transaction.ID,
			Currency:        transaction.Currency,
			Amount:          floatValue(transaction.TransactionAmount),
			Type:            transaction.Type,
			Product:         transaction.Product,
			OrderID:         transaction.OrderID,
			TransactionType: transaction.TransactionType,
			TransactionID:   transaction.TransactionID,
			CreatedAt:       transaction.CreatedAt,
		})
	}

	return result, nil
}

func floatValue(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}
