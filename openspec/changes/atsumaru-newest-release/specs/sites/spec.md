# Spec Delta

## ADDED Requirements

### Requirement: Duplicate chapter releases select the newest

For titles with multiple scanlation releases of the same chapter number, `chapters` SHALL return the release with the greatest `release_at`. All duplicate releases SHALL remain individually addressable through their composite `id` at the `pages` boundary.

#### Scenario: Newest release wins
- **WHEN** `AJDs` has Chapter 1 releases `qRWvEm` and `2V2CqE` with increasing timestamps
- **THEN** `chapters` returns `AJDs:2V2CqE` for Chapter 1

#### Scenario: Old reader links canonicalize
- **WHEN** a reader link carries the composite id of an older duplicate
- **THEN** the reader resolves the chapter number to the listed release and shows its pages

#### Scenario: Older release stays stored
- **WHEN** duplicate releases exist for one chapter number
- **THEN** both ids remain in the chapter store and only the listed one is used by reader links
