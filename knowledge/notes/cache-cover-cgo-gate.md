---
type: Fact
title: Cover transcode gate passes CGO_ENABLED=0 (default and nodynamic variants)
description: The ResizeCover webp pipeline in internal/cache passes its tests under CGO_ENABLED=0 in both the default purego-dynamic and nodynamic transpiled build variants.
tags: [cache, cover.go, webp, no-cgo]
generated: { by: agent/mcp, at: "2026-11-22T00:56:17Z" }
status: stable
---

## Fact

The cover transcode pipeline (`internal/cache/cover.go`: `ResizeCover` via gen2brain/webp v0.6.4 encode + golang.org/x/image v0.48.0 draw.CatmullRom) builds and tests green with cgo disabled, in both build variants of the webp package:

```text
$ CGO_ENABLED=0 go test ./internal/cache/...
ok      goisekai2/internal/cache    1.102s        -> exit 0
$ CGO_ENABLED=0 go test -tags nodynamic ./internal/cache/...
ok      goisekai2/internal/cache    1.624s        -> exit 0
$ go vet ./internal/cache/... ; gofmt -l internal/cache
clean, exit 0
```

Default build uses the purego dynamic path with automatic fallback; `-tags nodynamic` forces the WASM-transpiled path. Both encode a real 2000x3000 -> 533x800 webp inside the test.

## Consequence

The no-cgo rule (`decisions/dependency-stack`, `decisions/image-pipeline`) holds for the cover path: gates can stay tagless, and CI that wants the deterministic transpiled encoder can add `-tags nodynamic` without code changes.

## Related

- [Image transcode without libvips](../decisions/image-pipeline.md): the decision this evidence confirms.
