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

// ReplacedProxiesService exposes the replaced proxy operations.
type ReplacedProxiesService struct {
	client *Client
}

// ReplacedProxy records one proxy that was replaced in the proxy list.
type ReplacedProxy struct {
	// ID is the unique identifier of the replaced proxy instance.
	ID int `json:"id"`
	// Reason is why this proxy was replaced.
	Reason ReplacementReason `json:"reason"`
	// Proxy is the IP address of the replaced proxy.
	Proxy string `json:"proxy"`
	// ProxyPort is the port of the replaced proxy.
	ProxyPort int `json:"proxy_port"`
	// ProxyCountryCode is the country code of the replaced proxy.
	ProxyCountryCode string `json:"proxy_country_code"`
	// ReplacedWith is the IP address of the new proxy.
	ReplacedWith string `json:"replaced_with"`
	// ReplacedWithPort is the port of the new proxy.
	ReplacedWithPort int `json:"replaced_with_port"`
	// ReplacedWithCountryCode is the country code of the new proxy.
	ReplacedWithCountryCode string `json:"replaced_with_country_code"`
	// CreatedAt is when this instance was created.
	CreatedAt time.Time `json:"created_at"`
}

// ReplacedProxyListParams are the parameters for ReplacedProxiesService.List.
type ReplacedProxyListParams struct {
	// ProxyListReplacement filters by a specific proxy replacement ID.
	ProxyListReplacement *int
	// PlanID targets a specific plan; otherwise the default plan is used.
	PlanID *int
	// Page is the page number.
	Page *int
	// PageSize is the number of results per page.
	PageSize *int
}

// List returns the replaced proxy list in paginated format.
func (s *ReplacedProxiesService) List(ctx context.Context, params ReplacedProxyListParams, opts ...RequestOption) (*Page[ReplacedProxy], error) {
	q := url.Values{}
	setInt(q, "proxy_list_replacement", params.ProxyListReplacement)
	setInt(q, "plan_id", params.PlanID)
	setInt(q, "page", params.Page)
	setInt(q, "page_size", params.PageSize)
	return getPage[ReplacedProxy](ctx, s.client, "/api/v2/proxy/list/replaced/", q, opts)
}

// ListAll returns a lazy iterator over every replaced proxy across all pages.
func (s *ReplacedProxiesService) ListAll(ctx context.Context, params ReplacedProxyListParams, opts ...RequestOption) iter.Seq2[ReplacedProxy, error] {
	return iterPages(ctx, func(ctx context.Context) (*Page[ReplacedProxy], error) {
		return s.List(ctx, params, opts...)
	})
}

// ReplacedProxyDownloadParams are the parameters for
// ReplacedProxiesService.Download.
type ReplacedProxyDownloadParams struct {
	// DownloadToken is the key obtained from DownloadTokens.Get with the
	// replaced_proxy scope. Required.
	DownloadToken string
	// CountryCodes limits the download to the given ISO 3166-1 alpha-2
	// country codes (validated, upper-cased and hyphen-joined). Empty means
	// all countries.
	CountryCodes []string
	// AuthenticationType is AuthMethodUsername or AuthMethodSourceIP.
	AuthenticationType AuthenticationMethod
	// Mode is ModeDirect or ModeBackbone. Must be ModeBackbone when the
	// plan's pool_filter is residential.
	Mode ConnectionMode
	// Search holds optional search terms.
	Search string
	// ProxyListReplacement filters by a specific proxy replacement ID.
	ProxyListReplacement *int
	// ProxyProtocol fills the proxy protocol query parameter, typically
	// "any".
	ProxyProtocol string
}

// Download fetches the replaced proxy list as plain text; each line is
// new_address:new_port:username:password:replaced_address. This endpoint is
// unauthenticated (it uses the download token from the replaced_proxy
// scope); no Authorization header is sent.
func (s *ReplacedProxiesService) Download(ctx context.Context, params ReplacedProxyDownloadParams, opts ...RequestOption) (string, error) {
	if params.DownloadToken == "" {
		return "", errors.New("webshare: replaced proxy download: download token is required")
	}
	q := url.Values{}
	q.Set("download_token", params.DownloadToken)
	if len(params.CountryCodes) > 0 {
		normalized, err := normalizeCountryCodes(params.CountryCodes)
		if err != nil {
			return "", fmt.Errorf("webshare: replaced proxy download: %w", err)
		}
		q.Set("country_codes", strings.Join(normalized, "-"))
	}
	setString(q, "authentication_type", string(params.AuthenticationType))
	setString(q, "mode", string(params.Mode))
	setString(q, "search", params.Search)
	setInt(q, "proxy_list_replacement", params.ProxyListReplacement)
	setString(q, "proxy_protocol", params.ProxyProtocol)
	return s.client.doText(ctx, http.MethodGet, "/api/v2/proxy/list/replaced/download/", q, withOptions(opts, withAuthMode(authNone)))
}
