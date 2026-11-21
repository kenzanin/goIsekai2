# Spec Delta

## MODIFIED Requirements

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

#### Scenario: Successful save reports the write count
- **WHEN** a save submits two changed keys
- **THEN** the response body is `{"ok": true, "written": 2}` and both new values are on disk

#### Scenario: Rejected body names the offending key
- **WHEN** a save submits an unknown key or a non-integer for an integer knob
- **THEN** the response is `{"ok": false, "error": "<message>"}` naming that key and no file is written

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

#### Scenario: No live applier yet means restart required
- **WHEN** a knob whose owning subsystem has not booted (cache cap before the cache exists) is saved
- **THEN** the value is persisted, no live apply is attempted, and the response reports restart required

#### Scenario: Event follows the write, never precedes it
- **WHEN** the atomic write fails or validation rejects the body
- **THEN** no `config.changed` is published and no snapshot is swapped

## ADDED Requirements

### Requirement: Navigation lists only implemented destinations

The shared layout SHALL render a top navigation containing exactly the pages present in the running
build, marking the current page active. `Settings` SHALL sit immediately after `Plugins` when both
exist and SHALL be the last entry otherwise. A destination that is not implemented SHALL NOT render
as a dead link.

#### Scenario: Settings only
- **WHEN** the build serves only the settings surface and the user opens `/settings`
- **THEN** the navigation contains a `Settings` link marked active and no other entries

#### Scenario: No dead links
- **WHEN** a page that the design reserves for a later capability (library, plugins) does not exist in this build
- **THEN** the navigation does not render it at all
