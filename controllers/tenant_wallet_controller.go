package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"com.jetapcglobal.b2b.com/services"
	"github.com/gorilla/mux"
)

type TenantWalletController interface {
	GetWalletByTenantID(w http.ResponseWriter, r *http.Request)
}

type tenantWalletController struct {
	svc services.TenantWalletService
}

func NewTenantWalletController(svc services.TenantWalletService) TenantWalletController {
	return &tenantWalletController{svc: svc}
}

func (c *tenantWalletController) GetWalletByTenantID(w http.ResponseWriter, r *http.Request) {
	tenantIDParam := mux.Vars(r)["tenant_id"]
	if tenantIDParam == "" {
		writeError(w, http.StatusBadRequest, "tenant_id path parameter is required")
		return
	}

	tenantID, err := strconv.ParseInt(tenantIDParam, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "tenant_id must be a valid int64")
		return
	}

	result, err := c.svc.GetWalletSummaryByTenantID(tenantID)
	if err != nil {
		if errors.Is(err, services.ErrTenantWalletNotFound) {
			writeError(w, http.StatusNotFound, "tenant wallet not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch tenant wallet")
		return
	}

	writeJSON(w, http.StatusOK, Response{
		Data:    result,
		Message: "success",
	})
}
