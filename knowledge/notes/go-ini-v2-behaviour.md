---
type: Fact
title: unknwon.dev/ini/v2 round-trip behaviour
description: "v2 preserves comments and key order semantically, NewKey updates an existing key in place, and an empty value reads back as unset."
tags: [internal/config, ini, dependency]
generated: { by: agent/cli, at: "2026-11-21T14:17:12Z" }
status: stable
---

- `ini.LoadSources([]any{path}, IgnoreInlineComment(false), SpaceBeforeInlineComment(true))` keeps `;` comments and inline comments; a value containing the comment symbol is not truncated on read.
- `Section(s).NewKey(k, v)` on an existing key **updates the value in place** (verified: repeated Save calls produced one key, last value wins), so `Ini.Set` needs no exists-check.
- `WriteTo` realigns `=` and may relocate inline comments, so round-trip tests assert semantic preservation (key set, values, comments present), never byte equality.
- An empty value read back via `Key(k).String()` returns "", which the resolver treats as unset — a knob whose default is empty therefore resolves to source=default after a dump/load cycle (verified in `TestDumpRoundTripsThroughIni`).
- v2 requires `go 1.26` in go.mod.

# Related Concepts
- [INI config file handled by unknwon.dev/ini/v2](../decisions/config-file-format.md): records the library behaviour the config-file-format decision depends on
