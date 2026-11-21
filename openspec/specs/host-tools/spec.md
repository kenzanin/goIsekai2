# host-tools Specification

## Purpose

Defines the complete plugin-facing tool surface: one `host.*` namespace with identical names,
argument order and return shapes in every runtime, plus the two-value and convenience helpers that
plugins rely on. Tool tables (TLS profiles, VRF tables) stay plugin-side; the host only provides
mechanisms. A behaviour difference between runtimes is a defect, not a runtime trait.

### Parity matrix

| Tool | Lua | JS | WASM |
|---|---|---|---|
| `now` | yes | yes | yes |
| `log.info` / `log.error` | yes | yes | yes (`log_info` with a level byte) |
| `http.fetch` | yes | yes | yes |
| `http.fetch_body` (2-value) | yes | yes | text-in/text-out only |
| `http.get_json` | yes | yes | JSON text or not-found |
| `storage.get` / `storage.set` | yes | yes | yes |
| `fs.read` | yes | yes | yes |
| `config.get` | yes | yes | out of scope |
| `url_encode` / `url_decode` | yes | yes | yes |
| `crypto.vrf_sign` | yes | yes | out of scope |
| `re2.*` core (`match`, `find`) | yes | yes | yes |
| `re2` promoted (`capture`, `find_all`, `gmatch`, `replace`, `quote`, `find_index`) | yes | yes | no |
| `html.select` / `html.attr` | yes | yes | yes |
| `html.xpath` | yes | yes | no |
| `text.*` promoted (`json_blob`, `date_epoch`, `clean`) | yes | yes | no |
| `url_abs` | yes | yes | no |
| `require` (plugin dir + shared lib) | yes | n/a | n/a |

A gap marked absent SHALL stay inert: the runtime links and runs with its existing import set and
gains no new required import.

## Requirements

### Requirement: One implementation behind every binding

Every tool SHALL have exactly one host implementation; runtime bindings SHALL be thin marshals over
it. A tool exposed in one runtime and missing in another, where the matrix says it exists, SHALL be
treated as a defect.

#### Scenario: Same tool, same result
- **WHEN** a Lua plugin and a JS plugin call `host.now`, `host.http.fetch` and `host.storage.set`/`get`
- **THEN** both resolve the same tool set and produce identical JSON shapes

#### Scenario: WASM guest stays linkable
- **WHEN** an existing WASM plugin is loaded after a promoted helper lands
- **THEN** it links and runs with its previous imports and no new required import

### Requirement: http.fetch

`http.fetch(url, options)` SHALL perform the request through the host HTTP client with browser TLS
fingerprinting and return `{status, body, headers}`. `options` MAY contain `timeout_ms`, `profile`,
`method` (default `GET`), `body` and `headers`; the options headers SHALL be applied after the
built-in headers so a plugin can override `Accept`/`User-Agent`. Absent `timeout_ms`/`profile` SHALL
fall back to configuration defaults.

#### Scenario: Fetch with defaults
- **WHEN** a plugin calls `http.fetch("https://api.example.com/manga")` with no options
- **THEN** the request uses the configured timeout and TLS profile and returns `status`, `body` and `headers`

#### Scenario: POST with form body and custom header
- **WHEN** a plugin sends `method = POST`, a urlencoded `body`, and `Content-Type` plus `Authorization` headers
- **THEN** the request carries exactly that body and those headers and the response arrives as a normal fetch result

### Requirement: Per-call profile and cookie jar

A `profile` in the request options SHALL override the configured default for that call only; an
unknown profile SHALL fall back to the default with a warning, never a failed request. Changing the
configured default SHALL apply to later calls without restart. An opt-in per-plugin cookie jar SHALL
persist at `data/plugin-state/<id>.cookies.txt`, survive restarts, and never be shared between
plugins; without the option the request SHALL behave as if no jar existed.

#### Scenario: Per-call override
- **WHEN** a plugin passes `profile: "edge101"` while the default is `chrome116`
- **THEN** that request impersonates edge101 and later requests without the option use chrome116

#### Scenario: Session survives restart
- **WHEN** a plugin with an enabled jar receives session cookies and the host restarts
- **THEN** its next request reuses the stored cookies

#### Scenario: Isolation
- **WHEN** two plugins with jars fetch from the same site
- **THEN** neither observes the other's cookies

### Requirement: Network failure is detectable

A transport failure or timeout SHALL be returned as a failure the plugin can detect, and the host
SHALL log the URL and the error tagged with the plugin id. A non-2xx status SHALL NOT be a panic or
a host error; it is data.

#### Scenario: Unreachable host
- **WHEN** the target host is unreachable
- **THEN** the call returns a failure carrying the transport message and the plugin continues

#### Scenario: Upstream 410
- **WHEN** the upstream answers 410
- **THEN** `status` is 410 and the body is returned as-is

### Requirement: Two-value body fetch

`http.fetch_body(url, options)` SHALL return body and error, never both nil: success yields the body
with a nil error, a non-200 yields a nil body with `upstream status <code>`, and a transport failure
yields a nil body with the transport message.

#### Scenario: Upstream error is diagnosable
- **WHEN** `http.fetch_body(url)` hits a 410
- **THEN** the body is nil and the error is `upstream status 410`

#### Scenario: Transport failure
- **WHEN** the host cannot connect to the upstream host
- **THEN** the body is nil and the error is a non-empty transport message logged with the plugin id

### Requirement: JSON helpers

`http.get_json(url, options)` SHALL accept the same options as `http.fetch` and return the decoded
value plus an error; a table with none of the known option keys SHALL be treated as a headers map.
`http.post_json(url, body, headers)` SHALL send a JSON body and decode the reply.

#### Scenario: Legacy headers map still works
- **WHEN** a plugin calls `http.get_json(url, {Accept = "application/json"})`
- **THEN** the table is used as request headers and no option error occurs

### Requirement: Logging and time

`log.info(message)` and `log.error(message)` SHALL enter the host log pipeline tagged with the plugin
id and SHALL reach the log stream. `now()` SHALL return current unix seconds as an integer,
consistent with every host-generated timestamp.

#### Scenario: Plugin logs progress
- **WHEN** a plugin calls `log.info("fetched 20 titles")`
- **THEN** the message appears in the host log and the log stream associated with that plugin

### Requirement: Plugin config access

`config.get(key)` SHALL read from a `config.json` in the plugin's own folder, written by the user and
never read by the loader. A missing file or key SHALL return nil. The tool SHALL NOT grant file-write
or general file-read access.

#### Scenario: Credentials present
- **WHEN** the plugin's `config.json` holds `{"client_id": "c"}` and the plugin calls `config.get("client_id")`
- **THEN** it receives `"c"` and an absent key returns nil

### Requirement: Plugin state storage

`storage.get(key)` / `storage.set(key, value)` SHALL persist string values scoped to the calling
plugin id through the store, surviving reload and restart. A plugin SHALL NOT be able to read another
plugin's state, and SHALL NOT reach the database any other way.

#### Scenario: State survives restart
- **WHEN** a plugin sets `refresh_token` and the host restarts
- **THEN** the value is returned by `storage.get`

#### Scenario: Cross-plugin read is impossible
- **WHEN** a plugin calls `storage.get` with a key written by another plugin
- **THEN** it receives nil

### Requirement: Confined read-only file access

`fs.read(relpath)` SHALL return the contents of a regular file resolved inside the calling plugin's
own folder. Absolute paths, `..` traversal and missing files SHALL yield a not-found result, never an
error response and never a read outside the folder — identically in every runtime that exposes it.

#### Scenario: Plugin reads its own file
- **WHEN** a plugin calls `fs.read("config.json")` and the file exists in its folder
- **THEN** the contents are returned

#### Scenario: Escape and missing file
- **WHEN** a plugin calls `fs.read("../other/secret")` or `fs.read("nope.json")`
- **THEN** both report not-found and nothing outside the folder is read

### Requirement: Module resolution

`require(name)` SHALL resolve in the plugin folder first, then in the shared library root
`<plugins dir>/../plugins_lib`, both read-only, with absolute paths and `.`/`..` segments rejected in
both roots and a per-plugin module cache. A plugin-local module SHALL shadow a shared module of the
same name.

#### Scenario: Shared module resolves
- **WHEN** `mangabay/main.lua` requires `dle` and only `plugins_lib/dle.lua` exists
- **THEN** the module loads from the shared root

#### Scenario: Local shadows shared
- **WHEN** both the plugin folder and `plugins_lib` hold `dle.lua`
- **THEN** the plugin's own copy is loaded

#### Scenario: Traversal refused
- **WHEN** a plugin requires `../etc/passwd`
- **THEN** the load fails with `require: bad module name` and no file outside the roots is read

### Requirement: Regex match and find

`re2.match(pattern, text)` SHALL return capture groups as an array with element 1 = whole match and
group *n* = element *n+1*, and `re2.find` SHALL return the first match in the same form; no match
SHALL yield an empty array. Patterns SHALL be evaluated with linear-time RE2, and an invalid pattern
SHALL surface as a runtime error naming the compile failure.

#### Scenario: Match and find on fixture text
- **WHEN** a plugin calls `match("ch(\\d+)", "ch1050")` and `find("(\\w+)@(\\w+)", "user@host other")`
- **THEN** match returns `["ch1050","1050"]` and find returns `["user@host","user","host"]`

#### Scenario: Invalid pattern is loud
- **WHEN** the pattern fails to compile
- **THEN** the calling script receives a runtime error naming the compile failure

### Requirement: Regex single capture

`re2.capture(pattern, text)` SHALL return group 1 when the pattern declares groups, the whole match
when it declares none, the empty string for an empty group, and nil when the pattern does not match.

#### Scenario: Group present
- **WHEN** a plugin calls `capture('<a href="([^"]+)"', '<a href="/x">')`
- **THEN** the result is `/x`

#### Scenario: No match
- **WHEN** the pattern does not occur in the text
- **THEN** the result is nil and no error is raised

### Requirement: Regex bulk helpers

`re2.find_all(pattern, text)` SHALL return every match with its captures, `re2.gmatch(pattern, text)`
SHALL iterate matches, `re2.replace(pattern, repl, text)` SHALL accept goIsekai's `$N`/`${N}`
backreference syntax by translating it to RE2's `\N` inside the one shared core, `re2.quote(str)`
SHALL escape metacharacters, and `re2.find_index(pattern, text)` SHALL return the byte offset of the
first match or a negative value.

#### Scenario: find_all returns every match
- **WHEN** a plugin calls `find_all("\\d+", "abc 123456")`
- **THEN** the result holds `123` and `456` as separate matches

#### Scenario: Dollar backreference
- **WHEN** a plugin calls `replace("(\\w+)@\\w+", "$1", "bob@example")`
- **THEN** the result is `bob`

### Requirement: HTML tools

`html.select(html, css)` SHALL return the trimmed text of every element matching the selector, in
document order; `html.attr(html, css, attr)` SHALL return the attribute values of matching elements
that have it, skipping the rest. `html.xpath(expr, html)` SHALL return matching text/attribute values
in document order and nil for a scalar result or invalid expression. No match or malformed HTML SHALL
yield an empty array or nil, never an exception, and parsing SHALL be lenient.

#### Scenario: Select text from fixture HTML
- **WHEN** a plugin calls `html.select(fixture, "div.title > a")`
- **THEN** the array holds the trimmed text of each matched anchor in document order

#### Scenario: Attribute extraction skips missing
- **WHEN** only some matched elements carry the requested attribute
- **THEN** only those contribute values, in document order

#### Scenario: Legacy list helpers map onto select/attr
- **WHEN** a ported goIsekai plugin calls `find_list_text(html, css)` or `find_list_attr(html, css, attr)`
- **THEN** it is served by `select(html, css)` / `attr(html, css, attr)`; no separate list tool exists

### Requirement: Text normalization tools

`text.normalize_status(s)` SHALL map the status vocabulary and common synonyms to canonical
`Ongoing`, `Completed`, `Hiatus`, `Dropped`, `Upcoming`. `text.chapter_num(s)` SHALL return the
chapter number as a float (`"Ch. 12.5"` → `12.5`). `text.trim` SHALL be UTF-8 aware,
`text.unescape` SHALL decode entities, and `text.strip_html` SHALL remove tags, then trim, then
unescape.

#### Scenario: Status synonym
- **WHEN** a plugin calls `normalize_status("releasing")`
- **THEN** the result is `Ongoing`

#### Scenario: Chapter number
- **WHEN** a plugin calls `chapter_num("Ch. 12.5")`
- **THEN** the result is `12.5`

#### Scenario: Strip a fragment
- **WHEN** a plugin calls `strip_html("<p>Hello <b>world</b></p>")`
- **THEN** the result is `Hello world`

### Requirement: Date parsing

`text.date_to_iso(s)` SHALL parse absolute and relative human dates into `YYYY-MM-DD`.
`text.date_epoch(s)` SHALL use the same parser and return unix seconds, treating zone-less input as
UTC. The accepted layouts SHALL be `YYYY-MM-DD`, ISO date-time with optional `Z`, offset or
fractional seconds, `Mon DD, YYYY`, `DD Mon YYYY`, their `HH:MM` variants, and relative forms such as
`3 days ago`. An unparsable value SHALL return nil and SHALL NOT fall back to the current time.

#### Scenario: ISO date
- **WHEN** a plugin calls `date_epoch("2026-03-04")`
- **THEN** the result is the unix timestamp of 2026-03-04T00:00:00Z

#### Scenario: Human date
- **WHEN** a plugin calls `date_epoch("Mar 4, 2026")`
- **THEN** the result equals the ISO form above

#### Scenario: Relative date
- **WHEN** a plugin calls `date_to_iso("3 days ago")`
- **THEN** the result is the ISO date three days before today

#### Scenario: Garbage input
- **WHEN** the value is empty, `unknown`, or site placeholder text
- **THEN** the result is nil and the plugin decides the fallback

### Requirement: Bounded JSON blob extraction

`text.json_blob(text, marker)` SHALL return the first balanced `{...}` or `[...]` beginning at or
after `marker`, scanning escape-aware (`"` toggles string state, `\` escapes the next byte) and
tracking depth. It SHALL return nil when the marker is absent, the body is unbalanced, or the input
is truncated, and SHALL return raw JSON text — decoding stays with `json.decode`.

#### Scenario: State blob after marker
- **WHEN** the page contains `window.__DATA__ = {"a":{"b":"}}"};`
- **THEN** the result is exactly `{"a":{"b":"}}"}` — braces inside strings do not close it

#### Scenario: Truncated payload
- **WHEN** the document ends before the JSON body is balanced
- **THEN** the result is nil, not a partial blob

### Requirement: Display text cleanup

`text.clean(s)` SHALL turn scraped HTML text into display-ready text: decode entities to a fixed
point of at most 4 passes, then trim. Multiply-encoded forms and hex/decimal numeric references
SHALL resolve; an invalid codepoint (0, surrogate, above U+10FFFF) SHALL leave the literal reference
in place. Input without `&` SHALL skip decoding entirely.

#### Scenario: Single and nested encoding
- **WHEN** a plugin calls `clean("So What&#39;s Wrong")` and `clean("A&amp;#39;s")`
- **THEN** both results are `So What's Wrong` and `A's` respectively

#### Scenario: Real ampersand preserved
- **WHEN** a plugin calls `clean("https://x/api?a=1&b=2")`
- **THEN** the string is returned unchanged

#### Scenario: Pathological nesting terminates
- **WHEN** the input has more than four layers of `&amp;`
- **THEN** the call finishes within four passes without throwing

### Requirement: URL helpers

`url_abs(base, ref)` SHALL return `ref` unchanged when it is already `http://`/`https://`, prepend
the base scheme for `//` references, prepend the base origin for `/` references, and otherwise
resolve relative to the base path truncated at its last `/`; dot-segment normalization and
query/fragment merging are out of scope. `url_encode(s)` and `url_decode(s)` SHALL round-trip the
reserved characters they claim to handle.

#### Scenario: Protocol-relative and root-relative
- **WHEN** a plugin resolves `//cdn.site.tld/a.jpg` and `/img/a.jpg` against `https://site.tld/manga/x`
- **THEN** the results are `https://cdn.site.tld/a.jpg` and `https://site.tld/img/a.jpg`

#### Scenario: Encode round-trip
- **WHEN** a plugin encodes and decodes `a b&c=d%?`
- **THEN** the decoded value equals the original

### Requirement: Codecs

`codecs.base64_encode`, `base64_decode`, `hex_encode` and `hex_decode` SHALL be available in every
runtime that carries the tool, with a decode returning nil for invalid input (bad alphabet, bad
quantum, odd-length hex).

#### Scenario: Known vector
- **WHEN** a plugin calls `base64_encode("hello")`
- **THEN** the result is `aGVsbG8=`

#### Scenario: Invalid base64
- **WHEN** the input has invalid characters or an impossible quantum
- **THEN** the decode returns nil

### Requirement: Crypto primitives

`crypto.sha256_hex`, `crypto.md5_hex` and `crypto.hmac_sha256_hex` SHALL hash their input and return
lowercase hex. `crypto.aes_gcm_decrypt(iv, ct, tag, key)` SHALL verify the GCM tag before emitting
any plaintext and SHALL return nil on invalid hex, wrong key length or authentication failure.

#### Scenario: SHA-256 known answer
- **WHEN** a plugin calls `sha256_hex("abc")`
- **THEN** the result is `ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad`

#### Scenario: Tampered tag
- **WHEN** the GCM tag is corrupted
- **THEN** the call returns nil and no plaintext is emitted

### Requirement: Request signing

`crypto.vrf_sign(api_path, params, stages)` SHALL trim a leading `/api` from the path, serialize
params sorted by key as raw `k=v` pairs, and apply the plugin-declared stages in order as
`out[i] = tbl[(d ^ key[i % len] ^ prev) & 0xFF]` with `prev` seeded to `iv`. It SHALL return
base64url without padding, omit the `?` separator when params are empty, and fail with
`vrf_sign: stage <n>: ...` on invalid stage data without crashing the host.

#### Scenario: Deterministic sign
- **WHEN** the plugin signs the same path, params and stage set twice
- **THEN** both calls return the same base64url string as the captured test vector

#### Scenario: Invalid stage
- **WHEN** a stage `key` or `tbl` is not valid standard base64
- **THEN** the plugin receives a detectable failure naming the stage and the host logs it with the plugin id

#### Scenario: Tables stay plugin-side
- **WHEN** the plugin rebuilds its table set after a site rotation
- **THEN** only the plugin file changes and the host binary is not recompiled

### Requirement: Manifest params

`plugin.param(key)` SHALL return the value of `key` from the plugin's own `main.json` `params`
object, parsed at load time, and nil when the key is absent.

#### Scenario: Param lookup
- **WHEN** a plugin whose manifest has `params: {"site_url": "https://x"}` calls `plugin.param("site_url")`
- **THEN** the result is `https://x`

### Requirement: Bounded cost and no ambient authority

Every tool SHALL operate within the call budget and SHALL have bounded work: linear-time regex, no
unbounded decode loops, capped blob scanning. No tool SHALL give a plugin the database handle, the
event bus, the host filesystem outside its folder, or process control.

#### Scenario: Budget covers tool work
- **WHEN** a plugin spends its whole budget inside a sequence of tool calls
- **THEN** the call ends with a budget error rather than continuing past the bound

#### Scenario: No ambient handles
- **WHEN** a plugin inspects its environment for a database or bus handle
- **THEN** none exists and only the documented tools are reachable

### Requirement: Host tool injection

The daemon SHALL construct exactly one host tool set and inject it into every plugin runtime. No
runtime SHALL receive the store handle, the event bus, an ambient filesystem root or an ambient HTTP
client that bypasses the tool set. A tool exposed in Lua and JS SHALL resolve to the same Go
implementation with the same JSON shape. WASM SHALL keep its existing import set; promoted helpers
that the matrix marks absent SHALL remain inert rather than becoming required imports.

#### Scenario: Same tool set in every runtime
- **WHEN** Lua and JS plugins call `now`, `log.info`, `storage.set` and `http.fetch`
- **THEN** both resolve the same Go implementation and produce identical return shapes

#### Scenario: No ambient authority
- **WHEN** a plugin tries to reach the SQLite file, the event bus or a path outside its own folder
- **THEN** no such capability is present in its tool environment and the attempt fails inside the plugin

#### Scenario: Existing WASM plugin still links
- **WHEN** a promoted helper is added to Lua and JS
- **THEN** an existing WASM module links and runs with its previous imports unchanged
