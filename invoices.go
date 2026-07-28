package webshare

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// InvoicesService exposes the invoice operations.
type InvoicesService struct {
	client *Client
}

// Download returns the invoice for the given subscription transaction as PDF
// bytes. The ID is the Transaction.ID of a subscription payment.
func (s *InvoicesService) Download(ctx context.Context, subscriptionTransactionID int, opts ...RequestOption) ([]byte, error) {
	q := url.Values{}
	q.Set("subscription_transaction_id", strconv.Itoa(subscriptionTransactionID))
	return s.client.doBytes(ctx, http.MethodGet, "/api/v2/invoices/download", q, opts)
}
