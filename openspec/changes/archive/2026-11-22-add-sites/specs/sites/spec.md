# Spec Delta

## ADDED Requirements

### Requirement: Site corpus wiring

The daemon SHALL validate that a plugin folder name equals its manifest id and SHALL serve a declared logo at `/plugin/<id>/logo` with the file's content type. The shipped corpus SHALL include the mangadex reference plugin and SHALL record every baseline plugin as working or as an omission with a reason.

#### Scenario: Id mismatch rejected
- **WHEN** a manifest id differs from its folder name
- **THEN** the loader skips the folder and logs the reason

#### Scenario: Logo served
- **WHEN** a browser requests `/plugin/<id>/logo` for a declared logo
- **THEN** the response carries the image bytes and correct content type
