package webshare

import (
	"context"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// NotificationsService exposes the notification (activity feed) operations.
type NotificationsService struct {
	client *Client
}

// Notification is one account notification. Known types include
// too_much_bandwidth_too_little_proxies, unlimited_bandwidth_gets_throttled,
// subscription_renew_failed, subscription_cc_will_expire_soon,
// reminder_to_use_proxy, projected_proxy_usage_over_80,
// projected_proxy_usage_over_100, high_concurrency_error,
// 100_percent_bandwidth_used, proxies_are_unallocated and question_is_added.
type Notification struct {
	// ID is the unique identifier of the notification.
	ID int `json:"id"`
	// Type is the notification type.
	Type string `json:"type"`
	// IsDismissable reports whether the notification can be dismissed.
	IsDismissable bool `json:"is_dismissable"`
	// Context carries additional type-specific context (for example plan,
	// plan_limit, effect, projected_bandwidth_gbs).
	Context map[string]any `json:"context"`
	// CreatedAt is when this instance was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when this instance was last updated.
	UpdatedAt time.Time `json:"updated_at"`
	// DismissedAt is when the notification was dismissed. May be nil.
	DismissedAt *time.Time `json:"dismissed_at"`
}

// NotificationListParams are the parameters for NotificationsService.List.
type NotificationListParams struct {
	// DismissedAtIsNull filters by dismissal state.
	DismissedAtIsNull *bool
	// Ordering orders by id, created_at or dismissed_at; the default is
	// -created_at.
	Ordering string
	// Type filters by notification type.
	Type string
	// Page is the page number.
	Page *int
	// PageSize is the number of results per page.
	PageSize *int
}

// List returns the account notifications in paginated format.
func (s *NotificationsService) List(ctx context.Context, params NotificationListParams, opts ...RequestOption) (*Page[Notification], error) {
	q := url.Values{}
	setBool(q, "dismissed_at__isnull", params.DismissedAtIsNull)
	setString(q, "ordering", params.Ordering)
	setString(q, "type", params.Type)
	setInt(q, "page", params.Page)
	setInt(q, "page_size", params.PageSize)
	return getPage[Notification](ctx, s.client, "/api/v2/notification/", q, opts)
}

// ListAll returns a lazy iterator over every notification across all pages.
func (s *NotificationsService) ListAll(ctx context.Context, params NotificationListParams, opts ...RequestOption) iter.Seq2[Notification, error] {
	return iterPages(ctx, func(ctx context.Context) (*Page[Notification], error) {
		return s.List(ctx, params, opts...)
	})
}

// Get retrieves a notification.
func (s *NotificationsService) Get(ctx context.Context, id int, opts ...RequestOption) (*Notification, error) {
	out := &Notification{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/notification/"+strconv.Itoa(id)+"/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// Dismiss dismisses a notification and returns it with DismissedAt set.
func (s *NotificationsService) Dismiss(ctx context.Context, id int, opts ...RequestOption) (*Notification, error) {
	out := &Notification{}
	if err := s.client.doJSON(ctx, http.MethodPost, "/api/v2/notification/"+strconv.Itoa(id)+"/dismiss/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// Restore restores a dismissed notification and returns it with DismissedAt
// cleared.
func (s *NotificationsService) Restore(ctx context.Context, id int, opts ...RequestOption) (*Notification, error) {
	out := &Notification{}
	if err := s.client.doJSON(ctx, http.MethodPost, "/api/v2/notification/"+strconv.Itoa(id)+"/restore/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}
