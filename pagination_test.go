package webshare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

// pagedProxyHandler serves a three-page proxy list keyed by the page query
// parameter, with envelope next URLs pointing at the test server.
func pagedProxyHandler(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/proxy/list/" {
			t.Errorf("path = %q, want /api/v2/proxy/list/", r.URL.Path)
		}
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}
		var next any
		switch page {
		case "1":
			next = "http://" + r.Host + "/api/v2/proxy/list/?mode=direct&page=2"
		case "2":
			next = "http://" + r.Host + "/api/v2/proxy/list/?mode=direct&page=3"
		default:
			next = nil
		}
		writeJSON(t, w, http.StatusOK, map[string]any{
			"count":    6,
			"next":     next,
			"previous": nil,
			"results": []map[string]any{
				{"id": fmt.Sprintf("d-%s-1", page), "port": 8000},
				{"id": fmt.Sprintf("d-%s-2", page), "port": 8001},
			},
		})
	})
}

func TestListMultiPage(t *testing.T) {
	client := newTestClient(t, pagedProxyHandler(t))
	ctx := context.Background()
	page, err := client.Proxies.List(ctx, ProxyListParams{Mode: ModeDirect})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if page.Count != 6 {
		t.Errorf("Count = %d, want 6", page.Count)
	}
	if len(page.Results) != 2 || page.Results[0].ID != "d-1-1" {
		t.Fatalf("unexpected first page results: %+v", page.Results)
	}
	if !page.HasNextPage() {
		t.Fatal("HasNextPage() = false, want true")
	}

	second, err := page.NextPage(ctx)
	if err != nil {
		t.Fatalf("NextPage: %v", err)
	}
	if second.Results[0].ID != "d-2-1" {
		t.Errorf("second page first ID = %q, want d-2-1", second.Results[0].ID)
	}

	third, err := second.NextPage(ctx)
	if err != nil {
		t.Fatalf("NextPage: %v", err)
	}
	if third.HasNextPage() {
		t.Error("third page HasNextPage() = true, want false")
	}
	last, err := third.NextPage(ctx)
	if err != nil || last != nil {
		t.Errorf("NextPage past the end = (%v, %v), want (nil, nil)", last, err)
	}
}

func TestListAllCrossesPages(t *testing.T) {
	client := newTestClient(t, pagedProxyHandler(t))
	var ids []string
	for proxy, err := range client.Proxies.ListAll(context.Background(), ProxyListParams{Mode: ModeDirect}) {
		if err != nil {
			t.Fatalf("ListAll: %v", err)
		}
		ids = append(ids, proxy.ID)
	}
	want := []string{"d-1-1", "d-1-2", "d-2-1", "d-2-2", "d-3-1", "d-3-2"}
	if len(ids) != len(want) {
		t.Fatalf("ids = %v, want %v", ids, want)
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("ids = %v, want %v", ids, want)
		}
	}
}

func TestListAllIsLazyAndStoppable(t *testing.T) {
	requests := 0
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		pagedProxyHandler(t).ServeHTTP(w, r)
	}))
	seq := client.Proxies.ListAll(context.Background(), ProxyListParams{Mode: ModeDirect})
	if requests != 0 {
		t.Fatalf("requests before iteration = %d, want 0 (lazy)", requests)
	}
	for _, err := range seq {
		if err != nil {
			t.Fatalf("ListAll: %v", err)
		}
		break // stop after the first item
	}
	if requests != 1 {
		t.Errorf("requests after early break = %d, want 1", requests)
	}
}

func TestListAllHonorsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := newTestClient(t, pagedProxyHandler(t))
	var count int
	var lastErr error
	for _, err := range client.Proxies.ListAll(ctx, ProxyListParams{Mode: ModeDirect}) {
		if err != nil {
			lastErr = err
			break
		}
		count++
		if count == 2 {
			cancel() // cancel after consuming the first page
		}
	}
	if lastErr == nil {
		t.Fatal("expected an error after cancelling the context mid-iteration")
	}
	if count != 2 {
		t.Errorf("items consumed = %d, want 2", count)
	}
}

func TestStartingAfterPagination(t *testing.T) {
	// The proxy activity list paginates with starting_after; the SDK follows
	// the envelope next URL verbatim, whatever parameters it carries.
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/proxy/activity/" {
			t.Errorf("path = %q, want /api/v2/proxy/activity/", r.URL.Path)
		}
		if r.URL.Query().Get("starting_after") == "" {
			writeJSON(t, w, http.StatusOK, map[string]any{
				"count":    2,
				"next":     "http://" + r.Host + "/api/v2/proxy/activity/?starting_after=2024-01-01T00%3A00%3A00Z",
				"previous": nil,
				"results":  []map[string]any{{"timestamp": "2024-01-01T00:00:00Z", "protocol": "http"}},
			})
			return
		}
		writeJSON(t, w, http.StatusOK, map[string]any{
			"count":    2,
			"next":     nil,
			"previous": nil,
			"results":  []map[string]any{{"timestamp": "2024-01-02T00:00:00Z", "protocol": "socks"}},
		})
	}))
	var protocols []string
	for activity, err := range client.ProxyActivity.ListAll(context.Background(), ProxyActivityListParams{}) {
		if err != nil {
			t.Fatalf("ListAll: %v", err)
		}
		protocols = append(protocols, activity.Protocol)
	}
	if len(protocols) != 2 || protocols[0] != "http" || protocols[1] != "socks" {
		t.Errorf("protocols = %v, want [http socks]", protocols)
	}
}
