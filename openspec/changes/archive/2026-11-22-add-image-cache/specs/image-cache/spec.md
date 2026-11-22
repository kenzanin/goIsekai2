# Spec Delta

## ADDED Requirements

### Requirement: Image cache route wiring

The daemon SHALL register `/cache/img` before it opens the HTTP port. The route SHALL require full
placement context in the query; a request missing plugin, slug, chapter or page context SHALL be
refused with 400 and SHALL NOT create any file. Cache reads SHALL update LRU recency, and a reindex on
restart SHALL rebuild that recency from file modification times. Prefetch SHALL run as a direct call
from the reader service, never as an event.

#### Scenario: Missing context is refused
- **WHEN** `/cache/img` receives only a URL query
- **THEN** it answers 400 and the cache directory is unchanged

#### Scenario: Restart rebuilds recency
- **WHEN** the daemon restarts with a populated cache
- **THEN** cached files still serve and the LRU order is rebuilt from their modification times
