---
type: Decision
title: "Pages payload: raw JSON string array, persisted on success"
description: "reader consumes Host.Call results as raw payload: pages is a bare JSON array of URL strings, persisted on success, stale-served on failure."
tags: [internal/reader, internal/plugin, service.go]
generated: { by: agent/mcp, at: "2026-11-22T00:19:32Z" }
status: stable
---

generated: { by: "kit agent", at: "2026-11-22" }

`plugin.Host.Call` returns the plugin's `handle()` result bytes verbatim (lua.go/js.go return the
raw string; nothing unwraps a response envelope on the Go side). `internal/reader` therefore fixes
the pages payload as a bare JSON array of URL strings, requested with params `{"id": <chapter_id>}`,
and persists a successful list via `store.ReplacePages` so the stale fallback has data; a failed,
empty or unparsable reply falls back to stored `chapter_pages` in idx order.

Alternatives: envelope-wrapped responses (`{"ok":..,"data":..}` like cpp `envelope_data`) rejected
because no Go host code strips such a wrapper today; a reader-level unwrap would duplicate cpp
semantics with no producer.

Evidence:
- `CGO_ENABLED=0 go test ./internal/reader/...` -> `ok goisekai2/internal/reader 0.043s`, exit 0
  (`TestPagesFallsBackToStoredPages`: unbooted host -> stored `[u1 u2]`; `TestPagesDuplicateNumbers`
  exercises the same path for every number lookup).
- `go run scripts/godoc-check.go` exit 0.

Consequence: plugin authors emit the bare array from `handle()` for `pages`; if a response envelope
is ever introduced, `Service.pageList` is the single place to unwrap.

# Related Concepts
- [Plugin host contract v1 implemented and gated in internal/plugin](plugin-envelope-v1.md): Pages payload and fallback behaviour build on the plugin host ABI recorded in notes/plugin-envelope-v1.
