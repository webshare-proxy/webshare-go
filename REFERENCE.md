# webshare-go reference

Every public method of the SDK, grouped by service, with a link to the
official API documentation for the underlying endpoint. All methods take
`context.Context` first and accept trailing `...RequestOption`; both are
elided from the signatures below for brevity.

## Client construction and options

`webshare.NewClient(opts ...RequestOption) (*Client, error)` builds a client;
credentials come from `WithAPIKey`, `WithTokenSource` or the
`WEBSHARE_API_KEY` environment variable (see
[apidocs.webshare.io](https://apidocs.webshare.io) for how API keys work).
Options, usable on the client and on any individual call:

| Option | Purpose |
|---|---|
| `WithAPIKey(key)` | Authenticate with an API key |
| `WithTokenSource(ts)` | Plug in a custom credential provider (e.g. OAuth) |
| `WithUnauthenticated()` | Build a client without credentials (pre-login flows) |
| `WithBaseURL(url)` | Override the API base URL |
| `WithHTTPClient(hc)` | Use a custom `*http.Client` |
| `WithMaxRetries(n)` | Change the retry budget (default 2) |
| `WithRetryNonIdempotent()` | Opt POST/PATCH requests into retries |
| `WithTimeout(d)` | Per-attempt timeout (default 60s) |
| `WithSubuser(id)` | Act as a sub-user (X-Subuser header) |
| `WithSource(s)` | Replace the X-Webshare-Source identification header |
| `WithFederatedUser(id)` | Admin federated access header |
| `WithHeader(k, v)` | Add an extra request header |

## Errors

| Type | Meaning |
|---|---|
| `*webshare.Error` | API error: `StatusCode`, `Code`, `RequestID`, `Detail`, `FieldErrors`, `RetryAfter`, raw `Body` |
| `*webshare.RequestError` | Transport failure (wraps the underlying error) |
| `*webshare.ResponseDecodeError` | 2xx response body that could not be decoded |
| `*webshare.CrossOriginError` | Refused cross-origin pagination URL |

All are matched with `errors.As`. See the error handling section of the
[README](README.md#error-handling).

## Pagination

`List` methods return `*webshare.Page[T]` (`Results`, `Count`, `Next`,
`Previous`, `HasNextPage()`, `NextPage(ctx)`). `ListAll` methods return a
lazy `iter.Seq2[T, error]` that follows the server's `next` URL across
pages. See the pagination section of the [README](README.md#pagination).

## Helpers

| Function | Purpose | Docs |
|---|---|---|
| `webshare.ProxyURL(ProxyURLParams) (string, error)` | Build direct or backbone proxy connection URLs (country, city, sticky sessions, rotation) | [proxy-connection](https://apidocs.webshare.io/proxy-connection) |
| `client.Proxies.DownloadURL(ProxyDownloadParams) (string, error)` | Build the path-style proxy list download URL | [proxy-list/download](https://apidocs.webshare.io/proxy-list/download) |

## client.Proxies — Proxy list

| Method | Description | Docs |
|---|---|---|
| `List(ctx, ProxyListParams) (*Page[Proxy], error)` | List a plan's proxies (paginated; `Mode` is required). | [proxy-list/list](https://apidocs.webshare.io/proxy-list/list) |
| `ListAll(ctx, ProxyListParams) iter.Seq2[Proxy, error]` | Lazily iterate every proxy across all pages. | [proxy-list/list](https://apidocs.webshare.io/proxy-list/list) |
| `Refresh(ctx, ProxyRefreshParams) error` | Replace the entire proxy list on demand. | [proxy-list/ondemand_refresh](https://apidocs.webshare.io/proxy-list/ondemand_refresh) |
| `Download(ctx, ProxyDownloadParams) (string, error)` | Download the proxy list as `address:port:username:password` text. | [proxy-list/download](https://apidocs.webshare.io/proxy-list/download) |
| `DownloadURL(ProxyDownloadParams) (string, error)` | Build the shareable path-style download URL without a request. | [proxy-list/download](https://apidocs.webshare.io/proxy-list/download) |

## client.ProxyConfig — Proxy configuration

| Method | Description | Docs |
|---|---|---|
| `Get(ctx, planID int) (*ProxyConfigV3, error)` | Read a plan's proxy config, including the download token. | [proxy-config/get_proxy_config](https://apidocs.webshare.io/proxy-config/get_proxy_config) |
| `GetStats(ctx, planID int) (*ProxyListStats, error)` | Read the proxy list composition (countries, ranges, ASNs). | [proxy-config/get_proxy_stats](https://apidocs.webshare.io/proxy-config/get_proxy_stats) |
| `GetStatus(ctx, planID int) (*ProxyListStatus, error)` | Read the proxy list readiness state and credentials. | [proxy-config/get_proxy_status](https://apidocs.webshare.io/proxy-config/get_proxy_status) |
| `Update(ctx, ProxyConfigUpdateParams) (*ProxyConfig, error)` | Partially update the proxy config. | [proxy-config/update](https://apidocs.webshare.io/proxy-config/update) |
| `AllocateUnallocatedCountries(ctx, AllocateUnallocatedCountriesParams) (*ProxyConfig, error)` | Allocate proxies stuck in unallocated countries. | [proxy-config/allocate_unallocated_countries](https://apidocs.webshare.io/proxy-config/allocate_unallocated_countries) |

## client.ProxyReplacements — Proxy replacements

| Method | Description | Docs |
|---|---|---|
| `List(ctx, ProxyReplacementListParams) (*Page[ProxyReplacement], error)` | List proxy replacements. | [proxy-replacement/proxy_replacement/proxy_replacement_list](https://apidocs.webshare.io/proxy-replacement/proxy_replacement/proxy_replacement_list) |
| `ListAll(ctx, ProxyReplacementListParams) iter.Seq2[ProxyReplacement, error]` | Lazily iterate every replacement across all pages. | [proxy-replacement/proxy_replacement/proxy_replacement_list](https://apidocs.webshare.io/proxy-replacement/proxy_replacement/proxy_replacement_list) |
| `Create(ctx, ProxyReplacementCreateParams) (*ProxyReplacement, error)` | Start an asynchronous replacement; poll `Get` until completed. Supports dry runs. | [proxy-replacement/proxy_replacement/proxy_replacement_create](https://apidocs.webshare.io/proxy-replacement/proxy_replacement/proxy_replacement_create) |
| `Get(ctx, id int, ProxyReplacementGetParams) (*ProxyReplacement, error)` | Poll a replacement's state. | [proxy-replacement/proxy_replacement/proxy_replacement_retrieve](https://apidocs.webshare.io/proxy-replacement/proxy_replacement/proxy_replacement_retrieve) |

## client.ReplacedProxies — Replaced proxies

| Method | Description | Docs |
|---|---|---|
| `List(ctx, ReplacedProxyListParams) (*Page[ReplacedProxy], error)` | List proxies that were replaced. | [proxy-replacement/replaced_proxy/list_replaced_proxy](https://apidocs.webshare.io/proxy-replacement/replaced_proxy/list_replaced_proxy) |
| `ListAll(ctx, ReplacedProxyListParams) iter.Seq2[ReplacedProxy, error]` | Lazily iterate every replaced proxy across all pages. | [proxy-replacement/replaced_proxy/list_replaced_proxy](https://apidocs.webshare.io/proxy-replacement/replaced_proxy/list_replaced_proxy) |
| `Download(ctx, ReplacedProxyDownloadParams) (string, error)` | Download the replaced proxy list as text (needs a `replaced_proxy` download token). | [proxy-replacement/replaced_proxy/download](https://apidocs.webshare.io/proxy-replacement/replaced_proxy/download) |

## client.Stats — Usage statistics

| Method | Description | Docs |
|---|---|---|
| `List(ctx, StatsListParams) ([]ProxyStat, error)` | Hourly proxy stats for a period (bare array, not paginated). | [proxystats/list_stats](https://apidocs.webshare.io/proxystats/list_stats) |
| `Aggregate(ctx, StatsListParams) (*AggregateStats, error)` | Aggregate proxy usage for a period. | [proxystats/aggregate](https://apidocs.webshare.io/proxystats/aggregate) |

## client.ProxyActivity — Proxy activity

| Method | Description | Docs |
|---|---|---|
| `List(ctx, ProxyActivityListParams) (*Page[ProxyActivity], error)` | List proxy requests (paginates with `starting_after`). | [proxystats/list_activity](https://apidocs.webshare.io/proxystats/list_activity) |
| `ListAll(ctx, ProxyActivityListParams) iter.Seq2[ProxyActivity, error]` | Lazily iterate every activity across all pages. | [proxystats/list_activity](https://apidocs.webshare.io/proxystats/list_activity) |
| `Download(ctx, ProxyActivityDownloadParams) (string, error)` | Download activities as CSV (needs an `activity` download token). | [proxystats/download_activity](https://apidocs.webshare.io/proxystats/download_activity) |

## client.DownloadTokens — Download tokens

| Method | Description | Docs |
|---|---|---|
| `Get(ctx, DownloadTokenScope) (*DownloadToken, error)` | Fetch the download token for a scope (`proxy_list`, `replaced_proxy`, `activity`). | [downloads/get_download_token](https://apidocs.webshare.io/downloads/get_download_token) |
| `Reset(ctx, DownloadTokenScope) (*DownloadToken, error)` | Rotate the download token for a scope. | [downloads/reset_download_token](https://apidocs.webshare.io/downloads/reset_download_token) |

## client.IPAuthorizations — IP authorizations

| Method | Description | Docs |
|---|---|---|
| `List(ctx, IPAuthorizationListParams) (*Page[IPAuthorization], error)` | List authorized IP addresses. | [ipauthorization/list](https://apidocs.webshare.io/ipauthorization/list) |
| `ListAll(ctx, IPAuthorizationListParams) iter.Seq2[IPAuthorization, error]` | Lazily iterate every IP authorization across all pages. | [ipauthorization/list](https://apidocs.webshare.io/ipauthorization/list) |
| `Create(ctx, IPAuthorizationCreateParams) (*IPAuthorization, error)` | Authorize an IP address. | [ipauthorization/create](https://apidocs.webshare.io/ipauthorization/create) |
| `Get(ctx, id int, IPAuthorizationGetParams) (*IPAuthorization, error)` | Retrieve an IP authorization. | [ipauthorization/retrieve](https://apidocs.webshare.io/ipauthorization/retrieve) |
| `Delete(ctx, id int, IPAuthorizationGetParams) error` | Remove an IP authorization. | [ipauthorization/delete](https://apidocs.webshare.io/ipauthorization/delete) |
| `WhatsMyIP(ctx) (*WhatsMyIPResult, error)` | Return the caller's public IP address. | [ipauthorization/whatsmyip](https://apidocs.webshare.io/ipauthorization/whatsmyip) |

## client.Subusers — Sub-users

| Method | Description | Docs |
|---|---|---|
| `List(ctx, SubuserListParams) (*Page[Subuser], error)` | List sub-users. | [subuser/list](https://apidocs.webshare.io/subuser/list) |
| `ListAll(ctx, SubuserListParams) iter.Seq2[Subuser, error]` | Lazily iterate every sub-user across all pages. | [subuser/list](https://apidocs.webshare.io/subuser/list) |
| `Create(ctx, SubuserCreateParams) (*Subuser, error)` | Create a sub-user. | [subuser/create](https://apidocs.webshare.io/subuser/create) |
| `Get(ctx, id int, SubuserGetParams) (*Subuser, error)` | Retrieve a sub-user. | [subuser/retrieve](https://apidocs.webshare.io/subuser/retrieve) |
| `Update(ctx, id int, SubuserUpdateParams) (*Subuser, error)` | Partially update a sub-user. | [subuser/update](https://apidocs.webshare.io/subuser/update) |
| `Delete(ctx, id int, SubuserGetParams) error` | Delete a sub-user. | [subuser/delete](https://apidocs.webshare.io/subuser/delete) |
| `RefreshProxyList(ctx, id int) (*Subuser, error)` | Refresh a sub-user's custom proxy list. | [subuser/refresh_proxy_list](https://apidocs.webshare.io/subuser/refresh_proxy_list) |

## client.Profile — User profile

| Method | Description | Docs |
|---|---|---|
| `Get(ctx) (*Profile, error)` | Retrieve the user profile. | [userprofile/retrieve](https://apidocs.webshare.io/userprofile/retrieve) |
| `Update(ctx, ProfileUpdateParams) (*Profile, error)` | Partially update the user profile. | [userprofile/update](https://apidocs.webshare.io/userprofile/update) |
| `GetPreferences(ctx) (*ProfilePreferences, error)` | Retrieve the user preferences. | [userprofile/retrivePreferences](https://apidocs.webshare.io/userprofile/retrivePreferences) |
| `UpdatePreferences(ctx, ProfilePreferencesUpdateParams) (*ProfilePreferences, error)` | Partially update the user preferences. | [userprofile/updatePreferences](https://apidocs.webshare.io/userprofile/updatePreferences) |

## client.Notifications — Notifications

| Method | Description | Docs |
|---|---|---|
| `List(ctx, NotificationListParams) (*Page[Notification], error)` | List account notifications. | [notifications/list](https://apidocs.webshare.io/notifications/list) |
| `ListAll(ctx, NotificationListParams) iter.Seq2[Notification, error]` | Lazily iterate every notification across all pages. | [notifications/list](https://apidocs.webshare.io/notifications/list) |
| `Get(ctx, id int) (*Notification, error)` | Retrieve a notification. | [notifications/retrieve](https://apidocs.webshare.io/notifications/retrieve) |
| `Dismiss(ctx, id int) (*Notification, error)` | Dismiss a notification. | [notifications/dismiss](https://apidocs.webshare.io/notifications/dismiss) |
| `Restore(ctx, id int) (*Notification, error)` | Restore a dismissed notification. | [notifications/restore](https://apidocs.webshare.io/notifications/restore) |

## client.IDVerification — ID verification (read-only)

| Method | Description | Docs |
|---|---|---|
| `Get(ctx) (*IDVerification, error)` | Read the ID verification status (read-only; verifications complete via the dashboard). | [idverification/retrieve](https://apidocs.webshare.io/idverification/retrieve) |

## client.Verification.Flows — Account verification: flows

| Method | Description | Docs |
|---|---|---|
| `List(ctx, VerificationFlowListParams) (*Page[VerificationFlow], error)` | List verification flows. | [verification/list](https://apidocs.webshare.io/verification/list) |
| `ListAll(ctx, VerificationFlowListParams) iter.Seq2[VerificationFlow, error]` | Lazily iterate every flow across all pages. | [verification/list](https://apidocs.webshare.io/verification/list) |
| `Get(ctx, id int) (*VerificationFlow, error)` | Retrieve a verification flow. | [verification/retrieve](https://apidocs.webshare.io/verification/retrieve) |
| `SubmitEvidence(ctx, id int, SubmitEvidenceParams) (*VerificationFlow, error)` | Submit evidence with file uploads (multipart). | [verification/submit_evidence](https://apidocs.webshare.io/verification/submit_evidence) |
| `SubmitSecurityCode(ctx, id int, SubmitSecurityCodeParams) (*VerificationFlow, error)` | Submit the bank-statement security code. | [verification/submit_security_code](https://apidocs.webshare.io/verification/submit_security_code) |

## client.Verification.Questions — Account verification: questions

| Method | Description | Docs |
|---|---|---|
| `List(ctx, VerificationQuestionListParams) (*Page[VerificationQuestion], error)` | List compliance questions. | [verification/list_questions](https://apidocs.webshare.io/verification/list_questions) |
| `ListAll(ctx, VerificationQuestionListParams) iter.Seq2[VerificationQuestion, error]` | Lazily iterate every question across all pages. | [verification/list_questions](https://apidocs.webshare.io/verification/list_questions) |
| `SubmitAnswer(ctx, questionID int, SubmitAnswerParams) (*VerificationAnswer, error)` | Answer a question with optional attachments (multipart). | [verification/submit_answer](https://apidocs.webshare.io/verification/submit_answer) |

## client.Verification.Appeals — Account verification: appeals

| Method | Description | Docs |
|---|---|---|
| `List(ctx, VerificationAppealListParams) (*Page[VerificationAppeal], error)` | List suspension appeals. | [verification/list_appeals](https://apidocs.webshare.io/verification/list_appeals) |
| `ListAll(ctx, VerificationAppealListParams) iter.Seq2[VerificationAppeal, error]` | Lazily iterate every appeal across all pages. | [verification/list_appeals](https://apidocs.webshare.io/verification/list_appeals) |
| `Create(ctx, VerificationAppealCreateParams) (*VerificationAppeal, error)` | Submit a suspension appeal (one at a time). | [verification/submit_appeal](https://apidocs.webshare.io/verification/submit_appeal) |

## client.Verification.AbuseReports — Account verification: abuse reports

| Method | Description | Docs |
|---|---|---|
| `List(ctx, AbuseReportListParams) (*Page[AbuseReport], error)` | List abuse reports against the account. | [verification/list_abuse_reports](https://apidocs.webshare.io/verification/list_abuse_reports) |
| `ListAll(ctx, AbuseReportListParams) iter.Seq2[AbuseReport, error]` | Lazily iterate every abuse report across all pages. | [verification/list_abuse_reports](https://apidocs.webshare.io/verification/list_abuse_reports) |

## client.Verification — Account verification: singletons

| Method | Description | Docs |
|---|---|---|
| `GetSuspension(ctx) (*Suspension, error)` | Read when and why the account was suspended (works while suspended). | [verification/view_suspension](https://apidocs.webshare.io/verification/view_suspension) |
| `GetCategories(ctx) (map[string]VerificationCategory, error)` | Read the verification categories (map keyed by category). | [verification/categories](https://apidocs.webshare.io/verification/categories) |
| `GetLimits(ctx) (*VerificationLimits, error)` | Read the current proxy limit state. | [verification/limits](https://apidocs.webshare.io/verification/limits) |
| `GetThresholds(ctx) (map[string]VerificationThreshold, error)` | Read the verification thresholds (map keyed by category). | [verification/thresholds](https://apidocs.webshare.io/verification/thresholds) |

## client.Billing — Billing information

| Method | Description | Docs |
|---|---|---|
| `GetInfo(ctx) (*BillingInfo, error)` | Read the billing information singleton. | [billing/billing](https://apidocs.webshare.io/billing/billing) |
| `UpdateInfo(ctx, BillingInfoUpdateParams) (*BillingInfo, error)` | Update the billing information. | [billing/billing](https://apidocs.webshare.io/billing/billing) |

## client.PaymentMethods — Payment methods

| Method | Description | Docs |
|---|---|---|
| `List(ctx, PaymentMethodListParams) (*Page[PaymentMethod], error)` | List payment methods (polymorphic on `Type`). | [billing/payment_methods](https://apidocs.webshare.io/billing/payment_methods) |
| `ListAll(ctx, PaymentMethodListParams) iter.Seq2[PaymentMethod, error]` | Lazily iterate every payment method across all pages. | [billing/payment_methods](https://apidocs.webshare.io/billing/payment_methods) |
| `Get(ctx, id int) (*PaymentMethod, error)` | Retrieve a payment method. | [billing/payment_methods](https://apidocs.webshare.io/billing/payment_methods) |

## client.PendingPayments — Pending payments

| Method | Description | Docs |
|---|---|---|
| `List(ctx, PendingPaymentListParams) (*Page[PendingPayment], error)` | List pending payments. | [billing/pending_payments](https://apidocs.webshare.io/billing/pending_payments) |
| `ListAll(ctx, PendingPaymentListParams) iter.Seq2[PendingPayment, error]` | Lazily iterate every pending payment across all pages. | [billing/pending_payments](https://apidocs.webshare.io/billing/pending_payments) |
| `Get(ctx, id int) (*PendingPayment, error)` | Poll a pending payment after confirming with Stripe. | [billing/pending_payments](https://apidocs.webshare.io/billing/pending_payments) |

## client.Transactions — Transactions

| Method | Description | Docs |
|---|---|---|
| `List(ctx, TransactionListParams) (*Page[Transaction], error)` | List transactions. | [billing/transactions](https://apidocs.webshare.io/billing/transactions) |
| `ListAll(ctx, TransactionListParams) iter.Seq2[Transaction, error]` | Lazily iterate every transaction across all pages. | [billing/transactions](https://apidocs.webshare.io/billing/transactions) |
| `Get(ctx, id int) (*Transaction, error)` | Retrieve a transaction. | [billing/transactions](https://apidocs.webshare.io/billing/transactions) |

## client.Subscription — Subscription

| Method | Description | Docs |
|---|---|---|
| `Get(ctx) (*Subscription, error)` | Read the account's subscription singleton. | [subscription](https://apidocs.webshare.io/subscription) |
| `GetAvailableAssets(ctx) (map[string]map[string]AssetInfo, error)` | Read the assets available per proxy category and subtype. | [subscription/assets](https://apidocs.webshare.io/subscription/assets) |
| `Customize(ctx, SubscriptionCustomizeParams) (*SubscriptionCustomization, error)` | Read the customization limits for a plan. | [subscription/customize](https://apidocs.webshare.io/subscription/customize) |
| `Pricing(ctx, SubscriptionPricingParams) (*SubscriptionPricing, error)` | Price a custom plan. | [subscription/pricing](https://apidocs.webshare.io/subscription/pricing) |
| `EnableAutoRenewal(ctx) (*Subscription, error)` | Enable auto-renewal (payment method must be on file). | [subscription/auto_renewal](https://apidocs.webshare.io/subscription/auto_renewal) |
| `CancelAutoRenewal(ctx) (*Subscription, error)` | Cancel auto-renewal; removes the payment method. | [subscription/auto_renewal](https://apidocs.webshare.io/subscription/auto_renewal) |

## client.Plans — Plans

| Method | Description | Docs |
|---|---|---|
| `List(ctx, PlanListParams) (*Page[Plan], error)` | List the account's plans, including cancelled ones. | [subscription/plan](https://apidocs.webshare.io/subscription/plan) |
| `ListAll(ctx, PlanListParams) iter.Seq2[Plan, error]` | Lazily iterate every plan across all pages. | [subscription/plan](https://apidocs.webshare.io/subscription/plan) |
| `Get(ctx, id int) (*Plan, error)` | Retrieve a plan. | [subscription/plan](https://apidocs.webshare.io/subscription/plan) |
| `Update(ctx, id int, PlanUpdateParams) (*Plan, error)` | Update a plan's automatic refresh schedule. | [subscription/plan](https://apidocs.webshare.io/subscription/plan) |
| `Cancel(ctx, id int) (*PlanCancelResult, error)` | Cancel a plan; credits the remainder. | [subscription/plan](https://apidocs.webshare.io/subscription/plan) |

## client.Invoices — Invoices

| Method | Description | Docs |
|---|---|---|
| `Download(ctx, subscriptionTransactionID int) ([]byte, error)` | Download an invoice as PDF bytes; takes the `Transaction.ID`. | [subscription/download_invoice](https://apidocs.webshare.io/subscription/download_invoice) |

## client.Referral — Referral and coupon codes

| Method | Description | Docs |
|---|---|---|
| `GetConfig(ctx) (*ReferralConfig, error)` | Read the referral configuration. | [referral](https://apidocs.webshare.io/referral) |
| `UpdateConfig(ctx, ReferralConfigUpdateParams) (*ReferralConfig, error)` | Update the referral mode or PayPal payout email. | [referral](https://apidocs.webshare.io/referral) |
| `GetCouponCode(ctx) (*CouponCode, error)` | Read the coupon code applied to the account. | [referral/coupon_code](https://apidocs.webshare.io/referral/coupon_code) |
| `ApplyCouponCode(ctx, code string) (*CouponCode, error)` | Apply a coupon code (own 5/min rate limit; documented 400 codes). | [referral/coupon_code](https://apidocs.webshare.io/referral/coupon_code) |
| `RemoveCouponCode(ctx) error` | Remove the applied coupon code. | [referral/coupon_code](https://apidocs.webshare.io/referral/coupon_code) |
| `ListChannels(ctx) ([]ReferralChannel, error)` | List owned referral channels (bare array, not paginated). | [referral/referral_channel](https://apidocs.webshare.io/referral/referral_channel) |
| `ListCredits(ctx, ReferralCreditListParams) (*Page[ReferralCredit], error)` | List referral credits. | [referral/referral_credit](https://apidocs.webshare.io/referral/referral_credit) |
| `ListAllCredits(ctx, ReferralCreditListParams) iter.Seq2[ReferralCredit, error]` | Lazily iterate every referral credit across all pages. | [referral/referral_credit](https://apidocs.webshare.io/referral/referral_credit) |
| `GetCredit(ctx, id int) (*ReferralCredit, error)` | Retrieve a referral credit. | [referral/referral_credit](https://apidocs.webshare.io/referral/referral_credit) |
| `ListEarnouts(ctx, ReferralEarnoutListParams) (*Page[ReferralEarnout], error)` | List earn-outs. | [referral/referral_earnout](https://apidocs.webshare.io/referral/referral_earnout) |
| `ListAllEarnouts(ctx, ReferralEarnoutListParams) iter.Seq2[ReferralEarnout, error]` | Lazily iterate every earn-out across all pages. | [referral/referral_earnout](https://apidocs.webshare.io/referral/referral_earnout) |
| `GetEarnout(ctx, id int) (*ReferralEarnout, error)` | Retrieve an earn-out. | [referral/referral_earnout](https://apidocs.webshare.io/referral/referral_earnout) |
| `GetCodeInfo(ctx, referralCode string) (*ReferralCodeInfo, error)` | Read the public information of a referral code (unauthenticated). | [referral/referral_info](https://apidocs.webshare.io/referral/referral_info) |
