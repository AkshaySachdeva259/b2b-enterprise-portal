package controllers

import (
	"net/http"
	"strconv"

	"com.jetapcglobal.b2b.com/services"
	"github.com/gorilla/mux"
)

type PackInventoryController interface {
	ListByTenantID(w http.ResponseWriter, r *http.Request)
}

type packInventoryController struct {
	svc services.PackInventoryService
}

func NewPackInventoryController(svc services.PackInventoryService) PackInventoryController {
	return &packInventoryController{svc: svc}
}

func (c *packInventoryController) ListByTenantID(w http.ResponseWriter, r *http.Request) {
	tenantIDParam := mux.Vars(r)["tenant_id"]
	if tenantIDParam == "" {
		writeError(w, http.StatusBadRequest, "tenant_id path parameter is required")
		return
	}

	tenantID, err := strconv.ParseInt(tenantIDParam, 10, 64)
	if err != nil || tenantID <= 0 {
		writeError(w, http.StatusBadRequest, "tenant_id must be a valid int64")
		return
	}

	inventory, err := c.svc.ListByTenantID(tenantID, r.URL.Query().Get("status"))
	if err != nil {
		if err == services.ErrInvalidPackInventoryStatusFilter {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch pack inventory")
		return
	}

	writeJSON(w, http.StatusOK, Response{Data: inventory, Message: "success"})
}
