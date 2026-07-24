package webshare

import (
	"context"
	"net/http"
	"time"
)

// IDVerificationService exposes the read-only ID verification status.
// Webshare uses Stripe Identity; verifications are completed from the
// dashboard.
type IDVerificationService struct {
	client *Client
}

// IDVerificationState is the state of the ID verification.
type IDVerificationState string

// ID verification states.
const (
	// IDVerificationNotRequired means no ID verification is needed.
	IDVerificationNotRequired IDVerificationState = "not-required"
	// IDVerificationRequested means the account must verify via the
	// dashboard.
	IDVerificationRequested IDVerificationState = "requested"
	// IDVerificationPending means a verification was started and the client
	// secret is available for Stripe JS.
	IDVerificationPending IDVerificationState = "pending"
	// IDVerificationProcessing means Stripe is processing the verification.
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
