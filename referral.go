package webshare

import (
	"context"
	"encoding/json"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// ReferralService exposes the referral, affiliate and coupon code
// operations.
type ReferralService struct {
	client *Client
}

// ReferralMode is whether referral earnings convert to account credits or
// PayPal payouts.
type ReferralMode string

// Referral modes.
const (
	// ReferralModePayout sends earn-outs to the PayPal account.
	ReferralModePayout ReferralMode = "payout"
	// ReferralModeCredits converts earn-outs to account credits.
	ReferralModeCredits ReferralMode = "credits"
)

// ReferralConfig is the referral configuration of the account.
type ReferralConfig struct {
	// ID is the unique identifier of the referral config instance.
	ID int `json:"id"`
	// Mode selects payout or credits mode.
	Mode ReferralMode `json:"mode"`
	// PayPalPayoutEmail is the PayPal address receiving payouts. May be nil
	// in credits mode; must be set in payout mode.
	PayPalPayoutEmail *string `json:"paypal_payout_email"`
	// IDVerificationRequired reports whether ID verification must be
	// completed before a payout.
	IDVerificationRequired bool `json:"id_verification_required"`
	// CreditsEarned is the USD pending conversion to account credits at the
	// next earn-out.
	CreditsEarned float64 `json:"credits_earned"`
	// PayoutsEarned is the USD pending payout to PayPal at the next
	// earn-out.
	PayoutsEarned float64 `json:"payouts_earned"`
	// TotalCreditsEarned is the total credits earned so far.
	TotalCreditsEarned float64 `json:"total_credits_earned"`
	// TotalPayoutsEarned is the total payouts earned so far.
	TotalPayoutsEarned float64 `json:"total_payouts_earned"`
	// NumberOfUsersReferred is the number of unique users referred.
	NumberOfUsersReferred int `json:"number_of_users_referred"`
	// NumberOfUsersUpgraded is the number of unique referred users who
	// upgraded.
	NumberOfUsersUpgraded int `json:"number_of_users_upgraded"`
	// EarnOutFrequency is the earn-out frequency in Django duration format
	// ("[DD] [HH:MM:SS]").
	EarnOutFrequency string `json:"earn_out_frequency"`
	// NextEarnOutDate is when pending earnings are next converted or paid.
	NextEarnOutDate time.Time `json:"next_earn_out_date"`
	// MinimumEarnOutAmount is the minimum USD required for an earn-out.
	MinimumEarnOutAmount float64 `json:"minimum_earn_out_amount"`
	// ReferralCode is the user's unique referral code.
	ReferralCode string `json:"referral_code"`
	// ReferralURL is an example referral URL to the Webshare home page.
	ReferralURL string `json:"referral_url"`
	// ReferralMaximumCredits is the maximum credits or payouts earnable
	// from a single referral.
	ReferralMaximumCredits float64 `json:"referral_maximum_credits"`
	// ReferralCreditRatio is the ratio of earnings per referral purchase
	// (0.25 means $25 per $100 spend).
	ReferralCreditRatio float64 `json:"referral_credit_ratio"`
	// ReferralPaymentPendingDays is the grace period before a referral
	// credit becomes available, in Django duration format.
	ReferralPaymentPendingDays string `json:"referral_payment_pending_days"`
	// PromoType is the promotion type referrals receive. May be nil.
	PromoType *SubscriptionPromoType `json:"promo_type"`
	// PromoValue is the promotion value (10 or 20). Nil when PromoType is
	// nil.
	PromoValue *int `json:"promo_value"`
	// CreatedAt is when this instance was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when this instance was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// GetConfig retrieves the referral config. Switching to the affiliate
// program via the API indicates consent to the affiliate agreement.
func (s *ReferralService) GetConfig(ctx context.Context, opts ...RequestOption) (*ReferralConfig, error) {
	out := &ReferralConfig{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/referral/config/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// ReferralConfigUpdateParams are the parameters for
// ReferralService.UpdateConfig. Only Mode and PayPalPayoutEmail are
// writable.
type ReferralConfigUpdateParams struct {
	// Mode selects payout or credits mode.
	Mode ReferralMode
	// PayPalPayoutEmail sets the PayPal payout address; it must be set when
	// Mode is payout.
	PayPalPayoutEmail Nullable[string]
}

// MarshalJSON implements json.Marshaler, sending only the set fields.
func (p ReferralConfigUpdateParams) MarshalJSON() ([]byte, error) {
	body := map[string]any{}
	if p.Mode != "" {
		body["mode"] = p.Mode
	}
	if p.PayPalPayoutEmail.isPresent() {
		body["paypal_payout_email"] = p.PayPalPayoutEmail
	}
	return json.Marshal(body)
}

// UpdateConfig updates the referral config.
func (s *ReferralService) UpdateConfig(ctx context.Context, params ReferralConfigUpdateParams, opts ...RequestOption) (*ReferralConfig, error) {
	out := &ReferralConfig{}
	if err := s.client.doJSON(ctx, http.MethodPatch, "/api/v2/referral/config/", nil, params, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// CouponPromoType is the discount type of a coupon code or referral
// channel. This vocabulary is distinct from SubscriptionPromoType.
type CouponPromoType string

// Coupon promotion types.
const (
	// CouponPercentOff discounts a fraction of the price ("0.20" = 20%).
	CouponPercentOff CouponPromoType = "percent_off"
	// CouponValueOff discounts an absolute USD amount.
	CouponValueOff CouponPromoType = "value_off"
)

// CouponCode is the coupon code applied to the user. Every field is nil
// when no code is applied.
type CouponCode struct {
	// Code is the applied coupon code. Nil when no code is applied.
	Code *string `json:"code"`
	// PromoType is percent_off or value_off. Nil when no code is applied.
	PromoType *CouponPromoType `json:"promo_type"`
	// PromoValue is the discount magnitude as a decimal string. Nil when
	// no code is applied.
	PromoValue *string `json:"promo_value"`
	// Description is the human-readable channel description. Nil when no
	// code is applied.
	Description *string `json:"description"`
	// IsRecurring reports whether the discount applies to every renewal.
	// Nil when no code is applied.
	IsRecurring *bool `json:"is_recurring"`
}

// GetCouponCode retrieves the coupon code currently applied to the user. At
// most one coupon code can be applied at a time.
func (s *ReferralService) GetCouponCode(ctx context.Context, opts ...RequestOption) (*CouponCode, error) {
	out := &CouponCode{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/referral/coupon-code/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// ApplyCouponCode applies a coupon code to the user, replacing any
// previously applied code. Documented 400 error codes: not_found,
// code_inactive, code_expired, self_referral, already_redeemed. This
// endpoint has its own rate limit of 5 requests per minute per user.
func (s *ReferralService) ApplyCouponCode(ctx context.Context, code string, opts ...RequestOption) (*CouponCode, error) {
	body := struct {
		Code string `json:"code"`
	}{Code: code}
	out := &CouponCode{}
	if err := s.client.doJSON(ctx, http.MethodPost, "/api/v2/referral/coupon-code/", nil, body, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// RemoveCouponCode removes the coupon code currently applied to the user.
// Already-redeemed codes are not affected.
func (s *ReferralService) RemoveCouponCode(ctx context.Context, opts ...RequestOption) error {
	return s.client.doJSON(ctx, http.MethodDelete, "/api/v2/referral/coupon-code/", nil, nil, nil, opts)
}

// ReferralChannel is a coupon code owned by the user, created and
// configured by the Webshare team.
type ReferralChannel struct {
	// ID is the unique identifier of the referral channel.
	ID int `json:"id"`
	// Code is the coupon code other users apply. Matched
	// case-insensitively.
	Code string `json:"code"`
	// Description is the human-readable channel description.
	Description string `json:"description"`
	// PromoType is percent_off or value_off.
	PromoType CouponPromoType `json:"promo_type"`
	// PromoValue is the discount magnitude as a decimal string.
	PromoValue string `json:"promo_value"`
	// IsActive reports whether new users can still apply the code.
	IsActive bool `json:"is_active"`
	// IsRecurring reports whether the discount applies to every renewal.
	IsRecurring bool `json:"is_recurring"`
	// StartDate is when the code becomes applicable.
	StartDate time.Time `json:"start_date"`
	// EndDate is when the code stops being applicable. May be nil for
	// non-expiring channels.
	EndDate *time.Time `json:"end_date"`
	// TotalUses is the number of distinct redemptions across all users.
	TotalUses int `json:"total_uses"`
	// TotalCommission is the total USD commission earned through the
	// channel.
	TotalCommission float64 `json:"total_commission"`
	// TotalRevenue is the total USD revenue generated by redeemers, net of
	// refunds.
	TotalRevenue float64 `json:"total_revenue"`
	// InitialRate is the commission rate (0-1, decimal string) during the
	// initial period.
	InitialRate string `json:"initial_rate"`
	// OngoingRate is the commission rate (0-1, decimal string) after the
	// initial period.
	OngoingRate string `json:"ongoing_rate"`
	// InitialRatePeriodDays is the length of the initial commission period
	// in days.
	InitialRatePeriodDays int `json:"initial_rate_period_days"`
	// MaxEarningsPerReferred is the maximum USD commission (decimal
	// string) earnable from a single referred user.
	MaxEarningsPerReferred string `json:"max_earnings_per_referred"`
	// CreatedAt is when this instance was created.
	CreatedAt time.Time `json:"created_at"`
}

// ListChannels returns the referral channels owned by the user. This
// endpoint is not paginated: it returns a plain array ordered by ID
// ascending.
func (s *ReferralService) ListChannels(ctx context.Context, opts ...RequestOption) ([]ReferralChannel, error) {
	var out []ReferralChannel
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/referral/channel/", nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// ReferralCreditStatus is the status of a referral credit.
type ReferralCreditStatus string

// Referral credit statuses.
const (
	// CreditPending means the credit is inside the grace period.
	CreditPending ReferralCreditStatus = "pending"
	// CreditAvailable means the credit is available.
	CreditAvailable ReferralCreditStatus = "available"
	// CreditReverted means the referral payment was reversed and the
	// credit reverted.
	CreditReverted ReferralCreditStatus = "reverted"
)

// ReferralCredit is one credit earned from a referral's spend.
type ReferralCredit struct {
	// ID is the unique identifier of the referral credit instance.
	ID int `json:"id"`
	// UserID is the referred user who spent money on Webshare.
	UserID int `json:"user_id"`
	// Mode is whether the credit is in payout or credits mode.
	Mode string `json:"mode"`
	// Amount is the amount earned in USD.
	Amount float64 `json:"amount"`
	// Status is the credit status; it becomes available after the
	// referral payment pending period.
	Status ReferralCreditStatus `json:"status"`
	// ReferralChannel is the channel that produced the credit. Nil for
	// credits earned through a referral link.
	ReferralChannel *int `json:"referral_channel"`
	// ReferralChannelCode is the code of the producing channel. Nil for
	// credits earned through a referral link.
	ReferralChannelCode *string `json:"referral_channel_code"`
	// CreatedAt is when this instance was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when this instance was last updated.
	UpdatedAt time.Time `json:"updated_at"`
	// RevertedAt is when the credit was reverted. Nil unless Status is
	// reverted.
	RevertedAt *time.Time `json:"reverted_at"`
}

// ReferralCreditListParams are the parameters for
// ReferralService.ListCredits.
type ReferralCreditListParams struct {
	// Page is the page number.
	Page *int
	// PageSize is the number of results per page.
	PageSize *int
	// Mode filters by payout or credits.
	Mode ReferralMode
	// Status filters by credit status.
	Status ReferralCreditStatus
	// ReferralChannel filters by the ID of an owned referral channel.
	ReferralChannel *int
	// Ordering orders by id, mode, amount, status, created_at, updated_at
	// or reverted_at; prefix with "-" for descending order.
	Ordering string
}

// ListCredits returns the referral credits in paginated format.
func (s *ReferralService) ListCredits(ctx context.Context, params ReferralCreditListParams, opts ...RequestOption) (*Page[ReferralCredit], error) {
	q := url.Values{}
	setInt(q, "page", params.Page)
	setInt(q, "page_size", params.PageSize)
	setString(q, "mode", string(params.Mode))
	setString(q, "status", string(params.Status))
	setInt(q, "referral_channel", params.ReferralChannel)
	setString(q, "ordering", params.Ordering)
	return getPage[ReferralCredit](ctx, s.client, "/api/v2/referral/credit/", q, opts)
}

// ListAllCredits returns a lazy iterator over every referral credit across
// all pages.
func (s *ReferralService) ListAllCredits(ctx context.Context, params ReferralCreditListParams, opts ...RequestOption) iter.Seq2[ReferralCredit, error] {
	return iterPages(ctx, func(ctx context.Context) (*Page[ReferralCredit], error) {
		return s.ListCredits(ctx, params, opts...)
	})
}

// GetCredit retrieves a referral credit.
func (s *ReferralService) GetCredit(ctx context.Context, id int, opts ...RequestOption) (*ReferralCredit, error) {
	out := &ReferralCredit{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/referral/credit/"+strconv.Itoa(id)+"/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// EarnoutStatus is the status of an earn-out.
type EarnoutStatus string

// Earn-out statuses.
const (
	// EarnoutProcessing means the earn-out is being processed.
	EarnoutProcessing EarnoutStatus = "processing"
	// EarnoutCompleted means the earn-out completed.
	EarnoutCompleted EarnoutStatus = "completed"
	// EarnoutFailed means the earn-out failed; ErrorReason is set.
	EarnoutFailed EarnoutStatus = "failed"
)

// ReferralEarnout is one conversion of pending referral earnings into
// account credits or a PayPal payout.
type ReferralEarnout struct {
	// ID is the unique identifier of the earn-out instance.
	ID int `json:"id"`
	// Mode is whether the earn-out is a payout or credits.
	Mode string `json:"mode"`
	// PayPalPayoutEmail is the PayPal address that received the funds in
	// payout mode. May be nil.
	PayPalPayoutEmail *string `json:"paypal_payout_email"`
	// Amount is the amount earned out in USD.
	Amount float64 `json:"amount"`
	// Status is the earn-out status.
	Status EarnoutStatus `json:"status"`
	// ErrorReason describes the failure when Status is failed. May be nil.
	ErrorReason *string `json:"error_reason"`
	// CreatedAt is when this instance was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when this instance was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// ReferralEarnoutListParams are the parameters for
// ReferralService.ListEarnouts.
type ReferralEarnoutListParams struct {
	// Page is the page number.
	Page *int
	// PageSize is the number of results per page.
	PageSize *int
}

// ListEarnouts returns the earn-outs in paginated format. Note the endpoint
// path has no trailing slash in the docs.
func (s *ReferralService) ListEarnouts(ctx context.Context, params ReferralEarnoutListParams, opts ...RequestOption) (*Page[ReferralEarnout], error) {
	q := url.Values{}
	setInt(q, "page", params.Page)
	setInt(q, "page_size", params.PageSize)
	return getPage[ReferralEarnout](ctx, s.client, "/api/v2/referral/earnout", q, opts)
}

// ListAllEarnouts returns a lazy iterator over every earn-out across all
// pages.
func (s *ReferralService) ListAllEarnouts(ctx context.Context, params ReferralEarnoutListParams, opts ...RequestOption) iter.Seq2[ReferralEarnout, error] {
	return iterPages(ctx, func(ctx context.Context) (*Page[ReferralEarnout], error) {
		return s.ListEarnouts(ctx, params, opts...)
	})
}

// GetEarnout retrieves an earn-out.
func (s *ReferralService) GetEarnout(ctx context.Context, id int, opts ...RequestOption) (*ReferralEarnout, error) {
	out := &ReferralEarnout{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/referral/earnout/"+strconv.Itoa(id)+"/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// ReferralCodeInfo is the publicly available information of a referral
// code.
type ReferralCodeInfo struct {
	// ReferralCode is the referral code.
	ReferralCode string `json:"referral_code"`
	// PromoType is the promotion type. May be nil.
	PromoType *SubscriptionPromoType `json:"promo_type"`
	// PromoValue is the promotion value (10 or 20). Nil when PromoType is
	// nil.
	PromoValue *int `json:"promo_value"`
}

// GetCodeInfo retrieves the public information of a referral code. This is
// the only unauthenticated endpoint in the referral group; no Authorization
// header is sent.
func (s *ReferralService) GetCodeInfo(ctx context.Context, referralCode string, opts ...RequestOption) (*ReferralCodeInfo, error) {
	q := url.Values{}
	q.Set("referral_code", referralCode)
	out := &ReferralCodeInfo{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/referral/code/info/", q, nil, out, withOptions(opts, withAuthMode(authNone))); err != nil {
		return nil, err
	}
	return out, nil
}
