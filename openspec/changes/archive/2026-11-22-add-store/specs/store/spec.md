# Spec Delta

## ADDED Requirements

### Requirement: Daemon store lifecycle

The daemon SHALL open and migrate the configured SQLite library before it starts listening on the
HTTP port. If a migration fails, the daemon SHALL exit non-zero, name the failing version and leave
the previous `schema_version` intact. The `--dump-config` and `--help` paths SHALL NOT open or create
the database. On graceful shutdown the daemon SHALL close the store after in-flight HTTP handlers
finish.

#### Scenario: Migration failure blocks listening
- **WHEN** a stored migration fails during daemon boot
- **THEN** the daemon exits non-zero naming the version, no HTTP listener is opened, and the database
  stays at the previous version

#### Scenario: Config dump does not touch the database
- **WHEN** `goisekai2 --dump-config <path>` runs against a data directory with no database
- **THEN** it writes the ini and exits without creating any SQLite file

#### Scenario: Store closes after HTTP shutdown
- **WHEN** the daemon receives SIGTERM while a request is in flight
- **THEN** the request completes, the HTTP server shuts down, and the store connection is closed last

#### Scenario: Fresh boot is usable
- **WHEN** the daemon starts against an empty data directory
- **THEN** all migrations run in order, `schema_version` reaches the latest number, and `/settings`
  becomes reachable
