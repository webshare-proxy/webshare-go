package webshare

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ProxiesService exposes the proxy list operations.
type ProxiesService struct {
	client *Client
}

// Proxy is one entry of the proxy list.
type Proxy struct {
	// ID is the unique identifier of the proxy instance. Unlike other
	// resources, proxy IDs are strings (for example "d-10513").
	ID string `json:"id"`
	// Username is the proxy username.
	Username string `json:"username"`
	// Password is the proxy password.
	Password string `json:"password"`
	// ProxyAddress is the IP address of the proxy. In direct connection mode
	// connect to this address; in backbone mode connect to p.webshare.io.
	// Nil when the plan's pool_filter is residential.
	ProxyAddress *string `json:"proxy_address"`
	// Port is the port used to connect to the proxy. In backbone mode the
	// port is always set for IP authorization.
	Port int `json:"port"`
	// Valid reports whether the proxy is working as expected. Proxies are
	// checked once every 30 seconds.
	Valid bool `json:"valid"`
	// LastVerification is the last time the proxy was checked.
	LastVerification time.Time `json:"last_verification"`
	// CountryCode is the ISO 3166-1 alpha-2 country code of the proxy.
	CountryCode string `json:"country_code"`
	// CityName is the city name of the proxy.
	CityName string `json:"city_name"`
	// CreatedAt is when this instance was created.
	CreatedAt time.Time `json:"created_at"`
}

// ProxyListParams are the parameters for ProxiesService.List.
type ProxyListParams struct {
	// Mode must be ModeDirect or ModeBackbone. Required. Must be
	// ModeBackbone when the plan's pool_filter is residential.
	Mode ConnectionMode
	// PlanID targets a specific plan; otherwise the default plan is used.
	PlanID *int
	// Page is the page number.
	Page *int
	// PageSize is the number of results per page.
	PageSize *int
	// CountryCodeIn filters by ISO 3166-1 alpha-2 country codes.
	CountryCodeIn []string
	// Search filters by a search phrase. Does not work in backbone mode.
	Search string
	// Ordering is a comma separated list of ordering fields; prefix a field
	// with "-" for descending order. Not supported in backbone mode.
	Ordering string
	// CreatedAt filters by proxy create date. Does not work in backbone mode.
	CreatedAt string
	// ProxyAddress filters by a specific proxy address. Does not work in
	// backbone mode.
	ProxyAddress string
	// ProxyAddressIn filters by proxy addresses. Does not work in backbone
	// mode.
	ProxyAddressIn []string
	// Valid filters by proxy validity. Does not work in backbone mode.
	Valid *bool
	// ASNNumber filters by the proxy ASN number. Does not work in backbone
	// mode.
	ASNNumber string
	// ASNName filters by the proxy ASN name. Does not work in backbone mode.
	ASNName string
}

func (p ProxyListParams) values() url.Values {
	q := url.Values{}
	setString(q, "mode", string(p.Mode))
	setInt(q, "plan_id", p.PlanID)
	setInt(q, "page", p.Page)
	setInt(q, "page_size", p.PageSize)
	setStringList(q, "country_code__in", p.CountryCodeIn)
	setString(q, "search", p.Search)
	setString(q, "ordering", p.Ordering)
	setString(q, "created_at", p.CreatedAt)
	setString(q, "proxy_address", p.ProxyAddress)
	setStringList(q, "proxy_address__in", p.ProxyAddressIn)
	setBool(q, "valid", p.Valid)
	setString(q, "asn_number", p.ASNNumber)
	setString(q, "asn_name", p.ASNName)
	return q
}

// List returns the proxy list in paginated format. The Mode parameter is
// required.
func (s *ProxiesService) List(ctx context.Context, params ProxyListParams, opts ...RequestOption) (*Page[Proxy], error) {
	return getPage[Proxy](ctx, s.client, "/api/v2/proxy/list/", params.values(), opts)
}

// ListAll returns a lazy iterator over every proxy across all pages.
func (s *ProxiesService) ListAll(ctx context.Context, params ProxyListParams, opts ...RequestOption) iter.Seq2[Proxy, error] {
	return iterPages(ctx, func(ctx context.Context) (*Page[Proxy], error) {
		return s.List(ctx, params, opts...)
	})
}

// ProxyRefreshParams are the parameters for ProxiesService.Refresh.
type ProxyRefreshParams struct {
	// PlanID targets a specific plan; otherwise the default plan is used.
	PlanID *int
}

// Refresh replaces the entire proxy list on demand. Only available when the
// plan has on_demand_refreshes_available.
func (s *ProxiesService) Refresh(ctx context.Context, params ProxyRefreshParams, opts ...RequestOption) error {
	q := url.Values{}
	setInt(q, "plan_id", params.PlanID)
	return s.client.doJSON(ctx, http.MethodPost, "/api/v2/proxy/list/refresh/", q, nil, nil, opts)
}

// AuthenticationMethod selects how downloaded proxies are authenticated.
type AuthenticationMethod string

// Authentication methods for proxy list downloads.
const (
	// AuthMethodUsername authenticates with username and password.
	AuthMethodUsername AuthenticationMethod = "username"
	// AuthMethodSourceIP authenticates by source IP address.
	AuthMethodSourceIP AuthenticationMethod = "sourceip"
)

// ProxyDownloadParams are the parameters for ProxiesService.Download and
// ProxiesService.DownloadURL.
type ProxyDownloadParams struct {
	// Token is the proxy_list_download_token from the proxy config API.
	// Required.
	Token string
	// CountryCodes limits the download to the given ISO 3166-1 alpha-2
	// country codes (hyphen-joined in the URL). Empty means all countries.
	CountryCodes []string
	// Protocol fills the literal proxy protocol slot of the download path.
	// Defaults to "any".
	Protocol string
	// AuthenticationMethod is AuthMethodUsername or AuthMethodSourceIP.
	// Required.
	AuthenticationMethod AuthenticationMethod
	// EndpointMode is ModeDirect or ModeBackbone. Required. Must be
	// ModeBackbone when the plan's pool_filter is residential.
	EndpointMode ConnectionMode
	// Search holds optional search terms. Empty means no search terms.
	Search string
	// PlanID targets a specific plan; otherwise the default plan is used.
	PlanID *int
}

func (p ProxyDownloadParams) path() (string, error) {
	if p.Token == "" {
		return "", errors.New("webshare: proxy download: token is required")
	}
	switch p.AuthenticationMethod {
	case AuthMethodUsername, AuthMethodSourceIP:
	case "":
		return "", errors.New("webshare: proxy download: authentication method is required (username or sourceip)")
	default:
		return "", fmt.Errorf("webshare: proxy download: invalid authentication method %q", p.AuthenticationMethod)
	}
	switch p.EndpointMode {
	case ModeDirect, ModeBackbone:
	case "":
		return "", errors.New("webshare: proxy download: endpoint mode is required (direct or backbone)")
	default:
		return "", fmt.Errorf("webshare: proxy download: invalid endpoint mode %q", p.EndpointMode)
	}
	countries := "-"
	if len(p.CountryCodes) > 0 {
		countries = strings.Join(p.CountryCodes, "-")
	}
	protocol := p.Protocol
	if protocol == "" {
		protocol = "any"
	}
	search := "-"
	if p.Search != "" {
		search = p.Search
	}
	return "/api/v2/proxy/list/download/" +
		url.PathEscape(p.Token) + "/" +
		url.PathEscape(countries) + "/" +
		url.PathEscape(protocol) + "/" +
		string(p.AuthenticationMethod) + "/" +
		string(p.EndpointMode) + "/" +
		url.PathEscape(search) + "/", nil
}

// Download fetches the proxy list as plain text, one proxy per line in
// address:port:username:password format. This endpoint is unauthenticated
// (the URL embeds the download token); no Authorization header is sent.
func (s *ProxiesService) Download(ctx context.Context, params ProxyDownloadParams, opts ...RequestOption) (string, error) {
	path, err := params.path()
	if err != nil {
		return "", err
	}
	q := url.Values{}
	setInt(q, "plan_id", params.PlanID)
	return s.client.doText(ctx, http.MethodGet, path, q, withOptions(opts, withAuthMode(authNone)))
}

// DownloadURL builds the shareable path-style proxy list download URL without
// performing a request.
func (s *ProxiesService) DownloadURL(params ProxyDownloadParams) (string, error) {
	path, err := params.path()
	if err != nil {
		return "", err
	}
	q := url.Values{}
	setInt(q, "plan_id", params.PlanID)
	cfg := s.client.cfg
	return cfg.resolve(path, q), nil
}
