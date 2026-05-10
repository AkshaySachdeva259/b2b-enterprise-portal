package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"com.jetapcglobal.b2b.com/models"
	"com.jetapcglobal.b2b.com/services"
)

type CartController interface {
	LoadCart(w http.ResponseWriter, r *http.Request)
	UpdateCart(w http.ResponseWriter, r *http.Request)
	DeleteCartItems(w http.ResponseWriter, r *http.Request)
}

type cartController struct {
	svc services.CartService
}

type updateCartRequest struct {
	UserID   string                       `json:"user_id"`
	Currency string                       `json:"currency,omitempty"`
	Items    []models.CartUpdateItemInput `json:"items"`
}

type deleteCartItemsRequest struct {
	UserID     string  `json:"user_id"`
	Currency   string  `json:"currency,omitempty"`
	CatalogIDs []string `json:"catalog_ids"`
}

func NewCartController(svc services.CartService) CartController {
	return &cartController{svc: svc}
}

func (c *cartController) LoadCart(w http.ResponseWriter, r *http.Request) {
	cart, err := c.svc.LoadCart(
		firstNonEmpty(r.URL.Query().Get("user_id")),
		r.URL.Query().Get("currency"),
	)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCartUserIDRequired), errors.Is(err, services.ErrCartCurrencyNotSupported):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, services.ErrCartCatalogNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to load cart")
		}
		return
	}

	writeJSON(w, http.StatusOK, Response{Data: cart, Message: "success"})
}

func (c *cartController) UpdateCart(w http.ResponseWriter, r *http.Request) {
	var req updateCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cart, err := c.svc.UpdateCart(firstNonEmpty(req.UserID), req.Currency, req.Items)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCartUserIDRequired),
			errors.Is(err, services.ErrCartCatalogIDRequired),
			errors.Is(err, services.ErrCartInvalidItemQuantity),
			errors.Is(err, services.ErrCartCurrencyNotSupported):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, services.ErrCartCatalogNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to update cart")
		}
		return
	}

	writeJSON(w, http.StatusOK, Response{Data: cart, Message: "success"})
}

func (c *cartController) DeleteCartItems(w http.ResponseWriter, r *http.Request) {
	var req deleteCartItemsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cart, err := c.svc.DeleteCartItems(firstNonEmpty(req.UserID), req.Currency, req.CatalogIDs)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCartUserIDRequired),
			errors.Is(err, services.ErrCartCatalogIDsRequired),
			errors.Is(err, services.ErrCartCatalogIDRequired),
			errors.Is(err, services.ErrCartCurrencyNotSupported):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, services.ErrCartCatalogNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to delete cart items")
		}
		return
	}

	writeJSON(w, http.StatusOK, Response{Data: cart, Message: "success"})
}
