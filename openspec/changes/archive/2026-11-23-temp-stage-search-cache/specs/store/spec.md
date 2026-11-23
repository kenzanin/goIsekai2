# Spec Delta

## ADDED Requirements

### Requirement: JSON1 cache view

The store SHALL expose a read-only `api_cache_view` created at open time. It SHALL project `params_json` and `payload` into queryable columns with SQLite JSON1 functions without duplicating stored payload bytes. The view is a V2 convenience object and SHALL NOT change the frozen table inventory.

#### Scenario: View exposes JSON fields
- **WHEN** an `api_cache` row is written using JSON params and payload
- **THEN** `api_cache_view` reports its `query_text`, `ok`, `total` and `bytes`

#### Scenario: View survives restart
- **WHEN** the daemon restarts with the same database
- **THEN** `api_cache_view` still exists because it is created idempotently at open
