package webshare

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Error is returned when the Webshare API responds with a non-success status
// code. Use errors.As to inspect it:
//
//	var apiErr *webshare.Error
//	if errors.As(err, &apiErr) {
//		switch apiErr.Code {
//		case "2fa_needed":
//			// ...
//		}
//	}
type Error struct {
	// StatusCode is the HTTP status code of the response.
	StatusCode int
	// Code is the machine-readable API error code when present (for example
	// "2fa_needed", "account_suspended" or "account_deleted"). Empty when the
	// response carried no code.
	Code string
	// RequestID is the value of the X-Request-ID response header when present.
	RequestID string
	// Detail is the human-readable error message.
	Detail string
	// FieldErrors maps field names to validation messages for request
	// validation failures, e.g. {"mode": ["This field is required."]}.
	FieldErrors map[string][]string
	// Body is the raw response body, capped at 1 MiB.
	Body []byte
	// RetryAfter is the wait parsed from the Retry-After header (delta
	// seconds or HTTP-date), when present and valid. Nil when the header was
	// absent or unparsable. It is exposed so callers can self-throttle calls
	// the SDK does not retry (for example a 429 on a POST).
	RetryAfter *time.Duration
}

// Error implements the error interface.
func (e *Error) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "webshare: API error (status %d", e.StatusCode)
	if e.Code != "" {
		fmt.Fprintf(&b, ", code %s", e.Code)
	}
	if e.RequestID != "" {
		fmt.Fprintf(&b, ", request id %s", e.RequestID)
	}
	b.WriteString(")")
	if e.Detail != "" {
		b.WriteString(": ")
		b.WriteString(truncate(e.Detail, maxDetailLen))
	}
	if len(e.FieldErrors) > 0 {
		fields := make([]string, 0, len(e.FieldErrors))
		for field, msgs := range e.FieldErrors {
			fields = append(fields, field+": "+strings.Join(msgs, "; "))
		}
		fmt.Fprintf(&b, " [%s]", strings.Join(fields, ", "))
	}
	return b.String()
}

// RequestError is returned when a request could not be completed at the
// transport level (connection failures, timeouts, cancellation). It wraps the
// underlying error for use with errors.Is and errors.Unwrap.
type RequestError struct {
	// Err is the underlying transport error.
	Err error
}

// Error implements the error interface.
func (e *RequestError) Error() string {
	return "webshare: request failed: " + e.Err.Error()
}

// Unwrap returns the underlying transport error.
func (e *RequestError) Unwrap() error {
	return e.Err
}

// ResponseDecodeError is returned when a success (2xx) response body cannot
// be decoded as the expected JSON shape, for example a non-JSON body or a
// malformed pagination envelope. It carries the status and the raw body
// (capped at 1 MiB) so callers never see a bare JSON error.
type ResponseDecodeError struct {
	// StatusCode is the HTTP status code of the response.
	StatusCode int
	// Body is the raw response body, capped at 1 MiB.
	Body []byte
	// Err is the underlying decode error.
	Err error
}

// Error implements the error interface. The body is rendered as a short
// snippet; the capped raw body remains available on Body.
func (e *ResponseDecodeError) Error() string {
	return fmt.Sprintf("webshare: decoding response body (status %d): %v: %s",
		e.StatusCode, e.Err, truncate(strings.TrimSpace(string(e.Body)), maxDetailLen))
}

// Unwrap returns the underlying decode error.
func (e *ResponseDecodeError) Unwrap() error {
	return e.Err
}

// CrossOriginError is returned when a URL taken from a response envelope
// (such as a pagination next link) points at a different origin than the
// client base URL. The SDK refuses to follow such URLs so credentials are
// never sent cross-origin.
type CrossOriginError struct {
	// BaseOrigin is the scheme://host origin of the client base URL.
	BaseOrigin string
	// TargetOrigin is the scheme://host origin of the refused URL.
	TargetOrigin string
}

// Error implements the error interface.
func (e *CrossOriginError) Error() string {
	return fmt.Sprintf("webshare: refusing to follow cross-origin URL: response points at %s but the client base URL is %s", e.TargetOrigin, e.BaseOrigin)
}

const (
	// maxErrorBody caps how much of an error (or undecodable) response body
	// is captured on error values.
	maxErrorBody = 1 << 20 // 1 MiB
	// maxDetailLen caps the human-facing detail rendered by Error().
	maxDetailLen = 2048
)

// truncate shortens s to at most n bytes, marking the cut.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "... (truncated)"
}

// parseAPIError builds an *Error from a non-success response. Body parsing is
// tolerant: it accepts {"detail": "..."} objects, DRF field-error maps, bare
// JSON string bodies and non-JSON bodies.
func parseAPIError(statusCode int, requestID string, body []byte) *Error {
	apiErr := &Error{
		StatusCode: statusCode,
		RequestID:  requestID,
		Body:       body,
	}
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return apiErr
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(body, &obj); err == nil {
		for key, raw := range obj {
			// "detail" and "code" are strings in API error envelopes, but
			// DRF validation errors may reuse the same keys as field names
			// with a list of messages (e.g. {"code": ["Invalid promo
			// code."]}); handle both shapes.
			var s string
			if json.Unmarshal(raw, &s) == nil {
				switch key {
				case "detail":
					apiErr.Detail = s
					continue
				case "code":
					apiErr.Code = s
					continue
				}
			}
			var msgs []string
			if json.Unmarshal(raw, &msgs) == nil {
				if apiErr.FieldErrors == nil {
					apiErr.FieldErrors = make(map[string][]string)
				}
				apiErr.FieldErrors[key] = msgs
			}
		}
		return apiErr
	}

	var s string
	if err := json.Unmarshal(body, &s); err == nil {
		apiErr.Detail = s
		return apiErr
	}

	apiErr.Detail = trimmed
	return apiErr
}
