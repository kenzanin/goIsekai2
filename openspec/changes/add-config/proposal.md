# Proposal

## Why

`config` is first in the porting order because every later capability boots from a resolved
configuration snapshot: the store needs `paths.data_dir`, the plugin host needs `plugin.plugins` and the
fetch profile, the cache needs its cap and format. If those subsystems each read the environment or
the ini on their own, precedence drifts and the dump/save surface stops being the single source of
truth the baseline guarantees.

Doing it first also forces the two pieces the other ten changes reuse: the minimal daemon HTTP shell
(router, base template, navigation) and the typed event bus, because saving config is the first
observable domain event (`config.changed`).

## What Changes

- Declarative knob registry becomes the only place a default, a doc line, an allowed range or a
  flag/env name is written; `--dump-config`, validation and help are generated from it.
- Layered resolution default < ini < `GOISEKAI_*` env < explicit flag, recording per-key provenance;
  boot is strict (bad value exits non-zero) while the read projection is lenient so the settings page
  can repair a broken ini.
- Hand-written INI reader/writer that preserves comments, key order and unknown keys, with an atomic
  write (temp file, fsync, rename). No third-party ini dependency.
- `--config <path>` is required (usage + exit 2 otherwise), `--dump-config [path]` never clobbers,
  `--help` works without `--config`.
- Resolved configuration is an immutable snapshot; `POST /config` persists, publishes
  `config.changed`, then applies the knobs that have a registered live applier. Knobs whose
  subsystem does not exist yet report restart-required instead of pretending to apply.
- Minimal daemon shell: `go-chi` router, one `html/template` base layout with a navigation bar,
  embedded assets, `GET /settings` (server-rendered knob rows with provenance and restart badges),
  `GET /config` (JSON projection), `POST /config` (merge-save).
- Typed event facade over `gookit/event` behind one owner goroutine: sync dispatch, panic isolation,
  async API banned, a recorder for tests. First topic: `config.changed`.

**Out of scope:** store/SQLite, plugin host, host tools, cache, reader, FSM lifecycles, the stream
hub and `/ws/log`, the tray binary, and any live applier whose subsystem does not exist yet
(cache cap, proxy, fetch timeout are registered when their subsystem lands).

## Capabilities

### New Capabilities

None. This change implements an existing capability.

### Modified Capabilities

- `config`: the save-response and live-apply contract is refined (staged live appliers, atomic write,
  response bodies), and the navigation/daemon-shell baseline for the settings page is added.

## Impact

- New packages: `internal/config` (registry, resolve, ini read/write, snapshot), `internal/events`
  (typed facade + actor), `internal/http` (mux, templates, handlers for the settings surface).
- `cmd/goisekai2`: gains the flag surface (`--config`, `--dump-config`, `--help`) and the boot
  sequence up to "serve".
- Dependencies added: `go-chi/chi/v5`, `gookit/event` — both already decided in
  `knowledge/decisions/http-framework.md` and `knowledge/decisions/event-dispatch.md`.
- No database, no plugin surface, no cache touched. `CGO_ENABLED=0 go build ./...` stays the gate.
