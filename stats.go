package webshare

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// StatsService exposes the proxy usage statistics operations.
type StatsService struct {
	client *Client
}

// StatErrorReason describes one class of failed proxy requests within a
// stats window.
type StatErrorReason struct {
	// Reason is the error code. The same code is present in the
	// X-Webshare-Error-Reason header on failed proxied requests.
	Reason string `json:"reason"`
	// Type is "configuration" or "connection".
	Type string `json:"type"`
	// HowToFix is an end-user guide for fixing the error.
	HowToFix string `json:"how_to_fix"`
	// HTTPStatus is the status the proxy endpoint may return for this
	// error. May be nil.
	HTTPStatus *int `json:"http_status"`
	// Count is the number of failed proxy requests with this reason.
	Count int `json:"count"`
}

// ProxyStat is one hourly aggregate of proxy usage.
type ProxyStat struct {
	// Timestamp is the start of the one-hour aggregation window.
	Timestamp time.Time `json:"timestamp"`
	// IsProjected indicates whether the stat is projected rather than real.
	IsProjected bool `json:"is_projected"`
	// BandwidthTotal is the total bandwidth use in bytes for the window.
	BandwidthTotal int64 `json:"bandwidth_total"`
	// BandwidthAverage is the average bandwidth in bytes per request.
	BandwidthAverage int64 `json:"bandwidth_average"`
	// RequestsTotal is the number of proxy requests made.
	RequestsTotal int64 `json:"requests_total"`
	// RequestsSuccessful is the number of successful proxy requests.
	RequestsSuccessful int64 `json:"requests_successful"`
	// RequestsFailed is the number of failed proxy requests.
	RequestsFailed int64 `json:"requests_failed"`
	// ErrorReasons lists the error reasons. Empty when IsProjected is true.
	ErrorReasons []StatErrorReason `json:"error_reasons"`
	// CountriesUsed maps country code to request count. Empty when
	// IsProjected is true.
	CountriesUsed map[string]int64 `json:"countries_used"`
	// NumberOfProxiesUsed is the estimated number of unique proxy addresses
	// used. Zero when IsProjected is true.
	NumberOfProxiesUsed int `json:"number_of_proxies_used"`
	// ProtocolsUsed maps proxy protocol (http, socks) to request count.
	// Empty when IsProjected is true.
	ProtocolsUsed map[string]int64 `json:"protocols_used"`
	// AverageConcurrency is the estimated average number of concurrent
	// proxy requests. Nil when IsProjected is true.
	AverageConcurrency *float64 `json:"average_concurrency"`
	// AverageRPS is the estimated average proxy requests per second. Nil
	// when IsProjected is true.
	AverageRPS *float64 `json:"average_rps"`
	// LastRequestSentAt is when the last proxy request in the window was
	// sent. Nil when IsProjected is true.
	LastRequestSentAt *time.Time `json:"last_request_sent_at"`
}

// AggregateStats is the aggregated proxy usage for a period. It is also the
// shape of a sub-user's aggregate_stats field.
type AggregateStats struct {
	// BandwidthProjected is the projected bandwidth for the period in bytes.
	BandwidthProjected int64 `json:"bandwidth_projected"`
	// BandwidthTotal is the total bandwidth use in bytes.
	BandwidthTotal int64 `json:"bandwidth_total"`
	// BandwidthAverage is the average bandwidth in bytes per request.
	BandwidthAverage int64 `json:"bandwidth_average"`
	// RequestsTotal is the number of proxy requests made.
	RequestsTotal int64 `json:"requests_total"`
	// RequestsSuccessful is the number of successful proxy requests.
	RequestsSuccessful int64 `json:"requests_successful"`
	// RequestsFailed is the number of failed proxy requests.
	RequestsFailed int64 `json:"requests_failed"`
	// ErrorReasons lists the error reasons.
	ErrorReasons []StatErrorReason `json:"error_reasons"`
	// CountriesUsed maps country code to request count.
	CountriesUsed map[string]int64 `json:"countries_used"`
	// NumberOfProxiesUsed is the estimated number of unique proxy addresses
	// used.
	NumberOfProxiesUsed int `json:"number_of_proxies_used"`
	// ProtocolsUsed maps proxy protocol (http, socks) to request count.
	ProtocolsUsed map[string]int64 `json:"protocols_used"`
	// AverageConcurrency is the estimated average number of concurrent
	// proxy requests.
	AverageConcurrency *float64 `json:"average_concurrency"`
	// AverageRPS is the estimated average proxy requests per second.
	AverageRPS *float64 `json:"average_rps"`
	// LastRequestSentAt is when the last proxy request was sent.
	LastRequestSentAt *time.Time `json:"last_request_sent_at"`
}

// StatsListParams are the parameters for StatsService.List and
// StatsService.Aggregate.
type StatsListParams struct {
	// TimestampLTE bounds stats to timestamps at or before this time. A
	// future time includes projected stats. Cannot be after the
	// subscription end date.
	TimestampLTE *time.Time
	// TimestampGTE bounds stats to timestamps at or after this time. Must
	// be before TimestampLTE and no older than 90 days.
	TimestampGTE *time.Time
	// PlanID targets a specific plan; otherwise the default plan is used.
	PlanID *int
}

func (p StatsListParams) values() url.Values {
	q := url.Values{}
	setTime(q, "timestamp__lte", p.TimestampLTE)
	setTime(q, "timestamp__gte", p.TimestampGTE)
	setInt(q, "plan_id", p.PlanID)
	return q
}

// List returns the hourly proxy stats within a time period. This endpoint is
// not paginated: it returns a bare array. Hours without proxy usage have no
// entry.
func (s *StatsService) List(ctx context.Context, params StatsListParams, opts ...RequestOption) ([]ProxyStat, error) {
	var out []ProxyStat
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/stats/", params.values(), nil, &out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// Aggregate returns the proxy stats aggregated over the given period.
func (s *StatsService) Aggregate(ctx context.Context, params StatsListParams, opts ...RequestOption) (*AggregateStats, error) {
	out := &AggregateStats{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/stats/aggregate/", params.values(), nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}
