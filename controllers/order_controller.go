package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"com.jetapcglobal.b2b.com/services"
	"github.com/gorilla/mux"
)

type OrderController interface {
	GetRecentOrdersByTenantID(w http.ResponseWriter, r *http.Request)
	GetOrderHistoryByTenantID(w http.ResponseWriter, r *http.Request)
	GetEsimOrderHistoryByTenantID(w http.ResponseWriter, r *http.Request)
	GetTodayOrderSummaryByTenantID(w http.ResponseWriter, r *http.Request)
}

type orderController struct {
	svc services.OrderService
}

func NewOrderController(svc services.OrderService) OrderController {
	return &orderController{svc: svc}
}

func (c *orderController) GetRecentOrdersByTenantID(w http.ResponseWriter, r *http.Request) {
	tenantID, limit, ok := parseTenantIDAndOrderLimit(w, r)
	if !ok {
		return
	}

	productType, ok := parseOrderProductTypeFilter(w, r)
	if !ok {
		return
	}

	orders, err := c.svc.GetRecentOrdersByTenantID(tenantID, limit, productType)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch recent orders")
		return
	}

	writeJSON(w, http.StatusOK, Response{Data: orders, Message: "success"})
}

func (c *orderController) GetOrderHistoryByTenantID(w http.ResponseWriter, r *http.Request) {
	tenantID, limit, ok := parseTenantIDAndOrderLimit(w, r)
	if !ok {
		return
	}

	productType, ok := parseOrderProductTypeFilter(w, r)
	if !ok {
		return
	}

	orders, err := c.svc.GetOrderHistoryByTenantID(tenantID, limit, productType)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch order history")
		return
	}

	writeJSON(w, http.StatusOK, Response{Data: orders, Message: "success"})
}

func (c *orderController) GetEsimOrderHistoryByTenantID(w http.ResponseWriter, r *http.Request) {
	tenantID, limit, ok := parseTenantIDAndOrderLimit(w, r)
	if !ok {
		return
	}

	orders, err := c.svc.GetEsimOrderHistoryByTenantID(tenantID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch esim order history")
		return
	}

	writeJSON(w, http.StatusOK, Response{Data: orders, Message: "success"})
}

func (c *orderController) GetTodayOrderSummaryByTenantID(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseTenantID(w, r)
	if !ok {
		return
	}

	location, err := resolveOrderSummaryLocation(firstNonEmpty(r.URL.Query().Get("timezone"), r.URL.Query().Get("tz")))
	if err != nil {
		writeError(w, http.StatusBadRequest, "timezone must be a valid IANA timezone")
		return
	}

	summary, err := c.svc.GetTodayOrderSummaryByTenantID(tenantID, location)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch order summary")
		return
	}

	writeJSON(w, http.StatusOK, Response{Data: summary, Message: "success"})
}

func parseTenantIDAndOrderLimit(w http.ResponseWriter, r *http.Request) (int64, int, bool) {
	tenantID, ok := parseTenantID(w, r)
	if !ok {
		return 0, 0, false
	}

	limit := 0
	if limitParam := firstNonEmpty(r.URL.Query().Get("limit")); limitParam != "" {
		parsedLimit, err := strconv.Atoi(limitParam)
		if err != nil || parsedLimit <= 0 {
			writeError(w, http.StatusBadRequest, "limit must be a positive integer")
			return 0, 0, false
		}
		limit = parsedLimit
	}

	return tenantID, limit, true
}

func parseTenantID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	tenantIDParam := mux.Vars(r)["tenant_id"]
	if tenantIDParam == "" {
		writeError(w, http.StatusBadRequest, "tenant_id path parameter is required")
		return 0, false
	}

	tenantID, err := strconv.ParseInt(tenantIDParam, 10, 64)
	if err != nil || tenantID <= 0 {
		writeError(w, http.StatusBadRequest, "tenant_id must be a valid int64")
		return 0, false
	}

	return tenantID, true
}

func resolveOrderSummaryLocation(timezoneName string) (*time.Location, error) {
	if timezoneName == "" {
		return time.Local, nil
	}
	return time.LoadLocation(timezoneName)
}

func parseOrderProductTypeFilter(w http.ResponseWriter, r *http.Request) (string, bool) {
	value := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("product_type")))
	switch value {
	case "", "catalog", "esim":
		return value, true
	default:
		writeError(w, http.StatusBadRequest, "product_type must be one of: catalog, esim")
		return "", false
	}
}
