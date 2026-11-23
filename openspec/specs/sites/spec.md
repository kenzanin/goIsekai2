# sites Specification

## Purpose

Defines what a shipped plugin is: the manifest, the folder contract, the ABI it must implement, the
mangadex reference plugin that fixes the canonical vocabulary, and the parity rule for the corpus
ported from the baseline. This capability is about site content and plugin packaging; the runtime that
executes plugins is `plugin-host`, and `host-tools` owns the tools they call.

Baseline corpus at porting time: 33 Lua plugin folders, 2 JS plugin folders and 6 wasm plugin folders
under the baseline `plugins_dev/`, plus roughly 40 Lua and 40 JS plugins proven in V1. Those are the
porting corpus, not a vendored dependency.

## Requirements

### Requirement: Plugin folder contract

Every plugin SHALL be a directory containing a manifest (`main.json`) that declares `api`, `id`,
`name`, `website`, optional `logo`, optional `image_referer`, optional `capabilities`, plus its code
entry (`main.lua`, `main.js`, or `main.wasm`) and any data files (`genres.lua`, `config.json`). The
folder name SHALL equal the manifest `id`.

#### Scenario: Missing manifest
- **WHEN** a folder has code but no readable `main.json`
- **THEN** the loader skips it, logs the reason, and the plugin never appears in the UI

#### Scenario: Id mismatch
- **WHEN** the manifest `id` differs from the folder name
- **THEN** the plugin is rejected with a logged reason instead of loading twice under two names

### Requirement: ABI method vocabulary

A plugin SHALL implement the baseline method names unchanged: `search`, `detail`, `chapters`, `pages`,
and opportunistically `genres`. It SHALL NOT invent method names, and any method it does not implement
SHALL answer `{"ok": false, "error": "unknown method <name>"}`.

#### Scenario: Required methods present
- **WHEN** a plugin is loaded
- **THEN** `search`, `detail`, `chapters` and `pages` are callable by name

#### Scenario: Optional genres
- **WHEN** a plugin does not implement `genres`
- **THEN** the UI's genre filter falls back to a free-text tag and reports no genres rather than failing

### Requirement: Canonical item shape

`search` SHALL return `{"ok": true, "data": {"items": [...], "total": <int>}}`, where each item carries
`id`, `title`, `cover_url`, optional `author`/`status`/`tags`, and `total` is an integer not smaller
than the item count. `detail` SHALL return the same canonical fields plus synopsis, author, artist, alt
titles and tags. `chapters` SHALL return items with `number`, `title`, and `release_at` as integer unix
seconds — never a date string. All timestamps SHALL be unix seconds at every boundary.

#### Scenario: Unix timestamps everywhere
- **WHEN** any plugin returns a chapter with `release_at`
- **THEN** the value is an integer unix second count, in JSON without quotes

#### Scenario: Empty search is success
- **WHEN** a search matches nothing
- **THEN** the plugin returns an ok envelope with an empty item list and `total: 0`

#### Scenario: `pages` returns absolute urls in reading order
- **WHEN** `pages` is called for a known chapter
- **THEN** the data array holds absolute https urls in reading order with one entry per page

### Requirement: Errors are data, never crashes

A plugin SHALL translate every upstream failure (HTTP error, parse failure, missing field, timeout,
auth rejection) into `{"ok": false, "error": "<message>"}`. A plugin SHALL NOT raise raw errors across
the host boundary, and SHALL NOT retry in a loop.

#### Scenario: Upstream 5xx
- **WHEN** the site answers 500
- **THEN** the plugin returns an error envelope with a meaningful message and the host logs it against the plugin

#### Scenario: Shape drift
- **WHEN** the site changes markup and a required field is missing
- **THEN** the plugin returns an error envelope instead of an item with empty fields

### Requirement: Network access only through host tools

A plugin SHALL perform all network access through `http.fetch`, all persistence through `storage.*`, and
all logging through `log.*`. Direct sockets, filesystem access and process spawning SHALL be
unavailable, and a plugin that attempts them SHALL fail with an explicit error.

#### Scenario: Direct socket attempt
- **WHEN** a Lua plugin requires `socket`
- **THEN** the call fails with a clear "not available" error and the plugin logs it

### Requirement: Mangadex is the reference plugin

The mangadex plugin SHALL stay the canonical-format reference: `search` ordered by `followedCount desc`
with an exact-title match first, `tags` accepted as comma-joined English names with AND semantics,
`detail`/`chapters` sourced from its API, `pages` from the at-home endpoint, and `genres` from its tag
list. MangaDex vocabulary SHALL NOT be translated.

#### Scenario: Ranking parity
- **WHEN** searching "one piece" on page 1
- **THEN** results are ordered by follower count with the exact title first

#### Scenario: Tags use AND
- **WHEN** `tags: "Isekai,Fantasy"` is passed
- **THEN** only titles carrying both tags return and `total` reflects the filtered query

#### Scenario: Page windows are disjoint
- **WHEN** the same query is asked with page 2
- **THEN** the window does not repeat page 1 and clamps at the last page

### Requirement: Mangadex optional authentication

With `client_id`, `client_secret`, `username` and `password` in the plugin folder's `config.json`, the
plugin SHALL authenticate with the password flow, store `access_token`, `refresh_token` and
`expires_at` (unix seconds) in plugin storage, SHALL refresh on expiry or on a 401 with a single
retry, and SHALL attach `Authorization: Bearer <token>` only to API-host requests — never to image
hosts. Without `config.json` it SHALL run unauthenticated exactly as before.

#### Scenario: No credentials
- **WHEN** `config.json` is absent
- **THEN** no auth attempt occurs and every method behaves unauthenticated

#### Scenario: Refresh on expiry
- **WHEN** the stored access token is past `expires_at`
- **THEN** the next API call refreshes once and succeeds without surfacing an error

#### Scenario: 401 mid-session
- **WHEN** a call returns 401 while the token looks unexpired
- **THEN** the plugin refreshes once, retries once, then returns the error envelope if it still fails

#### Scenario: Bad credentials
- **WHEN** the token endpoint rejects the login
- **THEN** the plugin returns an auth failure envelope and does not loop

#### Scenario: Token never leaks to image hosts
- **WHEN** an image host request is built
- **THEN** it carries no Authorization header

### Requirement: Plugin identity assets

The plugin folder SHALL carry a logo when the manifest declares one, and the host SHALL serve it at
`/plugin/<id>/logo` with the correct content type; wherever the plugin renders in the UI the logo SHALL
appear next to its name.

#### Scenario: Logo served
- **WHEN** a browser requests `/plugin/mangadex/logo`
- **THEN** the image bytes come back with the matching content type

#### Scenario: Missing logo
- **WHEN** a plugin declares no logo
- **THEN** the UI renders the name without a broken image

### Requirement: Corpus parity, no regression

Porting SHALL preserve behaviour: every plugin shipped in the baseline `plugins_dev/` set SHALL either
work in V2 with the same method results or be recorded as a documented, decision-backed omission.
Shipping fewer working plugins than the baseline without a recorded decision SHALL be treated as a
defect.

#### Scenario: Parity check
- **WHEN** the ported plugin set is compared against the baseline inventory
- **THEN** each baseline plugin is either working or listed as an omission with its reason

#### Scenario: Fixture conformance
- **WHEN** a plugin is ported
- **THEN** a conformance fixture replays a recorded upstream response through the plugin and asserts the envelope shape

### Requirement: Runtime coverage

The shipped set SHALL cover all three runtimes — Lua, JS and wasm — and the loader SHALL report the
runtime it selected per plugin so a missing engine is visible rather than silent.

#### Scenario: Runtime reported
- **WHEN** plugins are listed
- **THEN** each entry names its runtime (lua, js, wasm)

#### Scenario: Engine gaps surface
- **WHEN** a LuaJIT-only module is required under the Lua 5.1 engine
- **THEN** the plugin fails with a clear unsupported-feature error and the reason is logged

### Requirement: Site corpus wiring

The daemon SHALL validate that a plugin folder name equals its manifest id and SHALL serve a declared logo at `/plugin/<id>/logo` with the file's content type. The shipped corpus SHALL include the mangadex reference plugin and SHALL record every baseline plugin as working or as an omission with a reason.

#### Scenario: Id mismatch rejected
- **WHEN** a manifest id differs from its folder name
- **THEN** the loader skips the folder and logs the reason

#### Scenario: Logo served
- **WHEN** a browser requests `/plugin/<id>/logo` for a declared logo
- **THEN** the response carries the image bytes and correct content type

### Requirement: Duplicate chapter releases select the newest

For titles with multiple scanlation releases of the same chapter number, `chapters` SHALL return the release with the greatest `release_at`. All duplicate releases SHALL remain individually addressable through their composite `id` at the `pages` boundary.

#### Scenario: Newest release wins
- **WHEN** `AJDs` has Chapter 1 releases `qRWvEm` and `2V2CqE` with increasing timestamps
- **THEN** `chapters` returns `AJDs:2V2CqE` for Chapter 1

#### Scenario: Old reader links canonicalize
- **WHEN** a reader link carries the composite id of an older duplicate
- **THEN** the reader resolves the chapter number to the listed release and shows its pages

#### Scenario: Older release stays stored
- **WHEN** duplicate releases exist for one chapter number
- **THEN** both ids remain in the chapter store and only the listed one is used by reader links
