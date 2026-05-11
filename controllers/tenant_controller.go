package controllers

import (
	"errors"
	"net/http"
	"strings"

	"com.jetapcglobal.b2b.com/services"
)

type TenantController interface {
	CheckByEmail(w http.ResponseWriter, r *http.Request)
}

type tenantController struct {
	svc services.TenantService
}

func NewTenantController(svc services.TenantService) TenantController {
	return &tenantController{svc: svc}
}

func (c *tenantController) CheckByEmail(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.URL.Query().Get("email"))
	if email == "" {
		writeError(w, http.StatusBadRequest, "email query parameter is required")
		return
	}

	tenant, err := c.svc.FindByEmail(email)
	if err != nil {
		if errors.Is(err, services.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "no tenant registered")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to check tenant")
		return
	}

	writeJSON(w, http.StatusOK, Response{
		Data:    tenant,
		Message: "tenant registered",
	})
}
