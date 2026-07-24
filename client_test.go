package webshare

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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
	_, err := client.APIKeys.Create(context.Background(), APIKeyCreateParams{Label: "x"})
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
		writeJSON(t, w, http.StatusOK, map[string]any{"id": 9, "label": "x"})
	}))
	key, err := client.APIKeys.Create(context.Background(), APIKeyCreateParams{Label: "x"}, WithRetryNonIdempotent())
	if err != nil {
		t.Fatalf("APIKeys.Create: %v", err)
	}
	if key.ID != 9 {
		t.Errorf("key.ID = %d, want 9", key.ID)
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
