package services

import (
	"encoding/json"
	"errors"
	"math/big"
	"strings"

	"com.jetapcglobal.b2b.com/models"
	"com.jetapcglobal.b2b.com/repository"
	"github.com/google/uuid"
)

var (
	ErrCartUserIDRequired        = errors.New("user_id is required")
	ErrCartCatalogIDsRequired    = errors.New("catalog_ids is required")
	ErrCartCatalogIDRequired     = errors.New("catalog_id is required")
	ErrCartInvalidItemQuantity   = errors.New("quantity must be greater than zero")
	ErrCartCatalogNotFound       = errors.New("one or more catalog items were not found")
	ErrCartCurrencyNotSupported  = errors.New("currency is not supported for one or more catalog items")
	errCatalogPriceNotFound      = errors.New("catalog price not found")
	errInvalidCatalogPriceAmount = errors.New("invalid catalog price amount")
)

type CartService interface {
	LoadCart(userID string, currency string) (*models.CartDetail, error)
	UpdateCart(userID string, currency string, items []models.CartUpdateItemInput) (*models.CartDetail, error)
	DeleteCartItems(userID string, currency string, catalogIDs []int64) (*models.CartDetail, error)
}

type cartService struct {
	repo        repository.CartRepository
	catalogRepo repository.CatalogRepository
}

type catalogPriceEntry struct {
	Value      string `json:"value"`
	Symbol     string `json:"symbol"`
	ListAmount string `json:"list_amount"`
}

func NewCartService(repo repository.CartRepository, catalogRepo repository.CatalogRepository) CartService {
	return &cartService{repo: repo, catalogRepo: catalogRepo}
}

func (s *cartService) LoadCart(userID string, currency string) (*models.CartDetail, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrCartUserIDRequired
	}

	requestedCurrency := normalizeCartCurrency(currency)
	cart, err := s.getOrCreateCart(userID, requestedCurrency)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.ListItemsByCartID(cart.ID)
	if err != nil {
		return nil, err
	}

	effectiveCurrency := resolveCartCurrency(cart.Currency, requestedCurrency)
	catalogs, err := s.fetchCatalogsForCartItems(items)
	if err != nil {
		return nil, err
	}

	if err := s.ensureCurrencyAvailable(catalogs, effectiveCurrency); err != nil {
		return nil, err
	}

	if !strings.EqualFold(cart.Currency, effectiveCurrency) {
		if err := s.repo.UpdateCurrency(cart.ID, effectiveCurrency); err != nil {
			return nil, err
		}

		cart, err = s.repo.GetByUserID(userID)
		if err != nil {
			return nil, err
		}
	}

	return s.buildCartDetailFromCatalogs(cart, items, catalogs)
}

func (s *cartService) UpdateCart(userID string, currency string, items []models.CartUpdateItemInput) (*models.CartDetail, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrCartUserIDRequired
	}

	normalizedItems, err := normalizeCartItems(items)
	if err != nil {
		return nil, err
	}

	requestedCurrency := normalizeCartCurrency(currency)
	cart, err := s.getOrCreateCart(userID, "")
	if err != nil {
		return nil, err
	}

	effectiveCurrency := resolveCartCurrency(cart.Currency, requestedCurrency)
	catalogs, err := s.fetchCatalogsByIDs(cartItemCatalogIDs(normalizedItems))
	if err != nil {
		return nil, err
	}

	if err := s.ensureCurrencyAvailable(catalogs, effectiveCurrency); err != nil {
		return nil, err
	}

	if !strings.EqualFold(cart.Currency, effectiveCurrency) {
		if err := s.repo.UpdateCurrency(cart.ID, effectiveCurrency); err != nil {
			return nil, err
		}
	}

	repoItems := make([]models.CartItem, 0, len(normalizedItems))
	for _, item := range normalizedItems {
		repoItems = append(repoItems, models.CartItem{
			ID:        uuid.New(),
			CartID:    cart.ID,
			CatalogID: item.CatalogID,
			Quantity:  item.Quantity,
		})
	}

	if err := s.repo.ReplaceItems(cart.ID, repoItems); err != nil {
		return nil, err
	}

	cart, err = s.repo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	storedItems, err := s.repo.ListItemsByCartID(cart.ID)
	if err != nil {
		return nil, err
	}

	return s.buildCartDetailFromCatalogs(cart, storedItems, catalogs)
}

func (s *cartService) DeleteCartItems(userID string, currency string, catalogIDs []int64) (*models.CartDetail, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrCartUserIDRequired
	}

	normalizedCatalogIDs, err := normalizeCartCatalogIDs(catalogIDs)
	if err != nil {
		return nil, err
	}

	requestedCurrency := normalizeCartCurrency(currency)
	cart, err := s.getOrCreateCart(userID, requestedCurrency)
	if err != nil {
		return nil, err
	}

	if err := s.repo.DeleteItemsByCatalogIDs(cart.ID, normalizedCatalogIDs); err != nil {
		return nil, err
	}

	return s.LoadCart(userID, requestedCurrency)
}

func (s *cartService) getOrCreateCart(userID string, currency string) (*models.Cart, error) {
	cart, err := s.repo.GetByUserID(userID)
	if err == nil {
		if strings.TrimSpace(cart.Currency) == "" {
			defaultCurrency := resolveCartCurrency("", currency)
			if err := s.repo.UpdateCurrency(cart.ID, defaultCurrency); err != nil {
				return nil, err
			}

			return s.repo.GetByUserID(userID)
		}

		return cart, nil
	}

	if !errors.Is(err, repository.ErrCartNotFound) {
		return nil, err
	}

	return s.repo.Create(userID, resolveCartCurrency("", currency))
}

func (s *cartService) fetchCatalogsForCartItems(items []models.CartItem) (map[int64]models.Catalog, error) {
	if len(items) == 0 {
		return map[int64]models.Catalog{}, nil
	}

	catalogIDs := make([]int64, 0, len(items))
	for _, item := range items {
		catalogIDs = append(catalogIDs, item.CatalogID)
	}

	return s.fetchCatalogsByIDs(catalogIDs)
}

func (s *cartService) fetchCatalogsByIDs(catalogIDs []int64) (map[int64]models.Catalog, error) {
	if len(catalogIDs) == 0 {
		return map[int64]models.Catalog{}, nil
	}

	uniqueIDs := uniqueCatalogIDs(catalogIDs)
	catalogs, err := s.catalogRepo.GetByCatalogIDs(uniqueIDs)
	if err != nil {
		return nil, err
	}

	catalogMap := make(map[int64]models.Catalog, len(catalogs))
	for _, catalog := range catalogs {
		if catalog.CatalogID == nil {
			continue
		}
		catalogMap[*catalog.CatalogID] = catalog
	}

	if len(catalogMap) != len(uniqueIDs) {
		return nil, ErrCartCatalogNotFound
	}

	return catalogMap, nil
}

func (s *cartService) ensureCurrencyAvailable(catalogs map[int64]models.Catalog, currency string) error {
	for _, catalog := range catalogs {
		if _, err := extractCatalogPrice(catalog, currency); err != nil {
			if errors.Is(err, errCatalogPriceNotFound) || errors.Is(err, errInvalidCatalogPriceAmount) {
				return ErrCartCurrencyNotSupported
			}
			return err
		}
	}

	return nil
}

func (s *cartService) buildCartDetailFromCatalogs(cart *models.Cart, items []models.CartItem, catalogs map[int64]models.Catalog) (*models.CartDetail, error) {
	if catalogs == nil {
		var err error
		catalogs, err = s.fetchCatalogsForCartItems(items)
		if err != nil {
			return nil, err
		}
	}

	detail := &models.CartDetail{
		ID:       cart.ID,
		UserID:   cart.UserID,
		Currency: cart.Currency,
		Items:    make([]models.CartLineItem, 0, len(items)),
		Summary: models.CartSummary{
			Currency:  cart.Currency,
			Subtotal:  "0.00",
			ListTotal: "0.00",
			Savings:   "0.00",
		},
		CreatedAt: cart.CreatedAt,
		UpdatedAt: cart.UpdatedAt,
	}

	if len(items) == 0 {
		return detail, nil
	}

	subtotal := new(big.Rat)
	listTotal := new(big.Rat)

	for _, item := range items {
		catalog, ok := catalogs[item.CatalogID]
		if !ok {
			return nil, ErrCartCatalogNotFound
		}

		price, err := extractCatalogPrice(catalog, cart.Currency)
		if err != nil {
			if errors.Is(err, errCatalogPriceNotFound) || errors.Is(err, errInvalidCatalogPriceAmount) {
				return nil, ErrCartCurrencyNotSupported
			}
			return nil, err
		}

		catalog, err = withSelectedCatalogPrice(catalog, cart.Currency, price)
		if err != nil {
			return nil, err
		}

		unitAmount, err := parseDecimal(price.Value)
		if err != nil {
			return nil, ErrCartCurrencyNotSupported
		}

		unitListAmount, err := parseDecimal(price.ListAmount)
		if err != nil {
			return nil, ErrCartCurrencyNotSupported
		}

		lineAmount := multiplyDecimal(unitAmount, item.Quantity)
		lineListAmount := multiplyDecimal(unitListAmount, item.Quantity)
		lineSavings := subtractDecimal(lineListAmount, lineAmount)

		subtotal.Add(subtotal, lineAmount)
		listTotal.Add(listTotal, lineListAmount)

		detail.Items = append(detail.Items, models.CartLineItem{
			ID:        item.ID,
			CatalogID: item.CatalogID,
			Quantity:  item.Quantity,
			Catalog:   catalog,
			Pricing: models.CartPriceBreakdown{
				Currency:       cart.Currency,
				Symbol:         price.Symbol,
				UnitAmount:     formatDecimal(unitAmount),
				UnitListAmount: formatDecimal(unitListAmount),
				LineAmount:     formatDecimal(lineAmount),
				LineListAmount: formatDecimal(lineListAmount),
				LineSavings:    formatDecimal(lineSavings),
			},
		})
		detail.Summary.QuantityTotal += item.Quantity
		if detail.Summary.Symbol == "" {
			detail.Summary.Symbol = price.Symbol
		}
	}

	detail.Summary.ItemCount = len(detail.Items)
	detail.Summary.Subtotal = formatDecimal(subtotal)
	detail.Summary.ListTotal = formatDecimal(listTotal)
	detail.Summary.Savings = formatDecimal(subtractDecimal(listTotal, subtotal))

	return detail, nil
}

func normalizeCartItems(items []models.CartUpdateItemInput) ([]models.CartUpdateItemInput, error) {
	if len(items) == 0 {
		return []models.CartUpdateItemInput{}, nil
	}

	quantitiesByCatalogID := make(map[int64]int, len(items))
	orderedCatalogIDs := make([]int64, 0, len(items))

	for _, item := range items {
		if item.CatalogID <= 0 {
			return nil, ErrCartCatalogIDRequired
		}
		if item.Quantity <= 0 {
			return nil, ErrCartInvalidItemQuantity
		}

		if _, exists := quantitiesByCatalogID[item.CatalogID]; !exists {
			orderedCatalogIDs = append(orderedCatalogIDs, item.CatalogID)
		}
		quantitiesByCatalogID[item.CatalogID] += item.Quantity
	}

	normalized := make([]models.CartUpdateItemInput, 0, len(orderedCatalogIDs))
	for _, catalogID := range orderedCatalogIDs {
		normalized = append(normalized, models.CartUpdateItemInput{
			CatalogID: catalogID,
			Quantity:  quantitiesByCatalogID[catalogID],
		})
	}

	return normalized, nil
}

func normalizeCartCatalogIDs(catalogIDs []int64) ([]int64, error) {
	if len(catalogIDs) == 0 {
		return nil, ErrCartCatalogIDsRequired
	}

	uniqueCatalogIDs := make([]int64, 0, len(catalogIDs))
	seen := make(map[int64]struct{}, len(catalogIDs))

	for _, catalogID := range catalogIDs {
		if catalogID <= 0 {
			return nil, ErrCartCatalogIDRequired
		}
		if _, exists := seen[catalogID]; exists {
			continue
		}

		seen[catalogID] = struct{}{}
		uniqueCatalogIDs = append(uniqueCatalogIDs, catalogID)
	}

	return uniqueCatalogIDs, nil
}

func cartItemCatalogIDs(items []models.CartUpdateItemInput) []int64 {
	catalogIDs := make([]int64, 0, len(items))
	for _, item := range items {
		catalogIDs = append(catalogIDs, item.CatalogID)
	}
	return catalogIDs
}

func uniqueCatalogIDs(catalogIDs []int64) []int64 {
	seen := make(map[int64]struct{}, len(catalogIDs))
	unique := make([]int64, 0, len(catalogIDs))

	for _, catalogID := range catalogIDs {
		if _, exists := seen[catalogID]; exists {
			continue
		}
		seen[catalogID] = struct{}{}
		unique = append(unique, catalogID)
	}

	return unique
}

func extractCatalogPrice(catalog models.Catalog, currency string) (catalogPriceEntry, error) {
	priceMap := make(map[string]catalogPriceEntry)
	if err := json.Unmarshal(catalog.Prices, &priceMap); err != nil {
		return catalogPriceEntry{}, err
	}

	price, ok := priceMap[currency]
	if !ok {
		return catalogPriceEntry{}, errCatalogPriceNotFound
	}

	price.Value = strings.TrimSpace(price.Value)
	price.ListAmount = strings.TrimSpace(price.ListAmount)
	price.Symbol = strings.TrimSpace(price.Symbol)

	if price.Value == "" {
		return catalogPriceEntry{}, errCatalogPriceNotFound
	}
	if price.ListAmount == "" {
		price.ListAmount = price.Value
	}

	if _, err := parseDecimal(price.Value); err != nil {
		return catalogPriceEntry{}, errInvalidCatalogPriceAmount
	}
	if _, err := parseDecimal(price.ListAmount); err != nil {
		return catalogPriceEntry{}, errInvalidCatalogPriceAmount
	}

	return price, nil
}

func parseDecimal(value string) (*big.Rat, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "0"
	}

	decimal := new(big.Rat)
	if _, ok := decimal.SetString(value); !ok {
		return nil, errInvalidCatalogPriceAmount
	}

	return decimal, nil
}

func multiplyDecimal(value *big.Rat, quantity int) *big.Rat {
	result := new(big.Rat).Set(value)
	return result.Mul(result, big.NewRat(int64(quantity), 1))
}

func subtractDecimal(left *big.Rat, right *big.Rat) *big.Rat {
	result := new(big.Rat).Set(left)
	return result.Sub(result, right)
}

func formatDecimal(value *big.Rat) string {
	if value == nil {
		return "0.00"
	}
	return value.FloatString(2)
}

func normalizeCartCurrency(currency string) string {
	return strings.ToLower(strings.TrimSpace(currency))
}

func resolveCartCurrency(current string, requested string) string {
	requested = normalizeCartCurrency(requested)
	if requested != "" {
		return requested
	}

	return models.DefaultCartCurrency
}

func withSelectedCatalogPrice(catalog models.Catalog, currency string, price catalogPriceEntry) (models.Catalog, error) {
	selectedPrice, err := json.Marshal(map[string]catalogPriceEntry{
		currency: price,
	})
	if err != nil {
		return models.Catalog{}, err
	}

	catalog.Prices = selectedPrice
	return catalog, nil
}
