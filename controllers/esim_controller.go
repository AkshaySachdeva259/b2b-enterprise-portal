package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"com.jetapcglobal.b2b.com/models"
	"com.jetapcglobal.b2b.com/services"
	"github.com/gorilla/mux"
)

type EsimController interface {
	GetInventoryByTenantID(w http.ResponseWriter, r *http.Request)
	OrderEsims(w http.ResponseWriter, r *http.Request)
	AssignCatalog(w http.ResponseWriter, r *http.Request)
}

type esimController struct {
	svc services.EsimService
}

type orderEsimRequest struct {
	TenantID string `json:"tenant_id"`
	Quantity int    `json:"quantity"`
}

type orderEsimResponse struct {
	TenantID string        `json:"tenant_id"`
	Quantity int           `json:"quantity"`
	Esims    []models.Esim `json:"esims"`
}

type assignCatalogRequest struct {
	TenantID         string `json:"tenant_id"`
	ReceiverUserID   string `json:"receiver_user_id"`
	CatalogID        int64  `json:"catalog_id"`
	ICCID            string `json:"iccid,omitempty"`
	AutoAllocateEsim bool   `json:"auto_allocate_esim,omitempty"`
}

type assignCatalogResponse struct {
	TenantID          string                `json:"tenant_id"`
	ReceiverUserID    string                `json:"receiver_user_id"`
	CatalogID         int64                 `json:"catalog_id"`
	ICCID             string                `json:"iccid"`
	AutoAllocatedEsim bool                  `json:"auto_allocated_esim"`
	Esim              *models.Esim          `json:"esim"`
	Allocation        *models.B2BAllocation `json:"allocation"`
}

func NewEsimController(svc services.EsimService) EsimController {
	return &esimController{svc: svc}
}

func (c *esimController) GetInventoryByTenantID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := firstNonEmpty(
		vars["tenant_id"],
		r.URL.Query().Get("tenant_id"),
	)

	esims, err := c.svc.GetInventoryByTenantID(tenantID, r.URL.Query().Get("status"))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrTenantIDRequired), errors.Is(err, services.ErrInvalidEsimInventoryFilter):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to fetch esim inventory")
		}
		return
	}

	writeJSON(w, http.StatusOK, Response{Data: esims, Message: "success"})
}

func (c *esimController) OrderEsims(w http.ResponseWriter, r *http.Request) {
	var req orderEsimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := firstNonEmpty(req.TenantID)
	esims, err := c.svc.OrderEsims(tenantID, req.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrTenantIDRequired), errors.Is(err, services.ErrInvalidEsimQuantity), errors.Is(err, services.ErrTenantIDMustBeInt8):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, services.ErrTenantCreditLimitNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, services.ErrInsufficientCreditLimit):
			writeError(w, http.StatusConflict, err.Error())
		case errors.Is(err, services.ErrInsufficientEsimInventory):
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to order esims")
		}
		return
	}

	writeJSON(w, http.StatusCreated, Response{
		Data: orderEsimResponse{
			TenantID: tenantID,
			Quantity: len(esims),
			Esims:    esims,
		},
		Message: "success",
	})
}

func (c *esimController) AssignCatalog(w http.ResponseWriter, r *http.Request) {
	var req assignCatalogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := firstNonEmpty(req.TenantID)
	receiverUserID := firstNonEmpty(req.ReceiverUserID)
	esim, allocation, autoAllocatedEsim, err := c.svc.AssignCatalog(tenantID, receiverUserID, req.CatalogID, req.ICCID, req.AutoAllocateEsim)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrTenantIDRequired), errors.Is(err, services.ErrReceiverUserIDRequired), errors.Is(err, services.ErrCatalogIDRequired):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, services.ErrCatalogNotFound), errors.Is(err, services.ErrTenantEsimNotFound), errors.Is(err, services.ErrTenantHasNoEsims):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, services.ErrInsufficientEsimInventory):
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to assign catalog")
		}
		return
	}

	writeJSON(w, http.StatusCreated, Response{
		Data: assignCatalogResponse{
			TenantID:          tenantID,
			ReceiverUserID:    receiverUserID,
			CatalogID:         req.CatalogID,
			ICCID:             esim.ICCID,
			AutoAllocatedEsim: autoAllocatedEsim,
			Esim:              esim,
			Allocation:        allocation,
		},
		Message: "success",
	})
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
