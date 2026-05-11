package router

import (
	"net/http"

	"com.jetapcglobal.b2b.com/controllers"
	"com.jetapcglobal.b2b.com/middleware"
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
	tenantWalletRepo := repository.NewTenantWalletRepository(db)
	packInventoryRepo := repository.NewPackInventoryRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	cartRepo := repository.NewCartRepository(db)

	destSvc := services.NewDestinationService(destRepo)
	catalogSvc := services.NewCatalogService(catalogRepo, destRepo)
	esimSvc := services.NewEsimService(esimRepo, catalogRepo)
	cartSvc := services.NewCartService(cartRepo, catalogRepo)
	tenantWalletSvc := services.NewTenantWalletService(tenantWalletRepo)
	packInventorySvc := services.NewPackInventoryService(packInventoryRepo)
	orderSvc := services.NewOrderService(orderRepo)
	emailNotifier := services.NewEmailNotifierFromEnv()
	packAssignmentSvc := services.NewPackAssignmentService(esimRepo, catalogRepo, emailNotifier)

	destCtrl := controllers.NewDestinationController(destSvc)
	catalogCtrl := controllers.NewCatalogController(catalogSvc)
	esimCtrl := controllers.NewEsimController(esimSvc)
	tenantCreditLimitSvc := services.NewTenantCreditLimitService(tenantCreditLimitRepo)
	tenantCreditLimitCtrl := controllers.NewTenantCreditLimitController(tenantCreditLimitSvc)
	cartCtrl := controllers.NewCartController(cartSvc)
	tenantWalletCtrl := controllers.NewTenantWalletController(tenantWalletSvc)
	packInventoryCtrl := controllers.NewPackInventoryController(packInventorySvc)
	orderCtrl := controllers.NewOrderController(orderSvc)
	packAssignmentCtrl := controllers.NewPackAssignmentController(packAssignmentSvc)

	// Handle preflight OPTIONS for all routes
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}).Methods(http.MethodOptions)

	// Public routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/pages", destCtrl.GetDestinations).Methods(http.MethodGet)
	api.HandleFunc("/destinations", destCtrl.GetDestinations).Methods(http.MethodGet)
	api.HandleFunc("/catalog", catalogCtrl.ListPacks).Methods(http.MethodGet)
	api.HandleFunc("/catalog/{catalog_id}", catalogCtrl.GetPackDetail).Methods(http.MethodGet)
	api.HandleFunc("/catalog/assign", packAssignmentCtrl.AssignPack).Methods(http.MethodPost)

	api.HandleFunc("/packs/assign", packAssignmentCtrl.AssignPack).Methods(http.MethodPost)
	api.HandleFunc("/packs", catalogCtrl.ListPacks).Methods(http.MethodGet)
	api.HandleFunc("/packs/{catalog_id}", catalogCtrl.GetPackDetail).Methods(http.MethodGet)

	api.HandleFunc("/tenants/{tenant_id}/orders", orderCtrl.GetOrderHistoryByTenantID).Methods(http.MethodGet)
	api.HandleFunc("/tenants/{tenant_id}/orders/recent", orderCtrl.GetRecentOrdersByTenantID).Methods(http.MethodGet)
	api.HandleFunc("/tenants/{tenant_id}/orders/esims", orderCtrl.GetEsimOrderHistoryByTenantID).Methods(http.MethodGet)
	api.HandleFunc("/tenants/{tenant_id}/orders/summary", orderCtrl.GetTodayOrderSummaryByTenantID).Methods(http.MethodGet)
	api.HandleFunc("/tenants/{tenant_id}/packs/inventory", packInventoryCtrl.ListByTenantID).Methods(http.MethodGet)

	api.HandleFunc("/cart/load", cartCtrl.LoadCart).Methods(http.MethodGet)
	api.HandleFunc("/cart/update", cartCtrl.UpdateCart).Methods(http.MethodPut)
	api.HandleFunc("/cart/items", cartCtrl.DeleteCartItems).Methods(http.MethodDelete)

	api.HandleFunc("/esims/inventory", esimCtrl.GetInventoryByTenantID).Methods(http.MethodGet)
	api.HandleFunc("/esims/inventory/{tenant_id}", esimCtrl.GetInventoryByTenantID).Methods(http.MethodGet)
	api.HandleFunc("/esims/order", esimCtrl.OrderEsims).Methods(http.MethodPost)
	api.HandleFunc("/esims/assign-catalog", esimCtrl.AssignCatalog).Methods(http.MethodPost)

	api.HandleFunc("/tenants/{tenant_id}/credit-limit", tenantCreditLimitCtrl.GetCurrentByTenantID).Methods(http.MethodGet)
	api.HandleFunc("/tenants/{tenant_id}/wallet", tenantWalletCtrl.GetWalletByTenantID).Methods(http.MethodGet)

	return middleware.CORS(r)
}
