# config Specification

## Purpose

Defines the host configuration surface: the required `--config` boot argument, one typed
registry that is the only source of truth for defaults and help text, the precedence
default < ini < `GOISEKAI_*` env < explicit flag, a dump generated from the registry, a
comment-preserving save endpoint, the `GET /config` projection and the `/settings` page that
edits through it. Config is an immutable snapshot swapped on reload; no
component reads the environment or the file directly after boot.

## Requirements

### Requirement: Required config argument

The host SHALL require `--config <path>` on every boot except `--help` and
`--dump-config`. With no arguments, or with arguments lacking `--config`, it SHALL print a
usage line naming the ini path and a `--help` hint, exit with code 2, and leave both the
data directory and the database untouched.

#### Scenario: No arguments
- **WHEN** the binary runs with no arguments
- **THEN** usage plus a `--help` hint is printed, the exit code is 2, and no `data/` is created

#### Scenario: Arguments without config
- **WHEN** flags are passed without `--config` and neither `--help` nor `--dump-config` is used
- **THEN** the same usage path runs and no socket is opened and no database is opened

### Requirement: Typed registry as the single source of truth

Every knob SHALL be declared once in a registry with section, key, Go type, default value,
allowed values or range where applicable, and a one-line doc comment. Defaults, validation
and `--dump-config` output SHALL all be derived from that declaration; no default may be
duplicated in code that consumes the knob.

#### Scenario: New knob needs one declaration
- **WHEN** a developer adds a knob to the registry
- **THEN** it appears in `--dump-config`, is validated on load, and is rejected with a clear error when its ini or env value has the wrong type

#### Scenario: Unknown ini keys survive
- **WHEN** the ini contains a key the registry does not know
- **THEN** loading ignores it with a warning and saving preserves it verbatim

### Requirement: Layered configuration precedence

Every knob SHALL resolve as: built-in default, overridden by the ini value, overridden by
the `GOISEKAI_*` environment variable (upper-snake, section and key joined by `_`, e.g.
`GOISEKAI_SERVER_PORT`), overridden by an explicitly passed flag. A flag that was not
passed SHALL NOT override the ini. Resolution SHALL record provenance per key so
`--dump-config --resolved` can show where each value came from.

#### Scenario: Flag beats env beats ini
- **WHEN** `server.port` is set in the ini, `GOISEKAI_SERVER_PORT` in the environment, and `--port 9000` on the command line
- **THEN** the listener binds 9000; without `--port` it binds the env value; without the env it binds the ini value

#### Scenario: Unset flag keeps ini
- **WHEN** `cache_max_mb` is set in the ini and no `--cache-max-mb` flag is passed
- **THEN** the cache cap comes from the ini, not from the built-in default

#### Scenario: Invalid value is not silently ignored
- **WHEN** an ini or env value cannot be parsed as the declared type, or falls outside the allowed set
- **THEN** the host reports the key, the offending value and the layer, and exits non-zero instead of booting with a guess

### Requirement: Self-describing dump

`--dump-config [path]` SHALL emit every registry entry as its ini section, its doc comment,
and `key = default`, generated from the registry and never hand-maintained. It SHALL run
without `--config`, print to stdout when no path is given, write the file when a path is
given, and never overwrite an existing file.

#### Scenario: Dump lists every knob
- **WHEN** `--dump-config` runs with no path
- **THEN** stdout contains every registry key under its section, each preceded by its doc comment and showing its default

#### Scenario: Dump does not clobber
- **WHEN** `--dump-config existing.ini` runs and that file already exists
- **THEN** the command fails with a clear message and the file is unchanged

### Requirement: Comment-preserving save endpoint

`POST /config` SHALL merge submitted values into the configured ini, preserving existing
comments, key order and unknown keys; when the file does not exist it SHALL be created from
the dump template populated with the submitted values. The write SHALL be atomic (temporary
file in the same directory, fsync, rename). A value failing validation SHALL be rejected
before the file is touched.

#### Scenario: Save keeps comments
- **WHEN** the ini carries comments and `POST /config` sets `server.port`
- **THEN** the file keeps its comments and layout with only that value changed

#### Scenario: Save creates a missing file
- **WHEN** the configured ini does not exist and a save arrives
- **THEN** the file is created from the dump template containing the submitted values

#### Scenario: Rejected value leaves the file intact
- **WHEN** a save submits `cache_max_mb = -1`
- **THEN** the response reports the validation error and the ini is byte-identical to before

### Requirement: Immutable snapshot with explicit reload

The resolved configuration SHALL be an immutable value handed to every component at boot.
A save SHALL persist, then publish `config.changed`, then hand a new snapshot to the
adapters that can apply a change live (log level, cache cap, fetch timeout, proxy). Knobs
that cannot be applied live SHALL report that a restart is required rather than pretending
to take effect.

#### Scenario: Live-applicable knob
- **WHEN** `log.level` is saved
- **THEN** the new level takes effect without restart and `config.changed` names the key, old and new value

#### Scenario: Restart-only knob
- **WHEN** `server.port` is saved
- **THEN** the response says a restart is required, the value is persisted, and the running listener is unchanged

### Requirement: Read projection endpoint

`GET /config` SHALL return a JSON projection of the registry: one entry per knob carrying
`section`, `key`, `fullKey`, `doc`, `type`, `flag`, `env`, `default`, `value` and `source`, where
`value` is the resolved value and `source` is the winning layer (`default` | `ini` | `env` | `flag`).
It SHALL re-resolve from the file on every request, never mutate the ini, and never return 500 for a
malformed value: a knob whose ini text cannot be parsed is reported with the registry default,
`source: default` and `invalid: true`. (Divergence from boot, which is strict by design: boot refuses
to start on a bad value, the projection tolerates it so the settings page can repair it.)

#### Scenario: Projection reflects the on-disk ini
- **WHEN** the ini is edited on disk and `GET /config` is called
- **THEN** the returned `value` reflects the new file contents without a daemon restart

#### Scenario: Malformed value is reported, not fatal
- **WHEN** a numeric knob holds an unparseable ini value
- **THEN** `GET /config` returns 200, that knob carries the registry default with `invalid: true`, and every other knob keeps its resolved value

#### Scenario: Projection does not write
- **WHEN** `GET /config` is called repeatedly
- **THEN** the ini file's bytes and mtime are unchanged

### Requirement: Settings page projects the registry

`GET /settings` SHALL render one row per registry knob (`section.key`) grouped by section, showing its
doc text, resolved value and winning layer. Rows won by `env` or `flag` SHALL render read-only with a hint
naming the source (`GOISEKAI_*` / `--flag`) instead of offering an edit that would not apply; rows won by
`ini` or `default` SHALL be editable and submit through the existing `POST /config` merge-save.

#### Scenario: Every knob appears exactly once
- **WHEN** the user opens `/settings`
- **THEN** every registry knob (`server.port`, `server.proxy`, `plugin.*`, `cache.*`, `paths.*`, `log.level`) appears once, grouped by section, with doc text and current value

#### Scenario: Env-owned knob is read-only
- **WHEN** a knob is supplied through its `GOISEKAI_*` variable or an explicit flag
- **THEN** its row is read-only and names the winning source

#### Scenario: Save round-trip
- **WHEN** the user edits an ini-backed knob and submits
- **THEN** `POST /config` reports the write, the page re-reads `GET /config`, and the row shows the post-save value

### Requirement: Restart hints and navigation entry

Rows whose effect requires a restart SHALL carry a restart badge, and knobs declared live-applicable
(`log.level`, `cache.cache_max_mb`, fetch timeout, proxy, `plugin.fetch_profile`) SHALL not. The top
navigation SHALL include `Settings` linking to `/settings`, marked active on that page.

#### Scenario: Restart-only row
- **WHEN** the user edits `server.port`
- **THEN** the save succeeds and the row shows the restart badge until the daemon reboots

#### Scenario: Live row has no badge
- **WHEN** the user edits `plugin.fetch_profile`
- **THEN** the running host applies it without a restart and the row shows no badge

#### Scenario: Nav entry
- **WHEN** any page renders the top bar
- **THEN** a `Settings` link is present and marked active on `/settings`

### Requirement: Secret handling

Knobs holding credentials or tokens (proxy URL with userinfo, enrichment API keys) SHALL
support an environment-variable source in preference to the ini, SHALL never be echoed in
logs, and SHALL be redacted in `--dump-config --resolved` output.

#### Scenario: Secret from environment
- **WHEN** a secret knob is supplied via its `GOISEKAI_*` variable
- **THEN** it takes effect, `--dump-config` shows the placeholder value, and no log line contains the secret
