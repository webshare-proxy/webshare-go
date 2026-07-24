package webshare

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	retryBaseDelay    = 500 * time.Millisecond
	retryMaxDelay     = 8 * time.Second
	retryAfterCap     = 60 * time.Second
	headerRequestID   = "X-Request-ID"
	headerSubuser     = "X-Subuser"
	headerFederation  = "X-Webshare-Federated-Access"
	contentTypeJSON   = "application/json"
	acceptJSONDefault = "application/json"
)

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
	u.Path, u.RawPath = joinPath(u.Path, path)
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}
	return u.String()
}

// joinPath appends an API path (which may contain pre-encoded segments) to a
// base path, preserving trailing-slash semantics exactly.
func joinPath(base, path string) (unescaped, escaped string) {
	for len(base) > 0 && base[len(base)-1] == '/' {
		base = base[:len(base)-1]
	}
	escaped = base + path
	unescaped, err := url.PathUnescape(escaped)
	if err != nil {
		unescaped = escaped
	}
	return unescaped, escaped
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
	data, err := c.doRaw(ctx, method, path, query, payload, contentType, "", opts)
	if err != nil {
		return err
	}
	return decodeJSON(data, out)
}

// doText performs a request that returns a plain-text response body.
func (c *Client) doText(ctx context.Context, method, path string, query url.Values, opts []RequestOption) (string, error) {
	data, err := c.doRaw(ctx, method, path, query, nil, "", "*/*", opts)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// doBytes performs a request that returns a binary response body.
func (c *Client) doBytes(ctx context.Context, method, path string, query url.Values, opts []RequestOption) ([]byte, error) {
	return c.doRaw(ctx, method, path, query, nil, "", "*/*", opts)
}

// doMultipart performs a request with a multipart/form-data body and decodes
// a JSON response into out.
func (c *Client) doMultipart(ctx context.Context, method, path string, body []byte, contentType string, out any, opts []RequestOption) error {
	data, err := c.doRaw(ctx, method, path, nil, body, contentType, "", opts)
	if err != nil {
		return err
	}
	return decodeJSON(data, out)
}

// doURL performs a GET against an absolute URL (used to follow pagination
// "next" links verbatim) and decodes the JSON response into out.
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
	data, err := c.execute(ctx, cfg, http.MethodGet, u.String(), nil, "", "")
	if err != nil {
		return err
	}
	return decodeJSON(data, out)
}

func decodeJSON(data []byte, out any) error {
	if out == nil || len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("webshare: decoding response body: %w", err)
	}
	return nil
}

func (c *Client) doRaw(ctx context.Context, method, path string, query url.Values, body []byte, contentType, accept string, opts []RequestOption) ([]byte, error) {
	cfg, err := c.requestConfigFor(opts)
	if err != nil {
		return nil, err
	}
	return c.execute(ctx, cfg, method, cfg.resolve(path, query), body, contentType, accept)
}

// execute runs the request with the retry policy applied and returns the
// response body, mapping non-success statuses to *Error and transport
// failures to *RequestError.
func (c *Client) execute(ctx context.Context, cfg requestConfig, method, rawURL string, body []byte, contentType, accept string) ([]byte, error) {
	retryAllowed := cfg.retryNonIdempotent || isIdempotent(method)
	attempts := 1
	if retryAllowed {
		attempts = cfg.maxRetries + 1
	}

	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			if err := sleepContext(ctx, backoffDelay(attempt-1, lastErr)); err != nil {
				return nil, &RequestError{Err: err}
			}
		}
		data, retryable, err := c.attempt(ctx, cfg, method, rawURL, body, contentType, accept)
		if err == nil {
			return data, nil
		}
		if ctx.Err() != nil {
			return nil, &RequestError{Err: ctx.Err()}
		}
		if !retryable {
			return nil, err
		}
		lastErr = err
	}
	return nil, lastErr
}

// attempt performs a single HTTP round trip. The returned boolean reports
// whether the failure is retryable.
func (c *Client) attempt(ctx context.Context, cfg requestConfig, method, rawURL string, body []byte, contentType, accept string) ([]byte, bool, error) {
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
		return nil, false, &RequestError{Err: err}
	}

	token, err := cfg.tokenSource.Token(attemptCtx)
	if err != nil {
		return nil, false, &RequestError{Err: fmt.Errorf("obtaining credentials: %w", err)}
	}
	req.Header.Set("Authorization", token.authorizationHeader())
	req.Header.Set("User-Agent", "webshare-go/"+Version)
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
			return nil, false, &RequestError{Err: ctx.Err()}
		}
		return nil, true, &RequestError{Err: err}
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		if ctx.Err() != nil {
			return nil, false, &RequestError{Err: ctx.Err()}
		}
		return nil, true, &RequestError{Err: err}
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return data, false, nil
	}
	apiErr := parseAPIError(resp.StatusCode, resp.Header.Get(headerRequestID), data)
	apiErr.retryAfter = parseRetryAfter(resp.Header.Get("Retry-After"))
	return nil, isRetryableStatus(resp.StatusCode), apiErr
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
	if errors.As(lastErr, &apiErr) && apiErr.retryAfter > 0 {
		return min(apiErr.retryAfter, retryAfterCap)
	}
	ceiling := retryBaseDelay << retry
	if ceiling > retryMaxDelay || ceiling <= 0 {
		ceiling = retryMaxDelay
	}
	return time.Duration(rand.Float64() * float64(ceiling))
}

// parseRetryAfter parses a Retry-After header value in either delta-seconds
// or HTTP-date form. It returns 0 when absent or unparsable.
func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}
	if secs, err := strconv.Atoi(value); err == nil {
		if secs < 0 {
			return 0
		}
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(value); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
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
