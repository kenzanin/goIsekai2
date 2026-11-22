# Spec Delta

## ADDED Requirements

### Requirement: Reader route wiring

The daemon SHALL register the reader routes before it opens the HTTP port. Page-list prefetch SHALL be
a direct service call, never an event, and SHALL never block the rendering request. A failed plugin
refresh SHALL fall back to the last stored page list for that exact chapter id while logging the
failure with plugin and chapter id. Position writes SHALL keep absent fields and SHALL be persisted
before any prefetch starts.

#### Scenario: Prefetch is a direct call
- **WHEN** a chapter finishes rendering
- **THEN** the host calls the page-list resolver directly for next then previous chapter, without publishing an event

#### Scenario: Stale list survives a failed refresh
- **WHEN** the plugin page call fails and a stored page list exists for that chapter id
- **THEN** the stored list is returned and the failure is logged with plugin and chapter id

#### Scenario: Position fields are preserved
- **WHEN** a position POST carries only chapter and page
- **THEN** stored mode, direction and zoom stay unchanged before prefetch begins
