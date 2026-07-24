package webshare

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Int returns a pointer to v, for use in optional request fields.
func Int(v int) *int { return &v }

// Int64 returns a pointer to v, for use in optional request fields.
func Int64(v int64) *int64 { return &v }

// Float64 returns a pointer to v, for use in optional request fields.
func Float64(v float64) *float64 { return &v }

// Bool returns a pointer to v, for use in optional request fields.
func Bool(v bool) *bool { return &v }

// String returns a pointer to v, for use in optional request fields.
func String(v string) *string { return &v }

// Time returns a pointer to v, for use in optional request fields.
func Time(v time.Time) *time.Time { return &v }

// Nullable represents an optional request field that distinguishes between
// being omitted, set to an explicit JSON null, and set to a concrete value.
// The zero value is omitted from the request.
type Nullable[T any] struct {
	present bool
	null    bool
	value   T
}

// NullableOf returns a Nullable holding the given value.
func NullableOf[T any](v T) Nullable[T] {
	return Nullable[T]{present: true, value: v}
}

// Null returns a Nullable that serializes as an explicit JSON null.
func Null[T any]() Nullable[T] {
	return Nullable[T]{present: true, null: true}
}

// isPresent reports whether the field was set (to a value or to null).
func (n Nullable[T]) isPresent() bool { return n.present }

// MarshalJSON implements json.Marshaler.
func (n Nullable[T]) MarshalJSON() ([]byte, error) {
	if n.null || !n.present {
		return []byte("null"), nil
	}
	return json.Marshal(n.value)
}

// ASNInfo describes an ASN entry in the proxy configuration maps. On the wire
// it is a heterogeneous two-element array: ["ASN NAME", 105].
type ASNInfo struct {
	// Name is the ASN name.
	Name string
	// Count is the number of proxies in the ASN.
	Count int
}

// UnmarshalJSON implements json.Unmarshaler for the wire tuple form.
func (a *ASNInfo) UnmarshalJSON(data []byte) error {
	var tuple []json.RawMessage
	if err := json.Unmarshal(data, &tuple); err != nil {
		return fmt.Errorf("ASN entry: %w", err)
	}
	if len(tuple) != 2 {
		return fmt.Errorf("ASN entry: expected 2 elements, got %d", len(tuple))
	}
	if err := json.Unmarshal(tuple[0], &a.Name); err != nil {
		return fmt.Errorf("ASN name: %w", err)
	}
	if err := json.Unmarshal(tuple[1], &a.Count); err != nil {
		return fmt.Errorf("ASN count: %w", err)
	}
	return nil
}

// MarshalJSON implements json.Marshaler, producing the wire tuple form.
func (a ASNInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal([2]any{a.Name, a.Count})
}

// query building helpers: unset optional values are omitted; booleans are
// rendered as true/false; lists are comma-joined.

func setString(q url.Values, key, value string) {
	if value != "" {
		q.Set(key, value)
	}
}

func setStringList(q url.Values, key string, values []string) {
	if len(values) > 0 {
		q.Set(key, strings.Join(values, ","))
	}
}

func setInt(q url.Values, key string, value *int) {
	if value != nil {
		q.Set(key, strconv.Itoa(*value))
	}
}

func setBool(q url.Values, key string, value *bool) {
	if value != nil {
		q.Set(key, strconv.FormatBool(*value))
	}
}

func setTime(q url.Values, key string, value *time.Time) {
	if value != nil {
		q.Set(key, value.Format(time.RFC3339Nano))
	}
}
