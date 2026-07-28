package webshare

import (
	"context"
	"net/http"
	"testing"
	"time"
)

// TestResourceMethodRouting sweeps every thin resource wrapper not covered by
// a dedicated test, asserting each hits the expected method and path. A typo
// in a wrapper's route fails here instead of shipping invisibly.
func TestResourceMethodRouting(t *testing.T) {
	const envelope = `{"count":0,"next":null,"previous":null,"results":[]}`
	now := time.Now()

	tests := []struct {
		name   string
		method string
		path   string
		body   string
		call   func(ctx context.Context, c *Client) error
	}{
		{"Proxies.Refresh", http.MethodPost, "/api/v2/proxy/list/refresh/", ``, func(ctx context.Context, c *Client) error {
			return c.Proxies.Refresh(ctx, ProxyRefreshParams{})
		}},
		{"ProxyConfig.GetStats", http.MethodGet, "/api/v3/proxy/list/stats", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.ProxyConfig.GetStats(ctx, 5)
			return err
		}},
		{"ProxyConfig.GetStatus", http.MethodGet, "/api/v3/proxy/list/status", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.ProxyConfig.GetStatus(ctx, 5)
			return err
		}},
		{"ProxyConfig.AllocateUnallocatedCountries", http.MethodPost, "/api/v2/proxy/config/allocate_unallocated_countries/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.ProxyConfig.AllocateUnallocatedCountries(ctx, AllocateUnallocatedCountriesParams{NewCountries: map[string]int{"US": 1}})
			return err
		}},
		{"ProxyReplacements.List", http.MethodGet, "/api/v3/proxy/replace/", envelope, func(ctx context.Context, c *Client) error {
			_, err := c.ProxyReplacements.List(ctx, ProxyReplacementListParams{})
			return err
		}},
		{"ReplacedProxies.List", http.MethodGet, "/api/v2/proxy/list/replaced/", envelope, func(ctx context.Context, c *Client) error {
			_, err := c.ReplacedProxies.List(ctx, ReplacedProxyListParams{})
			return err
		}},
		{"Stats.Aggregate", http.MethodGet, "/api/v2/stats/aggregate/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Stats.Aggregate(ctx, StatsListParams{})
			return err
		}},
		{"ProxyActivity.List", http.MethodGet, "/api/v2/proxy/activity/", envelope, func(ctx context.Context, c *Client) error {
			_, err := c.ProxyActivity.List(ctx, ProxyActivityListParams{})
			return err
		}},
		{"ProxyActivity.Download", http.MethodGet, "/api/v2/proxy/activity/download/", `Time,Hostname`, func(ctx context.Context, c *Client) error {
			_, err := c.ProxyActivity.Download(ctx, ProxyActivityDownloadParams{DownloadToken: "tok"})
			return err
		}},
		{"DownloadTokens.Reset", http.MethodPost, "/api/v2/download_token/activity/reset/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.DownloadTokens.Reset(ctx, ScopeActivity)
			return err
		}},
		{"IPAuthorizations.List", http.MethodGet, "/api/v2/proxy/ipauthorization/", envelope, func(ctx context.Context, c *Client) error {
			_, err := c.IPAuthorizations.List(ctx, IPAuthorizationListParams{})
			return err
		}},
		{"IPAuthorizations.Get", http.MethodGet, "/api/v2/proxy/ipauthorization/3/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.IPAuthorizations.Get(ctx, 3, IPAuthorizationGetParams{})
			return err
		}},
		{"IPAuthorizations.Delete", http.MethodDelete, "/api/v2/proxy/ipauthorization/3/", ``, func(ctx context.Context, c *Client) error {
			return c.IPAuthorizations.Delete(ctx, 3, IPAuthorizationGetParams{})
		}},
		{"Subusers.List", http.MethodGet, "/api/v2/subuser/", envelope, func(ctx context.Context, c *Client) error {
			_, err := c.Subusers.List(ctx, SubuserListParams{})
			return err
		}},
		{"Subusers.Get", http.MethodGet, "/api/v2/subuser/5/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Subusers.Get(ctx, 5, SubuserGetParams{})
			return err
		}},
		{"Subusers.Update", http.MethodPatch, "/api/v2/subuser/5/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Subusers.Update(ctx, 5, SubuserUpdateParams{Label: String("l")})
			return err
		}},
		{"Subusers.Delete", http.MethodDelete, "/api/v2/subuser/5/", ``, func(ctx context.Context, c *Client) error {
			return c.Subusers.Delete(ctx, 5, SubuserGetParams{})
		}},
		{"Subusers.RefreshProxyList", http.MethodPost, "/api/v2/subuser/5/refresh/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Subusers.RefreshProxyList(ctx, 5)
			return err
		}},
		{"Profile.Update", http.MethodPatch, "/api/v2/profile/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Profile.Update(ctx, ProfileUpdateParams{Timezone: String("UTC")})
			return err
		}},
		{"Profile.GetPreferences", http.MethodGet, "/api/v2/profile/preferences/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Profile.GetPreferences(ctx)
			return err
		}},
		{"Profile.UpdatePreferences", http.MethodPatch, "/api/v2/profile/preferences/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Profile.UpdatePreferences(ctx, ProfilePreferencesUpdateParams{OnboardingActivityPageViewedAt: Time(now)})
			return err
		}},
		{"Notifications.List", http.MethodGet, "/api/v2/notification/", envelope, func(ctx context.Context, c *Client) error {
			_, err := c.Notifications.List(ctx, NotificationListParams{})
			return err
		}},
		{"Notifications.Get", http.MethodGet, "/api/v2/notification/13/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Notifications.Get(ctx, 13)
			return err
		}},
		{"Notifications.Restore", http.MethodPost, "/api/v2/notification/13/restore/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Notifications.Restore(ctx, 13)
			return err
		}},
		{"Verification.GetSuspension", http.MethodGet, "/api/v2/verification/suspension/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Verification.GetSuspension(ctx)
			return err
		}},
		{"Verification.GetLimits", http.MethodGet, "/api/v2/verification/limits/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Verification.GetLimits(ctx)
			return err
		}},
		{"Verification.GetThresholds", http.MethodGet, "/api/v2/verification/thresholds/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Verification.GetThresholds(ctx)
			return err
		}},
		{"Verification.Flows.List", http.MethodGet, "/api/v2/verification/flow/", envelope, func(ctx context.Context, c *Client) error {
			_, err := c.Verification.Flows.List(ctx, VerificationFlowListParams{})
			return err
		}},
		{"Verification.Flows.Get", http.MethodGet, "/api/v2/verification/flow/4/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Verification.Flows.Get(ctx, 4)
			return err
		}},
		{"Verification.Flows.SubmitSecurityCode", http.MethodPost, "/api/v2/verification/flow/4/submit_verification_code/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Verification.Flows.SubmitSecurityCode(ctx, 4, SubmitSecurityCodeParams{SecurityCode: "AB"})
			return err
		}},
		{"Verification.Questions.List", http.MethodGet, "/api/v2/verification/question/", envelope, func(ctx context.Context, c *Client) error {
			_, err := c.Verification.Questions.List(ctx, VerificationQuestionListParams{})
			return err
		}},
		{"Verification.Appeals.List", http.MethodGet, "/api/v2/verification/appeal/", envelope, func(ctx context.Context, c *Client) error {
			_, err := c.Verification.Appeals.List(ctx, VerificationAppealListParams{})
			return err
		}},
		{"Verification.Appeals.Create", http.MethodPost, "/api/v2/verification/appeal/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Verification.Appeals.Create(ctx, VerificationAppealCreateParams{Appeal: "please"})
			return err
		}},
		{"Verification.AbuseReports.List", http.MethodGet, "/api/v2/verification/abuse_report/", envelope, func(ctx context.Context, c *Client) error {
			_, err := c.Verification.AbuseReports.List(ctx, AbuseReportListParams{})
			return err
		}},
		{"Billing.UpdateInfo", http.MethodPatch, "/api/v2/subscription/billing_info/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Billing.UpdateInfo(ctx, BillingInfoUpdateParams{Name: String("n")})
			return err
		}},
		{"PaymentMethods.Get", http.MethodGet, "/api/v2/payment/method/2/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.PaymentMethods.Get(ctx, 2)
			return err
		}},
		{"PendingPayments.List", http.MethodGet, "/api/v2/payment/pending/", envelope, func(ctx context.Context, c *Client) error {
			_, err := c.PendingPayments.List(ctx, PendingPaymentListParams{})
			return err
		}},
		{"Transactions.List", http.MethodGet, "/api/v2/payment/transaction/", envelope, func(ctx context.Context, c *Client) error {
			_, err := c.Transactions.List(ctx, TransactionListParams{})
			return err
		}},
		{"Subscription.Get", http.MethodGet, "/api/v2/subscription/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Subscription.Get(ctx)
			return err
		}},
		{"Subscription.GetAvailableAssets", http.MethodGet, "/api/v2/subscription/available_assets/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Subscription.GetAvailableAssets(ctx)
			return err
		}},
		{"Subscription.Customize", http.MethodGet, "/api/v2/subscription/customize/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Subscription.Customize(ctx, SubscriptionCustomizeParams{})
			return err
		}},
		{"Subscription.EnableAutoRenewal", http.MethodPut, "/api/v2/subscription/renewal/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Subscription.EnableAutoRenewal(ctx)
			return err
		}},
		{"Plans.List", http.MethodGet, "/api/v2/subscription/plan/", envelope, func(ctx context.Context, c *Client) error {
			_, err := c.Plans.List(ctx, PlanListParams{})
			return err
		}},
		{"Plans.Update", http.MethodPatch, "/api/v2/subscription/plan/2/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Plans.Update(ctx, 2, PlanUpdateParams{})
			return err
		}},
		{"Plans.Cancel", http.MethodPost, "/api/v2/subscription/plan/2/cancel/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Plans.Cancel(ctx, 2)
			return err
		}},
		{"Referral.GetConfig", http.MethodGet, "/api/v2/referral/config/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Referral.GetConfig(ctx)
			return err
		}},
		{"Referral.UpdateConfig", http.MethodPatch, "/api/v2/referral/config/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Referral.UpdateConfig(ctx, ReferralConfigUpdateParams{Mode: ReferralModeCredits})
			return err
		}},
		{"Referral.GetCouponCode", http.MethodGet, "/api/v2/referral/coupon-code/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Referral.GetCouponCode(ctx)
			return err
		}},
		{"Referral.RemoveCouponCode", http.MethodDelete, "/api/v2/referral/coupon-code/", ``, func(ctx context.Context, c *Client) error {
			return c.Referral.RemoveCouponCode(ctx)
		}},
		{"Referral.ListChannels", http.MethodGet, "/api/v2/referral/channel/", `[]`, func(ctx context.Context, c *Client) error {
			_, err := c.Referral.ListChannels(ctx)
			return err
		}},
		{"Referral.ListCredits", http.MethodGet, "/api/v2/referral/credit/", envelope, func(ctx context.Context, c *Client) error {
			_, err := c.Referral.ListCredits(ctx, ReferralCreditListParams{})
			return err
		}},
		{"Referral.GetCredit", http.MethodGet, "/api/v2/referral/credit/8/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Referral.GetCredit(ctx, 8)
			return err
		}},
		{"Referral.ListEarnouts", http.MethodGet, "/api/v2/referral/earnout/", envelope, func(ctx context.Context, c *Client) error {
			_, err := c.Referral.ListEarnouts(ctx, ReferralEarnoutListParams{})
			return err
		}},
		{"Referral.GetEarnout", http.MethodGet, "/api/v2/referral/earnout/8/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Referral.GetEarnout(ctx, 8)
			return err
		}},
		{"Referral.GetCodeInfo", http.MethodGet, "/api/v2/referral/code/info/", `{}`, func(ctx context.Context, c *Client) error {
			_, err := c.Referral.GetCodeInfo(ctx, "abc")
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requireRequest(t, r, tt.method, tt.path)
				w.Header().Set("Content-Type", "application/json")
				if tt.body == "" {
					w.WriteHeader(http.StatusNoContent)
					return
				}
				if _, err := w.Write([]byte(tt.body)); err != nil {
					t.Errorf("writing response: %v", err)
				}
			}))
			if err := tt.call(context.Background(), client); err != nil {
				t.Fatalf("%s: %v", tt.name, err)
			}
		})
	}
}
