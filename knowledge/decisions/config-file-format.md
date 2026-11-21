---
type: decision
title: INI config file handled by gopkg.in/ini.v1
description: "Use gopkg.in/ini.v1 v1.68.0 (repo github.com/go-ini/ini) for reading, merging and saving the goIsekai2 ini instead of a hand-written ordered document model; comments, key order and unknown keys survive, byte-identical formatting does not."
tags: [config, ini, dependency, stack, save]
generated: { by: agent/cli, at: "2026-11-21T13:59:41Z" }
status: stable
---

## Context

`openspec/changes/add-config` first planned a hand-written ordered document model (~150 LOC) so that a
`POST /config` save could rewrite touched values and re-emit every other byte unchanged. The user rejected
that plan on 2026-11-21 and asked for the library instead: "config pakai .ini bro dan gunakan library ini
https://github.com/go-ini/ini".

Module path check: `github.com/go-ini/ini` declares its module path as `gopkg.in/ini.v1` (v1.68.0), so
`go get github.com/go-ini/ini` fails with a path mismatch and the import path is `gopkg.in/ini.v1`. That is
the same package goIsekai V1 used (V1 pinned v1.67.3).

## Decision

Read, merge and write the ini through `gopkg.in/ini.v1`:

- `ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: false, SpaceBeforeInlineComment: true}, raw)`
- dotted knob names map to `file.Section(section).Key(key)`
- `Key.SetValue` only for the knobs being saved
- `WriteTo` into the atomic temp file (same dir, mode 0600, fsync file, rename, fsync dir)

## Evidence (measured 2026-11-21, v1.68.0, CGO_ENABLED=0)

Fixture: banner comment, inline comment, empty `proxy =`, unknown key. After load → `SetValue` →
`WriteTo`:

| Property | Result |
|---|---|
| banner comment | preserved verbatim |
| inline comment | preserved but moved onto its own line above the key |
| key order | preserved |
| unknown keys | preserved |
| empty value | preserved |
| `=` alignment | realigned per section |

So the save contract is **semantic** preservation, not byte equality. Test assertions are updated to
match; a golden byte-identical round-trip is explicitly not a requirement.

## Consequences

- `go.mod` gains `gopkg.in/ini.v1 v1.68.0` — pure Go, builds under `CGO_ENABLED=0`, so the no-cgo gate is
  unaffected.
- `openspec/changes/add-config` D1 and task group 3 now describe the library wrapper, not the document
  model.
- The atomic write protocol (D8) is untouched: the library writes into our temp file, we own fsync and
  rename.
- Risk: formatting normalisation touches the whole file on the first save (alignment), so a
  `git diff`-style review of the ini shows more churn than the values changed. Accepted.

# Related Concepts
- [Stack: libraries vs hand-written code](../architecture/stack.md): supplies the ini parser row in the stack review and replaces the hand-written INI plan
- [Dependency stack: pure Go only (modernc sqlite, tls-client, gopher-lua, quickjs, wazero)](dependency-stack.md): adds gopkg.in/ini.v1 to the approved pure-Go dependency set
