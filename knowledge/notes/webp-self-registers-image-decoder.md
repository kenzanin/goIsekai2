---
type: Fact
title: gen2brain/webp registers its decoder with image on import
description: "Importing github.com/gen2brain/webp v0.6.4 registers a \"webp\" handler (RIFF????WEBPVP8) with image.RegisterFormat, so image.Decode parses webp wherever the package is linked."
tags: [cache, cover.go, webp]
generated: { by: agent/mcp, at: "2026-11-22T00:56:17Z" }
status: stable
---

## Fact

`github.com/gen2brain/webp` v0.6.4 registers itself as an `image` decoder on import (module source, `webp.go:257`):

```go
image.RegisterFormat("webp", "RIFF????WEBPVP8", decodeWrapper, DecodeConfig)
```

Proven behaviourally: `TestResizeCoverDownscalesLongestEdge` (internal/cache/cover_test.go) feeds the encoder output of `ResizeCover` back through plain `image.Decode` and asserts `format == "webp"`; the full package gate is green:

```text
$ CGO_ENABLED=0 go test ./internal/cache/...
ok      goisekai2/internal/cache    1.102s    -> exit 0
```

## Consequence

`internal/cache/cover.go` needs no extra blank import to re-decode webp: importing the package for `Encode` (which the transcode path does) enables `image.Decode` for cached webp everywhere the cache package is linked. Only jpeg/png need explicit blank imports there.

## Related

- [Image cache: v3 layout](../architecture/cache-layout.md): consumers of the stored webp bytes.
