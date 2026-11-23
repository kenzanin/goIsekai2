# Spec Delta

## ADDED Requirements

### Requirement: Chapters sort newest first

Detail and chapter display lists SHALL order chapters by stored release time descending, then chapter number descending. The reader's internal navigation order is unchanged.

#### Scenario: Newest chapter on top
- **WHEN** a manga has chapters 1 through 25 stored
- **THEN** the detail chapter list starts at chapter 25 and page 2 continues from lower numbers

#### Scenario: Duplicate numbers
- **WHEN** two releases share a chapter number
- **THEN** the newer release is listed first

### Requirement: Detail cover payload stays raw

The detail save toggle SHALL send the upstream cover URL, never the rendered `/cache/img` URL. The save endpoint SHALL unwrap any proxy layer before persisting.

#### Scenario: Unsaved then saved covers
- **WHEN** a manga is unsaved and saved again from the detail toggle
- **THEN** the stored cover remains the upstream URL and the next render shows one proxy layer
