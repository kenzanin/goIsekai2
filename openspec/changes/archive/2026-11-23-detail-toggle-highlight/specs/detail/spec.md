# Spec Delta

## MODIFIED Requirements

### Requirement: Compact saved indicator and resume control

The detail page SHALL render one saved-state toggle below the synopsis: `☆ Save` in the normal state and `★ Saved` highlighted in amber when saved, with `aria-pressed` reflecting membership. Clicking it toggles membership in place. Start/Continue and Fetch/Force/Reset SHALL share that action row, distributed evenly; no action buttons SHALL be clustered under the title.

#### Scenario: Saved indicator placement
- **WHEN** a saved manga detail page is rendered
- **THEN** no action buttons appear under the title and a highlighted `★ Saved` toggle appears below the synopsis

#### Scenario: Unsaved toggle is normal
- **WHEN** an unsaved manga detail page is rendered
- **THEN** the toggle reads `☆ Save` without the amber highlight and has `aria-pressed="false"`

#### Scenario: Toggle lights up after save
- **WHEN** the user clicks `☆ Save` and the view refreshes
- **THEN** the button reads `★ Saved`, is highlighted, and has `aria-pressed="true"`

#### Scenario: Resume from stored position
- **WHEN** a stored reader position exists at chapter 12 page 3
- **THEN** the detail page shows `Continue CH12 P3` linking to that chapter id

#### Scenario: Start without position
- **WHEN** chapters exist but no reader position is stored
- **THEN** the detail page shows `Start reading CH1` linking to the first chapter
