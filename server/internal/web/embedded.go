package web

import "embed"

// distFS contains the built SPA. It is populated by the Dockerfile's client-build
// stage which copies client/dist into ./dist before `go build`.
//
// For local Go runs without Docker, a stub dist/index.html is committed so the
// embed compiles and tests pass; the real (built) bundle is what gets served in
// production.
//
//go:embed all:dist
var distFS embed.FS
