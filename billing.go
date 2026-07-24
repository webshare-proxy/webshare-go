package webshare

import (
	"context"
	"net/http"
	"time"
)

// BillingService exposes the billing information operations.
type BillingService struct {
	client *Client
}

// BillingInfo is the billing information singleton of the account.
type BillingInfo struct {
	// ID is the unique identifier of the billing information instance.
	ID int `json:"id"`
	// Name appears on invoices; it can be a company name.
	Name string `json:"name"`
	// Address appears on invoices; it can be a corporate address.
	Address string `json:"address"`
	// BillingEmail appears on invoices.
	BillingEmail string `json:"billing_email"`
	// CreatedAt is when the account was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when this instance was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// GetInfo returns the billing information associated with the account. There
// is exactly one billing information object per account.
func (s *BillingService) GetInfo(ctx context.Context, opts ...RequestOption) (*BillingInfo, error) {
	out := &BillingInfo{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/subscription/billing_info/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// BillingInfoUpdateParams are the parameters for BillingService.UpdateInfo.
// Only set fields are sent.
type BillingInfoUpdateParams struct {
	// Name sets the invoice name.
	Name *string `json:"name,omitempty"`
	// Address sets the invoice address.
	Address *string `json:"address,omitempty"`
	// BillingEmail sets the invoice email address.
	BillingEmail *string `json:"billing_email,omitempty"`
}

// UpdateInfo partially updates the billing information.
func (s *BillingService) UpdateInfo(ctx context.Context, params BillingInfoUpdateParams, opts ...RequestOption) (*BillingInfo, error) {
	out := &BillingInfo{}
	if err := s.client.doJSON(ctx, http.MethodPatch, "/api/v2/subscription/billing_info/", nil, params, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}
