package controllers

import (
	"net/http"

	"com.jetapcglobal.b2b.com/services"
)

type CatalogController interface {
	GetByPageName(w http.ResponseWriter, r *http.Request)
}

type catalogController struct {
	svc services.CatalogService
}

func NewCatalogController(svc services.CatalogService) CatalogController {
	return &catalogController{svc: svc}
}

func (c *catalogController) GetByPageName(w http.ResponseWriter, r *http.Request) {
	pageName := r.URL.Query().Get("page_name")
	if pageName == "" {
		writeError(w, http.StatusBadRequest, "page_name query parameter is required")
		return
	}

	catalogs, err := c.svc.GetByPageName(pageName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch catalog")
		return
	}
	writeJSON(w, http.StatusOK, Response{Data: catalogs, Message: "success"})
}
