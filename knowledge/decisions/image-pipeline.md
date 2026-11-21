---
type: decision
title: "Image transcode without libvips: gen2brain WASM-transpiled webp/avif/jxl, pure-Go resize"
description: "Replaces the baseline libvips pipeline with cgo-free gen2brain/webp, gen2brain/avif and gen2brain/jpegxl (libwebp/libavif/libjxl compiled to WASM and transpiled to Go), x/image draw for resize, nodynamic build tag for determinism, plus parity and benchmark gates."
tags: [decision, cache, transcode, webp, avif, jpegxl, no-cgo, image]
generated: { by: agent/mcp, at: "2026-11-21T12:19:03Z" }
status: stable
---

## Problem

The baseline transcodes images with **libvips**: buffer in, buffer out, `image_format` = `webp|avif|jxl|none`.
libvips is C. Under the project's no-cgo rule it cannot be linked, and shelling out to `vips`/`cwebp` binaries would
put a runtime dependency on the host that the pure-Go promise excludes.

## Decision

Use pure-Go codecs that wrap the reference C libraries compiled to WASM and transpiled to Go:

| Format | Package | Mechanism |
|---|---|---|
| WebP | `github.com/gen2brain/webp` v0.6.4 | libwebp -> WASM -> transpiled Go via wasm2go; optional purego dynamic lib, falls back automatically |
| AVIF | `github.com/gen2brain/avif` v0.6.0 | same mechanism around libavif |
| JPEG XL | `github.com/gen2brain/jpegxl` | same mechanism around libjxl |
| Decode/resize | `image`, `golang.org/x/image` v0.48.0 | stdlib + `draw.CatmullRom` |

All MIT, all cgo-free, all maintained (2026). The full `image_format` vocabulary survives, so the config contract does
not change.

## Alternatives rejected

| Option | Why not |
|---|---|
| libvips via cgo | violates the no-cgo rule; also drags a large C dependency tree into every build |
| shell out to `cwebp`/`avifenc` | runtime host dependency, process spawn per image, error handling through exit codes |
| `HugoSmits86/nativewebp` (pure Go encoder) | viable for webp only, but then `avif`/`jxl` knobs silently stop working |
| `kolesa-team/go-webp` | cgo |
| No transcode at all (`none` only) | changes page weight by ~10x versus the baseline; not parity |

## Consequences and verification

1. Build with the `nodynamic` tag so the transpiled path is always used: deterministic across machines, no dlopen of
   a system libwebp that may differ in version.
2. Transcode is CPU-bound and one-time per image. A slow first fetch is acceptable; a per-request cost would not be.

3. Parity gate: take the same 1.79 MB PNG the baseline smoked (-> 103 KB webp), run it through the Go pipeline with
   the same target quality, and assert the output is within tolerance (size band, dimensions, ratio) -- not
   byte-identical, since the encoder settings differ.
4. Benchmark gate: one cover and one page, cold and warm, recorded in the cache package test; regressions block the
   merge. If Wasm-transpiled encode turns out too slow, the escape hatch is a build tag that swaps in a native
   encoder, keeping the same interface -- not a cgo dependency in the default build.
