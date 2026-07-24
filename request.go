package webshare

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	retryBaseDelay    = 500 * time.Millisecond
	retryMaxDelay     = 8 * time.Second
	retryAfterCap     = 60 * time.Second
	headerRequestID   = "X-Request-ID"
	headerSubuser     = "X-Subuser"
	headerFederation  = "X-Webshare-Federated-Access"
	headerSource      = "X-Webshare-Source"
	contentTypeJSON   = "application/json"
	acceptJSONDefault = "application/json"
)

// defaultSource is the default X-Webshare-Source header value, identifying
// the SDK and the Go runtime for API-side caller tracking.
var defaultSource = "WebshareSDK/" + Version + " (Go; " + runtime.Version() + ")"

// authMode describes how an operation authenticates.
type authMode int

const (
	// authRequired operations always send the Authorization header and fail
	// client-side when the client has no credentials.
	authRequired authMode = iota
	// authNone operations never send the Authorization header, even on a
	// credentialed client (spec security: []).
	authNone
)

// withAuthMode marks an operation's authentication mode. It is applied after
// caller options so resource methods control it.
func withAuthMode(mode authMode) RequestOption {
	return func(cfg *requestConfig) {
		cfg.authMode = mode
	}
}

// withOptions returns opts extended with extra, never mutating the caller's
// backing array.
func withOptions(opts []RequestOption, extra ...RequestOption) []RequestOption {
	return append(opts[:len(opts):len(opts)], extra...)
}

// requestConfigFor layers per-request options over the client defaults.
func (c *Client) requestConfigFor(opts []RequestOption) (requestConfig, error) {
	cfg := c.cfg.clone()
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.err != nil {
		return cfg, cfg.err
	}
	return cfg, nil
}

// resolve builds the absolute request URL for a path and query.
func (cfg *requestConfig) resolve(path string, query url.Values) string {
	u := *cfg.baseURL
	u.Path, u.RawPath = joinPath(u.EscapedPath(), path)
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}
	return u.String()
}

// joinPath appends an API path (which may contain pre-encoded segments) to an
// escaped base path, preserving trailing-slash semantics exactly.
func joinPath(escapedBase, path string) (unescaped, escaped string) {
	for len(escapedBase) > 0 && escapedBase[len(escapedBase)-1] == '/' {
		escapedBase = escapedBase[:len(escapedBase)-1]
	}
	escaped = escapedBase + path
	unescaped, err := url.PathUnescape(escaped)
	if err != nil {
		unescaped = escaped
	}
	return unescaped, escaped
}

// origin extracts the scheme://host origin of a URL for same-origin checks.
func origin(u *url.URL) string {
	return strings.ToLower(u.Scheme) + "://" + strings.ToLower(u.Host)
}

// doJSON performs a request with an optional JSON body and decodes a JSON
// response into out (which may be nil for empty responses).
func (c *Client) doJSON(ctx context.Context, method, path string, query url.Values, body any, out any, opts []RequestOption) error {
	var payload []byte
	contentType := ""
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("webshare: encoding request body: %w", err)
		}
		contentType = contentTypeJSON
	}
	data, status, err := c.doRaw(ctx, method, path, query, payload, contentType, "", opts)
	if err != nil {
		return err
	}
	return decodeJSON(data, status, out)
}

// doText performs a request that returns a plain-text response body.
func (c *Client) doText(ctx context.Context, method, path string, query url.Values, opts []RequestOption) (string, error) {
	data, _, err := c.doRaw(ctx, method, path, query, nil, "", "*/*", opts)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// doBytes performs a request that returns a binary response body.
func (c *Client) doBytes(ctx context.Context, method, path string, query url.Values, opts []RequestOption) ([]byte, error) {
	data, _, err := c.doRaw(ctx, method, path, query, nil, "", "*/*", opts)
	return data, err
}

// doMultipart performs a request with a multipart/form-data body and decodes
// a JSON response into out. Multipart bodies are buffered to bytes before
// sending so retries never replay an exhausted stream.
func (c *Client) doMultipart(ctx context.Context, method, path string, body []byte, contentType string, out any, opts []RequestOption) error {
	data, status, err := c.doRaw(ctx, method, path, nil, body, contentType, "", opts)
	if err != nil {
		return err
	}
	return decodeJSON(data, status, out)
}

// doURL performs a GET against an absolute URL (used to follow pagination
// "next" links verbatim) and decodes the JSON response into out. URLs whose
// origin differs from the client base URL are refused with a
// *CrossOriginError so credentials are never sent cross-origin.
func (c *Client) doURL(ctx context.Context, rawURL string, out any, opts []RequestOption) error {
	cfg, err := c.requestConfigFor(opts)
	if err != nil {
		return err
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("webshare: invalid pagination URL %q: %w", rawURL, err)
	}
	if !u.IsAbs() {
		u = cfg.baseURL.ResolveReference(u)
	}
	if got, want := origin(u), origin(cfg.baseURL); got != want {
		return &CrossOriginError{BaseOrigin: want, TargetOrigin: got}
	}
	data, status, err := c.execute(ctx, cfg, http.MethodGet, u.String(), nil, "", "")
	if err != nil {
		return err
	}
	return decodeJSON(data, status, out)
}

// decodeJSON unmarshals a success response body into out, mapping decode
// failures to *ResponseDecodeError.
func decodeJSON(data []byte, status int, out any) error {
	if out == nil || len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return &ResponseDecodeError{StatusCode: status, Body: capBody(data), Err: err}
	}
	return nil
}

// capBody bounds a captured body to the error-body cap.
func capBody(data []byte) []byte {
	if len(data) > maxErrorBody {
		return data[:maxErrorBody]
	}
	return data
}

func (c *Client) doRaw(ctx context.Context, method, path string, query url.Values, body []byte, contentType, accept string, opts []RequestOption) ([]byte, int, error) {
	cfg, err := c.requestConfigFor(opts)
	if err != nil {
		return nil, 0, err
	}
	return c.execute(ctx, cfg, method, cfg.resolve(path, query), body, contentType, accept)
}

// execute runs the request with the retry policy applied and returns the
// response body and status, mapping non-success statuses to *Error and
// transport failures to *RequestError.
func (c *Client) execute(ctx context.Context, cfg requestConfig, method, rawURL string, body []byte, contentType, accept string) ([]byte, int, error) {
	if cfg.authMode == authRequired && cfg.tokenSource == nil {
		return nil, 0, errors.New("webshare: this operation requires credentials: construct the client with webshare.WithAPIKey or webshare.WithTokenSource, or set the WEBSHARE_API_KEY environment variable")
	}

	retryAllowed := cfg.retryNonIdempotent || isIdempotent(method)
	attempts := 1
	if retryAllowed {
		attempts = cfg.maxRetries + 1
	}

	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			if err := sleepContext(ctx, backoffDelay(attempt-1, lastErr)); err != nil {
				return nil, 0, &RequestError{Err: err}
			}
		}
		data, status, retryable, err := c.attempt(ctx, cfg, method, rawURL, body, contentType, accept)
		if err == nil {
			return data, status, nil
		}
		if ctx.Err() != nil {
			return nil, 0, &RequestError{Err: ctx.Err()}
		}
		if !retryable {
			return nil, 0, err
		}
		lastErr = err
	}
	return nil, 0, lastErr
}

// attempt performs a single HTTP round trip. The returned boolean reports
// whether the failure is retryable.
func (c *Client) attempt(ctx context.Context, cfg requestConfig, method, rawURL string, body []byte, contentType, accept string) ([]byte, int, bool, error) {
	attemptCtx := ctx
	if cfg.timeout > 0 {
		var cancel context.CancelFunc
		attemptCtx, cancel = context.WithTimeout(ctx, cfg.timeout)
		defer cancel()
	}

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(attemptCtx, method, rawURL, reader)
	if err != nil {
		return nil, 0, false, &RequestError{Err: err}
	}

	// The token source is consulted on every attempt so refreshed
	// credentials are picked up between retries. Unauthenticated operations
	// never send the header.
	if cfg.authMode != authNone && cfg.tokenSource != nil {
		token, err := cfg.tokenSource.Token(attemptCtx)
		if err != nil {
			return nil, 0, false, &RequestError{Err: fmt.Errorf("obtaining credentials: %w", err)}
		}
		req.Header.Set("Authorization", token.authorizationHeader())
	}
	req.Header.Set("User-Agent", "webshare-go/"+Version)
	source := cfg.source
	if source == "" {
		source = defaultSource
	}
	req.Header.Set(headerSource, source)
	if accept == "" {
		accept = acceptJSONDefault
	}
	req.Header.Set("Accept", accept)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if cfg.subuserID != nil {
		req.Header.Set(headerSubuser, strconv.FormatInt(*cfg.subuserID, 10))
	}
	if cfg.federatedUserID != nil {
		req.Header.Set(headerFederation, strconv.FormatInt(*cfg.federatedUserID, 10))
	}
	for key, values := range cfg.headers {
		req.Header[key] = values
	}

	resp, err := cfg.httpClient.Do(req)
	if err != nil {
		// Connection errors and per-attempt timeouts are retryable unless
		// the parent context is done.
		if ctx.Err() != nil {
			return nil, 0, false, &RequestError{Err: ctx.Err()}
		}
		return nil, 0, true, &RequestError{Err: err}
	}
	defer resp.Body.Close()

	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	var bodyReader io.Reader = resp.Body
	if !success {
		// Error bodies are capped at 1 MiB; success bodies (e.g. proxy
		// list downloads) are read in full.
		bodyReader = io.LimitReader(resp.Body, maxErrorBody)
	}
	data, err := io.ReadAll(bodyReader)
	if err != nil {
		if ctx.Err() != nil {
			return nil, 0, false, &RequestError{Err: ctx.Err()}
		}
		return nil, 0, true, &RequestError{Err: err}
	}

	if success {
		return data, resp.StatusCode, false, nil
	}
	apiErr := parseAPIError(resp.StatusCode, resp.Header.Get(headerRequestID), data)
	apiErr.RetryAfter = parseRetryAfter(resp.Header.Get("Retry-After"))
	return nil, resp.StatusCode, isRetryableStatus(resp.StatusCode), apiErr
}

func isIdempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodHead:
		return true
	}
	return false
}

func isRetryableStatus(status int) bool {
	switch status {
	case http.StatusRequestTimeout, http.StatusTooManyRequests,
		http.StatusInternalServerError, http.StatusBadGateway,
		http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	}
	return false
}

// backoffDelay computes the delay before the next retry: exponential backoff
// with full jitter (base 0.5s, cap 8s), or the server-provided Retry-After
// value (capped at 60s) when present on the previous error.
func backoffDelay(retry int, lastErr error) time.Duration {
	var apiErr *Error
	if errors.As(lastErr, &apiErr) && apiErr.RetryAfter != nil && *apiErr.RetryAfter > 0 {
		return min(*apiErr.RetryAfter, retryAfterCap)
	}
	ceiling := retryBaseDelay << retry
	if ceiling > retryMaxDelay || ceiling <= 0 {
		ceiling = retryMaxDelay
	}
	return time.Duration(rand.Float64() * float64(ceiling))
}

// parseRetryAfter parses a Retry-After header value in either delta-seconds
// or HTTP-date form. Whitespace is trimmed; NaN, infinite, negative or
// otherwise unparsable values yield nil (the header is treated as absent).
// Zone-less HTTP-dates are interpreted as UTC (http.ParseTime semantics).
func parseRetryAfter(value string) *time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if secs, err := strconv.ParseFloat(value, 64); err == nil {
		if math.IsNaN(secs) || math.IsInf(secs, 0) || secs < 0 {
			return nil
		}
		d := time.Duration(secs * float64(time.Second))
		return &d
	}
	if t, err := http.ParseTime(value); err == nil {
		if d := time.Until(t); d > 0 {
			return &d
		}
	}
	return nil
}

// sleepContext waits for d, returning early with the context error when the
// context is cancelled mid-backoff.
func sleepContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
