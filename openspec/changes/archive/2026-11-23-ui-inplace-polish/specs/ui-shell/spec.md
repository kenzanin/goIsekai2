# Spec Delta

## MODIFIED Requirements

### Requirement: Canonical navigation order

Every SSR page using the shared shell SHALL render nav from canonical path lists in the baseline order. Internal navigation and in-place mutations SHALL swap only the `#page-main` element through htmx with history push, reinitialise Alpine, and update the active header link without a full document reload.

#### Scenario: Baseline order is stable
- **WHEN** all six shell pages are registered
- **THEN** the header reads Library, Search, Plugins, Settings ... Log, About in that order

#### Scenario: Missing page does not create a dead link
- **WHEN** a page in the canonical list is not registered
- **THEN** it is omitted from the nav and no other page order changes

#### Scenario: Menu opens in place
- **WHEN** a nav link is clicked
- **THEN** only the main element is replaced and the browser URL updates
