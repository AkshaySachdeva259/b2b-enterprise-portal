package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"com.jetapcglobal.b2b.com/services"
	"github.com/gorilla/mux"
)

type CatalogController interface {
	ListPacks(w http.ResponseWriter, r *http.Request)
	GetPackDetail(w http.ResponseWriter, r *http.Request)
}

type catalogController struct {
	svc services.CatalogService
}

func NewCatalogController(svc services.CatalogService) CatalogController {
	return &catalogController{svc: svc}
}

func (c *catalogController) ListPacks(w http.ResponseWriter, r *http.Request) {
	limit := 0
	if limitParam := firstNonEmpty(r.URL.Query().Get("limit")); limitParam != "" {
		parsedLimit, err := strconv.Atoi(limitParam)
		if err != nil || parsedLimit <= 0 {
			writeError(w, http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		limit = parsedLimit
	}

	catalogs, err := c.svc.ListPacks(
		firstNonEmpty(r.URL.Query().Get("page_name"), r.URL.Query().Get("destination")),
		limit,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch packs")
		return
	}
	writeJSON(w, http.StatusOK, Response{Data: catalogs, Message: "success"})
}

func (c *catalogController) GetPackDetail(w http.ResponseWriter, r *http.Request) {
	catalogIDParam := mux.Vars(r)["catalog_id"]
	if catalogIDParam == "" {
		writeError(w, http.StatusBadRequest, "catalog_id path parameter is required")
		return
	}

	catalog, err := c.svc.GetPackDetail(catalogIDParam)
	if err != nil {
		if errors.Is(err, services.ErrPackCatalogNotFound) {
			writeError(w, http.StatusNotFound, "catalog not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch catalog details")
		return
	}

	writeJSON(w, http.StatusOK, Response{Data: catalog, Message: "success"})
}
