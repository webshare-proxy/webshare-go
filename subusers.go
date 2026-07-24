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

// SubusersService exposes the sub-user operations. The sub-user API is only
// available after accepting additional terms for the Webshare sub-user
// portal (https://proxy.webshare.io/subuser/); expect 403 responses
// otherwise. To act as a sub-user on the proxy config, list, stats and
// activity APIs, pass WithSubuser.
type SubusersService struct {
	client *Client
}

// Subuser is one sub-user of the account.
type Subuser struct {
	// ID is the unique identifier of the sub-user. It never changes and is
	// never reused.
	ID int `json:"id"`
	// Label identifies the sub-user.
	Label string `json:"label"`
	// ProxyCountries maps country code to proxy count for the sub-user's
	// custom proxy list. The special code ZZ means any available country.
	// Nil when custom proxy lists are disabled.
	ProxyCountries map[string]int `json:"proxy_countries"`
	// ProxyLimit is the sub-user bandwidth limit in GB. Zero means
	// unlimited.
	ProxyLimit float64 `json:"proxy_limit"`
	// MaxThreadCount is the maximum proxy request concurrency for the
	// sub-user.
	MaxThreadCount int `json:"max_thread_count"`
	// AggregateStats holds the proxy stats for the sub-user in the same
	// shape as the aggregate stats API.
	AggregateStats AggregateStats `json:"aggregate_stats"`
	// CreatedAt is when the sub-user was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when the sub-user was last updated.
	UpdatedAt time.Time `json:"updated_at"`
	// BandwidthUseStartDate is when bandwidth use calculation starts.
	BandwidthUseStartDate time.Time `json:"bandwidth_use_start_date"`
	// BandwidthUseEndDate is when the bandwidth use resets. Read-only.
	BandwidthUseEndDate time.Time `json:"bandwidth_use_end_date"`
}

// SubuserListParams are the parameters for SubusersService.List.
type SubuserListParams struct {
	// Page is the page number.
	Page *int
	// PageSize is the number of results per page.
	PageSize *int
	// PlanID targets a specific plan; otherwise the default plan is used.
	PlanID *int
}

// List retrieves all sub-users in paginated format.
func (s *SubusersService) List(ctx context.Context, params SubuserListParams, opts ...RequestOption) (*Page[Subuser], error) {
	q := url.Values{}
	setInt(q, "page", params.Page)
	setInt(q, "page_size", params.PageSize)
	setInt(q, "plan_id", params.PlanID)
	return getPage[Subuser](ctx, s.client, "/api/v2/subuser/", q, opts)
}

// ListAll returns a lazy iterator over every sub-user across all pages.
func (s *SubusersService) ListAll(ctx context.Context, params SubuserListParams, opts ...RequestOption) iter.Seq2[Subuser, error] {
	return iterPages(ctx, func(ctx context.Context) (*Page[Subuser], error) {
		return s.List(ctx, params, opts...)
	})
}

// SubuserCreateParams are the parameters for SubusersService.Create.
type SubuserCreateParams struct {
	// PlanID targets a specific plan; otherwise the default plan is used.
	PlanID *int
	// Label identifies the sub-user.
	Label string
	// ProxyLimit is the bandwidth limit in GB. Zero means unlimited.
	ProxyLimit *float64
	// MaxThreadCount is the maximum proxy request concurrency.
	MaxThreadCount *int
	// ProxyCountries maps country code to proxy count for a custom proxy
	// list; ZZ means any available country. Null disables custom lists.
	ProxyCountries Nullable[map[string]int]
	// BandwidthUseStartDate sets when bandwidth use calculation starts.
	BandwidthUseStartDate *time.Time
}

// MarshalJSON implements json.Marshaler, sending only the set fields.
func (p SubuserCreateParams) MarshalJSON() ([]byte, error) {
	body := map[string]any{}
	if p.Label != "" {
		body["label"] = p.Label
	}
	if p.ProxyLimit != nil {
		body["proxy_limit"] = *p.ProxyLimit
	}
	if p.MaxThreadCount != nil {
		body["max_thread_count"] = *p.MaxThreadCount
	}
	if p.ProxyCountries.isPresent() {
		body["proxy_countries"] = p.ProxyCountries
	}
	if p.BandwidthUseStartDate != nil {
		body["bandwidth_use_start_date"] = p.BandwidthUseStartDate
	}
	return json.Marshal(body)
}

// Create creates a new sub-user.
func (s *SubusersService) Create(ctx context.Context, params SubuserCreateParams, opts ...RequestOption) (*Subuser, error) {
	q := url.Values{}
	setInt(q, "plan_id", params.PlanID)
	out := &Subuser{}
	if err := s.client.doJSON(ctx, http.MethodPost, "/api/v2/subuser/", q, params, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// SubuserGetParams are the optional parameters for SubusersService.Get and
// Delete.
type SubuserGetParams struct {
	// PlanID targets a specific plan; otherwise the default plan is used.
	PlanID *int
}

// Get retrieves a sub-user.
func (s *SubusersService) Get(ctx context.Context, id int, params SubuserGetParams, opts ...RequestOption) (*Subuser, error) {
	q := url.Values{}
	setInt(q, "plan_id", params.PlanID)
	out := &Subuser{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/subuser/"+strconv.Itoa(id)+"/", q, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// SubuserUpdateParams are the parameters for SubusersService.Update. Only
// set fields are sent.
type SubuserUpdateParams struct {
	// PlanID targets a specific plan; otherwise the default plan is used.
	PlanID *int
	// Label sets the sub-user label.
	Label *string
	// ProxyCountries maps country code to proxy count for a custom proxy
	// list; ZZ means any available country. Null disables custom lists.
	ProxyCountries Nullable[map[string]int]
	// ProxyLimit sets the bandwidth limit in GB. Zero means unlimited.
	ProxyLimit *float64
	// MaxThreadCount sets the maximum proxy request concurrency.
	MaxThreadCount *int
	// BandwidthUseStartDate sets when bandwidth use calculation starts.
	BandwidthUseStartDate *time.Time
}

// MarshalJSON implements json.Marshaler, sending only the set fields.
func (p SubuserUpdateParams) MarshalJSON() ([]byte, error) {
	body := map[string]any{}
	if p.Label != nil {
		body["label"] = *p.Label
	}
	if p.ProxyCountries.isPresent() {
		body["proxy_countries"] = p.ProxyCountries
	}
	if p.ProxyLimit != nil {
		body["proxy_limit"] = *p.ProxyLimit
	}
	if p.MaxThreadCount != nil {
		body["max_thread_count"] = *p.MaxThreadCount
	}
	if p.BandwidthUseStartDate != nil {
		body["bandwidth_use_start_date"] = p.BandwidthUseStartDate
	}
	return json.Marshal(body)
}

// Update partially updates a sub-user.
func (s *SubusersService) Update(ctx context.Context, id int, params SubuserUpdateParams, opts ...RequestOption) (*Subuser, error) {
	q := url.Values{}
	setInt(q, "plan_id", params.PlanID)
	out := &Subuser{}
	if err := s.client.doJSON(ctx, http.MethodPatch, "/api/v2/subuser/"+strconv.Itoa(id)+"/", q, params, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete removes a sub-user.
func (s *SubusersService) Delete(ctx context.Context, id int, params SubuserGetParams, opts ...RequestOption) error {
	q := url.Values{}
	setInt(q, "plan_id", params.PlanID)
	return s.client.doJSON(ctx, http.MethodDelete, "/api/v2/subuser/"+strconv.Itoa(id)+"/", q, nil, nil, opts)
}

// RefreshProxyList refreshes the proxy list of a sub-user. Only available
// when the sub-user has a custom proxy list.
func (s *SubusersService) RefreshProxyList(ctx context.Context, id int, opts ...RequestOption) (*Subuser, error) {
	out := &Subuser{}
	if err := s.client.doJSON(ctx, http.MethodPost, "/api/v2/subuser/"+strconv.Itoa(id)+"/refresh/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}
