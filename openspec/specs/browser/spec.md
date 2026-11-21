# browser Specification

## Purpose

Lets a plugin drive a real browser through the host so it can read sites whose content only exists
after client-side JavaScript runs. The host owns process supervision, session lifetime, serialization
and failure reporting; the plugin only names a browser tool and passes arguments. Grounded in the
baseline `browser-host` code path and its archived change.

Divergence from the baseline (recorded in `decisions/browser-host-transport`): Go supervises the child
with `os/exec` + `context` instead of `reproc`, and the transport sits behind one `Browser` interface so
an MCP client is the default backend and a CDP adapter can be swapped later without touching plugins.

## Requirements

### Requirement: Browser tool passthrough

The host SHALL expose exactly one plugin-accessible host tool, `browser.call(tool, args)`, that
forwards the named browser tool and a `tools/call`-shaped argument object to the configured browser
backend and returns the tool's result as text inside the standard host-tool envelope. Supported tool
names SHALL at least cover `goto`, `evaluate`, `waitForState`, `waitForSelector`, `waitForScript`,
`getCookies`, `session_new`, `session_list`, `session_close`, and `login`.

#### Scenario: Successful evaluate
- **WHEN** a plugin calls `browser.call("evaluate", {script: "return document.title"})` after an earlier `goto`
- **THEN** the call returns an ok envelope the plugin can consume as text, and the session still holds the page from that `goto`

#### Scenario: Unknown tool fails explicitly
- **WHEN** a plugin names a tool the backend does not expose
- **THEN** the host returns a failure the plugin can branch on and logs the tool name with the backend detail

### Requirement: Serialized multi-step sequences

The host SHALL serialize complete call sequences that share browser state (for example `goto` then
`evaluate`) so concurrent plugin invocations never interleave actions across the shared session. The
serialization SHALL be per backend session, not global across the whole daemon.

#### Scenario: Concurrent harvests do not interleave
- **WHEN** two plugin calls, each doing `goto` followed by `evaluate`, run concurrently
- **THEN** the second sequence starts only after the first finishes, and each sees its own page state

#### Scenario: Unrelated work is not blocked
- **WHEN** a slow browser sequence runs
- **THEN** ordinary plugin calls, HTTP handlers and the update batch continue

### Requirement: Session recovery on expiry

The host SHALL keep the backend session id in memory per process and SHALL recover exactly once from
a session-expired response by re-initializing the session and retrying the original tool call. A
second expiry in the same call SHALL surface as a failure.

#### Scenario: Session expires mid-call
- **WHEN** the backend reports the cached session no longer exists
- **THEN** the host creates a fresh session, retries that call once, and returns its result or the failure

#### Scenario: Repeated expiry is not a loop
- **WHEN** the retried call also reports expiry
- **THEN** the host returns the failure and does not retry again

### Requirement: Unreachable backend fails fast

The host SHALL report an explicit, loggable failure rather than hanging when the configured browser
URL refuses connections or times out, and plugin callers SHALL receive the ordinary host-tool failure
signal within the configured timeout.

#### Scenario: Daemon not running
- **WHEN** `browser.call` is invoked while nothing listens on the configured URL
- **THEN** the call fails within the timeout, the log names the URL, and the plugin receives a failure

### Requirement: In-process supervision of the browser daemon

On boot the daemon SHALL reuse an already-healthy browser instance answering on the configured URL
instead of spawning a second one, and otherwise SHALL spawn `lightpanda mcp --host 127.0.0.1 --port <port>`
as its own child, wait for health before binding the HTTP port, and record ownership. An instance the
daemon did not spawn SHALL be left running at exit.

#### Scenario: Startup with nothing running
- **WHEN** the daemon boots and the browser URL does not answer
- **THEN** it spawns the child, waits for health, then binds the HTTP port

#### Scenario: Startup reuses a healthy instance
- **WHEN** a healthy backend already answers on the configured URL
- **THEN** no child is spawned and the existing instance survives daemon exit

#### Scenario: Spawn failure is visible
- **WHEN** the browser binary is missing or exits immediately at boot
- **THEN** the daemon logs the reason and continues serving, with `browser.call` failing fast

### Requirement: Single respawn, then fail fast

On death of the child it spawned, the host SHALL respawn it once, re-initialize the session and retry
the interrupted call once; if that retry also fails the browser backend SHALL be marked dead for the
rest of the process and later calls SHALL fail immediately. On shutdown the host SHALL stop only the
child it owns, terminating and then killing after a grace period.

#### Scenario: Child crash mid-operation
- **WHEN** the owned child dies during a `browser.call`
- **THEN** the host respawns once, re-initializes the session, retries once, and returns the result or marks the backend dead

#### Scenario: Restart storm prevented
- **WHEN** repeated deaths keep failing after the single respawn
- **THEN** the backend stays dead until the daemon restarts and no further spawns occur

#### Scenario: Shutdown
- **WHEN** the daemon shuts down after spawning its own child
- **THEN** that child is terminated (killed after the grace period) and a foreign instance is untouched

### Requirement: Loopback bypass of the host proxy

The host SHALL NOT route browser-backend requests through the host-wide HTTP proxy when the configured
URL targets loopback, regardless of proxy config.

#### Scenario: Proxy configured, loopback direct
- **WHEN** an upstream proxy is configured and a plugin calls with the default `http://127.0.0.1:<port>`
- **THEN** backend requests go directly to loopback

### Requirement: Image referer sentinel

A plugin manifest SHALL support `image_referer: "none"`, which makes upstream image fetches for that
plugin send no `Referer` header. Any other value SHALL keep today's behavior: an explicit manifest
referer, otherwise the plugin origin derived from its website.

#### Scenario: Sentinel suppresses the header
- **WHEN** a plugin declaring `image_referer: "none"` serves reader images and the cache misses
- **THEN** the upstream request carries no Referer and the CDN response is accepted like a direct fetch

#### Scenario: Other values unchanged
- **WHEN** the manifest omits the field or sets a URL
- **THEN** the origin or the explicit value is sent as before

### Requirement: No browser access for plugins without the capability

A plugin whose manifest does not declare browser use SHALL NOT reach `browser.call`; the host SHALL
answer with the unknown-tool failure so a plugin cannot silently drive a shared session.

#### Scenario: Undeclared plugin
- **WHEN** a plugin without browser capability calls `browser.call`
- **THEN** the host returns a failure and does not spawn, reuse or touch a session
