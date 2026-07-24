package webshare

import (
	"context"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// TransactionsService exposes the transaction operations.
type TransactionsService struct {
	client *Client
}

// TransactionStatus is the status of a transaction. There is no failed
// state: failed payments live on PendingPayment.
type TransactionStatus string

// Transaction statuses.
const (
	// TransactionCompleted means the transaction completed.
	TransactionCompleted TransactionStatus = "completed"
	// TransactionRefunded means the transaction was (partially) refunded.
	TransactionRefunded TransactionStatus = "refunded"
)

// Transaction is a completed payment transaction.
type Transaction struct {
	// ID is the unique identifier of the transaction instance.
	ID int `json:"id"`
	// Status is the transaction status. Partial refunds show as refunded.
	Status TransactionStatus `json:"status"`
	// PaymentMethod is the nested payment method used for the transaction.
	PaymentMethod PaymentMethod `json:"payment_method"`
	// Reason describes the transaction.
	Reason string `json:"reason"`
	// Amount is the transaction amount in USD.
	Amount float64 `json:"amount"`
	// CreditsUsed is the credits used in the transaction.
	CreditsUsed float64 `json:"credits_used"`
	// CreditsGained is the credits gained in the transaction (for example
	// when downgrading).
	CreditsGained float64 `json:"credits_gained"`
	// RefundAmount is the amount refunded in USD.
	RefundAmount float64 `json:"refund_amount"`
	// RefundDate is when the last refund was issued. May be nil.
	RefundDate *time.Time `json:"refund_date"`
	// CreatedAt is when this instance was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when this instance was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// TransactionListParams are the parameters for TransactionsService.List.
type TransactionListParams struct {
	// Page is the page number.
	Page *int
	// PageSize is the number of results per page.
	PageSize *int
}

// List retrieves the transactions in paginated format.
func (s *TransactionsService) List(ctx context.Context, params TransactionListParams, opts ...RequestOption) (*Page[Transaction], error) {
	q := url.Values{}
	setInt(q, "page", params.Page)
	setInt(q, "page_size", params.PageSize)
	return getPage[Transaction](ctx, s.client, "/api/v2/payment/transaction/", q, opts)
}

// ListAll returns a lazy iterator over every transaction across all pages.
func (s *TransactionsService) ListAll(ctx context.Context, params TransactionListParams, opts ...RequestOption) iter.Seq2[Transaction, error] {
	return iterPages(ctx, func(ctx context.Context) (*Page[Transaction], error) {
		return s.List(ctx, params, opts...)
	})
}

// Get retrieves a transaction.
func (s *TransactionsService) Get(ctx context.Context, id int, opts ...RequestOption) (*Transaction, error) {
	out := &Transaction{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/payment/transaction/"+strconv.Itoa(id)+"/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}
