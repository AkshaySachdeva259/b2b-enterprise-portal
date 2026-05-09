package router

import (
	"net/http"

	"com.jetapcglobal.b2b.com/controllers"
	"com.jetapcglobal.b2b.com/repository"
	"com.jetapcglobal.b2b.com/services"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func New(db *gorm.DB) http.Handler {
	r := mux.NewRouter()

	destRepo := repository.NewDestinationRepository(db)
	catalogRepo := repository.NewCatalogRepository(db)
	esimRepo := repository.NewEsimRepository(db)
	tenantCreditLimitRepo := repository.NewTenantCreditLimitRepository(db)

	destSvc := services.NewDestinationService(destRepo)
	catalogSvc := services.NewCatalogService(catalogRepo)
	esimSvc := services.NewEsimService(esimRepo, catalogRepo, tenantCreditLimitRepo)

	destCtrl := controllers.NewDestinationController(destSvc)
	catalogCtrl := controllers.NewCatalogController(catalogSvc)
	esimCtrl := controllers.NewEsimController(esimSvc)
	tenantCreditLimitSvc := services.NewTenantCreditLimitService(tenantCreditLimitRepo)
	tenantCreditLimitCtrl := controllers.NewTenantCreditLimitController(tenantCreditLimitSvc)

	// Public routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/pages", destCtrl.GetAllPages).Methods(http.MethodGet)
	api.HandleFunc("/catalog", catalogCtrl.GetByPageName).Methods(http.MethodGet)
	api.HandleFunc("/esims/inventory", esimCtrl.GetInventoryByTenantID).Methods(http.MethodGet)
	api.HandleFunc("/esims/inventory/{tenant_id}", esimCtrl.GetInventoryByTenantID).Methods(http.MethodGet)
	api.HandleFunc("/esims/order", esimCtrl.OrderEsims).Methods(http.MethodPost)
	api.HandleFunc("/esims/assign-catalog", esimCtrl.AssignCatalog).Methods(http.MethodPost)
	api.HandleFunc("/tenants/{tenant_id}/credit-limit", tenantCreditLimitCtrl.GetCurrentByTenantID).Methods(http.MethodGet)

	return r
}
