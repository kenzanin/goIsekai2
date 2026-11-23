# Spec Delta

## MODIFIED Requirements

### Requirement: Search result preparation

The explicit preparation action SHALL write raw detail and chapter envelopes into the process-lifetime TEMP table `stage_detail`, not into persistent `api_cache`. The staged row SHALL be keyed by `(plugin_id, slug)` and SHALL be deleted after promotion.

#### Scenario: Prepare writes JSON only
- **WHEN** the preparation action succeeds for a new slug
- **THEN** `stage_detail` holds both envelopes and `saved_manga` is unchanged

#### Scenario: Prepare failure is non-fatal
- **WHEN** the source is unreachable during preparation
- **THEN** detail renders placeholders and the Save action still works

#### Scenario: Staged row expires with the process
- **WHEN** the daemon restarts
- **THEN** `stage_detail` is empty and a new search-result click stages data again
