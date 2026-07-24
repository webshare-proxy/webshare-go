# webshare-go

Official Go SDK for the [Webshare proxy API](https://apidocs.webshare.io).

[![CI](https://github.com/webshare-proxy/webshare-go/actions/workflows/ci.yml/badge.svg)](https://github.com/webshare-proxy/webshare-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/webshare-proxy/webshare-go.svg)](https://pkg.go.dev/github.com/webshare-proxy/webshare-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Install

```sh
go get github.com/webshare-proxy/webshare-go
```

Requires Go 1.23 or newer.

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"log"

	webshare "github.com/webshare-proxy/webshare-go"
)

func main() {
	client, err := webshare.NewClient() // reads WEBSHARE_API_KEY
	if err != nil {
		log.Fatal(err)
	}
	page, err := client.Proxies.List(context.Background(), webshare.ProxyListParams{Mode: webshare.ModeDirect})
	if err != nil {
		log.Fatal(err)
	}
	for _, proxy := range page.Results {
		// ProxyAddress is nil on residential plans, which connect through
		// the p.webshare.io backbone instead.
		address := webshare.BackboneHost
		if proxy.ProxyAddress != nil {
			address = *proxy.ProxyAddress
		}
		fmt.Printf("%s:%d (%s)\n", address, proxy.Port, proxy.CountryCode)
	}
}
```

## Authentication

The client reads the `WEBSHARE_API_KEY` environment variable by default, or
takes a key explicitly:

```go
client, err := webshare.NewClient(webshare.WithAPIKey("your-api-key"))
```

Credentials are pluggable. Anything that implements `TokenSource` can supply
tokens, including refreshing OAuth credentials:

```go
type TokenSource interface {
	Token(ctx context.Context) (webshare.Token, error)
}

client, err := webshare.NewClient(webshare.WithTokenSource(myOAuthSource))
```

The returned `Token` carries the value and the Authorization header scheme
(`Token` by default), so alternative schemes such as `Bearer` plug in without
changes elsewhere.

Unauthenticated operations (login, registration, password reset, referral
code info and the download endpoints) never send the Authorization header. To
build a client without credentials at all, for example before a login flow,
opt out explicitly:

```go
client, err := webshare.NewClient(webshare.WithUnauthenticated())
```

Calling an authenticated operation on such a client returns a clear error
naming the fix.

To act on behalf of a sub-user on proxy config, list, stats and activity
calls, or to use admin federated access, pass `webshare.WithSubuser(id)` or
`webshare.WithFederatedUser(id)` on the client or on any individual call.

## Pagination

Every list method returns a `*Page[T]` exposing the raw envelope (`Results`,
`Count`, `Next`, `Previous`) plus `NextPage(ctx)`. The `ListAll` variants
return a lazy `iter.Seq2[T, error]` that follows the server's `next` URL
across page boundaries:

```go
for proxy, err := range client.Proxies.ListAll(ctx, webshare.ProxyListParams{Mode: webshare.ModeDirect}) {
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(proxy.ID)
}
```

## Error handling

API failures are returned as `*webshare.Error` with the HTTP status, the API
error code, the `X-Request-ID` header, the human-readable detail and any
per-field validation messages. Transport failures are wrapped in
`*webshare.RequestError`:

```go
var apiErr *webshare.Error
if errors.As(err, &apiErr) {
	if apiErr.Code == "2fa_needed" {
		// submit a code with client.TwoFactorAuth.SubmitCode, then retry
	}
}
```

## Retries

Failed requests are retried up to 2 times by default with exponential
backoff and full jitter, honoring `Retry-After` on 429 responses. Retries
apply to connection errors, timeouts and 408/429/5xx responses, and only to
idempotent requests (GET, PUT, DELETE) unless opted in:

```go
client, err := webshare.NewClient(webshare.WithMaxRetries(5))
// or per request:
key, err := client.APIKeys.Create(ctx, params, webshare.WithRetryNonIdempotent())
```

## Timeouts

The timeout bounds a single HTTP attempt, including reading the response
body; the default is 60 seconds, configurable per client and per request.
With retries and Retry-After waits, the total call time can exceed it — use
the context to bound the whole call:

```go
client, err := webshare.NewClient(webshare.WithTimeout(10 * time.Second))
profile, err := client.Profile.Get(ctx, webshare.WithTimeout(2*time.Second))
```

Contexts are honored everywhere, including while waiting between retries.

## Proxy connection helper

`webshare.ProxyURL` builds proxy connection URLs for both connection modes,
including the backbone username grammar for country, city, sticky sessions
and rotation:

```go
proxyURL, err := webshare.ProxyURL(webshare.ProxyURLParams{
	Mode:         webshare.ModeBackbone,
	Username:     "myuser",
	Password:     "mypassword",
	CountryCodes: []string{"US"},
	SessionID:    "1234",
})
// http://myuser-us-1234:mypassword@p.webshare.io:80
```

`client.Proxies.DownloadURL` builds the shareable path-style proxy list
download URL, and `client.Proxies.Download` fetches the list as plain text.

Runnable programs for all of the above live in [examples](examples).

## Supported versions

| SDK | Go |
|---|---|
| 0.x | 1.23, 1.24, stable |

API documentation: https://apidocs.webshare.io

## License

MIT, see [LICENSE](LICENSE). Issues and pull requests are welcome.
