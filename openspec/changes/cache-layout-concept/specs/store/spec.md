# Spec Delta

## ADDED Requirements

### Requirement: SQLite api_cache is the sole response cache

Plugin, search and enrichment response payloads SHALL live in the SQLite `api_cache` table with TTL. The host SHALL NOT write response bodies to an on-disk `api/` tree.

#### Scenario: Search response cached
- **WHEN** `/api/search` stores a plugin response
- **THEN** the payload is queryable through `api_cache_view` and no disk API file is created

#### Scenario: Prefetch uses the image tree
- **WHEN** a reader window prefetches page bodies
- **THEN** they land in the structured image cache rather than `api/prefetch`
