---
type: Fact
title: Mangafire and mangahub need host tools the runtime never provided
description: "After the Lua 5.5 runtime swap, atsumaru works end to end; mangafire and mangahub still fail on host.crypto.vrf_sign, host.re2.capture and host.codecs.base64_decode, which no runtime binding or dispatcher ever implemented."
tags: [plugins, lua, host-tools, gaps]
generated: { by: agent/mcp, at: "2026-11-22T15:01:16Z" }
status: stable
---

## Evidence

Live daemon on :18092 after the modernc.org/lua port:

- `/api/search?q=one piece&plugin=atsumaru` -> ok, 20 items; refresh/detail/chapters/pages -> 1199 chapters and 57 page URLs.
- `/api/search?...&plugin=mangafire` -> `plugin:40: attempt to index a nil value (field 'crypto')`; line 40 calls `host.crypto.vrf_sign`, which has no dispatcher case.
- `/api/search?...&plugin=mangahub` -> `attempt to index a function value (field 're2')`; line 30 calls `host.re2.capture`, and line 38 calls `host.codecs.base64_decode`; neither exists.

## Scope

These are pre-existing host-tool gaps, not regressions from the Lua engine swap. They are recorded here rather than silently fixed inside the runtime change. Add the missing tools in a separate host-tools change if those sites are needed.

# Related Concepts
- [Lua runtime: modernc.org/lua v0.1.0 for Lua 5.5.1 plugins](../decisions/lua-runtime.md): E2E evidence and out-of-scope gaps observed during the Lua 5.5 port.
