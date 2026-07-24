package webshare

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// RequestOption configures a Client or an individual request. Every option
// can be passed to NewClient to set a default for all requests, and most can
// also be passed to any API method to override the client-level setting for
// that single call.
type RequestOption func(*requestConfig)

// requestConfig holds the effective settings for a request. The client keeps
// a base copy that per-request options are layered on top of.
type requestConfig struct {
	baseURL            *url.URL
	httpClient         *http.Client
	tokenSource        TokenSource
	unauthenticated    bool
	authMode           authMode
	maxRetries         int
	retryNonIdempotent bool
	timeout            time.Duration
	subuserID          *int64
	federatedUserID    *int64
	source             string
	headers            http.Header
	err                error
}

func (c requestConfig) clone() requestConfig {
	out := c
	if c.headers != nil {
		out.headers = c.headers.Clone()
	}
	return out
}

// WithAPIKey authenticates requests with the given Webshare API key. It is
// shorthand for WithTokenSource(StaticTokenSource(Token{Value: key})). An
// empty key is treated as absent: the client falls back to the
// WEBSHARE_API_KEY environment variable and then errors.
func WithAPIKey(key string) RequestOption {
	if key == "" {
		return func(*requestConfig) {}
	}
	return WithTokenSource(StaticTokenSource(Token{Value: key}))
}

// WithTokenSource sets the credential provider used to authenticate requests.
// Use this to plug in dynamic credentials such as OAuth tokens.
func WithTokenSource(ts TokenSource) RequestOption {
	return func(cfg *requestConfig) {
		cfg.tokenSource = ts
	}
}

// WithUnauthenticated constructs a client without credentials: NewClient
// neither requires a key nor reads WEBSHARE_API_KEY. Only unauthenticated
// operations (login, registration, password reset, referral code info, the
// download endpoints) can be called; authenticated operations fail with a
// clear client-side error until credentials are provided.
func WithUnauthenticated() RequestOption {
	return func(cfg *requestConfig) {
		cfg.unauthenticated = true
	}
}

// WithBaseURL overrides the API base URL. The default is
// "https://proxy.webshare.io" (the bare host: every operation path carries
// its full /api/vN/... prefix). The URL must carry a scheme and a host.
func WithBaseURL(rawURL string) RequestOption {
	return func(cfg *requestConfig) {
		u, err := url.Parse(rawURL)
		if err != nil {
			cfg.err = fmt.Errorf("webshare: invalid base URL %q: %w", rawURL, err)
			return
		}
		if u.Scheme == "" || u.Host == "" {
			cfg.err = fmt.Errorf("webshare: invalid base URL %q: scheme and host are required", rawURL)
			return
		}
		cfg.baseURL = u
	}
}

// WithHTTPClient sets the *http.Client used to execute requests.
func WithHTTPClient(hc *http.Client) RequestOption {
	return func(cfg *requestConfig) {
		cfg.httpClient = hc
	}
}

// WithMaxRetries sets how many times a failed request is retried (default 2,
// meaning up to 3 attempts in total). Retries apply to connection errors,
// timeouts and 408/429/5xx responses on idempotent requests.
func WithMaxRetries(n int) RequestOption {
	return func(cfg *requestConfig) {
		if n < 0 {
			n = 0
		}
		cfg.maxRetries = n
	}
}

// WithRetryNonIdempotent opts non-idempotent requests (POST and PATCH) into
// the retry policy. By default only GET, PUT and DELETE requests are retried.
func WithRetryNonIdempotent() RequestOption {
	return func(cfg *requestConfig) {
		cfg.retryNonIdempotent = true
	}
}

// WithTimeout sets the timeout for a single request attempt (default 60s).
// A zero or negative duration disables the timeout.
func WithTimeout(d time.Duration) RequestOption {
	return func(cfg *requestConfig) {
		cfg.timeout = d
	}
}

// WithSubuser adds the X-Subuser header so that proxy configuration, proxy
// list, proxy stats and proxy activity calls act on behalf of the given
// sub-user.
func WithSubuser(id int64) RequestOption {
	return func(cfg *requestConfig) {
		cfg.subuserID = &id
	}
}

// WithFederatedUser adds the X-Webshare-Federated-Access header so requests
// retrieve data as the given user. This is an admin-only feature.
func WithFederatedUser(id int64) RequestOption {
	return func(cfg *requestConfig) {
		cfg.federatedUserID = &id
	}
}

// WithSource replaces the X-Webshare-Source header value that identifies the
// caller for API-side tracking. The default is
// "WebshareSDK/<version> (Go; <runtime version>)"; products built on the SDK
// (such as a CLI) can send their own product string instead.
func WithSource(source string) RequestOption {
	return func(cfg *requestConfig) {
		cfg.source = source
	}
}

// WithHeader adds an extra header to requests.
func WithHeader(key, value string) RequestOption {
	return func(cfg *requestConfig) {
		if cfg.headers == nil {
			cfg.headers = make(http.Header)
		}
		cfg.headers.Set(key, value)
	}
}
