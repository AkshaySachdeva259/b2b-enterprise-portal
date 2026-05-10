package models

const (
	PackAssignmentEsimSourceExistingUser       = "existing_user"
	PackAssignmentEsimSourceTenantInventory    = "tenant_inventory"
	PackAssignmentEsimSourceAvailableInventory = "available_inventory"
)

type PackAssignmentResult struct {
	OrderID             string             `json:"order_id"`
	OrderStatus         string             `json:"order_status"`
	TenantID            int64              `json:"tenant_id"`
	ReceiverUserID      string             `json:"receiver_user_id"`
	CatalogID           string             `json:"catalog_id"`
	InvoiceID           string             `json:"invoice_id"`
	RequestID           string             `json:"request_id"`
	TransactionID       string             `json:"transaction_id"`
	AmountChargedUSD    string             `json:"amount_charged_usd"`
	WalletCurrency      string             `json:"wallet_currency"`
	WalletBalanceBefore string             `json:"wallet_balance_before"`
	WalletBalanceAfter  string             `json:"wallet_balance_after"`
	EsimSource          string             `json:"esim_source"`
	Esim                *Esim              `json:"esim"`
	Allocation          *B2BAllocation     `json:"allocation"`
	WalletTransaction   *WalletTransaction `json:"wallet_transaction"`
}
