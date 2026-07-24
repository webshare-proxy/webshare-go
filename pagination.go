package webshare

import (
	"context"
	"iter"
	"net/http"
	"net/url"
)

// Page is one page of results from a paginated list endpoint. It exposes the
// raw pagination envelope and can fetch the following page.
type Page[T any] struct {
	// Count is the total number of results across all pages.
	Count int `json:"count"`
	// Next is the URL of the next page, or nil on the last page.
	Next *string `json:"next"`
	// Previous is the URL of the previous page, or nil on the first page.
	Previous *string `json:"previous"`
	// Results holds the items on this page.
	Results []T `json:"results"`

	client *Client
	opts   []RequestOption
}

// HasNextPage reports whether a further page is available.
func (p *Page[T]) HasNextPage() bool {
	return p != nil && p.Next != nil && *p.Next != ""
}

// NextPage fetches the next page by following the envelope's next URL
// verbatim. It returns (nil, nil) when there is no next page.
func (p *Page[T]) NextPage(ctx context.Context) (*Page[T], error) {
	if !p.HasNextPage() {
		return nil, nil
	}
	return getPageURL[T](ctx, p.client, *p.Next, p.opts)
}

// getPage fetches the first page of a paginated list endpoint.
func getPage[T any](ctx context.Context, c *Client, path string, query url.Values, opts []RequestOption) (*Page[T], error) {
	page := &Page[T]{client: c, opts: opts}
	if err := c.doJSON(ctx, http.MethodGet, path, query, nil, page, opts); err != nil {
		return nil, err
	}
	return page, nil
}

// getPageURL fetches a page from an absolute pagination URL.
func getPageURL[T any](ctx context.Context, c *Client, rawURL string, opts []RequestOption) (*Page[T], error) {
	page := &Page[T]{client: c, opts: opts}
	if err := c.doURL(ctx, rawURL, page, opts); err != nil {
		return nil, err
	}
	return page, nil
}

// iterPages returns a lazy sequence over every item of a paginated list,
// fetching pages on demand by following the envelope's next URL. Iteration
// stops with an error when the context is cancelled or a page fetch fails.
func iterPages[T any](ctx context.Context, fetch func(context.Context) (*Page[T], error)) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		page, err := fetch(ctx)
		for {
			if err != nil {
				var zero T
				yield(zero, err)
				return
			}
			for _, item := range page.Results {
				if !yield(item, nil) {
					return
				}
			}
			if !page.HasNextPage() {
				return
			}
			if ctxErr := ctx.Err(); ctxErr != nil {
				var zero T
				yield(zero, &RequestError{Err: ctxErr})
				return
			}
			page, err = page.NextPage(ctx)
		}
	}
}
