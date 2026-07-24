package webshare

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// SubscriptionService exposes the subscription operations. Each account has
// exactly one subscription with a stable ID; a new plan object is created
// every time the customer re-customizes.
type SubscriptionService struct {
	client *Client
}

// SubscriptionPromoType is the type of promotion applied to a subscription
// or referral.
type SubscriptionPromoType string

// Subscription promotion types.
const (
	// PromoFirstTimeValueOff discounts a fixed value off the first
	// purchase.
	PromoFirstTimeValueOff SubscriptionPromoType = "first_time_value_off"
	// PromoFirstTimePercentOff discounts a percentage off the first
	// purchase.
	PromoFirstTimePercentOff SubscriptionPromoType = "first_time_percent_off"
	// PromoAlwaysValueOff discounts a fixed value off every purchase.
	PromoAlwaysValueOff SubscriptionPromoType = "always_value_off"
	// PromoAlwaysPercentOff discounts a percentage off every purchase.
	PromoAlwaysPercentOff SubscriptionPromoType = "always_percent_off"
)

// Subscription is the subscription object of the account.
type Subscription struct {
	// ID is the unique identifier of the subscription instance; it does not
	// change for the user.
	ID int `json:"id"`
	// Plan is the ID of the active plan instance; it changes whenever the
	// user re-customizes their plan.
	Plan int `json:"plan"`
	// PaymentMethod is the ID of the payment method on file. Nil when
	// auto-renewal is cancelled.
	PaymentMethod *int `json:"payment_method"`
	// FreeCredits is the free credits available for the account in USD.
	FreeCredits float64 `json:"free_credits"`
	// Term determines the amount charged at the next renewal.
	Term SubscriptionTerm `json:"term"`
	// StartDate is the start of the current renewal term. The difference
	// between end and start dates is always 30 days, even on yearly terms.
	StartDate time.Time `json:"start_date"`
	// EndDate is the end of the current renewal term.
	EndDate time.Time `json:"end_date"`
	// RenewalsPaid is the number of 30-day renewals paid. Yearly terms pay
	// for 12 renewals at once.
	RenewalsPaid int `json:"renewals_paid"`
	// RenewalsEnabled reports whether auto-renewal is enabled.
	RenewalsEnabled bool `json:"renewals_enabled"`
	// FailedPaymentTimes counts failed automated renewal payments.
	FailedPaymentTimes int `json:"failed_payment_times"`
	// AccountDiscountPercentage is the discount percentage for the account.
	AccountDiscountPercentage int `json:"account_discount_percentage"`
	// PromotionAvailableFirstTimeRenewal25Off reports whether the 25% off
	// first renewal promotion is available.
	PromotionAvailableFirstTimeRenewal25Off bool `json:"promotion_available_first_time_renewal_25_off"`
	// Customizable reports whether the subscription is customizable.
	Customizable bool `json:"customizable"`
	// Paused reports whether the subscription is paused.
	Paused bool `json:"paused"`
	// ReactivationDate is when a paused subscription resumes. May be nil.
	ReactivationDate *time.Time `json:"reactivation_date"`
	// ReactivationPeriodLeft is the period left in a paused subscription.
	// May be nil.
	ReactivationPeriodLeft *string `json:"reactivation_period_left"`
	// PromoType is the promotion type for the account. May be nil.
	PromoType *SubscriptionPromoType `json:"promo_type"`
	// PromoValue is the promotion value (10 or 20). Nil when PromoType is
	// nil.
	PromoValue *int `json:"promo_value"`
	// Throttled reports whether the subscription is throttled, usually due
	// to high bandwidth usage with few proxies.
	Throttled bool `json:"throttled"`
	// CreatedAt is when this instance was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when this instance was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// Get returns the subscription object associated with the account.
func (s *SubscriptionService) Get(ctx context.Context, opts ...RequestOption) (*Subscription, error) {
	out := &Subscription{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/subscription/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// AssetInfo describes the assets available for one proxy category and
// subtype.
type AssetInfo struct {
	// TotalSubnets is the number of subnets available under the category.
	TotalSubnets int `json:"total_subnets"`
	// AvailableCountries maps country code to the number of proxies
	// available in that country.
	AvailableCountries map[string]int `json:"available_countries"`
}

// GetAvailableAssets returns the available assets for each proxy category.
// The result maps proxy category (shared, semidedicated, dedicated) to proxy
// subtype to asset info.
func (s *SubscriptionService) GetAvailableAssets(ctx context.Context, opts ...RequestOption) (map[string]map[string]AssetInfo, error) {
	out := map[string]map[string]AssetInfo{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/subscription/available_assets/", nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// ProxyType is the category of proxies in a plan.
type ProxyType string

// Proxy types.
const (
	// ProxyTypeFree is the free proxy category.
	ProxyTypeFree ProxyType = "free"
	// ProxyTypeShared is the shared proxy category.
	ProxyTypeShared ProxyType = "shared"
	// ProxyTypeSemidedicated is the semi-dedicated proxy category.
	ProxyTypeSemidedicated ProxyType = "semidedicated"
	// ProxyTypeDedicated is the dedicated proxy category.
	ProxyTypeDedicated ProxyType = "dedicated"
)

// ProxySubtype is the sub-category of proxies in a plan. Not all proxy types
// have the same subtypes.
type ProxySubtype string

// Proxy subtypes.
const (
	// SubtypeDefault is the default proxy subtype.
	SubtypeDefault ProxySubtype = "default"
	// SubtypePremium is the premium proxy subtype.
	SubtypePremium ProxySubtype = "premium"
	// SubtypeISP is the ISP proxy subtype.
	SubtypeISP ProxySubtype = "isp"
	// SubtypeResidential is the residential proxy subtype.
	SubtypeResidential ProxySubtype = "residential"
	// SubtypeDatacenterAndISP is the datacenter-and-ISP proxy subtype.
	SubtypeDatacenterAndISP ProxySubtype = "datacenter_and_isp"
)

// SubscriptionCustomizeParams are the parameters for
// SubscriptionService.Customize. On the wire the request is JSON encoded
// into a single "query" GET parameter; the SDK handles that.
type SubscriptionCustomizeParams struct {
	// PlanID targets a specific plan; otherwise the default plan is used.
	PlanID *int `json:"-"`
	// ProxyType is the proxy category.
	ProxyType ProxyType `json:"proxy_type,omitempty"`
	// ProxySubtype is the proxy sub-category.
	ProxySubtype ProxySubtype `json:"proxy_subtype,omitempty"`
	// ProxyCountries maps country code to proxy count; ZZ means randomly
	// allocated. Other customizations are based on this field.
	ProxyCountries map[string]int `json:"proxy_countries,omitempty"`
	// RequiredSiteChecks lists site checks the proxy list must pass.
	RequiredSiteChecks []string `json:"required_site_checks,omitempty"`
	// HighQualityIPsOnly restricts to high quality IPs. Not available for
	// residential plans.
	HighQualityIPsOnly *bool `json:"high_quality_ips_only,omitempty"`
}

// SubscriptionCustomization describes the limits and options available when
// customizing a plan.
type SubscriptionCustomization struct {
	// ProxyType is the proxy category the options apply to.
	ProxyType ProxyType `json:"proxy_type"`
	// ProxySubtype is the proxy sub-category the options apply to.
	ProxySubtype ProxySubtype `json:"proxy_subtype"`
	// ProxyCountMax is the maximum number of proxies.
	ProxyCountMax int `json:"proxy_count_max"`
	// ProxyCountMin is the minimum number of proxies.
	ProxyCountMin int `json:"proxy_count_min"`
	// AvailableCountries maps country code to available proxy count.
	AvailableCountries map[string]int `json:"available_countries"`
	// OnDemandRefreshesMax is the maximum number of on-demand refreshes.
	OnDemandRefreshesMax int `json:"on_demand_refreshes_max"`
	// OnDemandRefreshesMin is the minimum number of on-demand refreshes.
	OnDemandRefreshesMin int `json:"on_demand_refreshes_min"`
	// AutomaticRefreshFrequencyMax is the maximum refresh frequency in
	// seconds.
	AutomaticRefreshFrequencyMax int `json:"automatic_refresh_frequency_max"`
	// AutomaticRefreshFrequencyMin is the minimum refresh frequency in
	// seconds.
	AutomaticRefreshFrequencyMin int `json:"automatic_refresh_frequency_min"`
	// ProxyReplacementsMax is the maximum number of proxy replacements.
	ProxyReplacementsMax int `json:"proxy_replacements_max"`
	// ProxyReplacementsMin is the minimum number of proxy replacements.
	ProxyReplacementsMin int `json:"proxy_replacements_min"`
	// BandwidthLimitMax is the maximum bandwidth limit in GB.
	BandwidthLimitMax int `json:"bandwidth_limit_max"`
	// BandwidthLimitMin is the minimum bandwidth limit in GB.
	BandwidthLimitMin int `json:"bandwidth_limit_min"`
	// SubusersMax is the maximum number of sub-users.
	SubusersMax int `json:"subusers_max"`
	// SubusersMin is the minimum number of sub-users.
	SubusersMin int `json:"subusers_min"`
	// AvailableFeatures lists the plan features that can be selected.
	AvailableFeatures []CustomizationFeature `json:"available_features"`
	// AvailableSiteChecks lists the site checks that can be required.
	AvailableSiteChecks []SiteCheck `json:"available_site_checks"`
	// Terms lists the available terms and their renewals paid.
	Terms []CustomizationTerm `json:"terms"`
}

// CustomizationFeature is one selectable plan feature.
type CustomizationFeature struct {
	// Feature is the feature name.
	Feature string `json:"feature"`
	// Required reports whether the feature is required.
	Required bool `json:"required"`
}

// SiteCheck is one available site check.
type SiteCheck struct {
	// Name is the site check name.
	Name string `json:"name"`
}

// CustomizationTerm is one available term option.
type CustomizationTerm struct {
	// Term is the term name.
	Term SubscriptionTerm `json:"term"`
	// RenewalsPaid is the number of renewals paid at once for the term.
	RenewalsPaid int `json:"renewals_paid"`
}

// queryParam JSON encodes params into the single "query" GET parameter used
// by the customize and pricing endpoints.
func queryParam(params any, planID *int) (url.Values, error) {
	encoded, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("webshare: encoding query parameter: %w", err)
	}
	q := url.Values{}
	q.Set("query", string(encoded))
	setInt(q, "plan_id", planID)
	return q, nil
}

// Customize returns the limits and options available to customize a plan.
func (s *SubscriptionService) Customize(ctx context.Context, params SubscriptionCustomizeParams, opts ...RequestOption) (*SubscriptionCustomization, error) {
	q, err := queryParam(params, params.PlanID)
	if err != nil {
		return nil, err
	}
	out := &SubscriptionCustomization{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/subscription/customize/", q, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// CheckoutBehavior selects how a purchased plan interacts with the
// subscription's existing plans.
type CheckoutBehavior string

// Checkout behaviors.
const (
	// BehaviorReplace replaces the old plan with the new one. Only
	// supported for subscriptions with a single plan. The default.
	BehaviorReplace CheckoutBehavior = "replace"
	// BehaviorAdd adds the plan to the subscription.
	BehaviorAdd CheckoutBehavior = "add"
	// BehaviorUpgrade upgrades the plan (used by the pricing endpoint).
	BehaviorUpgrade CheckoutBehavior = "upgrade"
)

// PlanConfiguration describes a plan being priced or purchased. It is
// shared by the pricing, purchase and upgrade endpoints.
type PlanConfiguration struct {
	// ProxyType is the proxy category.
	ProxyType ProxyType `json:"proxy_type,omitempty"`
	// ProxySubtype is the proxy sub-category.
	ProxySubtype ProxySubtype `json:"proxy_subtype,omitempty"`
	// ProxyCountries maps country code to proxy count; ZZ means randomly
	// allocated.
	ProxyCountries map[string]int `json:"proxy_countries,omitempty"`
	// BandwidthLimit is the bandwidth limit in GB. Zero means unlimited.
	BandwidthLimit float64 `json:"bandwidth_limit"`
	// OnDemandRefreshesTotal is the number of on-demand refreshes purchased.
	OnDemandRefreshesTotal int `json:"on_demand_refreshes_total"`
	// AutomaticRefreshFrequency refreshes the proxy list every N seconds.
	// Zero means no automatic refreshes.
	AutomaticRefreshFrequency int `json:"automatic_refresh_frequency"`
	// ProxyReplacementsTotal is the number of proxy replacements purchased.
	ProxyReplacementsTotal int `json:"proxy_replacements_total"`
	// SubusersTotal is the number of sub-users allowed.
	SubusersTotal int `json:"subusers_total"`
	// IsUnlimitedIPAuthorizations enables unlimited IP authorizations.
	IsUnlimitedIPAuthorizations bool `json:"is_unlimited_ip_authorizations"`
	// IsHighConcurrency enables high concurrency (3,000 concurrent
	// requests).
	IsHighConcurrency bool `json:"is_high_concurrency"`
	// Is2XConcurrency enables 2x concurrency (1,000 concurrent requests).
	Is2XConcurrency bool `json:"is_2x_concurrency"`
	// IsHighPriorityNetwork enables the high priority network.
	IsHighPriorityNetwork bool `json:"is_high_priority_network"`
	// HighQualityIPsOnly restricts to high quality IPs with perfect fraud
	// scores. Not available for residential plans; 30% price premium.
	HighQualityIPsOnly bool `json:"high_quality_ips_only"`
	// RequiredSiteChecks lists site checks the proxy list must pass.
	RequiredSiteChecks []string `json:"required_site_checks,omitempty"`
	// Term is the subscription term.
	Term SubscriptionTerm `json:"term,omitempty"`
}

// SubscriptionPricingParams are the parameters for
// SubscriptionService.Pricing. On the wire the request is JSON encoded into
// a single "query" GET parameter; the SDK handles that.
type SubscriptionPricingParams struct {
	// PlanID targets a specific plan; otherwise the default plan is used.
	PlanID *int `json:"-"`
	// Behavior selects replace (default), add or upgrade.
	Behavior CheckoutBehavior `json:"behavior,omitempty"`
	// PlanConfiguration describes the plan being priced.
	PlanConfiguration
	// WithTax includes a tax breakdown in the response.
	WithTax bool `json:"with_tax,omitempty"`
}

// PricingTier is one discount tier of the pricing response. From is
// exclusive and To is inclusive; a nil To extends to infinity.
type PricingTier struct {
	// From is the exclusive lower bound of the tier.
	From int `json:"from"`
	// To is the inclusive upper bound of the tier. May be nil.
	To *int `json:"to"`
	// DiscountPercentage is the discount for the tier.
	DiscountPercentage int `json:"discount_percentage"`
	// PerProxyPrice is the price per proxy in the tier.
	PerProxyPrice float64 `json:"per_proxy_price"`
	// PerGBPrice is the price per GB in the tier. May be nil.
	PerGBPrice *float64 `json:"per_gb_price"`
}

// FeaturePrice is one feature's price in the pricing response.
type FeaturePrice struct {
	// Feature is the feature name.
	Feature string `json:"feature"`
	// IsSelected reports whether the feature is part of the priced plan.
	IsSelected bool `json:"is_selected"`
	// Price is the feature price in USD.
	Price float64 `json:"price"`
}

// TaxEntry is one entry of the tax breakdown.
type TaxEntry struct {
	// Amount is the tax amount.
	Amount string `json:"amount"`
	// TaxRateDetails describes the applied tax rate.
	TaxRateDetails TaxRateDetails `json:"tax_rate_details"`
	// TaxableAmount is the amount the tax applies to.
	TaxableAmount string `json:"taxable_amount"`
}

// TaxRateDetails describes an applied tax rate.
type TaxRateDetails struct {
	// PercentageDecimal is the tax percentage as a decimal string.
	PercentageDecimal string `json:"percentage_decimal"`
	// TaxType is the kind of tax, for example "gst".
	TaxType string `json:"tax_type"`
}

// CouponDiscount summarizes the coupon code applied to the user.
type CouponDiscount struct {
	// Code is the applied coupon code.
	Code string `json:"code"`
	// PromoType is percent_off or value_off.
	PromoType CouponPromoType `json:"promo_type"`
	// PromoValue is the discount magnitude as a decimal string.
	PromoValue string `json:"promo_value"`
	// IsRecurring reports whether the discount applies to every renewal.
	IsRecurring bool `json:"is_recurring"`
}

// SubscriptionPricing is the pricing of a custom plan.
type SubscriptionPricing struct {
	// DiscountPercentage is the percentage discount applied to the final
	// price.
	DiscountPercentage int `json:"discount_percentage"`
	// NonDiscountedPrice is the original price before discounts.
	NonDiscountedPrice float64 `json:"non_discounted_price"`
	// Price is the price after discounts.
	Price float64 `json:"price"`
	// PaidToday is the amount to be paid today, after credits.
	PaidToday float64 `json:"paid_today"`
	// PromoDiscount is the USD value of the promo discount applied to the
	// price.
	PromoDiscount float64 `json:"promo_discount"`
	// CreditsAdded is the credits added to make this subscription change.
	CreditsAdded float64 `json:"credits_added"`
	// CreditsUsed is the total credits used to change the subscription.
	CreditsUsed float64 `json:"credits_used"`
	// ProxyCountDiscountTiers lists per-proxy discount tiers.
	ProxyCountDiscountTiers []PricingTier `json:"proxy_count_discount_tiers"`
	// BandwidthDiscountTiers lists per-GB price tiers.
	BandwidthDiscountTiers []PricingTier `json:"bandwidth_discount_tiers"`
	// Features lists the features and their prices.
	Features []FeaturePrice `json:"features"`
	// TaxBreakdown lists tax entries when WithTax was requested.
	TaxBreakdown []TaxEntry `json:"tax_breakdown"`
	// CouponDiscount summarizes the applied coupon code. May be nil.
	CouponDiscount *CouponDiscount `json:"coupon_discount"`
}

// Pricing returns the pricing for a custom plan.
func (s *SubscriptionService) Pricing(ctx context.Context, params SubscriptionPricingParams, opts ...RequestOption) (*SubscriptionPricing, error) {
	q, err := queryParam(params, params.PlanID)
	if err != nil {
		return nil, err
	}
	out := &SubscriptionPricing{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/subscription/pricing/", q, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// PaymentMethodRef selects the payment method for checkout endpoints. It is
// tri-modal on the wire: null (use the payment method on file), an integer
// Webshare payment method ID, or a Stripe PaymentMethod string ("pm_...").
// The zero value serializes as null.
type PaymentMethodRef struct {
	id       *int
	stripeID string
}

// PaymentMethodOnFile uses the payment method on file (JSON null).
func PaymentMethodOnFile() PaymentMethodRef {
	return PaymentMethodRef{}
}

// PaymentMethodID references an existing Webshare payment method by ID.
func PaymentMethodID(id int) PaymentMethodRef {
	return PaymentMethodRef{id: &id}
}

// StripePaymentMethod references a new Stripe PaymentMethod by its "pm_..."
// identifier.
func StripePaymentMethod(id string) PaymentMethodRef {
	return PaymentMethodRef{stripeID: id}
}

// MarshalJSON implements json.Marshaler.
func (r PaymentMethodRef) MarshalJSON() ([]byte, error) {
	switch {
	case r.id != nil:
		return json.Marshal(*r.id)
	case r.stripeID != "":
		return json.Marshal(r.stripeID)
	default:
		return []byte("null"), nil
	}
}

// CheckoutResult is the shared response of the purchase, upgrade and renew
// endpoints.
type CheckoutResult struct {
	// PaymentRequired reports whether additional payment steps are needed.
	// When false, the purchase is complete and the account has the new
	// plan.
	PaymentRequired bool `json:"payment_required"`
	// Plan is the ID of the new plan object.
	Plan int `json:"plan"`
	// PendingPayment is the ID of the pending payment instance. Only
	// present when PaymentRequired is true.
	PendingPayment int `json:"pending_payment,omitempty"`
	// StripeClientSecret is the client secret for the Stripe PaymentIntent.
	// Only present when PaymentRequired is true.
	StripeClientSecret string `json:"stripe_client_secret,omitempty"`
	// StripePaymentIntent is the ID of the Stripe PaymentIntent. Only
	// present when PaymentRequired is true.
	StripePaymentIntent string `json:"stripe_payment_intent,omitempty"`
	// StripePaymentMethod is the ID of the Stripe PaymentMethod. Only
	// present when PaymentRequired is true.
	StripePaymentMethod string `json:"stripe_payment_method,omitempty"`
}

// SubscriptionPurchaseParams are the parameters for
// SubscriptionService.Purchase.
type SubscriptionPurchaseParams struct {
	// Behavior selects replace (default) or add.
	Behavior CheckoutBehavior `json:"behavior,omitempty"`
	// PlanConfiguration describes the plan being purchased.
	PlanConfiguration
	// PaymentMethod selects the payment method.
	PaymentMethod PaymentMethodRef `json:"payment_method"`
	// Recaptcha is the recaptcha token. Only required when a payment is
	// required.
	Recaptcha string `json:"recaptcha,omitempty"`
}

// Purchase purchases a new plan, replacing the existing plan and updating
// the subscription start date. Recaptcha validation is required when a
// payment is required; with enough account credits no payment is needed.
// When a payment is required the docs mark this endpoint as usable only from
// the Webshare dashboard, not programmatically.
func (s *SubscriptionService) Purchase(ctx context.Context, params SubscriptionPurchaseParams, opts ...RequestOption) (*CheckoutResult, error) {
	out := &CheckoutResult{}
	if err := s.client.doJSON(ctx, http.MethodPost, "/api/v2/subscription/checkout/purchase/", nil, params, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// SubscriptionRenewParams are the parameters for SubscriptionService.Renew.
type SubscriptionRenewParams struct {
	// PaymentMethod selects the payment method: the payment on file (null)
	// or an existing payment method ID.
	PaymentMethod PaymentMethodRef `json:"payment_method"`
	// Term is the term to renew.
	Term SubscriptionTerm `json:"term,omitempty"`
	// Recaptcha is the recaptcha token. Only required when a payment is
	// required.
	Recaptcha string `json:"recaptcha,omitempty"`
}

// Renew renews the subscription, adding renewals paid. Recaptcha validation
// is required when a payment is required; with enough account credits no
// payment is needed. When a payment is required the docs mark this endpoint
// as usable only from the Webshare dashboard, not programmatically.
func (s *SubscriptionService) Renew(ctx context.Context, params SubscriptionRenewParams, opts ...RequestOption) (*CheckoutResult, error) {
	out := &CheckoutResult{}
	if err := s.client.doJSON(ctx, http.MethodPost, "/api/v2/subscription/checkout/renew/", nil, params, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// EnableAutoRenewal enables auto-renewal for the subscription. A payment
// method must already be on file. Auto-renewal cannot be modified for free
// plans. Returns the subscription with RenewalsEnabled set to true.
func (s *SubscriptionService) EnableAutoRenewal(ctx context.Context, opts ...RequestOption) (*Subscription, error) {
	out := &Subscription{}
	if err := s.client.doJSON(ctx, http.MethodPut, "/api/v2/subscription/renewal/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// CancelAutoRenewal cancels auto-renewal for the subscription and removes
// the payment method on file. Auto-renewal cannot be modified for free
// plans. Unlike most DELETE endpoints it returns 200 with the subscription
// object (RenewalsEnabled false, PaymentMethod nil).
func (s *SubscriptionService) CancelAutoRenewal(ctx context.Context, opts ...RequestOption) (*Subscription, error) {
	out := &Subscription{}
	if err := s.client.doJSON(ctx, http.MethodDelete, "/api/v2/subscription/renewal/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}
