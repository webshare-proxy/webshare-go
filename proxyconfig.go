package webshare

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

// ProxyConfigService exposes the proxy configuration operations.
type ProxyConfigService struct {
	client *Client
}

// ProxyConfigState indicates whether a proxy list is ready to use.
type ProxyConfigState string

// Proxy configuration states.
const (
	// ProxyConfigPending means the proxy list is not ready yet.
	ProxyConfigPending ProxyConfigState = "pending"
	// ProxyConfigProcessing means the proxy list is being prepared.
	ProxyConfigProcessing ProxyConfigState = "processing"
	// ProxyConfigCompleted means the proxy list is ready to use.
	ProxyConfigCompleted ProxyConfigState = "completed"
)

// ProxyConfig is the full proxy configuration object returned by the v2
// update and allocate endpoints.
type ProxyConfig struct {
	// ID is the unique identifier of the proxy configuration instance.
	ID int `json:"id"`
	// State indicates whether the proxy list is ready to use.
	State ProxyConfigState `json:"state"`
	// Countries maps country code to proxy count in the proxy list.
	Countries map[string]int `json:"countries"`
	// AvailableCountries maps country code to proxy count available outside
	// the current proxy list.
	AvailableCountries map[string]int `json:"available_countries"`
	// UnallocatedCountries maps country code to unallocated proxy count.
	UnallocatedCountries map[string]int `json:"unallocated_countries"`
	// IPRanges24 maps /24 CIDR to proxy count in the proxy list. Empty for
	// residential plans.
	IPRanges24 map[string]int `json:"ip_ranges_24"`
	// IPRanges16 maps /16 CIDR to proxy count in the proxy list.
	IPRanges16 map[string]int `json:"ip_ranges_16"`
	// IPRanges8 maps /8 CIDR to proxy count in the proxy list.
	IPRanges8 map[string]int `json:"ip_ranges_8"`
	// AvailableIPRanges24 maps /24 CIDR to available proxy count.
	AvailableIPRanges24 map[string]int `json:"available_ip_ranges_24"`
	// AvailableIPRanges16 maps /16 CIDR to available proxy count.
	AvailableIPRanges16 map[string]int `json:"available_ip_ranges_16"`
	// AvailableIPRanges8 maps /8 CIDR to available proxy count.
	AvailableIPRanges8 map[string]int `json:"available_ip_ranges_8"`
	// ASNs maps ASN number to ASN name and count. Empty for residential
	// plans.
	ASNs map[string]ASNInfo `json:"asns"`
	// AvailableASNs maps ASN number to ASN name and count for proxies
	// available outside the current list.
	AvailableASNs map[string]ASNInfo `json:"available_asns"`
	// Username is the proxy username (8-32 characters, alphanumeric).
	Username string `json:"username"`
	// Password is the proxy password (8-32 characters, alphanumeric).
	Password string `json:"password"`
	// RequestTimeout is the maximum number of seconds a proxy request can be
	// used.
	RequestTimeout int `json:"request_timeout"`
	// RequestIdleTimeout is the maximum number of seconds a proxy request
	// can stay idle.
	RequestIdleTimeout int `json:"request_idle_timeout"`
	// IPAuthorizationCountryCodes lists country codes served for IP
	// authorization in backbone mode. Nil means all countries.
	IPAuthorizationCountryCodes []string `json:"ip_authorization_country_codes"`
	// IPAuthorizationCity is the city for IP authorization geo targeting.
	// Residential plans only. Nil when disabled.
	IPAuthorizationCity *string `json:"ip_authorization_city"`
	// IPAuthorizationState is the state for IP authorization geo targeting.
	// Observed on the live API; only mentioned in passing in the docs. Nil
	// when disabled.
	IPAuthorizationState *string `json:"ip_authorization_state"`
	// IPAuthorizationPostalCode is the postal code for IP authorization geo
	// targeting. Observed on the live API; only mentioned in passing in the
	// docs. Nil when disabled.
	IPAuthorizationPostalCode *string `json:"ip_authorization_postalcode"`
	// IPAuthorizationASN is the ASN targeted for IP authorization requests.
	// Nil when disabled.
	IPAuthorizationASN *string `json:"ip_authorization_asn"`
	// AutoReplaceInvalidProxies replaces proxies invalid for 15 minutes.
	AutoReplaceInvalidProxies bool `json:"auto_replace_invalid_proxies"`
	// AutoReplaceLowCountryConfidenceProxies replaces proxies with low
	// country confidence.
	AutoReplaceLowCountryConfidenceProxies bool `json:"auto_replace_low_country_confidence_proxies"`
	// AutoReplaceOutOfRotationProxies replaces proxies performing slower
	// than usual.
	AutoReplaceOutOfRotationProxies bool `json:"auto_replace_out_of_rotation_proxies"`
	// AutoReplaceFailedSiteCheckProxies replaces proxies that no longer pass
	// site checks.
	AutoReplaceFailedSiteCheckProxies bool `json:"auto_replace_failed_site_check_proxies"`
	// ProxyListDownloadToken is the token used in proxy list download links.
	ProxyListDownloadToken string `json:"proxy_list_download_token"`
	// IsProxyUsed indicates whether a proxy has been used.
	IsProxyUsed bool `json:"is_proxy_used"`
	// CreatedAt is when this instance was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when this instance was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// ProxyConfigV3 is the editable/config subset returned by the v3 get
// endpoint. Counts, ranges and status live on GetStats and GetStatus.
type ProxyConfigV3 struct {
	// RequestTimeout is the maximum number of seconds a proxy request can be
	// used (min 15 seconds, max 7 days).
	RequestTimeout int `json:"request_timeout"`
	// RequestIdleTimeout is the maximum number of seconds a proxy request
	// can stay idle (min 15 seconds, max 2 hours).
	RequestIdleTimeout int `json:"request_idle_timeout"`
	// IPAuthorizationCountryCodes lists country codes served for IP
	// authorization in backbone mode. Nil means all countries.
	IPAuthorizationCountryCodes []string `json:"ip_authorization_country_codes"`
	// IPAuthorizationCity is the city for IP authorization geo targeting.
	IPAuthorizationCity *string `json:"ip_authorization_city"`
	// IPAuthorizationState is the state for IP authorization geo targeting.
	// Observed on the live API; only mentioned in passing in the docs.
	IPAuthorizationState *string `json:"ip_authorization_state"`
	// IPAuthorizationPostalCode is the postal code for IP authorization geo
	// targeting. Observed on the live API; only mentioned in passing in the
	// docs.
	IPAuthorizationPostalCode *string `json:"ip_authorization_postalcode"`
	// IPAuthorizationASN is the ASN targeted for IP authorization requests.
	IPAuthorizationASN *string `json:"ip_authorization_asn"`
	// AutoReplaceInvalidProxies replaces proxies invalid for 15 minutes.
	AutoReplaceInvalidProxies bool `json:"auto_replace_invalid_proxies"`
	// AutoReplaceLowCountryConfidenceProxies replaces proxies with low
	// country confidence.
	AutoReplaceLowCountryConfidenceProxies bool `json:"auto_replace_low_country_confidence_proxies"`
	// AutoReplaceOutOfRotationProxies replaces proxies performing slower
	// than usual.
	AutoReplaceOutOfRotationProxies bool `json:"auto_replace_out_of_rotation_proxies"`
	// AutoReplaceFailedSiteCheckProxies replaces proxies that no longer pass
	// site checks.
	AutoReplaceFailedSiteCheckProxies bool `json:"auto_replace_failed_site_check_proxies"`
	// ProxyListDownloadToken is the token used in proxy list download links.
	ProxyListDownloadToken string `json:"proxy_list_download_token"`
}

// Get retrieves the proxy config (v3 endpoint). The plan ID is required.
func (s *ProxyConfigService) Get(ctx context.Context, planID int, opts ...RequestOption) (*ProxyConfigV3, error) {
	q := url.Values{}
	setInt(q, "plan_id", &planID)
	out := &ProxyConfigV3{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v3/proxy/config", q, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// ProxyListStats is the proxy list composition returned by the v3 stats
// endpoint: available countries, IP ranges and ASNs of the current list.
type ProxyListStats struct {
	// AvailableCountries maps country code to available proxy count.
	AvailableCountries map[string]int `json:"available_countries"`
	// IPRanges24 maps /24 CIDR to proxy count in the proxy list.
	IPRanges24 map[string]int `json:"ip_ranges_24"`
	// IPRanges16 maps /16 CIDR to proxy count in the proxy list.
	IPRanges16 map[string]int `json:"ip_ranges_16"`
	// IPRanges8 maps /8 CIDR to proxy count in the proxy list.
	IPRanges8 map[string]int `json:"ip_ranges_8"`
	// AvailableIPRanges24 maps /24 CIDR to available proxy count.
	AvailableIPRanges24 map[string]int `json:"available_ip_ranges_24"`
	// AvailableIPRanges16 maps /16 CIDR to available proxy count.
	AvailableIPRanges16 map[string]int `json:"available_ip_ranges_16"`
	// AvailableIPRanges8 maps /8 CIDR to available proxy count.
	AvailableIPRanges8 map[string]int `json:"available_ip_ranges_8"`
	// ASNs maps ASN number to ASN name and count.
	ASNs map[string]ASNInfo `json:"asns"`
	// AvailableASNs maps ASN number to ASN name and count.
	AvailableASNs map[string]ASNInfo `json:"available_asns"`
}

// GetStats retrieves the proxy list composition (v3 endpoint). The plan ID
// is required. This is distinct from the usage statistics under Stats.
func (s *ProxyConfigService) GetStats(ctx context.Context, planID int, opts ...RequestOption) (*ProxyListStats, error) {
	q := url.Values{}
	setInt(q, "plan_id", &planID)
	out := &ProxyListStats{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v3/proxy/list/stats", q, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// ProxyListStatus is the proxy list readiness state returned by the v3
// status endpoint.
type ProxyListStatus struct {
	// State indicates whether the proxy list is ready to use.
	State ProxyConfigState `json:"state"`
	// Countries maps country code to proxy count in the proxy list.
	Countries map[string]int `json:"countries"`
	// UnallocatedCountries maps country code to unallocated proxy count.
	UnallocatedCountries map[string]int `json:"unallocated_countries"`
	// Username is the proxy username.
	Username string `json:"username"`
	// Password is the proxy password.
	Password string `json:"password"`
	// IsProxyUsed indicates whether a proxy has been used.
	IsProxyUsed bool `json:"is_proxy_used"`
}

// GetStatus retrieves the proxy status (v3 endpoint). The plan ID is
// required.
func (s *ProxyConfigService) GetStatus(ctx context.Context, planID int, opts ...RequestOption) (*ProxyListStatus, error) {
	q := url.Values{}
	setInt(q, "plan_id", &planID)
	out := &ProxyListStatus{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v3/proxy/list/status", q, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// ProxyConfigUpdateParams are the parameters for ProxyConfigService.Update.
// Only set fields are sent. The nullable geo-targeting fields accept an
// explicit null via webshare.Null to disable them.
type ProxyConfigUpdateParams struct {
	// PlanID targets a specific plan; otherwise the default plan is used.
	PlanID *int
	// Username sets the proxy username (8-32 characters, alphanumeric).
	Username *string
	// Password sets the proxy password (8-32 characters, alphanumeric).
	Password *string
	// RequestTimeout sets the maximum seconds a proxy request can be used.
	RequestTimeout *int
	// RequestIdleTimeout sets the maximum seconds a proxy request can idle.
	RequestIdleTimeout *int
	// IPAuthorizationCountryCodes sets the country codes served for IP
	// authorization in backbone mode. Null means all countries.
	IPAuthorizationCountryCodes Nullable[[]string]
	// IPAuthorizationCity sets the city for IP authorization geo targeting
	// (residential plans only). Null disables.
	IPAuthorizationCity Nullable[string]
	// IPAuthorizationState sets the state for IP authorization geo
	// targeting. Mutually exclusive with the other geo filters. Null
	// disables.
	IPAuthorizationState Nullable[string]
	// IPAuthorizationPostalCode sets the postal code for IP authorization
	// geo targeting. Mutually exclusive with the other geo filters. Null
	// disables.
	IPAuthorizationPostalCode Nullable[string]
	// IPAuthorizationASN sets the ASN for IP authorization targeting.
	// Mutually exclusive with the other geo filters. Null disables.
	IPAuthorizationASN Nullable[string]
	// AutoReplaceInvalidProxies toggles replacing invalid proxies. Cannot be
	// edited for free plans.
	AutoReplaceInvalidProxies *bool
	// AutoReplaceLowCountryConfidenceProxies toggles replacing proxies with
	// low country confidence. Cannot be edited for free plans.
	AutoReplaceLowCountryConfidenceProxies *bool
	// AutoReplaceOutOfRotationProxies toggles replacing slow proxies.
	AutoReplaceOutOfRotationProxies *bool
	// AutoReplaceFailedSiteCheckProxies toggles replacing proxies failing
	// site checks.
	AutoReplaceFailedSiteCheckProxies *bool
}

// MarshalJSON implements json.Marshaler, sending only the set fields.
func (p ProxyConfigUpdateParams) MarshalJSON() ([]byte, error) {
	body := map[string]any{}
	if p.Username != nil {
		body["username"] = *p.Username
	}
	if p.Password != nil {
		body["password"] = *p.Password
	}
	if p.RequestTimeout != nil {
		body["request_timeout"] = *p.RequestTimeout
	}
	if p.RequestIdleTimeout != nil {
		body["request_idle_timeout"] = *p.RequestIdleTimeout
	}
	if p.IPAuthorizationCountryCodes.isPresent() {
		body["ip_authorization_country_codes"] = p.IPAuthorizationCountryCodes
	}
	if p.IPAuthorizationCity.isPresent() {
		body["ip_authorization_city"] = p.IPAuthorizationCity
	}
	if p.IPAuthorizationState.isPresent() {
		body["ip_authorization_state"] = p.IPAuthorizationState
	}
	if p.IPAuthorizationPostalCode.isPresent() {
		body["ip_authorization_postalcode"] = p.IPAuthorizationPostalCode
	}
	if p.IPAuthorizationASN.isPresent() {
		body["ip_authorization_asn"] = p.IPAuthorizationASN
	}
	if p.AutoReplaceInvalidProxies != nil {
		body["auto_replace_invalid_proxies"] = *p.AutoReplaceInvalidProxies
	}
	if p.AutoReplaceLowCountryConfidenceProxies != nil {
		body["auto_replace_low_country_confidence_proxies"] = *p.AutoReplaceLowCountryConfidenceProxies
	}
	if p.AutoReplaceOutOfRotationProxies != nil {
		body["auto_replace_out_of_rotation_proxies"] = *p.AutoReplaceOutOfRotationProxies
	}
	if p.AutoReplaceFailedSiteCheckProxies != nil {
		body["auto_replace_failed_site_check_proxies"] = *p.AutoReplaceFailedSiteCheckProxies
	}
	return json.Marshal(body)
}

// Update partially updates the proxy config (v2 endpoint) and returns the
// full proxy config object.
func (s *ProxyConfigService) Update(ctx context.Context, params ProxyConfigUpdateParams, opts ...RequestOption) (*ProxyConfig, error) {
	q := url.Values{}
	setInt(q, "plan_id", params.PlanID)
	out := &ProxyConfig{}
	if err := s.client.doJSON(ctx, http.MethodPatch, "/api/v2/proxy/config/", q, params, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// AllocateUnallocatedCountriesParams are the parameters for
// ProxyConfigService.AllocateUnallocatedCountries.
type AllocateUnallocatedCountriesParams struct {
	// PlanID targets a specific plan; otherwise the default plan is used.
	PlanID *int `json:"-"`
	// NewCountries maps upper-case country code to the number of proxies to
	// allocate. The total must exactly match the number of unallocated
	// proxies. Required.
	NewCountries map[string]int `json:"new_countries"`
}

// AllocateUnallocatedCountries allocates the proxies in unallocated_countries
// state and returns the full proxy config object.
func (s *ProxyConfigService) AllocateUnallocatedCountries(ctx context.Context, params AllocateUnallocatedCountriesParams, opts ...RequestOption) (*ProxyConfig, error) {
	q := url.Values{}
	setInt(q, "plan_id", params.PlanID)
	out := &ProxyConfig{}
	if err := s.client.doJSON(ctx, http.MethodPost, "/api/v2/proxy/config/allocate_unallocated_countries/", q, params, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}
