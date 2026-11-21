---
type: decision
title: INI config file handled by unknwon.dev/ini/v2
description: "Use unknwon.dev/ini/v2 v2.0.0-beta.7 for reading, merging and saving the goIsekai2 ini instead of a hand-written ordered document model or the v1 package; comments, key order and unknown keys survive, byte-identical formatting does not."
tags: [config, ini, dependency, stack, save]
generated: { by: agent/cli, at: "2026-11-21T13:59:41Z" }
status: stable
---

## Context

`openspec/changes/add-config` first planned a hand-written ordered document model (~150 LOC) so that a
`POST /config` save could rewrite touched values and re-emit every other byte unchanged. The user rejected
that plan on 2026-11-21 and asked for the library instead: "config pakai .ini bro dan gunakan library ini
https://github.com/go-ini/ini".

## Decision

Read, merge and write the ini through `unknwon.dev/ini/v2` (v2.0.0-beta.7):

- `ini.LoadSources([]any{path}, ini.IgnoreInlineComment(false), ini.SpaceBeforeInlineComment(true))`
- dotted knob names map to `file.Section(section).Key(key)`
- `Section.NewKey` for the knobs being saved (it overwrites an existing key in place)
- `File.WriteTo` into the atomic temp file (same dir, mode 0600, fsync file, rename, fsync dir)

The user asked for v2 explicitly on 2026-11-21 (`go get unknwon.dev/ini/v2@latest`), which supersedes the
v1 plan in the first version of this concept. `github.com/go-ini/ini` declares its module path as
`gopkg.in/ini.v1` and continues the v1 line, so it is not the v2 source; v2 lives at `unknwon.dev/ini/v2`.
v2 replaces the `LoadOptions` struct with functional load options and takes a `[]any` source list, so the
wrapper's call shape changed but the document model did not.

## Evidence (measured 2026-11-21, `internal/config` tests against v2.0.0-beta.7, CGO_ENABLED=0)

Fixture: banner comment, inline comment, empty `proxy =`, unknown key. After load → `NewKey` →
`WriteTo` the tests assert:

| Property | Result |
|---|---|
| banner comment | preserved verbatim |
| inline comment | preserved, may be relocated by the writer |
| key order | preserved |
| unknown keys | preserved |
| empty value | preserved |
| `=` alignment | realigned per section |

So the save contract is **semantic** preservation, not byte equality. Test assertions match that; a
golden byte-identical round-trip is explicitly not a requirement.

## Consequences

- `go.mod` gains `unknwon.dev/ini/v2 v2.0.0-beta.7` — pure Go, builds under `CGO_ENABLED=0`, so the no-cgo
  gate is unaffected. It is a **beta** (`beta.7`), which is the trade-off of following the v2 line.
- The module's `go` directive rises to `1.26` because v2 requires it; the local toolchain is 1.27.3.
- `openspec/changes/add-config` D1 and task group 3 describe the v2 wrapper, not the document model.
- The atomic write protocol (D8) is untouched: the library writes into our temp file, we own fsync and
  rename.
- Risk: formatting normalisation touches the whole file on the first save (alignment), so a diff of the
  ini shows more churn than the values changed. Accepted.
- Risk: a v2 beta may change API before GA; the wrapper (`LoadIni`/`Get`/`Set`/`Save`) is the seam that
  keeps that contained to one file.

# Related Concepts
- [Stack: libraries vs hand-written code](../architecture/stack.md): supplies the ini parser row in the stack review and replaces the hand-written INI plan
- [Dependency stack: pure Go only (modernc sqlite, tls-client, gopher-lua, quickjs, wazero)](dependency-stack.md): adds gopkg.in/ini.v1 to the approved pure-Go dependency set

## Text handling rule (follow-up, user request 2026-11-21)

The user also asked to "add the unicode library". In Go that is the standard library package
`unicode/utf8` — nothing to add to `go.mod`. Rule recorded in `architecture/stack`: any length,
truncation or validation of user-visible text (title, slug, cache namespace, log line) goes through
`utf8.RuneCountInString` / `utf8.ValidString` / `[]rune` slicing, never byte indexing, because the
baseline corpus is full of non-ASCII titles.
