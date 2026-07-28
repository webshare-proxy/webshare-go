package webshare

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestIntegration exercises the real API. It is skipped unless the
// dedicated WEBSHARE_INTEGRATION_TEST environment variable is set (in
// addition to WEBSHARE_API_KEY), so a plain `go test ./...` with a real key
// exported never hits production by accident. The optional
// WEBSHARE_BASE_URL environment variable overrides the API base URL
// (defaults to production).
func TestIntegration(t *testing.T) {
	if os.Getenv("WEBSHARE_INTEGRATION_TEST") == "" {
		t.Skip("WEBSHARE_INTEGRATION_TEST is not set; skipping integration test")
	}
	if os.Getenv("WEBSHARE_API_KEY") == "" {
		t.Skip("WEBSHARE_API_KEY is not set; skipping integration test")
	}
	var opts []RequestOption
	if base := os.Getenv("WEBSHARE_BASE_URL"); base != "" {
		opts = append(opts, WithBaseURL(base))
	}
	client, err := NewClient(opts...)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	profile, err := client.Profile.Get(ctx)
	if err != nil {
		t.Fatalf("Profile.Get: %v", err)
	}
	if profile.Email == "" {
		t.Error("profile.Email is empty")
	}

	page, err := client.Proxies.List(ctx, ProxyListParams{Mode: ModeDirect, PageSize: Int(5)})
	if err != nil {
		t.Fatalf("Proxies.List: %v", err)
	}
	if page.Count > 0 && len(page.Results) == 0 {
		t.Error("proxy list count is positive but the first page is empty")
	}
}
