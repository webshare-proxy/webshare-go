package webshare

import (
	"context"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// PlansService exposes the plan operations.
type PlansService struct {
	client *Client
}

// PlanStatus is the status of a plan.
type PlanStatus string

// Plan statuses.
const (
	// PlanActive is the account's active plan.
	PlanActive PlanStatus = "active"
	// PlanCancelled is a plan that is no longer active.
	PlanCancelled PlanStatus = "cancelled"
)

// PlanBundleInfo is set when a plan is an addon in a bundle.
type PlanBundleInfo struct {
	// PrimaryPlanID is the ID of the bundle's primary plan. May be nil.
	PrimaryPlanID *int `json:"primary_plan_id"`
	// DiscountRate is the discount applied to the plan, e.g. 0.2 for 20%.
	DiscountRate float64 `json:"discount_rate"`
}

// PlanBundleAddon is one addon of a bundle's primary plan.
type PlanBundleAddon struct {
	// PlanID is the ID of the addon plan.
	PlanID int `json:"plan_id"`
	// DiscountRate is the discount applied to the addon.
	DiscountRate float64 `json:"discount_rate"`
}

// Plan is one plan created by the user, active or cancelled.
type Plan struct {
	// ID is the unique identifier of the plan instance.
	ID int `json:"id"`
	// Status is the plan status.
	Status PlanStatus `json:"status"`
	// BandwidthLimit is the bandwidth limit in GB. Zero means unlimited.
	BandwidthLimit float64 `json:"bandwidth_limit"`
	// MonthlyPrice is the USD price for the monthly term.
	MonthlyPrice float64 `json:"monthly_price"`
	// YearlyPrice is the USD price for the yearly term.
	YearlyPrice float64 `json:"yearly_price"`
	// ProxyType is the proxy category.
	ProxyType ProxyType `json:"proxy_type"`
	// ProxySubtype is the proxy sub-category.
	ProxySubtype ProxySubtype `json:"proxy_subtype"`
	// ProxyCount is the number of proxies in the plan.
	ProxyCount int `json:"proxy_count"`
	// ProxyCountries maps country code to proxy count; ZZ means randomly
	// allocated.
	ProxyCountries map[string]int `json:"proxy_countries"`
	// RequiredSiteChecks lists site checks the proxy list must pass.
	RequiredSiteChecks []string `json:"required_site_checks"`
	// OnDemandRefreshesTotal is the number of on-demand refreshes purchased.
	OnDemandRefreshesTotal int `json:"on_demand_refreshes_total"`
	// OnDemandRefreshesUsed counts refreshes used since the subscription
	// start date.
	OnDemandRefreshesUsed int `json:"on_demand_refreshes_used"`
	// OnDemandRefreshesAvailable is the number of refreshes available.
	OnDemandRefreshesAvailable int `json:"on_demand_refreshes_available"`
	// AutomaticRefreshFrequency refreshes the proxy list every N seconds.
	// Zero means no automatic refreshes.
	AutomaticRefreshFrequency int `json:"automatic_refresh_frequency"`
	// AutomaticRefreshLastAt is the last automatic refresh. Nil in the list
	// endpoint.
	AutomaticRefreshLastAt *time.Time `json:"automatic_refresh_last_at"`
	// AutomaticRefreshNextAt is the next automatic refresh. Nil in the list
	// endpoint.
	AutomaticRefreshNextAt *time.Time `json:"automatic_refresh_next_at"`
	// ProxyReplacementsTotal is the number of proxy replacements purchased.
	ProxyReplacementsTotal int `json:"proxy_replacements_total"`
	// ProxyReplacementsUsed counts replacements used since the subscription
	// start date.
	ProxyReplacementsUsed int `json:"proxy_replacements_used"`
	// ProxyReplacementsAvailable is the number of replacements available.
	ProxyReplacementsAvailable int `json:"proxy_replacements_available"`
	// SubusersTotal is the number of sub-users allowed in the plan.
	SubusersTotal int `json:"subusers_total"`
	// SubusersUsed is the number of sub-users in use.
	SubusersUsed int `json:"subusers_used"`
	// SubusersAvailable is the number of sub-users still available.
	SubusersAvailable int `json:"subusers_available"`
	// IsUnlimitedIPAuthorizations reports unlimited IP authorizations.
	IsUnlimitedIPAuthorizations bool `json:"is_unlimited_ip_authorizations"`
	// IsHighConcurrency reports high concurrency (3,000 concurrent
	// requests).
	IsHighConcurrency bool `json:"is_high_concurrency"`
	// Is2XConcurrency reports 2x concurrency (1,000 concurrent requests).
	Is2XConcurrency bool `json:"is_2x_concurrency"`
	// IsHighPriorityNetwork reports the high priority network.
	IsHighPriorityNetwork bool `json:"is_high_priority_network"`
	// HighQualityIPsOnly reports whether only high quality IPs are used.
	HighQualityIPsOnly bool `json:"high_quality_ips_only"`
	// BundleInfo is set when the plan is a bundle addon. May be nil.
	BundleInfo *PlanBundleInfo `json:"bundle_info"`
	// BundleAddons lists addons when the plan is a bundle's primary plan.
	BundleAddons []PlanBundleAddon `json:"bundle_addons"`
	// CreatedAt is when this instance was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when this instance was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// PlanListParams are the parameters for PlansService.List.
type PlanListParams struct {
	// Page is the page number.
	Page *int
	// PageSize is the number of results per page.
	PageSize *int
}

// List retrieves all plans created by the user, including non-active ones,
// in paginated format.
func (s *PlansService) List(ctx context.Context, params PlanListParams, opts ...RequestOption) (*Page[Plan], error) {
	q := url.Values{}
	setInt(q, "page", params.Page)
	setInt(q, "page_size", params.PageSize)
	return getPage[Plan](ctx, s.client, "/api/v2/subscription/plan/", q, opts)
}

// ListAll returns a lazy iterator over every plan across all pages.
func (s *PlansService) ListAll(ctx context.Context, params PlanListParams, opts ...RequestOption) iter.Seq2[Plan, error] {
	return iterPages(ctx, func(ctx context.Context) (*Page[Plan], error) {
		return s.List(ctx, params, opts...)
	})
}

// Get retrieves a plan. The active plan ID is available on the subscription
// object.
func (s *PlansService) Get(ctx context.Context, id int, opts ...RequestOption) (*Plan, error) {
	out := &Plan{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/subscription/plan/"+strconv.Itoa(id)+"/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// PlanUpdateParams are the parameters for PlansService.Update. Only the
// automatic refresh schedule is updatable.
type PlanUpdateParams struct {
	// AutomaticRefreshNextAt sets the next automatic proxy list refresh.
	AutomaticRefreshNextAt *time.Time `json:"automatic_refresh_next_at,omitempty"`
}

// Update updates an existing plan. Only AutomaticRefreshNextAt can be
// updated.
func (s *PlansService) Update(ctx context.Context, id int, params PlanUpdateParams, opts ...RequestOption) (*Plan, error) {
	out := &Plan{}
	if err := s.client.doJSON(ctx, http.MethodPatch, "/api/v2/subscription/plan/"+strconv.Itoa(id)+"/", nil, params, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// PlanUpgradeParams are the parameters for PlansService.Upgrade. They match
// the pricing parameters plus a payment method and recaptcha.
type PlanUpgradeParams struct {
	// PlanConfiguration describes the upgraded plan.
	PlanConfiguration
	// PaymentMethod selects the payment method.
	PaymentMethod PaymentMethodRef `json:"payment_method"`
	// Recaptcha is the recaptcha token. Only required when a payment is
	// required.
	Recaptcha string `json:"recaptcha,omitempty"`
}

// Upgrade upgrades an existing plan, crediting the subscription for the
// remainder of the current plan. Recaptcha validation is required when a
// payment is required; with enough account credits no payment is needed.
// When a payment is required the docs mark this endpoint as usable only from
// the Webshare dashboard, not programmatically.
func (s *PlansService) Upgrade(ctx context.Context, id int, params PlanUpgradeParams, opts ...RequestOption) (*CheckoutResult, error) {
	out := &CheckoutResult{}
	path := "/api/v2/subscription/plan/" + strconv.Itoa(id) + "/upgrade/"
	if err := s.client.doJSON(ctx, http.MethodPost, path, nil, params, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// PlanCancelResult is the response of PlansService.Cancel.
type PlanCancelResult struct {
	// Success reports whether the cancellation succeeded.
	Success bool `json:"success"`
	// Transaction is the ID of the credit transaction created by the
	// cancellation.
	Transaction int `json:"transaction"`
}

// Cancel cancels an existing plan, crediting the subscription for the
// duration and bandwidth left in the plan.
func (s *PlansService) Cancel(ctx context.Context, id int, opts ...RequestOption) (*PlanCancelResult, error) {
	out := &PlanCancelResult{}
	path := "/api/v2/subscription/plan/" + strconv.Itoa(id) + "/cancel/"
	if err := s.client.doJSON(ctx, http.MethodPost, path, nil, struct{}{}, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}
