package webshare

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

// InvoicesService exposes the invoice operations.
type InvoicesService struct {
	client *Client
}

// Download returns the invoice for the given subscription transaction as PDF
// bytes.
func (s *InvoicesService) Download(ctx context.Context, subscriptionTransactionID string, opts ...RequestOption) ([]byte, error) {
	if subscriptionTransactionID == "" {
		return nil, errors.New("webshare: invoice download: subscription transaction ID is required")
	}
	q := url.Values{}
	q.Set("subscription_transaction_id", subscriptionTransactionID)
	return s.client.doBytes(ctx, http.MethodGet, "/api/v2/invoices/download", q, opts)
}
