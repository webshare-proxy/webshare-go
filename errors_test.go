package webshare

import (
	"net/http"
	"strings"
	"testing"
)

func TestErrorMessageFieldOrderDeterministic(t *testing.T) {
	err := &Error{
		StatusCode: http.StatusBadRequest,
		FieldErrors: map[string][]string{
			"zebra": {"last"},
			"alpha": {"first"},
			"mode":  {"middle"},
		},
	}
	want := "[alpha: first, mode: middle, zebra: last]"
	for range 20 {
		if msg := err.Error(); !strings.Contains(msg, want) {
			t.Fatalf("Error() = %q, want it to contain %q", msg, want)
		}
	}
}

func TestTruncatePreservesRuneBoundaries(t *testing.T) {
	// "é" is 2 bytes; cutting at 3 bytes must back up to the boundary.
	if got := truncate("aéé", 3); got != "aé... (truncated)" {
		t.Errorf("truncate = %q, want %q", got, "aé... (truncated)")
	}
	if got := truncate("plain", 10); got != "plain" {
		t.Errorf("truncate = %q, want unchanged input", got)
	}
}

func TestTopLevelArrayErrorBody(t *testing.T) {
	apiErr := parseAPIError(http.StatusBadRequest, "", []byte(`["First problem.", "Second problem."]`))
	if apiErr.Detail != "First problem.; Second problem." {
		t.Errorf("Detail = %q, want the joined messages", apiErr.Detail)
	}
}
