---
type: Fact
title: "tools.ToolSet: storage delegates to plugin_state, missing keys return gorm not-found"
description: "internal/plugin/tools passes CGO_ENABLED=0 with ToolSet storage mapped onto store Put/GetPluginState keyed by PluginID, and StorageGet surfaces gorm.ErrRecordNotFound for absent keys."
tags: [internal/plugin/tools, internal/store, plugin-host]
generated: { by: agent/mcp, at: "2026-11-21T18:31:03Z" }
status: stable
---

Verified 2026-11-22 by implementing the host-tool surface and running the package gates.

- `ToolSet.StorageSet/StorageGet` delegate to `store.PutPluginState/GetPluginState(ctx, t.PluginID, key, ...)`; rows are scoped by the composite PK `(plugin_id, key)`, proven by the cross-plugin miss assertion in `TestStorageScopedPerPlugin`.
- `StorageGet` on a missing key returns the raw gorm error (`record not found`), visible in the test run:

```
CGO_ENABLED=0 go test ./internal/plugin/tools
2026/11/22 01:30:01 .../internal/store/positions.go:54 record not found
[0.064ms] [rows:0] SELECT * FROM `plugin_state` WHERE plugin_id = "other" AND key = "last_sync" ...
ok  goisekai2/internal/plugin/tools
```

Consequence: the host binding layer (js/lua/wasm lanes) must translate that not-found error into the plugin-visible "empty value" result; it is not pre-swallowed in tools. `ConfigGet` resolves via `Snapshot.Get` and returns `("", false)` for unknown knobs; `LogInfo/LogError` attribute lines with `plugin=<PluginID>` and fall back to `slog.Default()` when `Log` is nil.

# Related Concepts
- [Architecture overview: layers, import rules and frozen external contracts](../architecture/overview.md): Implements the host-tools slice of the plugin layer; ToolSet is the per-plugin capability bundle the runtimes bind.
- [Zero config.Value renders "0", not "" — absent-knob paths must early-return](config-zero-value-string.md): ToolSet.ConfigGet carries the explicit not-found branch because a zero config.Value renders "0".
