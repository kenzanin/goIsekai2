---
type: Fact
title: Config resolution gotchas found while building add-config
description: "Registry bounds use 0 as unbounded, empty ini/env values count as unset, and gofmt is not on PATH so the Makefile must resolve it through the GOROOT."
tags: [internal/config, make, tooling]
generated: { by: agent/cli, at: "2026-11-21T14:17:12Z" }
status: stable
---

- `Entry.Min`/`Max` use 0 to mean "no bound": `Min=1, Max=0` is valid, so the validator must test `Min != 0 && v < Min` and `Max != 0 && v > Max` separately. A combined `(Min!=0||Max!=0) && (v<Min||v>Max)` panics at table construction because `Knobs()` treats an invalid default as a programming error.
- `Resolve` treats an empty string from ini/env as "unset" (layered precedence: default < ini < env < explicit flag); only `flag.Visit` marks a flag explicit, so an unset flag keeps the ini value.
- `Knobs()` panics on an invalid table on purpose: a bad default is a build-time mistake, not a runtime condition, and every test that calls it fails loudly.
- `gofmt` is not on PATH in this environment (Go comes from `~/sdk`), so the `Makefile` uses `GOFMT := $(shell go env GOROOT)/bin/gofmt`; `make lint` fails if `gofmt -l .` prints anything.
- The daemon refuses to overwrite an existing `--dump-config` target (exit 1, file untouched) and requires `--config` (exit 2 with a usage hint) unless `--help`/`--dump-config` is given.

# Related Concepts
- [INI config file handled by unknwon.dev/ini/v2](../decisions/config-file-format.md): records implementation traps found while applying the config-file-format decision
- [unknwon.dev/ini/v2 round-trip behaviour](go-ini-v2-behaviour.md): both record verified behaviour of the config layer built for add-config
