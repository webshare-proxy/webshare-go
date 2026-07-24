// Package webshare provides the official Go SDK for the Webshare proxy API
// (https://apidocs.webshare.io).
//
// Construct a client with NewClient and call methods on its resource
// services:
//
//	client, err := webshare.NewClient(webshare.WithAPIKey("..."))
//	if err != nil {
//		log.Fatal(err)
//	}
//	page, err := client.Proxies.List(ctx, webshare.ProxyListParams{Mode: webshare.ModeDirect})
//
// When no credential option is given, the client reads the WEBSHARE_API_KEY
// environment variable.
package webshare

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Version is the SDK version, reported in the User-Agent header.
const Version = "0.1.0"

// DefaultBaseURL is the default API base URL. It is the bare host: every
// operation path carries its full /api/vN/... prefix.
const DefaultBaseURL = "https://proxy.webshare.io"

const (
	defaultTimeout    = 60 * time.Second
	defaultMaxRetries = 2
)

// Client is the Webshare API client. Create one with NewClient. All API
// operations are grouped into resource services exposed as fields.
type Client struct {
	cfg requestConfig

	// Proxies exposes the proxy list operations.
	Proxies *ProxiesService
	// ProxyConfig exposes the proxy configuration operations.
	ProxyConfig *ProxyConfigService
	// ProxyReplacements exposes the proxy replacement operations.
	ProxyReplacements *ProxyReplacementsService
	// ReplacedProxies exposes the replaced proxy operations.
	ReplacedProxies *ReplacedProxiesService
	// Stats exposes the proxy usage statistics operations.
	Stats *StatsService
	// ProxyActivity exposes the proxy activity operations.
	ProxyActivity *ProxyActivityService
	// DownloadTokens exposes the download token operations.
	DownloadTokens *DownloadTokensService
	// IPAuthorizations exposes the IP authorization operations.
	IPAuthorizations *IPAuthorizationsService
	// Subusers exposes the sub-user operations.
	Subusers *SubusersService
	// APIKeys exposes the API key operations.
	APIKeys *APIKeysService
	// Profile exposes the user profile operations.
	Profile *ProfileService
	// Notifications exposes the notification operations.
	Notifications *NotificationsService
	// Auth exposes session and account lifecycle operations.
	Auth *AuthService
	// TwoFactorAuth exposes the two-factor authentication operations.
	TwoFactorAuth *TwoFactorAuthService
	// IDVerification exposes the ID verification operations.
	IDVerification *IDVerificationService
	// Verification exposes the account verification operations.
	Verification *VerificationService
	// Billing exposes the billing information operations.
	Billing *BillingService
	// PaymentMethods exposes the payment method operations.
	PaymentMethods *PaymentMethodsService
	// PendingPayments exposes the pending payment operations.
	PendingPayments *PendingPaymentsService
	// Transactions exposes the transaction operations.
	Transactions *TransactionsService
	// Subscription exposes the subscription operations.
	Subscription *SubscriptionService
	// Plans exposes the plan operations.
	Plans *PlansService
	// Invoices exposes the invoice operations.
	Invoices *InvoicesService
	// Referral exposes the referral and coupon code operations.
	Referral *ReferralService
}

// NewClient creates a Webshare API client. Credentials are taken from
// WithAPIKey or WithTokenSource; when neither is given, the WEBSHARE_API_KEY
// environment variable is used. NewClient returns an error when no credential
// is available (unless WithUnauthenticated is passed) or an option is
// invalid.
func NewClient(opts ...RequestOption) (*Client, error) {
	base, _ := url.Parse(DefaultBaseURL)
	cfg := requestConfig{
		baseURL:    base,
		httpClient: &http.Client{},
		maxRetries: defaultMaxRetries,
		timeout:    defaultTimeout,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.err != nil {
		return nil, cfg.err
	}
	if cfg.tokenSource == nil && !cfg.unauthenticated {
		key := os.Getenv("WEBSHARE_API_KEY")
		if key == "" {
			return nil, errors.New("webshare: missing credentials: pass webshare.WithAPIKey or webshare.WithTokenSource, set the WEBSHARE_API_KEY environment variable, or opt out with webshare.WithUnauthenticated")
		}
		cfg.tokenSource = StaticTokenSource(Token{Value: key})
	}

	c := &Client{cfg: cfg}
	c.Proxies = &ProxiesService{client: c}
	c.ProxyConfig = &ProxyConfigService{client: c}
	c.ProxyReplacements = &ProxyReplacementsService{client: c}
	c.ReplacedProxies = &ReplacedProxiesService{client: c}
	c.Stats = &StatsService{client: c}
	c.ProxyActivity = &ProxyActivityService{client: c}
	c.DownloadTokens = &DownloadTokensService{client: c}
	c.IPAuthorizations = &IPAuthorizationsService{client: c}
	c.Subusers = &SubusersService{client: c}
	c.APIKeys = &APIKeysService{client: c}
	c.Profile = &ProfileService{client: c}
	c.Notifications = &NotificationsService{client: c}
	c.Auth = &AuthService{client: c}
	c.TwoFactorAuth = &TwoFactorAuthService{client: c}
	c.IDVerification = &IDVerificationService{client: c}
	c.Verification = &VerificationService{
		Flows:        &VerificationFlowsService{client: c},
		Questions:    &VerificationQuestionsService{client: c},
		Appeals:      &VerificationAppealsService{client: c},
		AbuseReports: &VerificationAbuseReportsService{client: c},
		client:       c,
	}
	c.Billing = &BillingService{client: c}
	c.PaymentMethods = &PaymentMethodsService{client: c}
	c.PendingPayments = &PendingPaymentsService{client: c}
	c.Transactions = &TransactionsService{client: c}
	c.Subscription = &SubscriptionService{client: c}
	c.Plans = &PlansService{client: c}
	c.Invoices = &InvoicesService{client: c}
	c.Referral = &ReferralService{client: c}
	return c, nil
}
