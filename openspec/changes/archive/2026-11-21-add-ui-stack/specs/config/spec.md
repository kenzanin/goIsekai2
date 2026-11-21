# Spec Delta

## MODIFIED Requirements

### Requirement: Settings page projects the registry

The settings page SHALL render one row per registry knob, grouped by section, showing the knob's
documentation, resolved value and winning source. Rows whose winning source is `env` or `flag`
SHALL be disabled and SHALL name that source; rows that cannot be applied without a restart SHALL
carry a restart badge. The form SHALL submit only fields the user changed and SHALL report the
save result in place without a full page reload. The page SHALL keep the active `Settings`
navigation link and SHALL NOT link to unimplemented pages. The rendered markup MAY change between
changes as long as these observable behaviors hold.

#### Scenario: Every knob appears exactly once
- **WHEN** the user opens `/settings`
- **THEN** every registry knob (`server.port`, `server.proxy`, `plugin.*`, `cache.*`, `paths.*`, `log.level`) appears once, grouped by section, with doc text and current value

#### Scenario: Env-owned knob is read-only
- **WHEN** a knob is supplied through its `GOISEKAI_*` variable or an explicit flag
- **THEN** its row is read-only and names the winning source

#### Scenario: Save round-trip
- **WHEN** the user edits an ini-backed knob and submits
- **THEN** `POST /config` reports the write, the page re-reads `GET /config`, and the row shows the post-save value

#### Scenario: Settings page shows provenance and restart hints
- **WHEN** `/settings` renders with `GOISEKAI_LOG_LEVEL=debug`
- **THEN** that row is disabled, names `env` as its source, and restart-only rows show a badge

#### Scenario: Dirty-field submit updates status in place
- **WHEN** the user changes one editable field and submits the form
- **THEN** only that field is sent to `POST /config` and the status region reports the saved count
  without navigating away from `/settings`

#### Scenario: Navigation stays exact
- **WHEN** `/settings` is rendered
- **THEN** the only navigation link is the active `Settings` link

## ADDED Requirements

### Requirement: Local-first UI assets

Every stylesheet and script required by the daemon's pages SHALL be served by the daemon from
embedded assets. The pages SHALL render and remain interactive with no outbound network access;
they SHALL NOT request CDN-hosted CSS, JavaScript, fonts or images. Generated assets SHALL be
committed so a fresh clone can build and serve the UI without a separate asset build step at
runtime.

#### Scenario: No external asset requests
- **WHEN** `/settings` is loaded in a browser with outbound network blocked
- **THEN** it renders with its stylesheet and interactive controls, and the browser makes no
  requests to third-party hosts

#### Scenario: Fresh clone serves the UI
- **WHEN** a clean checkout is built with `CGO_ENABLED=0 go build ./...` and the daemon starts
- **THEN** `/settings` renders without running npm, a CDN fetch, or a runtime asset generator
