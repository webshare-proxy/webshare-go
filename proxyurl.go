package webshare

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// ConnectionMode selects how proxies are accessed: connecting directly to the
// proxy address, or through the p.webshare.io backbone.
type ConnectionMode string

// Connection modes.
const (
	// ModeDirect connects to the proxy_address and port returned by the
	// proxy list API.
	ModeDirect ConnectionMode = "direct"
	// ModeBackbone connects through p.webshare.io. Required when the plan's
	// pool_filter is residential.
	ModeBackbone ConnectionMode = "backbone"
)

// BackboneHost is the host used for backbone-mode proxy connections.
const BackboneHost = "p.webshare.io"

// ProxyURLParams describes a proxy connection for ProxyURL.
type ProxyURLParams struct {
	// Mode is the connection mode: ModeDirect or ModeBackbone. Required.
	Mode ConnectionMode
	// Scheme is the URL scheme, "http" by default.
	Scheme string
	// Username is the proxy username. Leave empty together with Password to
	// build an IP-authorization URL without credentials.
	Username string
	// Password is the proxy password. Required when Username is set.
	Password string
	// Address is the proxy host. Required in direct mode; defaults to
	// BackboneHost in backbone mode.
	Address string
	// Port is the proxy port. Required in direct mode; defaults to 80 in
	// backbone mode.
	Port int

	// CountryCodes appends ISO 3166-1 alpha-2 country codes to the backbone
	// username (lowercased). Backbone username/password auth only.
	CountryCodes []string
	// City appends a city_{name} parameter to the backbone username. Letters
	// and underscores only; city targeting is available on residential plans.
	City string
	// SessionID appends a numeric sticky-session ID to the backbone
	// username. Mutually exclusive with Rotate.
	SessionID string
	// Rotate appends the rotate parameter to the backbone username so every
	// request uses a new IP. Mutually exclusive with SessionID.
	Rotate bool
}

// ProxyURL builds a proxy connection URL. In direct mode it produces
// "http://user:pass@address:port"; in backbone mode it targets p.webshare.io
// and encodes country, city and session parameters into the username using
// the {username}[-{cc}...][-city_{name}][-{session}|-rotate] grammar. When no
// credentials are given, an IP-authorization URL without user info is built.
func ProxyURL(params ProxyURLParams) (string, error) {
	scheme := params.Scheme
	if scheme == "" {
		scheme = "http"
	}

	hasCredentials := params.Username != "" || params.Password != ""
	if hasCredentials {
		if params.Username == "" {
			return "", errors.New("webshare: proxy URL: username is required when a password is set")
		}
		if params.Password == "" {
			return "", errors.New("webshare: proxy URL: password is required when a username is set")
		}
	}
	hasUsernameParams := len(params.CountryCodes) > 0 || params.City != "" || params.SessionID != "" || params.Rotate

	var host string
	var port int
	switch params.Mode {
	case ModeDirect:
		if params.Address == "" {
			return "", errors.New("webshare: proxy URL: address is required in direct mode")
		}
		if params.Port <= 0 {
			return "", errors.New("webshare: proxy URL: port is required in direct mode")
		}
		if hasUsernameParams {
			return "", errors.New("webshare: proxy URL: country, city, session and rotate parameters are only supported in backbone mode")
		}
		host, port = params.Address, params.Port
	case ModeBackbone:
		host, port = params.Address, params.Port
		if host == "" {
			host = BackboneHost
		}
		if port <= 0 {
			port = 80
		}
		if hasUsernameParams && !hasCredentials {
			return "", errors.New("webshare: proxy URL: country, city, session and rotate parameters require username/password authentication")
		}
	case "":
		return "", errors.New("webshare: proxy URL: mode is required (direct or backbone)")
	default:
		return "", fmt.Errorf("webshare: proxy URL: invalid mode %q (must be direct or backbone)", params.Mode)
	}

	u := &url.URL{
		Scheme: scheme,
		Host:   net.JoinHostPort(host, strconv.Itoa(port)),
	}
	if hasCredentials {
		username, err := backboneUsername(params)
		if err != nil {
			return "", err
		}
		u.User = url.UserPassword(username, params.Password)
	}
	return u.String(), nil
}

// backboneUsername renders the username with its optional parameters in the
// documented order: country codes, then city, then session or rotate.
func backboneUsername(params ProxyURLParams) (string, error) {
	if params.SessionID != "" && params.Rotate {
		return "", errors.New("webshare: proxy URL: session ID and rotate are mutually exclusive")
	}
	parts := []string{params.Username}
	for _, code := range params.CountryCodes {
		if len(code) != 2 || !isAlpha(code) {
			return "", fmt.Errorf("webshare: proxy URL: invalid country code %q (must be a 2-letter ISO 3166-1 alpha-2 code)", code)
		}
		parts = append(parts, strings.ToLower(code))
	}
	if params.City != "" {
		if !isCityName(params.City) {
			return "", fmt.Errorf("webshare: proxy URL: invalid city %q (must contain only letters and underscores)", params.City)
		}
		parts = append(parts, "city_"+params.City)
	}
	switch {
	case params.SessionID != "":
		if !isDigits(params.SessionID) {
			return "", fmt.Errorf("webshare: proxy URL: invalid session ID %q (must be numeric)", params.SessionID)
		}
		parts = append(parts, params.SessionID)
	case params.Rotate:
		parts = append(parts, "rotate")
	}
	return strings.Join(parts, "-"), nil
}

func isAlpha(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return false
		}
	}
	return len(s) > 0
}

func isCityName(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && r != '_' {
			return false
		}
	}
	return len(s) > 0
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}
