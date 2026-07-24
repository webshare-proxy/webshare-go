package webshare

import "context"

// DefaultAuthScheme is the Authorization header scheme used when a Token does
// not specify one. All Webshare API keys and login tokens use this scheme.
const DefaultAuthScheme = "Token"

// Token is a credential presented to the API in the Authorization header as
// "<Scheme> <Value>".
type Token struct {
	// Value is the secret token value.
	Value string
	// Scheme is the Authorization header scheme. When empty,
	// DefaultAuthScheme ("Token") is used. OAuth-issued credentials may use a
	// different scheme such as "Bearer".
	Scheme string
}

// authorizationHeader renders the token as an Authorization header value.
// This is the single place the header format is constructed.
func (t Token) authorizationHeader() string {
	scheme := t.Scheme
	if scheme == "" {
		scheme = DefaultAuthScheme
	}
	return scheme + " " + t.Value
}

// TokenSource supplies the credential used to authenticate API requests. It
// is called once per request, which is cheap for static keys and allows
// refreshing implementations (for example OAuth token sources) to always
// provide a current token.
type TokenSource interface {
	// Token returns the credential to use for a request.
	Token(ctx context.Context) (Token, error)
}

// StaticTokenSource returns a TokenSource that always yields the given token.
func StaticTokenSource(t Token) TokenSource {
	return staticTokenSource{t}
}

type staticTokenSource struct {
	t Token
}

func (s staticTokenSource) Token(context.Context) (Token, error) {
	return s.t, nil
}
