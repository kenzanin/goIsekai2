# Spec Delta

## ADDED Requirements

### Requirement: Browser daemon wiring

The daemon SHALL register `browser.call` only after the browser backend is constructed, and SHALL start and stop any owned browser child around listen/shutdown. A plugin without browser capability SHALL receive the unknown-tool failure without touching the backend. A plugin manifest with `image_referer: "none"` SHALL make image fetches omit the `Referer` header.

#### Scenario: Undeclared plugin is denied
- **WHEN** a plugin without browser capability calls `browser.call`
- **THEN** the host returns unknown-tool and does not touch a session

#### Scenario: Referer sentinel
- **WHEN** a plugin declares `image_referer: "none"` and its image cache misses
- **THEN** the upstream request carries no `Referer` header
