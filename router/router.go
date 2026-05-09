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

	destSvc := services.NewDestinationService(destRepo)
	catalogSvc := services.NewCatalogService(catalogRepo)

	destCtrl := controllers.NewDestinationController(destSvc)
	catalogCtrl := controllers.NewCatalogController(catalogSvc)

	// Public routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/pages", destCtrl.GetAllPages).Methods(http.MethodGet)
	api.HandleFunc("/catalog", catalogCtrl.GetByPageName).Methods(http.MethodGet)

	return r
}
