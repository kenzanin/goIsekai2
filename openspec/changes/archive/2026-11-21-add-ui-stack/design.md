# Design: add-ui-stack

## Context

See `proposal.md` for motivation. The frozen contract is the `config` capability: registry rows,
provenance, dirty-field save, restart badges and the exact navigation list. cppIsekai's
`settings-page` baseline defines those behaviors but does not freeze markup or a JavaScript
framework, so the engine swap is an internal divergence, not a contract change. The shipped
`html/template` page is the starting point; this change ports it without changing `GET/POST /config`.

## Goals / Non-Goals

- Goals: typed SSR components, single compiled stylesheet, vendored client libraries, repeatable
  local build, unchanged API and test coverage.
- Non-Goals: new pages, SPA routing, theme system, WebSocket console, store/schema work.

## Decisions

### D1 — templ for SSR components

Port `/settings` from `html/template` to `github.com/a-h/templ` v0.3.1118. `templ generate` emits
`*_templ.go` which is committed; `make build` depends on generation and a freshness gate fails when
generated files drift. Baseline: cppIsekai renders page markup server-side with crow; the visible
contract (routes, rows, badges) is unchanged — only the template engine changes.

Alternative: keep `html/template` (rejected by user), React/Vue (rejected as SPA toolchains that
duplicate server state).

### D2 — Tailwind CLI for deterministic CSS

Compile `tailwind.config` inputs to one static CSS file with the standalone Tailwind v4.3.5 CLI;
embed the output via `go:embed`. No Node in the runtime image and no CDN `<link>`. Baseline: the
C++ page ships its own stylesheet in the binary; this keeps that local-first property.

Alternative: hand-written CSS (loses the requested utility workflow), Vite/PostCSS (adds a Node
build chain the daemon does not need).

### D3 — Alpine for local state, htmx for server fragments

Vendor Alpine v3.17.5 and htmx v2.0.11 minified files under the HTTP package and embed them. Alpine
owns dirty-field tracking and status banners; htmx owns the form POST and in-place status swap; the
JSON `/config` endpoint stays the single save contract. Baseline: cppIsekai's settings page posts
to the daemon and reloads; this change keeps the same endpoint but removes the full reload — an
additive UX improvement, recorded as a MODIFIED config delta.

Alternative: one library for both (blurs state ownership), plain `fetch()` (the shipped version,
rejected by user), client-side routing (out of scope).

### D4 — Generated assets are part of the diff

`*_templ.go`, compiled CSS and vendored JS are committed. `make generate` is explicit and
idempotent; `make build` reads the committed assets so a fresh clone builds with plain
`CGO_ENABLED=0 go build ./...` and no Tailwind/templ CLI. `make lint` includes a freshness check that
fails when regeneration would change committed assets. This makes "fresh clone serves the UI"
testable rather than aspirational.

Alternative: generate only in CI (fails a local `go build` on a clean checkout).

## Risks / Trade-offs

- [templ codegen adds a build step] → `make generate` is wired into `make build` and guarded by a
  freshness check; generated files are committed.
- [Tailwind standalone binary must be available] → `make generate` prefers the standalone CLI and
  falls back to `npx --yes --package @tailwindcss/cli@4.3.5 tailwindcss`; the compiled CSS is
  committed, so `make build` never needs either.
- [Alpine and htmx overlap] → strict split: Alpine never issues save requests, htmx never holds
  form state.
- [Markup change breaks a consumer] → JSON routes and payloads are unchanged; httptest asserts the
  behavioral contract, not CSS class names.

## Migration Plan

1. Add codegen/tooling and verify `CGO_ENABLED=0 go build ./...` before touching behavior.
2. Port `/settings` to templ components, swap CSS/JS assets, keep handlers and JSON shapes.
3. Run the existing config HTTP/e2e tests unchanged, plus new no-CDN and freshness checks.
4. Archive only after all gates pass and the main `config` spec has the delta applied.

Rollback: revert the change commit; archived `add-config` remains the known-good UI.

## Open Questions

None. The user directed the stack and the change scope; no deferrable unknown remains.
