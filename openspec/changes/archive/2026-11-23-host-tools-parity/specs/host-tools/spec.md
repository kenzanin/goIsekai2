# Spec Delta

## MODIFIED Requirements

### Requirement: One implementation behind every binding

Every host name exposed to a plugin SHALL have exactly one implementation in the shared dispatcher,
or SHALL be a registered alias forwarding to it. Adding a host name SHALL NOT change
plugin-visible behavior in any runtime: Lua, JS and WASM remain thin marshals, and an alias returns
results identical to its canonical name.

#### Scenario: Every name resolves once
- **WHEN** any host name is called in Lua, JS or WASM
- **THEN** it resolves through exactly one dispatcher entry and no binding shadows it with a local copy

#### Scenario: Alias equals canonical
- **WHEN** a plugin calls an alias and then its canonical name with the same arguments
- **THEN** both return identical values, because the alias forwards to the same implementation

#### Scenario: Lua and JS expose the same host table
- **WHEN** the host tables of the Lua and JS runtimes are enumerated
- **THEN** both contain exactly the same names as the dispatcher registry

#### Scenario: Same tool, same result
- **WHEN** a Lua plugin and a JS plugin call `host.now`, `host.http.fetch` and `host.storage.set`/`get`
- **THEN** both resolve the same tool set and produce identical JSON shapes

#### Scenario: WASM guest stays linkable
- **WHEN** an existing WASM plugin is loaded after a parity name lands
- **THEN** it links and runs with its previous imports and no new required import

## ADDED Requirements

### Requirement: Baseline host tool name parity

The host surface SHALL cover the cppIsekai baseline vocabulary. Every baseline name SHALL resolve
through the dispatcher, either as a canonical implementation or as a registered alias. A missing
name SHALL be closed by adding the implementation under its baseline name or by registering the
alias; existing names SHALL NOT change behavior.

#### Scenario: Core, state and network names resolve
- **WHEN** the dispatcher is enumerated for `now`, `log.info`, `log.error`, `log.debug`, `storage.get`, `storage.set`, `plugin.param`, `config.get`, `url_encode`, `url_decode`, `url_abs`, `json.encode`, `json.decode`, `http.fetch`, `http.fetch_body`, `http.get_json`, `http.post_json`, `fs.read` and `fs.read_lib`
- **THEN** each name resolves exactly once, as an implementation or a registered alias

#### Scenario: Text, regex, codec, crypto and browser names resolve
- **WHEN** the dispatcher is enumerated for `text.normalize_status`, `text.chapter_num`, `text.date_to_iso`, `text.date_epoch`, `text.json_blob`, `text.trim`, `text.unescape`, `text.strip_html`, `text.clean`, `re2.find`, `re2.match`, `re2.find_all`, `re2.gmatch`, `re2.replace`, `re2.quote`, `re2.find_index`, `re2.capture`, `codecs.base64_encode`, `codecs.base64_decode`, `codecs.hex_encode`, `codecs.hex_decode`, `crypto.vrf_sign`, `crypto.sha256_hex`, `crypto.md5_hex`, `crypto.hmac_sha256_hex`, `crypto.aes_gcm_decrypt`, `browser.call`, `html.select`, `html.attr` and `html.xpath`
- **THEN** each name resolves exactly once, as an implementation or a registered alias

#### Scenario: Known parity gaps close without behavior change
- **WHEN** the recorded gap list is checked: `log.debug`, `fs.read_lib`, `text.normalize_status`, `text.chapter_num`, `text.date_to_iso`, `text.trim`, `text.unescape`, `text.strip_html`, the `codecs.*` names, `re2.gmatch`, `crypto.md5_hex`, `crypto.hmac_sha256_hex`, `crypto.vrf_sign`, `crypto.aes_gcm_decrypt`, and Lua/JS wrappers for `browser.call` and `html.*`
- **THEN** every gap is a new canonical dispatcher entry or a registered alias proven equivalent by vector tests, and no pre-existing name changes behavior

### Requirement: Confined io shim

The plugin environment SHALL expose a read-only `io` shim providing only `io.read` and `io.lines`,
confined to the calling plugin's directory. Paths outside that directory and `..` escapes SHALL
fail with a typed error. This deliberately diverges from the baseline, whose `io.lines` accepts
arbitrary paths.

#### Scenario: Confined read succeeds
- **WHEN** a plugin calls `io.read` or `io.lines` on a file inside its own directory
- **THEN** the contents are returned read-only, identical to `fs.read` on the same path

#### Scenario: Confined io.lines escape fails
- **WHEN** a plugin calls `io.lines("../secrets")` or an absolute path outside its directory
- **THEN** the call fails with a typed path error and no file outside the plugin directory is opened
