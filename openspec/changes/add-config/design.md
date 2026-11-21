# Design

## Context

See `proposal.md` — Why. Constraints that shape the approach:

- Behavioral baseline is `~/tmp/cpp/cppIsekai`: `src/config/registry.hpp` (declarative knobs),
  `--dump-config` generated from the registry, `POST /config` merge-save that keeps ini comments, and
  `GET /settings` as a projection of that registry.
- The target contract is already written: `openspec/specs/config/spec.md` (10 requirements),
  `knowledge/architecture/overview.md` (layers and import rules),
  `knowledge/architecture/concurrency-ownership.md` (one owner per resource),
  `knowledge/decisions/event-dispatch.md` and `knowledge/decisions/http-framework.md`.
- No application code exists yet, so this change also lays down the two shared packages everything
  else imports (`internal/events`, `internal/http`) and must not leave a shim behind.

## Goals / Non-Goals

**Goals:**

- One registry declaration drives defaults, validation, `--dump-config`, `--help` and the settings page.
- Provenance is observable: every knob reports which layer won.
- A save is all-or-nothing on disk and observable in-process (event, then live apply).
- The event bus and HTTP shell are correct enough that later capabilities wire into them instead of
  rewriting them.

**Non-Goals:**

- No store, plugin host, cache, reader or FSM code.
- No stream hub / `/ws/log`, no SSE — that belongs to the log/plugin work.
- No config file watching; reload is explicit (save through the endpoint or restart).
- No knob whose subsystem does not exist yet, beyond declaring it in the registry so the dump and
  settings page are complete.

## Decisions

### D1 — Hand-written INI reader/writer, not a library (baseline: `src/config` ini handling)

Ports from: the baseline keeps the ini as the single on-disk source and preserves comments on save.

`internal/config/ini.go` keeps an ordered document model: `[]section{name, []entry{comment lines, key,
rawValue, trailingComment}}`, so a save rewrites only the touched `rawValue` and re-emits everything
else byte-for-byte. The same model is what `POST /config` merges into.

Alternatives considered: `gopkg.in/ini.v1` (used by goIsekai V1) loses nothing on read but complicates
guaranteed comment/order preservation and pulls a dependency for ~150 lines of parsing; writing the
file from a template loses unknown keys (which the spec forbids). Rejected.

### D2 — stdlib `flag` with registry-derived bindings (baseline: `cli-config`)

Each registry entry declares `Flag` (name) and optional `Env` (derived as
`GOISEKAI_<SECTION>_<KEY>` upper-snake when not overridden). The flag set is generated from the
registry at boot, so an undeclared knob cannot have a flag and a declared one always has one.
`--config`, `--dump-config` and `--help` are the only hand-written flags. Explicit-pass detection uses
`flag.Visit` (only flags actually present override the ini).

Alternatives: `pflag`/`cobra` — no subcommands, no POSIX clustering needed; stdlib `flag` plus
`flag.Visit` is enough and keeps the CLI surface identical to the baseline's.

### D3 — Two resolution modes: strict boot, lenient projection (baseline: `cli-config` + `settings-page`)

Boot resolves once and fails non-zero on an unparseable or out-of-range value, naming the key, the
value and the layer. `GET /config` re-resolves the file on every request and reports a knob whose ini
text cannot be parsed with the registry default plus `invalid: true`, so the settings page can repair
a file that would block boot. Divergence recorded in the spec requirement: the baseline is lenient in
both places; strict boot is the deliberate difference (a bad value must not silently run).

### D4 — Immutable snapshot with a generation counter, not an FSM (baseline: immutable resolved config)

`Snapshot` is a struct of resolved values plus `Provenance map[string]Layer` and `Generation uint64`.
Components receive the snapshot by value at boot; the daemon holds an `atomic.Pointer[Snapshot]` that
a save swaps after the write succeeds. Per `knowledge/architecture/state-machines.md`, config is
explicitly *not* a state machine — a reload is a pointer swap, not a transition.

### D5 — LiveApplier registry, staged as subsystems arrive (baseline: `plugin.fetch_profile` applies live)

`config` exposes `RegisterLiveApplier(key string, func(old, new Value) error)`; the daemon calls
`ApplySnapshot(new)` after publishing, which walks the changed keys, invokes registered appliers, and
collects keys without an applier into the response as `restartRequired`. This change registers exactly
one applier (`log.level`, via a `slog.LevelVar`); cache cap, proxy and fetch timeout register from
their own changes later. The rule "no applier ⇒ restart required" is asserted in a test, so a future
subsystem that forgets to register fails visibly instead of silently not applying.

### D6 — Typed event facade over `gookit/event`, one owner goroutine

`internal/events` holds the manager behind a single actor goroutine: `Publish(Event)` sends a typed
struct over a channel; the actor dispatches synchronously to subscribers registered for that topic,
recovers panics per handler (counter + `slog.Warn`), and never blocks a subscriber that enqueues.
`gookit/event`'s `Async`/`FireAsync` are never called (they block the publisher when the buffer is
full). Topic names are constants; payloads are typed structs, no `event.M` type assertions at call
sites.

Alternatives: a hand-rolled hub (rejected earlier by the user's decision to use `gookit/event`) and
direct function calls (rejected because `config.changed` is a fact other subsystems react to, not a
command). See `knowledge/decisions/event-dispatch.md`.

### D7 — HTTP shell: `go-chi` + `html/template` with an embedded asset set (baseline: `settings-page` nav)

Routes are registered from a small `Pages` list (`{path, title, handler}`); the base layout renders
navigation from the same list and marks the active page, which is how "no dead links" is enforced
structurally rather than by template conditionals. Templates and static files are `//go:embed`ed;
no build step, no CDN. Handlers stay `func(w http.ResponseWriter, r *http.Request)`.

Alternatives: a full framework (gin/echo lock handlers to their context type) — rejected in
`knowledge/decisions/http-framework.md`.

### D8 — Atomic write protocol for the ini

Save path: resolve new values → validate every submitted key → render the merged document into a
temporary file in the same directory with mode `0600` (the file can hold proxy credentials and API
keys) → `fsync` the file → `rename` over the target → `fsync` the directory. Failure at any step
leaves the original untouched and yields `{"ok": false, "error": ...}`. The file is created from the
dump template when absent.

### D9 — Settings page renders provenance, submits only changed fields

The page is server-rendered. Each row shows section, key, doc, resolved value and the winning layer.
Rows won by `env` or `flag` render disabled with a hint naming the source; rows won by `ini` or
`default` render editable. Restart-only rows carry a badge from the same `restartRequired` rule the
save response uses, so the page and the API cannot disagree. The form posts only dirty fields as
`{"section.key": value}`.

### D10 — Test strategy (no framework beyond stdlib `testing`)

- Precedence matrix: table-driven over default/ini/env/flag combinations, asserting value *and*
  provenance layer.
- INI round-trip: golden files with comments, odd spacing, unknown keys, duplicates; assert byte
  equality after an unrelated save.
- Atomicity: write failure (read-only dir) leaves the original bytes intact.
- Boot strictness vs projection leniency: same malformed ini, two different outcomes.
- Event ordering: a recorder asserts `config.changed` is published only after the file content
  changed, and not at all when validation rejects.
- Actor: FIFO order, panic in one handler does not stop later events, goroutine-leak check on close.
- HTTP: `httptest` against the mux for `/settings`, `GET /config`, `POST /config`, including the
  read-only rendering of an env-won knob.

## Risks / Trade-offs

- [Hand-written INI may miss a formatting edge the baseline's parser accepts] → golden-file tests
  cover comments, blank lines, `;`/`#` comments, CRLF, duplicate keys and BOM; unknown syntax is
  preserved verbatim rather than normalised.
- [Strict boot breaks a workflow the baseline tolerated] → the save path and the projection stay
  lenient, so a broken ini is always repairable from `/settings`; the error message names the key.
- [Staged live appliers could silently leave a knob unapplied] → `ApplySnapshot` returns
  `restartRequired` for every key without an applier, and the response/scenario asserts it.
- [Event bus correctness is easy to get subtly wrong] → single actor (no concurrent access to the
  manager), panic isolation, and the async API banned by convention plus a test that the exported
  surface has no async entry point.
- [Settings page becomes a second config store] → the page has no state of its own; it reads
  `GET /config` and writes `POST /config`, and only dirty fields are submitted.

## Migration Plan

Greenfield: nothing to migrate. First boot with a missing ini writes nothing (boot only reads);
`--dump-config <path>` is the documented way to create one, and `POST /config` creates it from the
template if the page is used before that. Rollback is `git revert` — no data is written outside the
ini.

## Open Questions

- Whether `--dump-config` should also emit the flag/env names as comments. Deferrable: the dump's
  content is specified as "section, doc comment, `key = default`", so adding a comment line later is
  additive and does not change a requirement.
