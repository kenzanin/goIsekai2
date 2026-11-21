---
type: Fact
title: "Go filename suffix trap: *_js.go is an implicit GOOS=js constraint"
description: "A Go file stem ending in _js is an implicit GOOS=js build constraint, so bindings_js.go was skipped on linux; renamed to bindings_quickjs.go."
tags: [internal/plugin, go-build, quickjs, naming]
generated: { by: agent/mcp, at: "2026-11-21T19:22:47Z" }
status: stable
---

## Evidence

`internal/plugin/bindings_js.go` (named in the worker brief) was invisible to the toolchain: `go vet` failed with `js.go:63:13: undefined: bindJS` while the file existed, and `go list -f '{{.GoFiles}}' ./internal/plugin/` omitted it. Cause: `go/build` treats a trailing `_js` in a file stem as a GOOS constraint (`js/wasm` target), so the file only builds under `GOOS=js`. A `/tmp/suffixtest` experiment showed even an explicit `//go:build !ignore` line does NOT lift the suffix rule - constraints AND with the implicit suffix match.

Fix: renamed to `bindings_quickjs.go` / `bindings_quickjs_test.go` (`quickjs` is no GOOS/GOARCH token). After rename: `CGO_ENABLED=0 go test ./internal/plugin/...` -> `ok internal/plugin 0.231s`, exit 0.

## Consequence

When naming Go files in this repo, the token after the last underscore in the stem must not equal a GOOS/GOARCH (`js`, `wasm`, `linux`, `amd64`, ...) unless a platform fork is intended. Runtime-binding pairs follow `bindings_<engine>.go` (lua, quickjs).
