# plugin-host Specification

## Purpose

Defines how plugins are discovered, validated, executed and paused: one folder per site, one
manifest, one JSON envelope contract shared by every runtime, a persisted lifecycle machine that
decides whether a call is allowed, single-call-per-plugin serialization, per-call budgets, and the
management surface. The plugin sees host tools and nothing else — never the database, never the
event bus, never the filesystem outside its own folder.

### Runtimes

| Runtime | Implementation | Known limits that must fail loudly |
|---|---|---|
| `lua` | `gopher-lua` (Lua 5.1) | no LuaJIT `ffi`/`bit`; requiring them returns a clear "unsupported module" error |
| `js` | `quickjs` (modernc.org/quickjs) | ES2023 engine, pure Go; the host contract is still synchronous, so a plugin returning a `Promise` is an error, never a silent empty result |
| `wasm` | `wazero` (WASI preview1) | imports are flat `env.<group>_<fn>`; the module must export `handle` |

## Requirements

### Requirement: Discovery

The host SHALL discover plugins by scanning the configured plugins directory, one folder per
plugin, each containing a `main.json` manifest and the entry file named by the manifest. A folder
that fails validation SHALL be skipped with a warning, without affecting other plugins. When two
folders declare the same id, the first in scan order wins and the duplicate is logged.

#### Scenario: Valid plugin is discovered
- **WHEN** a folder holds a parseable `main.json` and its declared entry file
- **THEN** the plugin is listed with its manifest name, its folder name as id, its runtime and its capability chips

#### Scenario: Broken plugin does not break the host
- **WHEN** a manifest is unparseable, the entry file is missing, or a required field is absent
- **THEN** that plugin is skipped and logged, and every other plugin still loads

#### Scenario: Duplicate ids
- **WHEN** two folders declare id `mangadex`
- **THEN** the first folder in scan order is registered and the second is skipped with a warning naming both paths

### Requirement: Manifest fields

The manifest SHALL declare `name`, `id` (implied by the folder name when absent), `version`,
`runtime` (`lua`, `js`, `wasm`), `capabilities` (subset of `search`, `detail`, `chapters`,
`enrichment`) and `api` (integer contract version). The optional logo SHALL be one of `logo.png`,
`logo.svg`, `logo.jpg`, `logo.avif`.

#### Scenario: Manifest declares runtime and capabilities
- **WHEN** a manifest declares `runtime: "lua"`, `capabilities: ["search","detail","chapters"]` and `api: 1`
- **THEN** the plugin is registered for exactly those capabilities under the Lua runtime

### Requirement: Call envelope

Every call SHALL send `{"api": <host api>, "method": "<method>", "params": {...}}` and SHALL accept
only `{"ok": true, "data": ...}` or `{"ok": false, "error": "<message>"}`. The method vocabulary
SHALL be `search`, `detail`, `chapters`, `pages`, plus optional `genres` — the baseline ABI names,
never renamed. Calling a method the plugin does not implement SHALL return
`{"ok": false, "error": "unknown method <name>"}` without invoking plugin code.

#### Scenario: Method round-trip
- **WHEN** the host calls `method: "search"` with valid params
- **THEN** the envelope's `data` is returned to the caller and `error` is surfaced as a logged failure with plugin id and method

#### Scenario: Plugin from the future
- **WHEN** a manifest declares `api: 2` and the host supports 1
- **THEN** the plugin is skipped with a log message naming both versions and is never executed

#### Scenario: Malformed response
- **WHEN** a plugin returns neither valid JSON nor the ok/error shape
- **THEN** the call fails, the raw output is logged (truncated), and the host keeps serving

### Requirement: Runtime parity and explicit limits

The host SHALL expose one internal runtime interface (`Call(ctx, envelope) -> response`) implemented
by `lua`, `js` and `wasm`, so no runtime is reachable through a code path the others do not share. A
runtime that needs cgo, or a module the runtime does not provide, SHALL fail with an explicit error
naming the plugin rather than misbehave silently.

#### Scenario: JS plugin answers a call
- **WHEN** a registered `js` plugin defines `handle` and is called
- **THEN** the response shape matches any Lua plugin's response and the HTTP layer cannot tell the runtimes apart

#### Scenario: WASM plugin answers a call
- **WHEN** a registered `wasm` module exports `handle`
- **THEN** the envelope JSON is passed in and the module's response is returned unchanged in shape

#### Scenario: LuaJIT-only module
- **WHEN** a Lua plugin calls `require("ffi")` or uses `bit`
- **THEN** the call fails with an "unsupported module (Lua 5.1 runtime)" error naming the plugin, and the host process is unaffected

#### Scenario: Broken runtime entry
- **WHEN** a `js` or `wasm` entry fails to parse, compile or instantiate
- **THEN** the plugin is not registered, the failure is logged, and already-registered plugins keep serving

### Requirement: Persisted lifecycle machine

Each plugin instance SHALL have a persisted lifecycle with states `Disabled`,
`Enabled{Idle, Busy}` and `Failed`, driven by its owning actor only. A transition SHALL persist its
state, commit, then publish `plugin.lifecycle`. An illegal trigger SHALL return a typed error,
increment a counter, and log a warning — never panic and never be silently ignored.

#### Scenario: Disable while idle
- **WHEN** a plugin in `Enabled.Idle` receives `Disable`
- **THEN** the state row is persisted as `Disabled`, `plugin.lifecycle` is published after the commit, and subsequent calls are refused

#### Scenario: Disable while a call runs
- **WHEN** a plugin in `Enabled.Busy` receives `Disable`
- **THEN** the transition waits for the in-flight call to finish or time out, then proceeds; the call is never killed mid-flight

#### Scenario: Illegal trigger
- **WHEN** a `Disabled` plugin receives `CallStarted`
- **THEN** the call is refused with a typed error, the counter increments, and no state row changes

#### Scenario: Failed at boot
- **WHEN** a plugin is enabled in the database but its entry fails to load
- **THEN** it is marked `Failed`, listed with an error badge, and the rest of the boot proceeds

### Requirement: Boot reconciliation and parity API

At boot each plugin's lifecycle SHALL be reconciled from the database, and a plugin whose entry no
longer loads SHALL become `Failed` without aborting boot. The HTTP API SHALL keep exposing the
boolean `enabled` field for parity; the richer state is internal and shown as a page badge.

#### Scenario: Parity payload
- **WHEN** a client reads the plugin JSON
- **THEN** it receives `enabled: true|false` exactly as the baseline did, with no new required field

### Requirement: Single call per plugin

The host SHALL allow at most one plugin call in flight per plugin instance. Concurrent callers
SHALL queue in arrival order and each SHALL honour its own context deadline; a caller whose
deadline expires while queued SHALL fail with a timeout error without invoking the plugin. A reload
SHALL wait for the current call rather than interleaving with it.

#### Scenario: Queued caller times out
- **WHEN** two calls arrive and the first runs longer than the second's deadline
- **THEN** the second returns a timeout error and the plugin is called exactly once

#### Scenario: Reload waits for the busy plugin
- **WHEN** a reload is requested while a call is in flight
- **THEN** the reload takes effect after that call completes and the in-flight call sees the old code

### Requirement: Disabled plugins are inert

A disabled plugin SHALL be omitted from the search source picker and library filters, skipped by
the auto-update scheduler's candidate query, and SHALL refuse on-demand refresh and direct calls
with `{"ok": false, "error": "plugin disabled"}` and no plugin invocation. The plugin SHALL stay
installed and listed. Library, reader and saved-entry paths SHALL keep working for titles already
saved under it. Re-enabling SHALL take effect immediately with no reload.

#### Scenario: Search picker omits a disabled plugin
- **WHEN** a plugin is disabled and the search page renders its source picker
- **THEN** the plugin is not offered as a source

#### Scenario: Refresh refused
- **WHEN** `POST /api/refresh/:plugin/:slug` targets a disabled plugin
- **THEN** the response is `{"ok":false,"error":"plugin disabled"}` and no plugin call is made

#### Scenario: Saved titles still readable
- **WHEN** a plugin is disabled and a saved title under it is opened from the library
- **THEN** stored chapters and cached pages still serve; only network-backed actions are refused

### Requirement: Health is not a state

The host SHALL record the last call outcome per plugin in memory as `{ok, error, at}` and SHALL
NOT model call failure as a lifecycle transition. A restart SHALL reset health to "unknown, treated
as active" until the next call.

#### Scenario: Failing plugin visible
- **WHEN** a plugin's latest call failed with `SSL connect error`
- **THEN** its management row shows `error` with that message while its lifecycle state stays `Enabled`

#### Scenario: Recovery clears the badge
- **WHEN** a later call succeeds
- **THEN** the row returns to `active` without any state transition

### Requirement: Per-call budget

Every runtime SHALL be bounded by a computation budget derived from the plugin timeout
configuration: an instruction budget for Lua, an interrupt deadline for JS, and a context
deadline plus fuel/epoch limit for WASM. An overrun SHALL produce
`{"ok": false, "error": "...budget exceeded..."}` within the configured bound, never a hang.

#### Scenario: Runaway loop
- **WHEN** a plugin loops without returning in any runtime
- **THEN** the host returns an ok:false budget error and continues serving other plugins

### Requirement: Sandbox

Lua SHALL expose only `base`, `string`, `table`, `math`, `json`, and a read-only subset of `io`
(`io.lines`, `io.read`); `os` and `debug` SHALL be unavailable and writes SHALL be impossible. JS
SHALL expose no filesystem, network or process globals. WASM SHALL receive only WASI preview1 plus
the documented imports. Plugins SHALL NOT reach the database or the event bus in any runtime.

#### Scenario: Sandbox escape attempt
- **WHEN** a plugin calls `os.execute`, `debug.sethook`, `io.open(path, "w")`, or an equivalent JS/global escape
- **THEN** the call fails inside the plugin and is surfaced as a plugin error without harming the host

#### Scenario: Nothing but host tools
- **WHEN** a plugin attempts to open the SQLite file or dispatch an event
- **THEN** no such capability exists in its environment and the attempt fails inside the plugin

### Requirement: Reload

The host SHALL expose `POST /plugins/reload` (all plugins) and a per-plugin reload action. A
reload SHALL rescan with the same discovery rules as boot, keep the persisted lifecycle state, and
make the new code effective for subsequent calls. A plugin that fails to reload SHALL become
`Failed`; healthy plugins SHALL be unaffected.

#### Scenario: Reload after editing a plugin
- **WHEN** an entry file is edited and reload is invoked
- **THEN** the next call runs the edited code without a host restart

#### Scenario: Reload breaks one plugin
- **WHEN** the edited entry no longer parses
- **THEN** that plugin becomes `Failed`, keeps its saved titles, and the other plugins keep serving

### Requirement: Plugin management page

The page SHALL list every plugin row with logo (initial fallback), name, version, runtime badge,
capability chips, title count and a status badge (`active`, `error` with the message, `disabled`,
`failed`), a client-side filter by name or capability, a persisted enable/disable toggle, a reload
action and a title-count deep link to `/library?plugin=<id>`. With no plugins the page SHALL show
the plugins directory path and the expected folder layout instead of an empty table.

#### Scenario: Row renders
- **WHEN** the page loads with a MangaDex row
- **THEN** it shows the logo or initial, the name, version, runtime badge, capability chips, title count and status badge

#### Scenario: Title count click
- **WHEN** the title count is clicked
- **THEN** the library page lists exactly that plugin's titles

#### Scenario: Empty state
- **WHEN** no plugin rows exist
- **THEN** the guidance block with the plugins path and layout renders instead of a table

### Requirement: Removed plugin folders are purged

When a scan finds that a registered plugin id has no folder on disk, the host SHALL forget it in a
single transaction: `title`, `plugin_state`, `reader_positions`, `saved_manga`, `title_meta`,
`update_attempts` and `api_cache` rows for that plugin, then the plugin row last. A missing plugins
directory SHALL short-circuit before the prune, so an unmounted volume cannot wipe the registry.

#### Scenario: Folder removed between boots
- **WHEN** a previously registered plugin's folder is gone on the next scan
- **THEN** its rows are purged in one transaction and it disappears from the list

#### Scenario: Missing directory does not prune
- **WHEN** the configured plugins directory does not exist at boot
- **THEN** no plugin is forgotten and the condition is logged as an error

### Requirement: Plugin host boot order

The daemon SHALL finish store migrations before it scans the plugins directory, and SHALL reconcile
each plugin's persisted lifecycle before the HTTP port opens. A broken plugin SHALL be marked
`Failed` and logged without blocking other plugins or boot. A missing plugins directory SHALL be
logged as an error and SHALL NOT prune any persisted plugin row. The reload endpoint SHALL use the
same discovery and reconciliation path as boot.

#### Scenario: Boot with a broken plugin
- **WHEN** one plugin entry fails to load during boot
- **THEN** that plugin is listed as failed, every other plugin loads, and the daemon still listens

#### Scenario: Missing plugin directory does not prune
- **WHEN** the configured plugins directory is absent at boot
- **THEN** the condition is logged, no plugin row is deleted, and boot continues

#### Scenario: Reload reuses boot discovery
- **WHEN** `POST /plugins/reload` runs
- **THEN** manifests are rescanned with the same validation rules, persisted lifecycle states are kept,
  and only healthy plugins run new code
