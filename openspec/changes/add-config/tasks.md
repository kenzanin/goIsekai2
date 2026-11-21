# Tasks

## 1. Skeleton and dependencies

- [x] 1.1 Add `go-chi/chi/v5` and `gookit/event` to `go.mod` and verify `CGO_ENABLED=0 go build ./...` still exits 0 with an empty `internal/` tree
- [x] 1.2 Create `internal/config`, `internal/events`, `internal/http` package stubs with doc comments naming their owner in `knowledge/architecture/concurrency-ownership.md`, and verify `go build ./...` and `go vet ./...` pass
- [x] 1.3 Create `cmd/goisekai2` boot skeleton (parse args, resolve config, print the resolved listen address) and verify `go run ./cmd/goisekai2 --help` prints usage and exits 0

## 2. Knob registry

- [x] 2.1 Implement `internal/config/registry.go`: declarative `Entry{Section, Key, Kind, Default, Min, Max, Allowed, Flag, Env, Doc}` with typed kinds and a `Knobs()` table — verify a table test asserts every entry has a non-empty doc, a unique `section.key` and a derivable env name (typed constructors were not needed: struct literals in one table read better)
- [x] 2.2 Declare the knob set (server.port, server.proxy, plugin.plugins, plugin.fetch_profile, plugin.fetch_timeout_ms, cache.cache_max_mb, cache.image_format, cache.cover_max_px, paths.data_dir, paths.db, paths.log_file, log.level) — verify `--dump-config` lists each once with its default and doc, and the dump re-parses to the same defaults
- [x] 2.3 Implement type/range validation returning the offending key, value and layer — verify table tests reject `cache_max_mb = -1` (ini and env), `log.level = loud` and `server.port = http`

## 3. INI handling on `unknwon.dev/ini/v2` and atomic writer

- [x] 3.1 Wrap `unknwon.dev/ini/v2` v2.0.0-beta.7 in `internal/config/ini.go`: `LoadIni` with
      `IgnoreInlineComment(false), SpaceBeforeInlineComment(true)`, section/key access by dotted name,
      and a `WriteTo` sink — verify tests keep comments, key order, unknown keys and empty
      values after a read/write cycle (semantic preservation, not byte equality: the library realigns
      `=` and relocates inline comments)
- [x] 3.2 Implement merge-save: `Set` only the touched knobs, leave every other key untouched,
      and create the file from the dump template when it is absent — verify a test saving `server.port`
      sees every other key's value unchanged and the banner comment still present
- [x] 3.3 Implement the atomic write (temp file in the same dir, mode 0600, fsync file, rename, fsync dir)
      around the `WriteTo` sink — verify a test with a read-only directory leaves the original file
      intact and returns an error

## 4. Resolution and provenance

- [x] 4.1 Implement resolution `default < ini < GOISEKAI_* < explicit flag` using `flag.Visit` for
      explicit-pass detection, returning a `Snapshot{Values, Provenance, Generation}` — verify the
      precedence matrix test asserts value and winning layer for all four layers, including "flag not
      passed keeps ini"
- [x] 4.2 Make boot strict: unparseable, unknown-enum or out-of-range ini/env values exit non-zero
      naming key, value and layer, and leave the data dir and ini untouched — verify `cmd/goisekai2`
      tests assert exit 1 and that no file was created
- [x] 4.3 Implement the lenient projection resolver used by `GET /config` (re-reads the file, marks `invalid: true`, falls back to the default for that knob only) — verify a test feeds the same malformed ini as 4.2 and asserts 200 with exactly one invalid knob

## 5. CLI surface

- [x] 5.1 Require `--config <path>`: missing `--config` without `--help`/`--dump-config` prints usage
      plus a `--help` hint and exits 2 — verify a test asserts stderr text, exit code and that no `data/`
      directory appears
- [x] 5.2 Implement `--dump-config [path]` generated from the registry, stdout when no path, refuse to
      overwrite an existing file — verify tests cover both forms and that an existing file's bytes are
      unchanged on refusal
- [x] 5.3 Wire `--help` to work without `--config` and list registry-derived flag names — verify a test
      asserts help contains `--config`, `--dump-config` and one sample knob flag

## 6. Event facade and actor

- [x] 6.1 Implement `internal/events`: topic constants, typed payload structs, `Publish`, `Subscribe`, `Close`, single owner goroutine, synchronous dispatch, per-handler panic recovery with counter and warn log — verify the actor test asserts FIFO order, that a panicking handler does not stop later events, and that `go.uber.org/goleak`-free shutdown is observed via `runtime.NumGoroutine` before/after `Close`
- [x] 6.2 Expose no async entry point and no raw `*event.Manager` outside the package — verify `go doc ./internal/events` shows no `Async`/`FireAsync` and a compile-time test keeps the manager unexported
- [x] 6.3 Add the recorder used by tests (`Record()` sink filled by subscriptions) — verify a test publishes `config.changed` and reads it back with key, old and new values intact

## 7. HTTP shell

- [x] 7.1 Implement the mux and page registry (`internal/http/server.go`): chi router, `Pages` list driving both routing and navigation, active-page marking — verify a test asserts only registered pages are routed and the nav renders exactly those links
- [x] 7.2 Add the embedded template set and base layout (nav, flash/error line, form scaffolding) with static file serving — verify a test renders `/settings` and asserts the layout contains the active `Settings` link and no library/plugins link
- [x] 7.3 Implement graceful shutdown (signal → `Server.Shutdown` with timeout) and log the bound address — verify an integration test starts the server on port 0, requests `/settings`, and shuts down cleanly

## 8. Config handlers

- [x] 8.1 Implement `GET /config` JSON projection with `section, key, fullKey, doc, type, flag, env, default, value, source` (+ `invalid` when applicable) — verify a test edits the ini on disk mid-test and asserts the next response reflects it without restart
- [x] 8.2 Implement `POST /config` merge-save: changed keys only, `{"ok":true,"written":n}` on success, `{"ok":false,"error":...}` naming the key on rejection, no write on rejection — verify tests assert both bodies and that a rejected save leaves the file byte-identical
- [x] 8.3 Implement `GET /settings`: rows grouped by section with doc, value and winning layer; `env`/`flag`-won rows disabled with a source hint; restart-only rows carrying the badge — verify a test sets `GOISEKAI_LOG_LEVEL` and asserts that row renders disabled with the env hint while an ini-won row renders an editable input
- [x] 8.4 Implement the settings form submit (dirty fields only) and its post-save re-read — verify an end-to-end `httptest` test posts one changed field, then asserts `/settings` shows the new value and the restart badge where applicable

## 9. Live apply and the first domain event

- [x] 9.1 Implement `RegisterLiveApplier` and `ApplySnapshot` returning `restartRequired` for keys without an applier — verify a test saves a knob with no applier and asserts `restartRequired` names it and no apply function ran
- [x] 9.2 Register the `log.level` applier through a `slog.LevelVar` and wire the daemon's logger to it — verify a test saves `log.level=debug`, asserts the level var changed without restart, and that the response lists no restart requirement for it
- [x] 9.3 Publish `config.changed` after the successful atomic write, before applying appliers, with `{key, old, new, generation}` — verify the recorder test asserts ordering (file changed first, then event) and that a failed write or rejected body publishes nothing

## 10. Integration checks

- [x] 10.1 Boot the daemon against a temp data dir and ini, then verify `GET /settings`, `GET /config`, a `POST /config` round trip and the file's comment preservation in one scripted `httptest`/process test
- [x] 10.2 Verify `CGO_ENABLED=0 go build ./...`, `go vet ./...`, `gofmt -l .` (empty) and `go test ./...` all pass, and that `--dump-config` output equals the golden ini
- [ ] 10.3 Update `knowledge/architecture/stack.md` with the two new dependencies' actual versions and add a one-line note in `knowledge/process/porting-order.md` moving `config` to `archived` only after archive

## Workflow follow-up

- Archive the change with `openspec archive add-config` after the review gate passes.
- Flip the `config` row in `knowledge/process/porting-order.md` to `archived` and set `store` to `active` before starting `add-store`.
