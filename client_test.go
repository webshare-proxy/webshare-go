package webshare

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func newTestClient(t *testing.T, handler http.Handler, opts ...RequestOption) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	options := append([]RequestOption{
		WithAPIKey("test-api-key"),
		WithBaseURL(server.URL),
	}, opts...)
	client, err := NewClient(options...)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client
}

func writeJSON(t *testing.T, w http.ResponseWriter, status int, body any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body != nil {
		if err := json.NewEncoder(w).Encode(body); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}
}

func TestNewClientRequiresCredentials(t *testing.T) {
	t.Setenv("WEBSHARE_API_KEY", "")
	if _, err := NewClient(); err == nil {
		t.Fatal("expected an error when no credentials are available")
	}
}

func TestNewClientReadsEnvironmentKey(t *testing.T) {
	t.Setenv("WEBSHARE_API_KEY", "env-key")
	var got string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("Authorization")
		writeJSON(t, w, http.StatusOK, map[string]any{"id": 1})
	})
	server := httptest.NewServer(handler)
	defer server.Close()
	client, err := NewClient(WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := client.Profile.Get(context.Background()); err != nil {
		t.Fatalf("Profile.Get: %v", err)
	}
	if got != "Token env-key" {
		t.Fatalf("Authorization = %q, want %q", got, "Token env-key")
	}
}

func TestAuthorizationHeaderAndUserAgent(t *testing.T) {
	var auth, userAgent string
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		userAgent = r.Header.Get("User-Agent")
		writeJSON(t, w, http.StatusOK, map[string]any{"id": 1})
	}))
	if _, err := client.Profile.Get(context.Background()); err != nil {
		t.Fatalf("Profile.Get: %v", err)
	}
	if auth != "Token test-api-key" {
		t.Errorf("Authorization = %q, want %q", auth, "Token test-api-key")
	}
	if userAgent != "webshare-go/"+Version {
		t.Errorf("User-Agent = %q, want %q", userAgent, "webshare-go/"+Version)
	}
}

// oauthSource simulates a refreshing OAuth credential with a non-default
// header scheme, as used by the CLI's PKCE flow.
type oauthSource struct {
	calls atomic.Int32
}

func (s *oauthSource) Token(context.Context) (Token, error) {
	s.calls.Add(1)
	return Token{Value: "oauth-access-token", Scheme: "Bearer"}, nil
}

func TestTokenSourcePlugin(t *testing.T) {
	var auth string
	source := &oauthSource{}
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		writeJSON(t, w, http.StatusOK, map[string]any{"id": 1})
	}), WithTokenSource(source))
	if _, err := client.Profile.Get(context.Background()); err != nil {
		t.Fatalf("Profile.Get: %v", err)
	}
	if auth != "Bearer oauth-access-token" {
		t.Errorf("Authorization = %q, want %q", auth, "Bearer oauth-access-token")
	}
	if source.calls.Load() == 0 {
		t.Error("token source was never called")
	}
}

func TestSourceHeaderDefault(t *testing.T) {
	var got string
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Webshare-Source")
		writeJSON(t, w, http.StatusOK, map[string]any{"id": 1})
	}))
	if _, err := client.Profile.Get(context.Background()); err != nil {
		t.Fatalf("Profile.Get: %v", err)
	}
	// e.g. "WebshareSDK/0.1.0 (Go; go1.25.4)" — the runtime version varies.
	pattern := regexp.MustCompile(`^WebshareSDK/` + regexp.QuoteMeta(Version) + ` \(Go; .+\)$`)
	if !pattern.MatchString(got) {
		t.Errorf("X-Webshare-Source = %q, want to match %q", got, pattern)
	}
	if !strings.Contains(got, runtime.Version()) {
		t.Errorf("X-Webshare-Source = %q, want to contain runtime version %q", got, runtime.Version())
	}
}

func TestSourceHeaderOverride(t *testing.T) {
	var got string
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Webshare-Source")
		writeJSON(t, w, http.StatusOK, map[string]any{"id": 1})
	}), WithSource("WebshareCLI/2.0.0 (Go; test)"))
	if _, err := client.Profile.Get(context.Background()); err != nil {
		t.Fatalf("Profile.Get: %v", err)
	}
	if got != "WebshareCLI/2.0.0 (Go; test)" {
		t.Errorf("X-Webshare-Source = %q, want the WithSource override", got)
	}

	// A per-request header option wins over everything.
	if _, err := client.Profile.Get(context.Background(), WithHeader("X-Webshare-Source", "per-request")); err != nil {
		t.Fatalf("Profile.Get: %v", err)
	}
	if got != "per-request" {
		t.Errorf("X-Webshare-Source = %q, want the per-request header override", got)
	}
}

func TestSubuserAndFederatedHeaders(t *testing.T) {
	var subuser, federated string
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		subuser = r.Header.Get("X-Subuser")
		federated = r.Header.Get("X-Webshare-Federated-Access")
		writeJSON(t, w, http.StatusOK, map[string]any{"id": 1})
	}), WithSubuser(42))
	// The federated header is layered on per-request.
	if _, err := client.Profile.Get(context.Background(), WithFederatedUser(7)); err != nil {
		t.Fatalf("Profile.Get: %v", err)
	}
	if subuser != "42" {
		t.Errorf("X-Subuser = %q, want %q", subuser, "42")
	}
	if federated != "7" {
		t.Errorf("X-Webshare-Federated-Access = %q, want %q", federated, "7")
	}
}

func TestRetryOn429WithRetryAfterAnd500(t *testing.T) {
	var attempts atomic.Int32
	start := time.Now()
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch attempts.Add(1) {
		case 1:
			w.Header().Set("Retry-After", "1")
			writeJSON(t, w, http.StatusTooManyRequests, map[string]any{"detail": "throttled"})
		case 2:
			writeJSON(t, w, http.StatusInternalServerError, map[string]any{"detail": "boom"})
		default:
			writeJSON(t, w, http.StatusOK, map[string]any{"id": 3})
		}
	}))
	profile, err := client.Profile.Get(context.Background())
	if err != nil {
		t.Fatalf("Profile.Get: %v", err)
	}
	if profile.ID != 3 {
		t.Errorf("profile.ID = %d, want 3", profile.ID)
	}
	if got := attempts.Load(); got != 3 {
		t.Errorf("attempts = %d, want 3", got)
	}
	if elapsed := time.Since(start); elapsed < time.Second {
		t.Errorf("elapsed = %v, want at least 1s (Retry-After honored)", elapsed)
	}
}

func TestRetriesExhausted(t *testing.T) {
	var attempts atomic.Int32
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		writeJSON(t, w, http.StatusServiceUnavailable, map[string]any{"detail": "down"})
	}), WithMaxRetries(1))
	_, err := client.Profile.Get(context.Background())
	var apiErr *Error
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("err = %v, want *Error with status 503", err)
	}
	if got := attempts.Load(); got != 2 {
		t.Errorf("attempts = %d, want 2 (maxRetries 1)", got)
	}
}

func TestNoRetryOnPOST(t *testing.T) {
	var attempts atomic.Int32
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		writeJSON(t, w, http.StatusInternalServerError, map[string]any{"detail": "boom"})
	}))
	_, err := client.IPAuthorizations.Create(context.Background(), IPAuthorizationCreateParams{IPAddress: "10.1.2.3"})
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *Error", err)
	}
	if got := attempts.Load(); got != 1 {
		t.Errorf("attempts = %d, want 1 (POST is not retried by default)", got)
	}
}

func TestRetryNonIdempotentOptIn(t *testing.T) {
	var attempts atomic.Int32
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) == 1 {
			writeJSON(t, w, http.StatusInternalServerError, map[string]any{"detail": "boom"})
			return
		}
		writeJSON(t, w, http.StatusOK, map[string]any{"id": 9, "ip_address": "10.1.2.3"})
	}))
	auth, err := client.IPAuthorizations.Create(context.Background(), IPAuthorizationCreateParams{IPAddress: "10.1.2.3"}, WithRetryNonIdempotent())
	if err != nil {
		t.Fatalf("IPAuthorizations.Create: %v", err)
	}
	if auth.ID != 9 {
		t.Errorf("auth.ID = %d, want 9", auth.ID)
	}
	if got := attempts.Load(); got != 2 {
		t.Errorf("attempts = %d, want 2 (opt-in retry)", got)
	}
}

func TestErrorMapping(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", "req-123")
		writeJSON(t, w, http.StatusBadRequest, map[string]any{
			"detail": "Validation failed.",
			"code":   "invalid",
			"mode":   []string{"This field is required."},
		})
	}))
	_, err := client.Proxies.List(context.Background(), ProxyListParams{})
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *Error", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode)
	}
	if apiErr.Code != "invalid" {
		t.Errorf("Code = %q, want %q", apiErr.Code, "invalid")
	}
	if apiErr.RequestID != "req-123" {
		t.Errorf("RequestID = %q, want %q", apiErr.RequestID, "req-123")
	}
	if apiErr.Detail != "Validation failed." {
		t.Errorf("Detail = %q, want %q", apiErr.Detail, "Validation failed.")
	}
	if got := apiErr.FieldErrors["mode"]; len(got) != 1 || got[0] != "This field is required." {
		t.Errorf("FieldErrors[mode] = %v, want [This field is required.]", got)
	}
	if len(apiErr.Body) == 0 {
		t.Error("Body is empty, want raw response body")
	}
}

func TestErrorMappingObjectFieldErrors(t *testing.T) {
	// The live API returns field errors as lists of objects, not the
	// documented lists of strings; both shapes must parse.
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		// Real dev-API response shape, verbatim.
		if _, err := io.WriteString(w, `{"mode":[{"message":"This field is required.","code":"required"}]}`); err != nil {
			t.Errorf("writing response: %v", err)
		}
	}))
	_, err := client.Proxies.List(context.Background(), ProxyListParams{})
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *Error", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode)
	}
	if got := apiErr.FieldErrors["mode"]; len(got) != 1 || got[0] != "This field is required." {
		t.Errorf("FieldErrors[mode] = %v, want [This field is required.]", got)
	}
}

func TestErrorCode2FANeeded(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusForbidden, map[string]any{
			"detail": "Two factor authentication is needed.",
			"code":   "2fa_needed",
		})
	}))
	_, err := client.Profile.Get(context.Background())
	var apiErr *Error
	if !errors.As(err, &apiErr) || apiErr.Code != "2fa_needed" {
		t.Fatalf("err = %v, want *Error with code 2fa_needed", err)
	}
}

func TestNonJSONErrorBody(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		if _, err := w.Write([]byte("upstream unavailable")); err != nil {
			t.Errorf("writing response: %v", err)
		}
	}), WithMaxRetries(0))
	_, err := client.Profile.Get(context.Background())
	var apiErr *Error
	if !errors.As(err, &apiErr) || apiErr.Detail != "upstream unavailable" {
		t.Fatalf("err = %v, want *Error with raw text detail", err)
	}
}

func TestContextCancellationMidBackoff(t *testing.T) {
	var attempts atomic.Int32
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.Header().Set("Retry-After", "30")
		writeJSON(t, w, http.StatusTooManyRequests, map[string]any{"detail": "throttled"})
	}))
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	_, err := client.Profile.Get(ctx)
	elapsed := time.Since(start)
	var reqErr *RequestError
	if !errors.As(err, &reqErr) {
		t.Fatalf("err = %v, want *RequestError", err)
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("errors.Is(err, context.Canceled) = false, want true")
	}
	if elapsed > 5*time.Second {
		t.Errorf("elapsed = %v, want prompt return on cancellation mid-backoff", elapsed)
	}
	if got := attempts.Load(); got != 1 {
		t.Errorf("attempts = %d, want 1", got)
	}
}

func TestTimeout(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(5 * time.Second):
		case <-r.Context().Done():
		}
	}), WithTimeout(50*time.Millisecond), WithMaxRetries(0))
	start := time.Now()
	_, err := client.Profile.Get(context.Background())
	var reqErr *RequestError
	if !errors.As(err, &reqErr) {
		t.Fatalf("err = %v, want *RequestError", err)
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Errorf("elapsed = %v, want timeout near 50ms", elapsed)
	}
}

func TestConnectionErrorWrapped(t *testing.T) {
	client, err := NewClient(
		WithAPIKey("k"),
		WithBaseURL("http://127.0.0.1:1"), // nothing listens here
		WithMaxRetries(0),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = client.Profile.Get(context.Background())
	var reqErr *RequestError
	if !errors.As(err, &reqErr) {
		t.Fatalf("err = %v, want *RequestError", err)
	}
	if reqErr.Unwrap() == nil {
		t.Error("Unwrap() = nil, want underlying transport error")
	}
}

func TestTokenSourceCalledPerAttempt(t *testing.T) {
	var attempts atomic.Int32
	source := &oauthSource{}
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) < 3 {
			writeJSON(t, w, http.StatusInternalServerError, map[string]any{"detail": "boom"})
			return
		}
		writeJSON(t, w, http.StatusOK, map[string]any{"id": 1})
	}), WithTokenSource(source))
	if _, err := client.Profile.Get(context.Background()); err != nil {
		t.Fatalf("Profile.Get: %v", err)
	}
	if got := source.calls.Load(); got != 3 {
		t.Errorf("token source calls = %d, want 3 (one per attempt)", got)
	}
}

func TestErrorBodyCappedAndDetailTruncated(t *testing.T) {
	huge := strings.Repeat("x", maxErrorBody+4096)
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := io.WriteString(w, huge); err != nil {
			t.Errorf("writing response: %v", err)
		}
	}))
	_, err := client.Profile.Get(context.Background())
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *Error", err)
	}
	if len(apiErr.Body) != maxErrorBody {
		t.Errorf("len(Body) = %d, want %d (capped)", len(apiErr.Body), maxErrorBody)
	}
	if msg := apiErr.Error(); len(msg) > maxDetailLen+256 {
		t.Errorf("len(Error()) = %d, want at most about %d (truncated detail)", len(msg), maxDetailLen)
	}
}

func TestErrorRetryAfterExposed(t *testing.T) {
	// A 429 on POST is not retried; RetryAfter lets callers self-throttle.
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "2")
		writeJSON(t, w, http.StatusTooManyRequests, map[string]any{"detail": "throttled"})
	}))
	_, err := client.IPAuthorizations.Create(context.Background(), IPAuthorizationCreateParams{IPAddress: "10.1.2.3"})
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *Error", err)
	}
	if apiErr.RetryAfter == nil || *apiErr.RetryAfter != 2*time.Second {
		t.Errorf("RetryAfter = %v, want 2s", apiErr.RetryAfter)
	}
}

func TestParseRetryAfterHardening(t *testing.T) {
	tests := []struct {
		value string
		want  *time.Duration
	}{
		{"5", durationPtr(5 * time.Second)},
		{"  5  ", durationPtr(5 * time.Second)},
		{"1.5", durationPtr(1500 * time.Millisecond)},
		{"-3", nil},
		{"NaN", nil},
		{"Inf", nil},
		{"soon", nil},
		{"", nil},
	}
	for _, tt := range tests {
		got := parseRetryAfter(tt.value)
		switch {
		case tt.want == nil && got != nil:
			t.Errorf("parseRetryAfter(%q) = %v, want nil", tt.value, *got)
		case tt.want != nil && (got == nil || *got != *tt.want):
			t.Errorf("parseRetryAfter(%q) = %v, want %v", tt.value, got, *tt.want)
		}
	}
	// HTTP-date form yields a positive duration for a future date.
	future := time.Now().UTC().Add(30 * time.Second).Format(http.TimeFormat)
	if got := parseRetryAfter(future); got == nil || *got <= 0 || *got > 31*time.Second {
		t.Errorf("parseRetryAfter(future date) = %v, want about 30s", got)
	}
	// Past dates are treated as absent.
	past := time.Now().UTC().Add(-time.Hour).Format(http.TimeFormat)
	if got := parseRetryAfter(past); got != nil {
		t.Errorf("parseRetryAfter(past date) = %v, want nil", *got)
	}
}

func durationPtr(d time.Duration) *time.Duration { return &d }

func TestSuccessNonJSONBodyIsDecodeError(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		if _, err := io.WriteString(w, "<html><body>maintenance page</body></html>"); err != nil {
			t.Errorf("writing response: %v", err)
		}
	}))
	_, err := client.Profile.Get(context.Background())
	var decodeErr *ResponseDecodeError
	if !errors.As(err, &decodeErr) {
		t.Fatalf("err = %v, want *ResponseDecodeError", err)
	}
	if decodeErr.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", decodeErr.StatusCode)
	}
	if !strings.Contains(string(decodeErr.Body), "maintenance") {
		t.Errorf("Body = %q, want the raw body captured", decodeErr.Body)
	}
}

func TestWithBaseURLValidation(t *testing.T) {
	if _, err := NewClient(WithAPIKey("k"), WithBaseURL("proxy.webshare.io")); err == nil {
		t.Error("expected an error for a base URL without a scheme")
	}
	if _, err := NewClient(WithAPIKey("k"), WithBaseURL("https://")); err == nil {
		t.Error("expected an error for a base URL without a host")
	}
}

func TestPerRequestHeaderOption(t *testing.T) {
	var got string
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Custom")
		writeJSON(t, w, http.StatusOK, map[string]any{"id": 1})
	}))
	if _, err := client.Profile.Get(context.Background(), WithHeader("X-Custom", "yes")); err != nil {
		t.Fatalf("Profile.Get: %v", err)
	}
	if got != "yes" {
		t.Errorf("X-Custom = %q, want %q", got, "yes")
	}
}
