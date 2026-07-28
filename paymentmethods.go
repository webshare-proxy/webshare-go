package webshare

import (
	"context"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// PaymentMethodsService exposes the payment method operations.
type PaymentMethodsService struct {
	client *Client
}

// PaymentMethodType identifies the kind of a payment method.
type PaymentMethodType string

// Payment method types.
const (
	// PaymentMethodStripeCard is a Stripe card payment method; it carries
	// the card-specific fields.
	PaymentMethodStripeCard PaymentMethodType = "StripeCard"
	// PaymentMethodLinkPayment is a Stripe Link payment method without
	// card-specific fields.
	PaymentMethodLinkPayment PaymentMethodType = "LinkPayment"
)

// PaymentMethod is a payment method associated with the account. Payment
// methods are polymorphic: discriminate on Type. The card fields (Brand,
// Last4, ExpirationYear, ExpirationMonth) are present only for StripeCard.
type PaymentMethod struct {
	// ID is the unique identifier of the payment method instance.
	ID int `json:"id"`
	// Type identifies the payment type, e.g. "StripeCard" or "LinkPayment".
	Type PaymentMethodType `json:"type"`
	// Brand is the card brand (StripeCard only).
	Brand string `json:"brand,omitempty"`
	// Last4 holds the last four digits of the card (StripeCard only).
	Last4 string `json:"last4,omitempty"`
	// Name is the cardholder name. May be nil.
	Name *string `json:"name"`
	// ExpirationYear is the card expiration year (StripeCard only).
	ExpirationYear int `json:"expiration_year,omitempty"`
	// ExpirationMonth is the card expiration month (StripeCard only).
	ExpirationMonth int `json:"expiration_month,omitempty"`
	// CreatedAt is when the payment method was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when the payment method was last updated.
	UpdatedAt time.Time `json:"updated_at"`
	// Line is the billing address line.
	Line string `json:"line,omitempty"`
	// City is the billing address city.
	City string `json:"city,omitempty"`
	// State is the billing address state. May be nil.
	State *string `json:"state"`
	// PostalCode is the billing address postal code.
	PostalCode string `json:"postal_code,omitempty"`
	// Country is the billing address country.
	Country string `json:"country,omitempty"`
}

// PaymentMethodListParams are the parameters for PaymentMethodsService.List.
type PaymentMethodListParams struct {
	// Page is the page number.
	Page *int
	// PageSize is the number of results per page.
	PageSize *int
}

// List retrieves the payment methods in paginated format.
func (s *PaymentMethodsService) List(ctx context.Context, params PaymentMethodListParams, opts ...RequestOption) (*Page[PaymentMethod], error) {
	q := url.Values{}
	setInt(q, "page", params.Page)
	setInt(q, "page_size", params.PageSize)
	return getPage[PaymentMethod](ctx, s.client, "/api/v2/payment/method/", q, opts)
}

// ListAll returns a lazy iterator over every payment method across all
// pages.
func (s *PaymentMethodsService) ListAll(ctx context.Context, params PaymentMethodListParams, opts ...RequestOption) iter.Seq2[PaymentMethod, error] {
	return iterPages(ctx, func(ctx context.Context) (*Page[PaymentMethod], error) {
		return s.List(ctx, params, opts...)
	})
}

// Get retrieves a payment method. The active payment method ID is available
// on the subscription object.
func (s *PaymentMethodsService) Get(ctx context.Context, id int, opts ...RequestOption) (*PaymentMethod, error) {
	out := &PaymentMethod{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/payment/method/"+strconv.Itoa(id)+"/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}
