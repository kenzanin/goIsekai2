---
type: Fact
title: QuickJS thrown-string errors keep exact text; match wrapped sentinels with Contains
description: "A string thrown inside quickjs Eval surfaces verbatim as Error.raw, so sentinels like \"promise return unsupported\" survive %w-wrapped chains only as substrings."
tags: [internal/plugin, js.go, quickjs, error-handling]
generated: { by: agent/mcp, at: "2026-11-21T17:45:08Z" }
status: stable
---

generated: { by: "kit agent", at: "2026-11-22" }

Verified while implementing `internal/plugin/js.go` (promise-result contract test).

Evidence:
- Test failure before the fix: `Call with Promise result = plugin: js call: promise return unsupported, want "promise return unsupported"` — the JS `throw"promise return unsupported"` reached Go with its exact text, prefixed only by the host's `%w` wrapper.
- Mechanism in modernc.org/quickjs v0.27.0 `quickjs.go`: `errFromException` → `ErrorFromValue` → `Error.raw = m.errString(v)`; for a thrown string, `errString`/`cString` returns the string verbatim, so `err.Error()` on the unwrapped error equals the thrown text.
- After wrapping (`fmt.Errorf("plugin: js call: %w", err)`), equality on the bare sentinel fails; `strings.Contains(err.Error(), "promise return unsupported")` passes (test now green).

Consequence: sentinel errors thrown inside QuickJS scripts must be matched with `strings.Contains` (or a dedicated unwrap), never `==`, anywhere a host layer wraps the eval error. If a caller ever needs `errors.Is` semantics, promote the sentinel to a typed Go error at the wrap site instead of relying on the JS string.
