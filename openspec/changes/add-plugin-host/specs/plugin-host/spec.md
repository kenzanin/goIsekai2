# Spec Delta

## ADDED Requirements

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
