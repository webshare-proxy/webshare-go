package webshare

import (
	"context"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// IPAuthorizationsService exposes the IP authorization operations.
type IPAuthorizationsService struct {
	client *Client
}

// IPAuthorization authorizes an IP address to use the proxies without
// username/password credentials.
type IPAuthorization struct {
	// ID is the unique identifier of the IP authorization.
	ID int `json:"id"`
	// IPAddress is the authorized IP address.
	IPAddress string `json:"ip_address"`
	// CreatedAt is when this instance was created.
	CreatedAt time.Time `json:"created_at"`
	// LastUsedAt is when this IP address was last used. May be nil.
	LastUsedAt *time.Time `json:"last_used_at"`
}

// IPAuthorizationListParams are the parameters for
// IPAuthorizationsService.List.
type IPAuthorizationListParams struct {
	// PlanID targets a specific plan; otherwise the default plan is used.
	PlanID *int
	// Page is the page number.
	Page *int
	// PageSize is the number of results per page.
	PageSize *int
}

// List returns the IP authorizations in paginated format.
func (s *IPAuthorizationsService) List(ctx context.Context, params IPAuthorizationListParams, opts ...RequestOption) (*Page[IPAuthorization], error) {
	q := url.Values{}
	setInt(q, "plan_id", params.PlanID)
	setInt(q, "page", params.Page)
	setInt(q, "page_size", params.PageSize)
	return getPage[IPAuthorization](ctx, s.client, "/api/v2/proxy/ipauthorization/", q, opts)
}

// ListAll returns a lazy iterator over every IP authorization across all
// pages.
func (s *IPAuthorizationsService) ListAll(ctx context.Context, params IPAuthorizationListParams, opts ...RequestOption) iter.Seq2[IPAuthorization, error] {
	return iterPages(ctx, func(ctx context.Context) (*Page[IPAuthorization], error) {
		return s.List(ctx, params, opts...)
	})
}

// IPAuthorizationCreateParams are the parameters for
// IPAuthorizationsService.Create.
type IPAuthorizationCreateParams struct {
	// PlanID targets a specific plan; otherwise the default plan is used.
	PlanID *int `json:"-"`
	// IPAddress is the IP address to authorize. Required. The API returns a
	// 400 error when the address is already authorized in the system.
	IPAddress string `json:"ip_address"`
}

// Create authorizes an IP address.
func (s *IPAuthorizationsService) Create(ctx context.Context, params IPAuthorizationCreateParams, opts ...RequestOption) (*IPAuthorization, error) {
	q := url.Values{}
	setInt(q, "plan_id", params.PlanID)
	out := &IPAuthorization{}
	if err := s.client.doJSON(ctx, http.MethodPost, "/api/v2/proxy/ipauthorization/", q, params, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// IPAuthorizationGetParams are the optional parameters for
// IPAuthorizationsService.Get and Delete.
type IPAuthorizationGetParams struct {
	// PlanID targets a specific plan; otherwise the default plan is used.
	PlanID *int
}

// Get retrieves an IP authorization.
func (s *IPAuthorizationsService) Get(ctx context.Context, id int, params IPAuthorizationGetParams, opts ...RequestOption) (*IPAuthorization, error) {
	q := url.Values{}
	setInt(q, "plan_id", params.PlanID)
	out := &IPAuthorization{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/proxy/ipauthorization/"+strconv.Itoa(id)+"/", q, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete removes an IP authorization.
func (s *IPAuthorizationsService) Delete(ctx context.Context, id int, params IPAuthorizationGetParams, opts ...RequestOption) error {
	q := url.Values{}
	setInt(q, "plan_id", params.PlanID)
	return s.client.doJSON(ctx, http.MethodDelete, "/api/v2/proxy/ipauthorization/"+strconv.Itoa(id)+"/", q, nil, nil, opts)
}

// WhatsMyIPResult is the response of IPAuthorizationsService.WhatsMyIP.
type WhatsMyIPResult struct {
	// IPAddress is the caller's public IP address.
	IPAddress string `json:"ip_address"`
}

// WhatsMyIP returns the caller's public IP address, useful before creating
// an IP authorization.
func (s *IPAuthorizationsService) WhatsMyIP(ctx context.Context, opts ...RequestOption) (*WhatsMyIPResult, error) {
	out := &WhatsMyIPResult{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/proxy/ipauthorization/whatsmyip/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}
