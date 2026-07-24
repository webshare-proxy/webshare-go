package webshare

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// authRecorder records whether an Authorization header arrived.
type authRecorder struct {
	got  string
	seen bool
}

func (a *authRecorder) handler(t *testing.T, status int, body any) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.got = r.Header.Get("Authorization")
		a.seen = true
		writeJSON(t, w, status, body)
	})
}

func TestUnauthenticatedOperationsNeverSendAuth(t *testing.T) {
	// Even on a credentialed client, spec security:[] operations must not
	// send the Authorization header.
	recorder := &authRecorder{}
	client := newTestClient(t, recorder.handler(t, http.StatusOK, map[string]any{"referral_code": "abc"}))
	if _, err := client.Referral.GetCodeInfo(context.Background(), "abc"); err != nil {
		t.Fatalf("GetCodeInfo: %v", err)
	}
	if !recorder.seen || recorder.got != "" {
		t.Errorf("Authorization = %q, want empty on unauthenticated operation", recorder.got)
	}

	recorder = &authRecorder{}
	client = newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder.got = r.Header.Get("Authorization")
		recorder.seen = true
		w.WriteHeader(http.StatusOK)
	}))
	_, err := client.Proxies.Download(context.Background(), ProxyDownloadParams{
		Token:                "tok",
		AuthenticationMethod: AuthMethodUsername,
		EndpointMode:         ModeDirect,
	})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if !recorder.seen || recorder.got != "" {
		t.Errorf("Authorization = %q, want empty on proxy list download", recorder.got)
	}
}

func TestUnauthenticatedClient(t *testing.T) {
	// A pre-login CLI flow builds a credential-less client explicitly; the
	// env key must not leak in.
	t.Setenv("WEBSHARE_API_KEY", "env-key-should-not-be-used")
	recorder := &authRecorder{}
	server := httptest.NewServer(recorder.handler(t, http.StatusOK, map[string]any{"token": "login-token"}))
	defer server.Close()
	client, err := NewClient(WithUnauthenticated(), WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	result, err := client.Auth.Login(context.Background(), LoginParams{Email: "a@b.c", Password: "pw", Recaptcha: "r"})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if result.Token != "login-token" {
		t.Errorf("Token = %q, want login-token", result.Token)
	}
	if recorder.got != "" {
		t.Errorf("Authorization = %q, want empty on unauthenticated client", recorder.got)
	}

	// Authenticated operations fail client-side with a clear error.
	recorder.seen = false
	_, err = client.Profile.Get(context.Background())
	if err == nil {
		t.Fatal("Profile.Get succeeded, want a client-side credentials error")
	}
	if !strings.Contains(err.Error(), "WithAPIKey") {
		t.Errorf("err = %v, want a message naming the fix", err)
	}
	if recorder.seen {
		t.Error("a request was sent for an authenticated operation without credentials")
	}
}

func TestEmptyAPIKeyTreatedAsAbsent(t *testing.T) {
	t.Setenv("WEBSHARE_API_KEY", "")
	if _, err := NewClient(WithAPIKey("")); err == nil {
		t.Fatal("expected an error for an explicit empty API key with no env fallback")
	}

	// With the env var set, an empty explicit key falls back to it.
	t.Setenv("WEBSHARE_API_KEY", "env-key")
	recorder := &authRecorder{}
	server := httptest.NewServer(recorder.handler(t, http.StatusOK, map[string]any{"id": 1}))
	defer server.Close()
	client, err := NewClient(WithAPIKey(""), WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := client.Profile.Get(context.Background()); err != nil {
		t.Fatalf("Profile.Get: %v", err)
	}
	if recorder.got != "Token env-key" {
		t.Errorf("Authorization = %q, want %q", recorder.got, "Token env-key")
	}
}

func TestAuthOptionalActivationComplete(t *testing.T) {
	// Credentialed client: the header is sent.
	recorder := &authRecorder{}
	client := newTestClient(t, recorder.handler(t, http.StatusOK, map[string]any{"token": "t"}))
	if _, err := client.Auth.CompleteActivation(context.Background(), ActivationCompleteParams{ActivationToken: "x"}); err != nil {
		t.Fatalf("CompleteActivation: %v", err)
	}
	if recorder.got != "Token test-api-key" {
		t.Errorf("Authorization = %q, want the credential on an auth-optional operation", recorder.got)
	}

	// Unauthenticated client: no header, but the call still works.
	t.Setenv("WEBSHARE_API_KEY", "")
	recorder = &authRecorder{}
	server := httptest.NewServer(recorder.handler(t, http.StatusOK, map[string]any{"token": "t"}))
	defer server.Close()
	anon, err := NewClient(WithUnauthenticated(), WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := anon.Auth.CompleteActivation(context.Background(), ActivationCompleteParams{ActivationToken: "x"}); err != nil {
		t.Fatalf("CompleteActivation: %v", err)
	}
	if recorder.got != "" {
		t.Errorf("Authorization = %q, want empty without credentials", recorder.got)
	}
}
