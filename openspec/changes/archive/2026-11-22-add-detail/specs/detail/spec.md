# Spec Delta

## ADDED Requirements

### Requirement: Detail route wiring

The daemon SHALL register detail and enrichment routes before it opens the HTTP port. Saved detail SHALL be served from its stored row without a plugin call, with `cached: true`; a cold saved row SHALL be materialized from a validated envelope first. Forced refresh SHALL replace both detail and chapter entries only when both validate, and SHALL stamp `update_attempts` in every outcome. Enrichment plugins SHALL load from a registry separate from scraper plugins.

#### Scenario: Saved row avoids plugin call
- **WHEN** a saved manga's detail is requested while the plugin is unreachable
- **THEN** the stored metadata is returned with `cached: true` and no plugin call is made

#### Scenario: Refresh is all-or-nothing
- **WHEN** one refresh envelope fails validation
- **THEN** neither cache entry is replaced and the attempt is stamped `error`

#### Scenario: Enrichment stays separate
- **WHEN** `data/enrichment/<id>/main.lua` exists at boot
- **THEN** it is available as an enrichment source and absent from the scraper plugin list
