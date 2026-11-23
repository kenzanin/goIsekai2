# Spec Delta

## ADDED Requirements

### Requirement: Compact saved indicator and resume control

For a saved manga, the detail page SHALL render the saved-state toggle below the synopsis using the short label `highlight` with a remove-from-library title. It SHALL also render a Start reading or Continue reading control, labelled `CH<n>` and `CH<n> P<m>` from the stored position, linking to the exact chapter id.

#### Scenario: Saved indicator placement
- **WHEN** a saved manga detail page is rendered
- **THEN** no remove button appears under the title and a `highlight` button appears below the synopsis

#### Scenario: Resume from stored position
- **WHEN** a stored reader position exists at chapter 12 page 3
- **THEN** the detail page shows `Continue CH12 P3` linking to that chapter id

#### Scenario: Start without position
- **WHEN** chapters exist but no reader position is stored
- **THEN** the detail page shows `Start reading CH1` linking to the first chapter
