# Spec Delta

## Purpose

Defines the shared SSR shell navigation inherited from the cppIsekai baseline so page order is a contract,
not an incidental property of handler registration order.

## ADDED Requirements

### Requirement: Canonical navigation order

Every SSR page using the shared shell SHALL render nav from canonical path lists: primary `/library`,
`/search`, `/plugins`, `/settings`; utility `/log`, `/about`. Unlisted registered pages are omitted; listed
unregistered pages are skipped silently. The full-screen reader is exempt because it owns its overlay.

#### Scenario: Baseline order is stable
- **WHEN** all six shell pages are registered
- **THEN** the header reads Library, Search, Plugins, Settings ... Log, About in that order

#### Scenario: Missing page does not create a dead link
- **WHEN** a page in the canonical list is not registered
- **THEN** it is omitted from the nav and no other page order changes
