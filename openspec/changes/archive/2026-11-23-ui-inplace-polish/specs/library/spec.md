# Spec Delta

## ADDED Requirements

### Requirement: Long pager window

The saved-library pager SHALL show a compact long format: first four pages, an ellipsis for a skipped range, and last five pages. The current page and its neighbours SHALL appear in the middle when it is outside both edge runs.

#### Scenario: Wide page range
- **WHEN** 14 pages are available and page 1 is current
- **THEN** the pager reads `1 2 3 4 … 10 11 12 13 14` with prev/next

#### Scenario: Middle page range
- **WHEN** page 7 of 14 is current
- **THEN** the pager shows the edge runs, an ellipsis around `6 7 8`, and the last five pages
