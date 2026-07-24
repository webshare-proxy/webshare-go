package webshare

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// requireRequest asserts the method and path of an incoming request.
func requireRequest(t *testing.T, r *http.Request, method, path string) {
	t.Helper()
	if r.Method != method {
		t.Errorf("method = %s, want %s", r.Method, method)
	}
	if r.URL.Path != path {
		t.Errorf("path = %q, want %q", r.URL.Path, path)
	}
}

func TestProxiesListQueryEncoding(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodGet, "/api/v2/proxy/list/")
		q := r.URL.Query()
		if q.Get("mode") != "backbone" {
			t.Errorf("mode = %q, want backbone", q.Get("mode"))
		}
		if q.Get("country_code__in") != "US,FR" {
			t.Errorf("country_code__in = %q, want US,FR", q.Get("country_code__in"))
		}
		if q.Get("valid") != "true" {
			t.Errorf("valid = %q, want true", q.Get("valid"))
		}
		if q.Has("search") {
			t.Error("unset search parameter was sent")
		}
		writeJSON(t, w, http.StatusOK, map[string]any{"count": 0, "next": nil, "previous": nil, "results": []any{}})
	}))
	_, err := client.Proxies.List(context.Background(), ProxyListParams{
		Mode:          ModeBackbone,
		CountryCodeIn: []string{"US", "FR"},
		Valid:         Bool(true),
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestProxiesDownloadText(t *testing.T) {
	const body = "10.1.2.3:9421:username:password\n10.1.2.4:6511:username:password\n"
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodGet, "/api/v2/proxy/list/download/tok/-/any/username/direct/-/")
		w.Header().Set("Content-Type", "text/plain")
		if _, err := io.WriteString(w, body); err != nil {
			t.Errorf("writing response: %v", err)
		}
	}))
	got, err := client.Proxies.Download(context.Background(), ProxyDownloadParams{
		Token:                "tok",
		AuthenticationMethod: AuthMethodUsername,
		EndpointMode:         ModeDirect,
	})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if got != body {
		t.Errorf("Download = %q, want %q", got, body)
	}
}

func TestProxyConfigGetV3NoTrailingSlash(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodGet, "/api/v3/proxy/config")
		if r.URL.Query().Get("plan_id") != "5" {
			t.Errorf("plan_id = %q, want 5", r.URL.Query().Get("plan_id"))
		}
		writeJSON(t, w, http.StatusOK, map[string]any{
			"request_timeout":           86400,
			"proxy_list_download_token": "aa87",
		})
	}))
	cfg, err := client.ProxyConfig.Get(context.Background(), 5)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if cfg.RequestTimeout != 86400 || cfg.ProxyListDownloadToken != "aa87" {
		t.Errorf("unexpected config: %+v", cfg)
	}
}

func TestProxyConfigUpdateWithASNsAndNull(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodPatch, "/api/v2/proxy/config/")
		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		if string(body["username"]) != `"new_username"` {
			t.Errorf("username = %s, want \"new_username\"", body["username"])
		}
		if string(body["ip_authorization_city"]) != "null" {
			t.Errorf("ip_authorization_city = %s, want explicit null", body["ip_authorization_city"])
		}
		if _, present := body["password"]; present {
			t.Error("unset password field was sent")
		}
		writeJSON(t, w, http.StatusOK, map[string]any{
			"id":       1,
			"state":    "completed",
			"username": "new_username",
			"asns":     map[string]any{"6137": []any{"ASN NAME", 105}},
		})
	}))
	cfg, err := client.ProxyConfig.Update(context.Background(), ProxyConfigUpdateParams{
		Username:            String("new_username"),
		IPAuthorizationCity: Null[string](),
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got := cfg.ASNs["6137"]; got.Name != "ASN NAME" || got.Count != 105 {
		t.Errorf("ASNs[6137] = %+v, want {ASN NAME 105}", got)
	}
}

func TestProxyReplacementCreateAndPoll(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v3/proxy/replace/":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decoding body: %v", err)
			}
			if body["dry_run"] != true {
				t.Errorf("dry_run = %v, want true", body["dry_run"])
			}
			writeJSON(t, w, http.StatusOK, map[string]any{"id": 98315, "state": "validating", "dry_run": true})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v3/proxy/replace/98315/":
			writeJSON(t, w, http.StatusOK, map[string]any{
				"id": 98315, "state": "validated", "dry_run": true,
				"proxies_removed": 1, "proxies_added": 1,
			})
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	ctx := context.Background()
	created, err := client.ProxyReplacements.Create(ctx, ProxyReplacementCreateParams{
		ToReplace:   ReplacementTarget{Type: ReplacementTargetIPRange, IPRanges: []string{"1.2.3.0/24"}},
		ReplaceWith: []ReplacementTarget{{Type: ReplacementTargetCountry, CountryCode: "US"}},
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.State != ReplacementValidating {
		t.Errorf("state = %q, want validating", created.State)
	}
	polled, err := client.ProxyReplacements.Get(ctx, created.ID, ProxyReplacementGetParams{})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if polled.State != ReplacementValidated || *polled.ProxiesRemoved != 1 {
		t.Errorf("unexpected replacement: %+v", polled)
	}
}

func TestReplacedProxiesDownloadRequiresToken(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodGet, "/api/v2/proxy/list/replaced/download/")
		if r.URL.Query().Get("download_token") != "key123" {
			t.Errorf("download_token = %q, want key123", r.URL.Query().Get("download_token"))
		}
		if _, err := io.WriteString(w, "10.1.2.3:9421:u:p:10.1.2.7\n"); err != nil {
			t.Errorf("writing response: %v", err)
		}
	}))
	if _, err := client.ReplacedProxies.Download(context.Background(), ReplacedProxyDownloadParams{}); err == nil {
		t.Fatal("expected error without a download token")
	}
	got, err := client.ReplacedProxies.Download(context.Background(), ReplacedProxyDownloadParams{DownloadToken: "key123"})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if !strings.HasPrefix(got, "10.1.2.3:9421") {
		t.Errorf("unexpected body: %q", got)
	}
}

func TestStatsListBareArray(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodGet, "/api/v2/stats/")
		writeJSON(t, w, http.StatusOK, []map[string]any{{
			"timestamp":       "2022-08-11T17:00:00-07:00",
			"bandwidth_total": 5000,
			"requests_total":  5,
		}})
	}))
	stats, err := client.Stats.List(context.Background(), StatsListParams{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(stats) != 1 || stats[0].BandwidthTotal != 5000 {
		t.Errorf("unexpected stats: %+v", stats)
	}
}

func TestDownloadTokensGet(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodGet, "/api/v2/download_token/activity/")
		writeJSON(t, w, http.StatusOK, map[string]any{
			"id": 56, "key": "abc", "scope": "activity",
			"expire_at": "2022-06-14T11:58:10.246406-07:00",
		})
	}))
	token, err := client.DownloadTokens.Get(context.Background(), ScopeActivity)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if token.Key != "abc" || token.Scope != ScopeActivity {
		t.Errorf("unexpected token: %+v", token)
	}
}

func TestWhatsMyIP(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodGet, "/api/v2/proxy/ipauthorization/whatsmyip/")
		writeJSON(t, w, http.StatusOK, map[string]any{"ip_address": "1.2.3.4"})
	}))
	result, err := client.IPAuthorizations.WhatsMyIP(context.Background())
	if err != nil {
		t.Fatalf("WhatsMyIP: %v", err)
	}
	if result.IPAddress != "1.2.3.4" {
		t.Errorf("IPAddress = %q, want 1.2.3.4", result.IPAddress)
	}
}

func TestSubuserCreate(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodPost, "/api/v2/subuser/")
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		if body["label"] != "newcustomer" || body["proxy_limit"] != float64(10) {
			t.Errorf("unexpected body: %v", body)
		}
		writeJSON(t, w, http.StatusOK, map[string]any{
			"id": 7, "label": "newcustomer", "proxy_limit": 10.0,
			"aggregate_stats": map[string]any{"bandwidth_total": 5000},
		})
	}))
	subuser, err := client.Subusers.Create(context.Background(), SubuserCreateParams{
		Label:      "newcustomer",
		ProxyLimit: Float64(10),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if subuser.ID != 7 || subuser.AggregateStats.BandwidthTotal != 5000 {
		t.Errorf("unexpected subuser: %+v", subuser)
	}
}

func TestNotificationDismiss(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodPost, "/api/v2/notification/13/dismiss/")
		writeJSON(t, w, http.StatusOK, map[string]any{
			"id": 13, "type": "reminder_to_use_proxy",
			"dismissed_at": "2022-06-14T11:58:10.246406-07:00",
		})
	}))
	notification, err := client.Notifications.Dismiss(context.Background(), 13)
	if err != nil {
		t.Fatalf("Dismiss: %v", err)
	}
	if notification.DismissedAt == nil {
		t.Error("DismissedAt = nil, want a timestamp")
	}
}

func TestAuthGetActivation(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodGet, "/api/v2/activation/")
		writeJSON(t, w, http.StatusOK, map[string]any{
			"email_is_verified":                       true,
			"last_time_email_verification_email_sent": nil,
			"created_at":                              "2019-05-09T23:34:00.095501-07:00",
			"updated_at":                              "2019-05-09T23:34:00.095501-07:00",
		})
	}))
	status, err := client.Auth.GetActivation(context.Background())
	if err != nil {
		t.Fatalf("GetActivation: %v", err)
	}
	if !status.EmailIsVerified || status.LastTimeEmailVerificationEmailSent != nil {
		t.Errorf("unexpected status: %+v", status)
	}
}

func TestTwoFactorChangeMethodSecretKey(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodPost, "/api/v2/twofactorauth/method/")
		writeJSON(t, w, http.StatusOK, map[string]any{
			"id": 137, "type": "device_totp", "active": false, "secret_key": "deadbeef",
		})
	}))
	method, err := client.TwoFactorAuth.ChangeMethod(context.Background(), TwoFactorMethodChangeParams{Type: MethodDeviceTOTP})
	if err != nil {
		t.Fatalf("ChangeMethod: %v", err)
	}
	if method.SecretKey != "deadbeef" {
		t.Errorf("SecretKey = %q, want deadbeef", method.SecretKey)
	}
}

func TestIDVerificationGet(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodGet, "/api/v2/idverification/")
		writeJSON(t, w, http.StatusOK, map[string]any{"id": 1, "state": "not-required", "client_secret": nil})
	}))
	verification, err := client.IDVerification.Get(context.Background())
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if verification.State != IDVerificationNotRequired {
		t.Errorf("State = %q, want not-required", verification.State)
	}
}

func TestSubmitEvidenceMultipart(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodPost, "/api/v2/verification/flow/1/submit_evidence/")
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("parsing multipart form: %v", err)
		}
		if got := r.FormValue("explanation"); got != "my explanation" {
			t.Errorf("explanation = %q, want %q", got, "my explanation")
		}
		files := r.MultipartForm.File["files"]
		if len(files) != 2 {
			t.Fatalf("len(files) = %d, want 2", len(files))
		}
		if files[0].Filename != "evidence.txt" {
			t.Errorf("first filename = %q, want evidence.txt", files[0].Filename)
		}
		f, err := files[0].Open()
		if err != nil {
			t.Fatalf("opening file: %v", err)
		}
		defer f.Close()
		content, _ := io.ReadAll(f)
		if string(content) != "file contents" {
			t.Errorf("file content = %q, want %q", content, "file contents")
		}
		writeJSON(t, w, http.StatusOK, map[string]any{"id": 1, "state": "inflow"})
	}))
	flow, err := client.Verification.Flows.SubmitEvidence(context.Background(), 1, SubmitEvidenceParams{
		Explanation: "my explanation",
		Files: []File{
			{Name: "evidence.txt", Reader: strings.NewReader("file contents")},
			{Name: "evidence2.txt", Reader: strings.NewReader("more contents")},
		},
	})
	if err != nil {
		t.Fatalf("SubmitEvidence: %v", err)
	}
	if flow.ID != 1 {
		t.Errorf("flow.ID = %d, want 1", flow.ID)
	}
}

func TestVerificationCategoriesMap(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodGet, "/api/v2/verification/categories/")
		writeJSON(t, w, http.StatusOK, map[string]any{
			"requests_to_financial_institutions": map[string]any{
				"description":              "ID verification needed.",
				"id_verification_required": true,
			},
		})
	}))
	categories, err := client.Verification.GetCategories(context.Background())
	if err != nil {
		t.Fatalf("GetCategories: %v", err)
	}
	category, ok := categories["requests_to_financial_institutions"]
	if !ok || !category.IDVerificationRequired {
		t.Errorf("unexpected categories: %+v", categories)
	}
}

func TestBillingGetInfo(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodGet, "/api/v2/subscription/billing_info/")
		writeJSON(t, w, http.StatusOK, map[string]any{"id": 1, "name": "Webshare Software"})
	}))
	info, err := client.Billing.GetInfo(context.Background())
	if err != nil {
		t.Fatalf("GetInfo: %v", err)
	}
	if info.Name != "Webshare Software" {
		t.Errorf("Name = %q, want Webshare Software", info.Name)
	}
}

func TestPaymentMethodsListPolymorphic(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodGet, "/api/v2/payment/method/")
		writeJSON(t, w, http.StatusOK, map[string]any{
			"count": 2, "next": nil, "previous": nil,
			"results": []map[string]any{
				{"id": 1, "type": "StripeCard", "brand": "visa", "last4": "4242"},
				{"id": 2, "type": "LinkPayment"},
			},
		})
	}))
	page, err := client.PaymentMethods.List(context.Background(), PaymentMethodListParams{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if page.Results[0].Type != PaymentMethodStripeCard || page.Results[0].Last4 != "4242" {
		t.Errorf("unexpected card: %+v", page.Results[0])
	}
	if page.Results[1].Type != PaymentMethodLinkPayment || page.Results[1].Last4 != "" {
		t.Errorf("unexpected link payment: %+v", page.Results[1])
	}
}

func TestPendingPaymentGet(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodGet, "/api/v2/payment/pending/1/")
		writeJSON(t, w, http.StatusOK, map[string]any{"id": 1, "status": "successful", "transaction": 5})
	}))
	payment, err := client.PendingPayments.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if payment.Status != PendingPaymentSuccessful || *payment.Transaction != 5 {
		t.Errorf("unexpected payment: %+v", payment)
	}
}

func TestTransactionGet(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodGet, "/api/v2/payment/transaction/1/")
		writeJSON(t, w, http.StatusOK, map[string]any{
			"id": 1, "status": "completed", "amount": 1.0,
			"payment_method": map[string]any{"id": 1, "brand": "visa"},
		})
	}))
	transaction, err := client.Transactions.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if transaction.PaymentMethod.Brand != "visa" {
		t.Errorf("unexpected transaction: %+v", transaction)
	}
}

func TestSubscriptionPricingQueryParamEncoding(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodGet, "/api/v2/subscription/pricing/")
		raw := r.URL.Query().Get("query")
		if raw == "" {
			t.Fatal("missing query parameter")
		}
		var decoded map[string]any
		if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
			t.Fatalf("query parameter is not JSON: %v", err)
		}
		if decoded["proxy_type"] != "shared" || decoded["term"] != "monthly" {
			t.Errorf("unexpected query object: %v", decoded)
		}
		if decoded["bandwidth_limit"] != float64(250) {
			t.Errorf("bandwidth_limit = %v, want 250", decoded["bandwidth_limit"])
		}
		if r.URL.Query().Get("plan_id") != "3" {
			t.Errorf("plan_id = %q, want 3", r.URL.Query().Get("plan_id"))
		}
		writeJSON(t, w, http.StatusOK, map[string]any{"price": 13.94, "paid_today": 8.94})
	}))
	pricing, err := client.Subscription.Pricing(context.Background(), SubscriptionPricingParams{
		PlanID: Int(3),
		PlanConfiguration: PlanConfiguration{
			ProxyType:      ProxyTypeShared,
			ProxySubtype:   SubtypeDefault,
			ProxyCountries: map[string]int{"US": 100},
			BandwidthLimit: 250,
			Term:           TermMonthly,
		},
	})
	if err != nil {
		t.Fatalf("Pricing: %v", err)
	}
	if pricing.Price != 13.94 {
		t.Errorf("Price = %v, want 13.94", pricing.Price)
	}
}

func TestCancelAutoRenewalReturnsSubscription(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodDelete, "/api/v2/subscription/renewal/")
		writeJSON(t, w, http.StatusOK, map[string]any{
			"id": 1, "plan": 2, "payment_method": nil, "renewals_enabled": false,
		})
	}))
	subscription, err := client.Subscription.CancelAutoRenewal(context.Background())
	if err != nil {
		t.Fatalf("CancelAutoRenewal: %v", err)
	}
	if subscription.RenewalsEnabled || subscription.PaymentMethod != nil {
		t.Errorf("unexpected subscription: %+v", subscription)
	}
}

func TestPlanGetParsesTimestamps(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodGet, "/api/v2/subscription/plan/2/")
		writeJSON(t, w, http.StatusOK, map[string]any{
			"id": 2, "status": "active",
			"created_at": "2022-06-14T11:58:10.246406-07:00",
		})
	}))
	plan, err := client.Plans.Get(context.Background(), 2)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	want := time.Date(2022, 6, 14, 11, 58, 10, 246406000, plan.CreatedAt.Location())
	if !plan.CreatedAt.Equal(want) {
		t.Errorf("CreatedAt = %v, want %v", plan.CreatedAt, want)
	}
}

func TestInvoiceDownloadPDFBytes(t *testing.T) {
	pdf := []byte("%PDF-1.7 fake")
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodGet, "/api/v2/invoices/download")
		if r.URL.Query().Get("subscription_transaction_id") != "tx-9" {
			t.Errorf("subscription_transaction_id = %q, want tx-9", r.URL.Query().Get("subscription_transaction_id"))
		}
		w.Header().Set("Content-Type", "application/pdf")
		if _, err := w.Write(pdf); err != nil {
			t.Errorf("writing response: %v", err)
		}
	}))
	got, err := client.Invoices.Download(context.Background(), "tx-9")
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if string(got) != string(pdf) {
		t.Errorf("Download = %q, want %q", got, pdf)
	}
}

func TestApplyCouponCodeFieldErrors(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodPost, "/api/v2/referral/coupon-code/")
		writeJSON(t, w, http.StatusBadRequest, map[string]any{"code": []string{"Invalid promo code."}})
	}))
	_, err := client.Referral.ApplyCouponCode(context.Background(), "NOPE")
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *Error", err)
	}
	if got := apiErr.FieldErrors["code"]; len(got) != 1 || got[0] != "Invalid promo code." {
		t.Errorf("FieldErrors[code] = %v, want [Invalid promo code.]", got)
	}
}
