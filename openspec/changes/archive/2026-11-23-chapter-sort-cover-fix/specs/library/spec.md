# Spec Delta

## ADDED Requirements

### Requirement: Save accepts only upstream cover URLs

`POST /api/library` SHALL unwrap `/cache/img` query layers before storing a cover, so a cover read from a rendered page can never be persisted as a proxy URL.

#### Scenario: Proxy payload sanitised
- **WHEN** a save request carries `/cache/img?...&url=<upstream>`
- **THEN** the stored cover is `<upstream>`

#### Scenario: Plain payload unchanged
- **WHEN** a save request carries a normal https cover URL
- **THEN** it is stored unchanged
