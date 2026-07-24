package webshare

import (
	"context"
	"net/http"
	"time"
)

// IDVerificationService exposes the ID verification operations. Webshare
// uses Stripe Identity: the Stripe JS library must be used with the returned
// client secret to complete the verification.
type IDVerificationService struct {
	client *Client
}

// IDVerificationState is the state of the ID verification.
type IDVerificationState string

// ID verification states.
const (
	// IDVerificationNotRequired means no ID verification is needed.
	IDVerificationNotRequired IDVerificationState = "not-required"
	// IDVerificationRequested means the account must verify; call Start.
	IDVerificationRequested IDVerificationState = "requested"
	// IDVerificationPending means Start succeeded and the client secret is
	// available for Stripe JS.
	IDVerificationPending IDVerificationState = "pending"
	// IDVerificationProcessing means Complete was called and Stripe is
	// processing the verification.
	IDVerificationProcessing IDVerificationState = "processing"
	// IDVerificationFailed means the verification failed.
	IDVerificationFailed IDVerificationState = "failed"
	// IDVerificationVerified means the verification succeeded.
	IDVerificationVerified IDVerificationState = "verified"
)

// IDVerification is the ID verification object of the account.
type IDVerification struct {
	// ID is the unique identifier of this instance.
	ID int `json:"id"`
	// State is the state of the current ID verification.
	State IDVerificationState `json:"state"`
	// ClientSecret is used with the Stripe JS API. Nil unless the
	// verification is pending.
	ClientSecret *string `json:"client_secret"`
	// VerificationFailureTimes counts failed ID verifications.
	VerificationFailureTimes int `json:"verification_failure_times"`
	// MaxVerificationFailureTimes is how many times a verification can fail
	// before new attempts are blocked.
	MaxVerificationFailureTimes int `json:"max_verification_failure_times"`
	// CreatedAt is the original registration date.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when the verification object was last updated.
	UpdatedAt time.Time `json:"updated_at"`
	// VerifiedAt is when the verification completed successfully. May be
	// nil.
	VerifiedAt *time.Time `json:"verified_at"`
}

// Get retrieves the ID verification object.
func (s *IDVerificationService) Get(ctx context.Context, opts ...RequestOption) (*IDVerification, error) {
	out := &IDVerification{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/idverification/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// Start starts an ID verification to receive the client secret. Allowed only
// from the requested state, or from failed while failure attempts remain. On
// success the state becomes pending.
func (s *IDVerificationService) Start(ctx context.Context, opts ...RequestOption) (*IDVerification, error) {
	out := &IDVerification{}
	if err := s.client.doJSON(ctx, http.MethodPost, "/api/v2/idverification/start/", nil, struct{}{}, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// Complete notifies the API to process a verification after finishing it
// with Stripe JS. The state must be pending; on success it becomes
// processing. Returns a 400 error when the Stripe JS verification was not
// completed.
func (s *IDVerificationService) Complete(ctx context.Context, opts ...RequestOption) (*IDVerification, error) {
	out := &IDVerification{}
	if err := s.client.doJSON(ctx, http.MethodPost, "/api/v2/idverification/complete/", nil, struct{}{}, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}
