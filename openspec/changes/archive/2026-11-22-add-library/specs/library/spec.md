# Spec Delta

## ADDED Requirements

### Requirement: Library route wiring

The daemon SHALL register the library routes before it opens the HTTP port. Every saving mutation
SHALL commit to the store before publishing its domain event, and SHALL never publish on a rejected
or failed write. The grid and search pages SHALL use only enabled plugins as search sources, while
saved titles from disabled plugins SHALL still list and open from stored data.

#### Scenario: Events follow the commit
- **WHEN** `POST /api/library` succeeds
- **THEN** the row is committed first and `library.saved` is published afterwards with plugin and slug

#### Scenario: Rejected save publishes nothing
- **WHEN** a save has an empty plugin or placeholder name that would overwrite a stored name
- **THEN** no row is written and no event is published

#### Scenario: Disabled plugin titles stay visible
- **WHEN** a plugin is disabled and its saved titles are listed
- **THEN** they appear in the library grid but are not offered as a search source
