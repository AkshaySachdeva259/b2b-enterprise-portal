package services

import (
	"encoding/json"
	"errors"
	"strings"

	"com.jetapcglobal.b2b.com/models"
	"com.jetapcglobal.b2b.com/repository"
	"gorm.io/gorm"
)

const defaultAllPackLimit = 35
const maxAllPackLimit = 35

var ErrPackCatalogNotFound = errors.New("catalog not found")

type CatalogService interface {
	ListPacks(pageName string, limit int) ([]models.PackSummary, error)
	GetPackDetail(catalogID string) (*models.PackDetail, error)
}

type catalogService struct {
	repo            repository.CatalogRepository
	destinationRepo repository.DestinationRepository
}

type catalogUSDPrice struct {
	Value string `json:"value"`
}

func NewCatalogService(repo repository.CatalogRepository, destinationRepo repository.DestinationRepository) CatalogService {
	return &catalogService{
		repo:            repo,
		destinationRepo: destinationRepo,
	}
}

func (s *catalogService) ListPacks(pageName string, limit int) ([]models.PackSummary, error) {
	normalizedPageName := normalizePackPageName(pageName)
	effectiveLimit := resolvePackLimit(normalizedPageName, limit)

	catalogs, err := s.repo.ListPacks(normalizedPageName, effectiveLimit)
	if err != nil {
		return nil, err
	}

	destinationNameMap, err := s.destinationNameMap()
	if err != nil {
		return nil, err
	}

	results := make([]models.PackSummary, 0, len(catalogs))
	for _, catalog := range catalogs {
		if catalog.CatalogID == nil {
			continue
		}

		results = append(results, models.PackSummary{
			ID:             *catalog.CatalogID,
			CountryName:    resolveCountryName(catalog.PageName, catalog.Name, destinationNameMap),
			DataInGB:       catalog.DataInGB,
			ValidityInDays: catalog.ValidityInDays,
			PriceUSD:       extractCatalogUSDPrice(catalog.Prices),
		})
	}

	return results, nil
}

func (s *catalogService) GetPackDetail(catalogID string) (*models.PackDetail, error) {
	catalog, err := s.repo.GetByCatalogID(catalogID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPackCatalogNotFound
		}
		return nil, err
	}

	destinationNameMap, err := s.destinationNameMap()
	if err != nil {
		return nil, err
	}

	return &models.PackDetail{
		ID:                 catalogID,
		CountryName:        resolveCountryName(catalog.PageName, catalog.Name, destinationNameMap),
		Vendor:             catalog.Vendor,
		PackageType:        catalog.PackageType,
		PackageTypeID:      catalog.PackageTypeID,
		Name:               catalog.Name,
		Description:        append([]string(nil), catalog.Description...),
		Callout:            catalog.Callout,
		DataInGB:           catalog.DataInGB,
		VoiceInMins:        catalog.VoiceInMins,
		SMS:                catalog.SMS,
		ValidityInDays:     catalog.ValidityInDays,
		SupportedCountries: append([]string(nil), catalog.Countries...),
		Visibility:         catalog.Visibility,
		Supports5G:         catalog.Supports5G,
		RedirectURL:        catalog.RedirectURL,
		Tags:               append([]string(nil), catalog.Tags...),
		Prices:             catalog.Prices,
		PriceUSD:           extractCatalogUSDPrice(catalog.Prices),
		EsimConfigID:       catalog.EsimConfigID,
		CatalogType:        catalog.CatalogType,
		CatalogMetadata:    catalog.CatalogMetadata,
		PageName:           catalog.PageName,
		CreatedAt:          catalog.CreatedAt,
		UpdatedAt:          catalog.UpdatedAt,
	}, nil
}

func (s *catalogService) destinationNameMap() (map[string]string, error) {
	destinations, err := s.destinationRepo.ListDestinations()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(destinations))
	for _, destination := range destinations {
		result[destination.Name] = destination.DisplayName
	}

	return result, nil
}

func normalizePackPageName(pageName string) string {
	pageName = strings.TrimSpace(pageName)
	if pageName == "" || strings.EqualFold(pageName, "all") {
		return ""
	}
	return pageName
}

func resolvePackLimit(pageName string, limit int) int {
	if pageName != "" {
		if limit > 0 {
			return limit
		}
		return 0
	}

	if limit <= 0 {
		return defaultAllPackLimit
	}
	if limit > maxAllPackLimit {
		return maxAllPackLimit
	}
	return limit
}

func resolveCountryName(pageName *string, fallback string, destinations map[string]string) string {
	if pageName != nil {
		normalizedPageName := strings.TrimSpace(*pageName)
		if normalizedPageName != "" {
			if displayName, ok := destinations[normalizedPageName]; ok && strings.TrimSpace(displayName) != "" {
				return displayName
			}
			return normalizedPageName
		}
	}

	return strings.TrimSpace(fallback)
}

func extractCatalogUSDPrice(prices json.RawMessage) string {
	if len(prices) == 0 {
		return ""
	}

	priceMap := make(map[string]catalogUSDPrice)
	if err := json.Unmarshal(prices, &priceMap); err != nil {
		return ""
	}

	price := priceMap["usd"]
	return strings.TrimSpace(price.Value)
}
