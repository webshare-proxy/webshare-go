# Releasing `github.com/webshare-proxy/webshare-go`

## Versioning policy

We follow [Semantic Versioning](https://semver.org) and Go's module versioning
rules. While the SDK is on `v0.x` the public API is not yet frozen: **minor**
bumps (`v0.1 → v0.2`) may contain breaking changes, **patch** bumps never do.

Once the API is settled we cut `v1.0.0`, after which breaking changes require a
new **major** version *and a new module path* (`.../webshare-go/v2`), per Go's
[semantic import versioning](https://go.dev/ref/mod#major-version-suffixes).

> ⚠️ The module proxy caches tags **immutably**. A published version can never be
> changed or deleted — only superseded by a higher version. Be deliberate about
> every tag, and especially about the first `v1`.

## How to cut a release

There is no registry to publish to — a semver **git tag** is the release. The
proxy and pkg.go.dev index it on first fetch.

1. Update `CHANGELOG.md`.
2. Tag and push:
   ```bash
   git tag v0.1.1
   git push origin v0.1.1
   ```
3. The **Release** workflow runs the tests, cuts a GitHub Release with generated
   notes, and warms `proxy.golang.org` so the version is immediately importable
   with `go get github.com/webshare-proxy/webshare-go@v0.1.1`. The docs render at
   <https://pkg.go.dev/github.com/webshare-proxy/webshare-go>.

## Credentials

None. Releases are authenticated by push access to the repo and the built-in
`GITHUB_TOKEN`; there are no registry secrets.
