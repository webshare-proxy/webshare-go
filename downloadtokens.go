package webshare

import (
	"context"
	"net/http"
	"time"
)

// DownloadTokensService exposes the download token operations. Download
// tokens authorize the replaced proxy and activity download endpoints.
type DownloadTokensService struct {
	client *Client
}

// DownloadTokenScope selects what a download token can download.
type DownloadTokenScope string

// Download token scopes.
const (
	// ScopeProxyList is the proxy_list download token scope.
	ScopeProxyList DownloadTokenScope = "proxy_list"
	// ScopeReplacedProxy is the replaced_proxy download token scope.
	ScopeReplacedProxy DownloadTokenScope = "replaced_proxy"
	// ScopeActivity is the activity download token scope.
	ScopeActivity DownloadTokenScope = "activity"
)

// DownloadToken authorizes record downloads for its scope until it expires.
type DownloadToken struct {
	// ID is the unique identifier of the download token.
	ID int `json:"id"`
	// Key is the token value passed as the download_token query parameter.
	Key string `json:"key"`
	// Scope is what this token can download.
	Scope DownloadTokenScope `json:"scope"`
	// ExpireAt is when the token expires.
	ExpireAt time.Time `json:"expire_at"`
}

// Get returns the download token for the given scope.
func (s *DownloadTokensService) Get(ctx context.Context, scope DownloadTokenScope, opts ...RequestOption) (*DownloadToken, error) {
	out := &DownloadToken{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/download_token/"+string(scope)+"/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// Reset rotates the download token for the given scope and returns the new
// token.
func (s *DownloadTokensService) Reset(ctx context.Context, scope DownloadTokenScope, opts ...RequestOption) (*DownloadToken, error) {
	out := &DownloadToken{}
	if err := s.client.doJSON(ctx, http.MethodPost, "/api/v2/download_token/"+string(scope)+"/reset/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}
