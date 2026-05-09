package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"com.jetapcglobal.b2b.com/services"
	"github.com/gorilla/mux"
)

type TenantCreditLimitController interface {
	GetCurrentByTenantID(w http.ResponseWriter, r *http.Request)
}

type tenantCreditLimitController struct {
	svc services.TenantCreditLimitService
}

func NewTenantCreditLimitController(svc services.TenantCreditLimitService) TenantCreditLimitController {
	return &tenantCreditLimitController{svc: svc}
}

func (c *tenantCreditLimitController) GetCurrentByTenantID(w http.ResponseWriter, r *http.Request) {
	tenantIDParam := mux.Vars(r)["tenant_id"]
	if tenantIDParam == "" {
		writeError(w, http.StatusBadRequest, "tenant_id path parameter is required")
		return
	}

	tenantID, err := strconv.ParseInt(tenantIDParam, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "tenant_id must be a valid int8")
		return
	}

	result, err := c.svc.GetCurrentByTenantID(tenantID)
	if err != nil {
		if errors.Is(err, services.ErrTenantCreditLimitNotFound) {
			writeError(w, http.StatusNotFound, "tenant credit limit not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch tenant credit limit")
		return
	}

	writeJSON(w, http.StatusOK, Response{
		Data:    result,
		Message: "success",
	})
}
