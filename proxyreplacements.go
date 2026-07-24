package webshare

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// ProxyReplacementsService exposes the proxy replacement operations (v3
// endpoints).
type ProxyReplacementsService struct {
	client *Client
}

// ReplacementState is the lifecycle state of a proxy replacement.
type ReplacementState string

// Proxy replacement states.
const (
	// ReplacementValidating means the dry run is in progress.
	ReplacementValidating ReplacementState = "validating"
	// ReplacementValidated means the dry run finished.
	ReplacementValidated ReplacementState = "validated"
	// ReplacementProcessing means the actual replacement is in progress.
	ReplacementProcessing ReplacementState = "processing"
	// ReplacementCompleted means the actual replacement finished.
	ReplacementCompleted ReplacementState = "completed"
	// ReplacementFailed means the replacement failed; ErrorCode and Error
	// are set.
	ReplacementFailed ReplacementState = "failed"
)

// ReplacementTargetType discriminates ReplacementTarget values.
type ReplacementTargetType string

// Replacement target types.
const (
	// ReplacementTargetIPRange targets proxies by CIDR ranges.
	ReplacementTargetIPRange ReplacementTargetType = "ip_range"
	// ReplacementTargetIPAddress targets proxies by IP address. Valid only
	// in ToReplace, not in ReplaceWith.
	ReplacementTargetIPAddress ReplacementTargetType = "ip_address"
	// ReplacementTargetASN targets proxies by ASN numbers.
	ReplacementTargetASN ReplacementTargetType = "asn"
	// ReplacementTargetCountry targets proxies by country code.
	ReplacementTargetCountry ReplacementTargetType = "country"
	// ReplacementTargetAny targets any proxies. Valid only in ReplaceWith.
	ReplacementTargetAny ReplacementTargetType = "any"
)

// ASNNumber is an ASN number that tolerates both string and numeric wire
// forms (the docs type it as string while some examples use integers). It
// always serializes as a string.
type ASNNumber string

// UnmarshalJSON implements json.Unmarshaler.
func (n *ASNNumber) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*n = ASNNumber(s)
		return nil
	}
	var i int64
	if err := json.Unmarshal(data, &i); err == nil {
		*n = ASNNumber(strconv.FormatInt(i, 10))
		return nil
	}
	return fmt.Errorf("ASN number: cannot decode %s", data)
}

// ReplacementTarget selects which proxies to replace, or what to replace
// them with, depending on Type. Unused selector fields are omitted.
type ReplacementTarget struct {
	// Type discriminates the target kind. Required.
	Type ReplacementTargetType `json:"type"`
	// IPRanges holds CIDR ranges for the ip_range type. Host bits may not
	// be set (10.0.0.0/24 is valid, 10.0.0.1/24 is not).
	IPRanges []string `json:"ip_ranges,omitempty"`
	// IPAddresses holds IP addresses for the ip_address type.
	IPAddresses []string `json:"ip_addresses,omitempty"`
	// ASNNumbers holds ASN numbers for the asn type.
	ASNNumbers []ASNNumber `json:"asn_numbers,omitempty"`
	// CountryCode holds the country code for the country type.
	CountryCode string `json:"country_code,omitempty"`
	// Count limits how many proxies the target matches.
	Count *int `json:"count,omitempty"`
}

// ReplacementReason explains why proxies were replaced.
type ReplacementReason string

// Replacement reasons.
const (
	// ReasonListUpdated means the proxy list was updated.
	ReasonListUpdated ReplacementReason = "list_updated"
	// ReasonProxyReplaced means a replacement was created manually.
	ReasonProxyReplaced ReplacementReason = "proxy_replaced"
	// ReasonAutoInvalidated means invalid proxies were auto-replaced.
	ReasonAutoInvalidated ReplacementReason = "auto_invalidated"
	// ReasonAutoOutOfRotation means slow proxies were auto-replaced.
	ReasonAutoOutOfRotation ReplacementReason = "auto_out_of_rotation"
	// ReasonAutoLowCountryConfidence means proxies with low country
	// confidence were auto-replaced.
	ReasonAutoLowCountryConfidence ReplacementReason = "auto_low_country_confidence"
	// ReasonAutoDeleted means deleted proxies were auto-replaced.
	ReasonAutoDeleted ReplacementReason = "auto_deleted"
	// ReasonAutoSiteCheck means proxies failing site checks were
	// auto-replaced.
	ReasonAutoSiteCheck ReplacementReason = "auto_site_check"
)

// ProxyReplacement is an asynchronous proxy replacement request.
type ProxyReplacement struct {
	// ID is the unique identifier of the proxy replacement instance.
	ID int `json:"id"`
	// ToReplace indicates which proxies to replace.
	ToReplace ReplacementTarget `json:"to_replace"`
	// ReplaceWith indicates which proxies to replace them with.
	ReplaceWith []ReplacementTarget `json:"replace_with"`
	// DryRun reports whether this replacement only computes the counts
	// without modifying the proxy list.
	DryRun bool `json:"dry_run"`
	// State is the replacement lifecycle state.
	State ReplacementState `json:"state"`
	// ProxiesRemoved is the number of proxies removed from the proxy list.
	ProxiesRemoved *int `json:"proxies_removed"`
	// ProxiesAdded is the number of proxies added to the proxy list.
	ProxiesAdded *int `json:"proxies_added"`
	// Reason is why the proxies were replaced.
	Reason ReplacementReason `json:"reason"`
	// ErrorCode is set when State is ReplacementFailed.
	ErrorCode *string `json:"error_code"`
	// Error is the error message when State is ReplacementFailed.
	Error *string `json:"error"`
	// CreatedAt is when this instance was created.
	CreatedAt time.Time `json:"created_at"`
	// DryRunCompletedAt is when the state became validated. May be nil.
	DryRunCompletedAt *time.Time `json:"dry_run_completed_at"`
	// CompletedAt is when the state became completed. May be nil.
	CompletedAt *time.Time `json:"completed_at"`
}

// ProxyReplacementListParams are the parameters for
// ProxyReplacementsService.List.
type ProxyReplacementListParams struct {
	// PlanID targets a specific plan; otherwise the default plan is used.
	PlanID *int
	// Page is the page number.
	Page *int
	// PageSize is the number of results per page.
	PageSize *int
	// Ordering orders by id (default), created_at or completed_at.
	Ordering string
	// DryRun filters replacements by whether they are dry runs.
	DryRun *bool
	// State filters replacements by state.
	State ReplacementState
}

// List retrieves the proxy replacements in paginated format.
func (s *ProxyReplacementsService) List(ctx context.Context, params ProxyReplacementListParams, opts ...RequestOption) (*Page[ProxyReplacement], error) {
	q := url.Values{}
	setInt(q, "plan_id", params.PlanID)
	setInt(q, "page", params.Page)
	setInt(q, "page_size", params.PageSize)
	setString(q, "ordering", params.Ordering)
	setBool(q, "dry_run", params.DryRun)
	setString(q, "state", string(params.State))
	return getPage[ProxyReplacement](ctx, s.client, "/api/v3/proxy/replace/", q, opts)
}

// ListAll returns a lazy iterator over every proxy replacement across all
// pages.
func (s *ProxyReplacementsService) ListAll(ctx context.Context, params ProxyReplacementListParams, opts ...RequestOption) iter.Seq2[ProxyReplacement, error] {
	return iterPages(ctx, func(ctx context.Context) (*Page[ProxyReplacement], error) {
		return s.List(ctx, params, opts...)
	})
}

// ProxyReplacementCreateParams are the parameters for
// ProxyReplacementsService.Create.
type ProxyReplacementCreateParams struct {
	// PlanID targets a specific plan; otherwise the default plan is used.
	PlanID *int `json:"-"`
	// ToReplace indicates which proxies to replace. Required.
	ToReplace ReplacementTarget `json:"to_replace"`
	// ReplaceWith indicates which proxies to replace them with. The
	// ip_address type cannot be used here. Required.
	ReplaceWith []ReplacementTarget `json:"replace_with"`
	// DryRun computes proxies removed/added without modifying the list.
	DryRun bool `json:"dry_run,omitempty"`
}

// Create starts a proxy replacement. This is an asynchronous API: it returns
// a replacement in the validating state; poll Get until it reaches
// completed (or failed). Not available when the plan's pool_filter is
// residential.
func (s *ProxyReplacementsService) Create(ctx context.Context, params ProxyReplacementCreateParams, opts ...RequestOption) (*ProxyReplacement, error) {
	q := url.Values{}
	setInt(q, "plan_id", params.PlanID)
	out := &ProxyReplacement{}
	if err := s.client.doJSON(ctx, http.MethodPost, "/api/v3/proxy/replace/", q, params, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// ProxyReplacementGetParams are the optional parameters for
// ProxyReplacementsService.Get.
type ProxyReplacementGetParams struct {
	// PlanID targets a specific plan; otherwise the default plan is used.
	PlanID *int
}

// Get retrieves a proxy replacement. Use it to poll the state of a
// replacement created with Create.
func (s *ProxyReplacementsService) Get(ctx context.Context, id int, params ProxyReplacementGetParams, opts ...RequestOption) (*ProxyReplacement, error) {
	q := url.Values{}
	setInt(q, "plan_id", params.PlanID)
	out := &ProxyReplacement{}
	path := "/api/v3/proxy/replace/" + strconv.Itoa(id) + "/"
	if err := s.client.doJSON(ctx, http.MethodGet, path, q, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}
