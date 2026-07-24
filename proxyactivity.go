package webshare

import (
	"context"
	"errors"
	"iter"
	"net/http"
	"net/url"
	"time"
)

// ProxyActivityService exposes the proxy activity operations.
type ProxyActivityService struct {
	client *Client
}

// ProxyActivity is one authenticated proxy request record.
type ProxyActivity struct {
	// Timestamp is when the proxy request was made.
	Timestamp time.Time `json:"timestamp"`
	// Protocol is the proxy protocol: "http" or "socks".
	Protocol string `json:"protocol"`
	// RequestDuration is the total proxy request duration in seconds.
	RequestDuration float64 `json:"request_duration"`
	// HandshakeDuration is the seconds spent authenticating and
	// establishing the proxy connection.
	HandshakeDuration float64 `json:"handshake_duration"`
	// TunnelDuration is the seconds the connection stayed active after the
	// handshake. May be nil.
	TunnelDuration *float64 `json:"tunnel_duration"`
	// ErrorReason is the error reason for the proxy request. May be nil.
	ErrorReason *string `json:"error_reason"`
	// ErrorReasonHowToFix is a user-friendly explanation of how to fix the
	// error. May be nil.
	ErrorReasonHowToFix *string `json:"error_reason_how_to_fix"`
	// AuthUsername is the proxy username used for this request. Only set
	// when ErrorReason is no_proxies_allocated. May be nil.
	AuthUsername *string `json:"auth_username"`
	// ProxyAddress is the IP address of the proxy used to access the target
	// site. Nil for residential plans.
	ProxyAddress *string `json:"proxy_address"`
	// Bytes is the number of bytes consumed by this proxy request.
	Bytes float64 `json:"bytes"`
	// ClientAddress is the IP address used to connect to the proxy server.
	ClientAddress string `json:"client_address"`
	// IPAddress is the IP address of the target site. May be nil.
	IPAddress *string `json:"ip_address"`
	// Hostname is the hostname of the target site. May be nil.
	Hostname *string `json:"hostname"`
	// Domain is the domain name of the target site. May be nil.
	Domain *string `json:"domain"`
	// Port is the port of the target site. May be nil.
	Port *int `json:"port"`
	// ProxyPort is the source port used to connect to the target site. May
	// be nil.
	ProxyPort *int `json:"proxy_port"`
	// ListenAddress is the IP address of the proxy server connected to.
	ListenAddress string `json:"listen_address"`
	// ListenPort is the port of the proxy server connected to.
	ListenPort int `json:"listen_port"`
}

// ProxyActivityListParams are the parameters for ProxyActivityService.List.
type ProxyActivityListParams struct {
	// TimestampLTE bounds activities to timestamps at or before this time.
	TimestampLTE *time.Time
	// TimestampGTE bounds activities to timestamps at or after this time.
	// No older than 90 days.
	TimestampGTE *time.Time
	// Search is a generic search query.
	Search string
	// ErrorReason matches only requests with the given error reason. Pass
	// "*" to match any request with an error.
	ErrorReason string
	// StartingAfter pages the list: pass the timestamp of the last activity
	// to retrieve the next page.
	StartingAfter *time.Time
	// PageSize is the number of results per page.
	PageSize *int
	// BytesGTE filters requests with bytes at or above the given value.
	BytesGTE string
	// BytesLTE filters requests with bytes at or below the given value.
	BytesLTE string
	// VerificationCategory filters by an account verification category.
	VerificationCategory string
	// PlanID targets a specific plan; otherwise the default plan is used.
	PlanID *int
}

// List returns the proxy activity within a time period. This endpoint
// paginates with starting_after and page_size instead of page numbers: pass
// the timestamp of the latest activity as StartingAfter to view the next
// page (NextPage and ListAll handle this automatically).
func (s *ProxyActivityService) List(ctx context.Context, params ProxyActivityListParams, opts ...RequestOption) (*Page[ProxyActivity], error) {
	q := url.Values{}
	setTime(q, "timestamp__lte", params.TimestampLTE)
	setTime(q, "timestamp__gte", params.TimestampGTE)
	setString(q, "search", params.Search)
	setString(q, "error_reason", params.ErrorReason)
	setTime(q, "starting_after", params.StartingAfter)
	setInt(q, "page_size", params.PageSize)
	setString(q, "bytes__gte", params.BytesGTE)
	setString(q, "bytes__lte", params.BytesLTE)
	setString(q, "verification_category", params.VerificationCategory)
	setInt(q, "plan_id", params.PlanID)
	return getPage[ProxyActivity](ctx, s.client, "/api/v2/proxy/activity/", q, opts)
}

// ListAll returns a lazy iterator over every proxy activity across all
// pages.
func (s *ProxyActivityService) ListAll(ctx context.Context, params ProxyActivityListParams, opts ...RequestOption) iter.Seq2[ProxyActivity, error] {
	return iterPages(ctx, func(ctx context.Context) (*Page[ProxyActivity], error) {
		return s.List(ctx, params, opts...)
	})
}

// ProxyActivityDownloadParams are the parameters for
// ProxyActivityService.Download.
type ProxyActivityDownloadParams struct {
	// DownloadToken is the key obtained from DownloadTokens.Get with the
	// activity scope. Required.
	DownloadToken string
	// TimestampLTE bounds activities to timestamps at or before this time.
	TimestampLTE *time.Time
	// TimestampGTE bounds activities to timestamps at or after this time.
	TimestampGTE *time.Time
	// Search is a generic search query.
	Search string
	// ErrorReason matches only requests with the given error reason. Pass
	// "*" to match any request with an error.
	ErrorReason string
	// StartingAfter passes the timestamp of the last activity.
	StartingAfter *time.Time
	// BytesGTE filters requests with bytes at or above the given value.
	BytesGTE string
	// BytesLTE filters requests with bytes at or below the given value.
	BytesLTE string
	// PlanID targets a specific plan; otherwise the default plan is used.
	PlanID *int
}

// Download fetches the proxy activities as a CSV file (columns: Time,
// Hostname, Destination Port, Bytes, Duration, Proxy, Your IP Address, Error
// Reason, Protocol). This endpoint is unauthenticated (it uses the download
// token from the activity scope); no Authorization header is sent.
func (s *ProxyActivityService) Download(ctx context.Context, params ProxyActivityDownloadParams, opts ...RequestOption) (string, error) {
	if params.DownloadToken == "" {
		return "", errors.New("webshare: proxy activity download: download token is required")
	}
	q := url.Values{}
	q.Set("download_token", params.DownloadToken)
	setTime(q, "timestamp__lte", params.TimestampLTE)
	setTime(q, "timestamp__gte", params.TimestampGTE)
	setString(q, "search", params.Search)
	setString(q, "error_reason", params.ErrorReason)
	setTime(q, "starting_after", params.StartingAfter)
	setString(q, "bytes__gte", params.BytesGTE)
	setString(q, "bytes__lte", params.BytesLTE)
	setInt(q, "plan_id", params.PlanID)
	return s.client.doText(ctx, http.MethodGet, "/api/v2/proxy/activity/download/", q, withOptions(opts, withAuthMode(authNone)))
}
