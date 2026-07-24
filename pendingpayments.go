package webshare

import (
	"context"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// PendingPaymentsService exposes the pending payment operations. A pending
// payment is a payment that has been initiated but not yet completed; poll
// Get after confirming a Stripe payment to learn the outcome.
type PendingPaymentsService struct {
	client *Client
}

// PendingPaymentStatus is the state of a pending payment.
type PendingPaymentStatus string

// Pending payment statuses.
const (
	// PendingPaymentPending means the payment was initiated.
	PendingPaymentPending PendingPaymentStatus = "pending"
	// PendingPaymentProcessing means the payment is being processed.
	PendingPaymentProcessing PendingPaymentStatus = "processing"
	// PendingPaymentSuccessful means the payment completed; Transaction is
	// set.
	PendingPaymentSuccessful PendingPaymentStatus = "successful"
	// PendingPaymentFailed means the payment failed; FailureReason is set.
	PendingPaymentFailed PendingPaymentStatus = "failed"
)

// SubscriptionTerm is the renewal term of a subscription or payment.
type SubscriptionTerm string

// Subscription terms.
const (
	// TermMonthly pays for one 30-day renewal at a time.
	TermMonthly SubscriptionTerm = "monthly"
	// TermYearly pays for twelve 30-day renewals at once.
	TermYearly SubscriptionTerm = "yearly"
)

// PendingPayment links a payment attempt to a plan and payment method.
type PendingPayment struct {
	// ID is the unique identifier of the pending payment object.
	ID int `json:"id"`
	// Status is the current state of the pending payment.
	Status PendingPaymentStatus `json:"status"`
	// FailureReason is a user-friendly message set when Status is failed.
	// May be nil.
	FailureReason *string `json:"failure_reason"`
	// PaymentMethod is the ID of the payment method used.
	PaymentMethod int `json:"payment_method"`
	// Plan is the ID of the plan being paid for.
	Plan int `json:"plan"`
	// Transaction is the resulting transaction ID; only set when Status is
	// successful. May be nil.
	Transaction *int `json:"transaction"`
	// IsRenewal reports whether the payment renews the subscription rather
	// than immediately changing it.
	IsRenewal bool `json:"is_renewal"`
	// Term is the term of the payment.
	Term SubscriptionTerm `json:"term"`
	// CreatedAt is when this instance was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when this instance was last updated.
	UpdatedAt time.Time `json:"updated_at"`
	// CompletedAt is when the payment completed. May be nil.
	CompletedAt *time.Time `json:"completed_at"`
}

// PendingPaymentListParams are the parameters for
// PendingPaymentsService.List.
type PendingPaymentListParams struct {
	// Page is the page number.
	Page *int
	// PageSize is the number of results per page.
	PageSize *int
}

// List retrieves the pending payments in paginated format.
func (s *PendingPaymentsService) List(ctx context.Context, params PendingPaymentListParams, opts ...RequestOption) (*Page[PendingPayment], error) {
	q := url.Values{}
	setInt(q, "page", params.Page)
	setInt(q, "page_size", params.PageSize)
	return getPage[PendingPayment](ctx, s.client, "/api/v2/payment/pending/", q, opts)
}

// ListAll returns a lazy iterator over every pending payment across all
// pages.
func (s *PendingPaymentsService) ListAll(ctx context.Context, params PendingPaymentListParams, opts ...RequestOption) iter.Seq2[PendingPayment, error] {
	return iterPages(ctx, func(ctx context.Context) (*Page[PendingPayment], error) {
		return s.List(ctx, params, opts...)
	})
}

// Get retrieves a pending payment.
func (s *PendingPaymentsService) Get(ctx context.Context, id int, opts ...RequestOption) (*PendingPayment, error) {
	out := &PendingPayment{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/payment/pending/"+strconv.Itoa(id)+"/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}
