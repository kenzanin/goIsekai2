# Design: host-tools-parity

## Context

See `proposal.md`. The `host-tools` spec froze the tool vocabulary; implementation and bindings
have since drifted (binding-local copies, a Lua-level remap, missing wrappers). cppIsekai
`src/plugin/lua/bindings/*` is the name authority for parity.

## Goals / Non-Goals

- Goals: every exposed host name resolves once; aliases are data, not code; Lua/JS host tables
  identical; gaps closed under baseline names; io shim confined and read-only.
- Non-Goals: new tool semantics, WASM import changes, plugin-side work, renaming existing tools.

## Decisions

### D1 — The dispatcher is the only implementation home

Every host name resolves through `internal/plugin/tools`. The binding-local `text.normalize_status`
and `plugin.param` copies move behind dispatcher entries and the copies are deleted. Binding-side
reimplementations are how the drift happened; they are removed, not tolerated.

Alternative: document the per-binding helpers as exceptions — rejected, the parity lock test
becomes meaningless if names can live outside the registry.

### D2 — Aliases are registered data with vector-proven equivalence

An alias registry (name → canonical entry) backs `codecs.*` over the crypto base64/hex helpers.
`crypto.sha256_hex`, `crypto.md5_hex`, `crypto.hmac_sha256_hex` and `crypto.vrf_sign` become
aliases of the existing `hash`/`hmac`/`sign` entries only where known-vector tests prove identical
output; otherwise they are new canonical entries. `crypto.aes_gcm_decrypt` has no equivalent today
and is canonical. Baseline: cppIsekai crypto host tools.

Alternative: duplicate implementations under both names — rejected, two bodies drift silently.

### D3 — Log levels share one sink

`log.debug` joins `log.info`/`log.error` as a name over the single log sink, parameterized by
level. The Lua-side remap of `log.debug` onto `log.info` is deleted; debug lines must be
distinguishable from info lines.

### D4 — Missing names land canonical under baseline names

`text.chapter_num`, `text.date_to_iso`, `text.trim`, `text.unescape`, `text.strip_html`,
`fs.read_lib` and `re2.gmatch` are new dispatcher entries with cppIsekai semantics; `fs.read_lib`
reuses `fs.read` confinement. Lua/JS wrappers marshal to them; no name is invented.

### D5 — Confined io shim, recorded divergence

`io.read`/`io.lines` are read-only and confined to the plugin directory, reusing the `fs.read`
confinement helper. Unlike baseline `io.lines`, arbitrary paths fail with a typed path error; the
divergence is recorded in the spec and proposal, per the project rule that divergences are
decisions, never silent.

### D6 — Parity is tested, not promised

A lock test enumerates the host tables of the Lua and JS runtimes and asserts equality with the
dispatcher registry (plus registered aliases). A name added to one runtime only fails CI.

## Risks / Trade-offs

- [Alias target changes semantics] → vector tests pin alias output to canonical output.
- [io divergence surprises ported plugins] → recorded in spec; escape failures are typed, not nil.
- [Binding dedup regressions] → parity lock plus before/after fixtures per moved helper.

## Migration Plan

Serial fixer lanes (dispatcher/aliases → crypto/codecs → binding wrappers → io shim), each
reviewed by the main session with gates green before the next; rollback is per-lane revert.

## Open Questions

None. Where alias-vs-canonical is conditional, the vector audit inside the lane decides and the
decision is visible in the lane's verification.
