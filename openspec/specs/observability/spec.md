# observability Specification

## Purpose
Collects the V2-only observability surfaces that were added after the original capability map: the About and
Log pages, the WebSocket log stream, and debug-only runtime instrumentation.

## Requirements

### Requirement: About page

`GET /about` SHALL render the project README with a locally vendored Mermaid.js bundle, without a CDN. The
README SHALL be embedded into the binary so the page works when the source tree is absent. Mermaid SHALL run
only on `pre code.language-mermaid` blocks and SHALL NOT auto-scan the whole document.

#### Scenario: About renders offline
- **WHEN** a browser requests `/about` from the daemon with no network access
- **THEN** the page returns 200, the README is visible, and `/assets/vendor/mermaid.min.js` is served locally

### Requirement: Log page and WebSocket stream

`GET /log` SHALL render a plain-JavaScript WebSocket client that connects to `GET /ws/log`, displays a bounded
list of recent log lines, and shows a connected/disconnected badge with automatic reconnect. `GET /ws/log`
SHALL replay a bounded snapshot to a new subscriber and then stream new lines through a drop-oldest fan-out
that never blocks the publisher. Frames SHALL be plain text.

#### Scenario: Snapshot replay
- **WHEN** a client opens `/ws/log` after the daemon has already logged lines
- **THEN** the most recent buffered lines are replayed before live lines

#### Scenario: Slow subscriber does not block logging
- **WHEN** a subscriber stops reading while new log lines arrive
- **THEN** the oldest buffered lines are dropped for that subscriber and the rest of the daemon keeps logging

### Requirement: Debug-only instrumentation

When `log.level` is `debug`, the host call path, Lua runtime call, tool dispatcher and search/detail handlers
SHALL emit verbose trace logs with request/result previews. At `info` and above those traces SHALL be absent.
The instrumentation SHALL be permanent and level-gated, not temporary debug code.

#### Scenario: Debug reveals the call chain
- **WHEN** `log.level=debug` and a plugin search runs
- **THEN** logs show plugin search, plugin call start, runtime request decode, result and call finish

#### Scenario: Info stays quiet
- **WHEN** `log.level=info` and the same search runs
- **THEN** none of the verbose trace lines are emitted
