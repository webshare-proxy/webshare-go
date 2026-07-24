package webshare

import (
	"context"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// APIKeysService exposes the API key operations. These endpoints are
// session-token-only: calling them with an API key returns a 403 with code
// "api_key_not_allowed" ("Cannot query this API endpoint with an API key.").
// Authenticate with a login token (for example from AuthService.Login) to
// manage API keys.
type APIKeysService struct {
	client *Client
}

// APIKey is one API key of the account. All API keys have the same
// permissions and full account access.
type APIKey struct {
	// ID is the unique identifier of the API key.
	ID int `json:"id"`
	// Key is the 40-character alphanumeric API key value.
	Key string `json:"key"`
	// Label describes the API key. Labels may be duplicated across keys.
	Label string `json:"label"`
	// CreatedAt is when this instance was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when this instance was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// APIKeyListParams are the parameters for APIKeysService.List.
type APIKeyListParams struct {
	// Page is the page number.
	Page *int
	// PageSize is the number of results per page.
	PageSize *int
}

// List returns the API keys in paginated format.
func (s *APIKeysService) List(ctx context.Context, params APIKeyListParams, opts ...RequestOption) (*Page[APIKey], error) {
	q := url.Values{}
	setInt(q, "page", params.Page)
	setInt(q, "page_size", params.PageSize)
	return getPage[APIKey](ctx, s.client, "/api/v2/apikey/", q, opts)
}

// ListAll returns a lazy iterator over every API key across all pages.
func (s *APIKeysService) ListAll(ctx context.Context, params APIKeyListParams, opts ...RequestOption) iter.Seq2[APIKey, error] {
	return iterPages(ctx, func(ctx context.Context) (*Page[APIKey], error) {
		return s.List(ctx, params, opts...)
	})
}

// APIKeyCreateParams are the parameters for APIKeysService.Create.
type APIKeyCreateParams struct {
	// Label describes the API key. May be duplicated across keys.
	Label string `json:"label,omitempty"`
}

// Create creates an API key. The response is the only place the full key
// value appears.
func (s *APIKeysService) Create(ctx context.Context, params APIKeyCreateParams, opts ...RequestOption) (*APIKey, error) {
	out := &APIKey{}
	if err := s.client.doJSON(ctx, http.MethodPost, "/api/v2/apikey/", nil, params, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// Get retrieves an API key.
func (s *APIKeysService) Get(ctx context.Context, id int, opts ...RequestOption) (*APIKey, error) {
	out := &APIKey{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/apikey/"+strconv.Itoa(id)+"/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// APIKeyUpdateParams are the parameters for APIKeysService.Update.
type APIKeyUpdateParams struct {
	// Label sets the API key label.
	Label string `json:"label,omitempty"`
}

// Update updates an API key.
func (s *APIKeysService) Update(ctx context.Context, id int, params APIKeyUpdateParams, opts ...RequestOption) (*APIKey, error) {
	out := &APIKey{}
	if err := s.client.doJSON(ctx, http.MethodPatch, "/api/v2/apikey/"+strconv.Itoa(id)+"/", nil, params, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete removes an API key.
func (s *APIKeysService) Delete(ctx context.Context, id int, opts ...RequestOption) error {
	return s.client.doJSON(ctx, http.MethodDelete, "/api/v2/apikey/"+strconv.Itoa(id)+"/", nil, nil, nil, opts)
}
