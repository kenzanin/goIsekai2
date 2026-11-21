---
type: architecture
title: "Stack: libraries vs hand-written code"
description: "One page answering what goIsekai2 takes from libraries and what it hand-writes, with the open slots where no decision exists yet."
tags: [stack, libraries, dependencies, handwritten, review]
generated: { by: agent/cli, at: "2026-11-21T12:35:23Z" }
status: stable
---

# Related Concepts
- [Dependency stack: pure Go only (modernc sqlite, tls-client, gopher-lua, goja, wazero)](../decisions/dependency-stack.md): stack page is the review index over the per-area dependency decisions
- [Image transcode without libvips: gen2brain WASM-transpiled webp/avif/jxl, pure-Go resize](../decisions/image-pipeline.md): image codecs are listed as library rows on the stack page
- [Adopt gookit/event behind a single-actor bus; separate stream plane](../decisions/event-dispatch.md): event bus and stream hub rows point at the dispatch decision
