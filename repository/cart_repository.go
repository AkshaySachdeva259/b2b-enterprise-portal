package repository

import (
	"errors"
	"time"

	"com.jetapcglobal.b2b.com/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrCartNotFound = errors.New("cart not found")

type CartRepository interface {
	GetByUserID(userID string) (*models.Cart, error)
	Create(userID string, currency string) (*models.Cart, error)
	UpdateCurrency(cartID uuid.UUID, currency string) error
	ListItemsByCartID(cartID uuid.UUID) ([]models.CartItem, error)
	DeleteItemsByCatalogIDs(cartID uuid.UUID, catalogIDs []string) error
	ReplaceItems(cartID uuid.UUID, items []models.CartItem) error
}

type cartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) GetByUserID(userID string) (*models.Cart, error) {
	var cart models.Cart

	err := r.db.
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		First(&cart).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCartNotFound
	}
	if err != nil {
		return nil, err
	}

	return &cart, nil
}

func (r *cartRepository) Create(userID string, currency string) (*models.Cart, error) {
	cart := &models.Cart{
		ID:       uuid.New(),
		UserID:   userID,
		Currency: currency,
	}

	if err := r.db.Create(cart).Error; err != nil {
		return nil, err
	}

	return cart, nil
}

func (r *cartRepository) UpdateCurrency(cartID uuid.UUID, currency string) error {
	return r.db.
		Model(&models.Cart{}).
		Where("id = ?", cartID).
		Updates(map[string]interface{}{
			"currency":   currency,
			"updated_at": time.Now(),
		}).Error
}

func (r *cartRepository) ListItemsByCartID(cartID uuid.UUID) ([]models.CartItem, error) {
	items := make([]models.CartItem, 0)

	err := r.db.
		Where("cart_id = ? AND deleted_at IS NULL", cartID).
		Order("created_at ASC").
		Find(&items).Error

	return items, err
}

func (r *cartRepository) DeleteItemsByCatalogIDs(cartID uuid.UUID, catalogIDs []string) error {
	if len(catalogIDs) == 0 {
		return nil
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Model(&models.Cart{}).
			Where("id = ?", cartID).
			Update("updated_at", time.Now()).Error; err != nil {
			return err
		}

		return tx.
			Where("cart_id = ? AND catalog_id IN ?", cartID, catalogIDs).
			Delete(&models.CartItem{}).Error
	})
}

func (r *cartRepository) ReplaceItems(cartID uuid.UUID, items []models.CartItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Model(&models.Cart{}).
			Where("id = ?", cartID).
			Update("updated_at", time.Now()).Error; err != nil {
			return err
		}

		if err := tx.
			Where("cart_id = ?", cartID).
			Delete(&models.CartItem{}).Error; err != nil {
			return err
		}

		if len(items) == 0 {
			return nil
		}

		return tx.Create(&items).Error
	})
}
