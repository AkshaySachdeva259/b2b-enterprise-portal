package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"com.jetapcglobal.b2b.com/services"
)

type PackAssignmentController interface {
	AssignPack(w http.ResponseWriter, r *http.Request)
}

type packAssignmentController struct {
	svc services.PackAssignmentService
}

type assignPackRequest struct {
	TenantID       string `json:"tenant_id"`
	ReceiverUserID string `json:"receiver_user_id"`
	ReceiverID     string `json:"receiver_id"`
	ReceiverEmail  string `json:"receiver_email"`
	CatalogID      int64  `json:"catalog_id"`
}

func NewPackAssignmentController(svc services.PackAssignmentService) PackAssignmentController {
	return &packAssignmentController{svc: svc}
}

func (c *packAssignmentController) AssignPack(w http.ResponseWriter, r *http.Request) {
	var req assignPackRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeDetailedError(w, http.StatusBadRequest, "invalid_request_body", "invalid request body", nil)
		return
	}

	receiverUserID := firstNonEmpty(req.ReceiverUserID, req.ReceiverID, req.ReceiverEmail)
	result, err := c.svc.AssignPack(req.TenantID, receiverUserID, req.CatalogID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrTenantIDRequired):
			writeDetailedError(w, http.StatusBadRequest, "tenant_id_required", err.Error(), nil)
		case errors.Is(err, services.ErrTenantIDMustBeInt64):
			writeDetailedError(w, http.StatusBadRequest, "invalid_tenant_id", err.Error(), nil)
		case errors.Is(err, services.ErrReceiverUserIDRequired):
			writeDetailedError(w, http.StatusBadRequest, "receiver_user_id_required", err.Error(), nil)
		case errors.Is(err, services.ErrCatalogIDRequired):
			writeDetailedError(w, http.StatusBadRequest, "catalog_id_required", err.Error(), nil)
		case errors.Is(err, services.ErrCatalogNotFound):
			writeDetailedError(w, http.StatusNotFound, "catalog_not_found", err.Error(), nil)
		case errors.Is(err, services.ErrCatalogUSDPriceUnavailable):
			writeDetailedError(w, http.StatusConflict, "catalog_price_unavailable", err.Error(), nil)
		case errors.Is(err, services.ErrTenantWalletNotFound):
			writeDetailedError(w, http.StatusNotFound, "tenant_wallet_not_found", err.Error(), nil)
		case errors.Is(err, services.ErrTenantWalletInactive):
			writeDetailedError(w, http.StatusConflict, "tenant_wallet_inactive", err.Error(), nil)
		case errors.Is(err, services.ErrUnsupportedTenantWalletCurrency):
			writeDetailedError(w, http.StatusConflict, "unsupported_wallet_currency", err.Error(), map[string]string{
				"expected_currency": "USD",
			})
		case errors.Is(err, services.ErrInsufficientEsimInventory):
			writeDetailedError(w, http.StatusConflict, "insufficient_esim_inventory", err.Error(), nil)
		default:
			var balanceErr *services.InsufficientWalletBalanceError
			if errors.As(err, &balanceErr) {
				writeDetailedError(w, http.StatusConflict, "insufficient_wallet_balance", balanceErr.Error(), map[string]string{
					"available_balance": balanceErr.AvailableBalance,
					"required_amount":   balanceErr.RequiredAmount,
					"currency":          balanceErr.Currency,
				})
				return
			}

			writeDetailedError(w, http.StatusInternalServerError, "pack_assignment_failed", "failed to assign pack", nil)
		}
		return
	}

	writeJSON(w, http.StatusCreated, Response{
		Data:    result,
		Message: "success",
	})
}
