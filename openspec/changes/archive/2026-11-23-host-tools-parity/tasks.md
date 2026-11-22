# Tasks: host-tools-parity

> Serial fixer lanes: one `@fixer` subagent at a time, reviewed by the main session, gates green
> before the next lane starts. No parallel lanes.

## 1. Lane 1 — dispatcher truth and alias registry

- [x] 1.1 Add a data-driven alias registry (name → canonical entry) to `internal/plugin/tools` and route every binding through dispatcher resolution — verify a table test resolves each alias to the same implementation as its target.
- [x] 1.2 Add dispatcher entries for `text.normalize_status` and `plugin.param`; lane 3 removes binding-local copies — verify identical results for the same inputs before and after.
- [x] 1.3 Add canonical entries `text.chapter_num`, `text.date_to_iso`, `text.trim`, `text.unescape`, `text.strip_html` and `re2.gmatch` with cppIsekai semantics — verify table tests against cppIsekai fixtures (unicode, empty, malformed input).
- [x] 1.4 Add `fs.read_lib` with `fs.read` confinement — verify escape attempts fail typed and a library file inside the plugin dir reads.
- [x] 1.5 Add `log.debug` over the shared log sink; lane 3 removes the Lua `log.debug → log.info` remap — verify a Lua fixture emits a debug-level line.

Lane gate: `go vet ./... && gofmt -l internal && CGO_ENABLED=0 go build ./... && go test ./internal/plugin/...`.

## 2. Lane 2 — crypto and codec surface

- [x] 2.1 Audit `crypto.sha256_hex`, `crypto.md5_hex`, `crypto.hmac_sha256_hex`, `crypto.vrf_sign` against the existing hash/hmac/sign entries; register aliases where vector tests prove identical output, else add canonical entries — verify known vectors for every name.
- [x] 2.2 Register `codecs.base64_encode`, `codecs.base64_decode`, `codecs.hex_encode`, `codecs.hex_decode` as aliases of the crypto base64/hex helpers — verify the alias-equals-canonical test.
- [x] 2.3 Implement `crypto.aes_gcm_decrypt` as a new canonical entry — verify a known key/nonce/ciphertext vector round-trips and a tampered tag fails typed.

Lane gate: lane 1 gate.

## 3. Lane 3 — binding wrappers and parity lock

- [x] 3.1 Expose `browser.call`, `html.select`, `html.attr`, `html.xpath`, `re2.gmatch` and the `codecs.*` names in the Lua layer as thin marshals — verify a Lua fixture exercises each against local fixtures.
- [x] 3.2 Expose the same names in the JS binding — verify a JS fixture parity run.
- [x] 3.3 Add the parity lock test: enumerate Lua and JS host tables, assert equality with the dispatcher registry — verify it fails when a name lands in one runtime only.

Lane gate: `make build && make test && go vet ./...`.

## 4. Lane 4 — confined io shim

- [x] 4.1 Implement read-only `io.read`/`io.lines` in Lua and JS, confined to the plugin directory via the `fs.read` helper — verify in-dir reads match `fs.read` output.
- [x] 4.2 Typed path error for `..` escapes, absolute paths and outside-dir reads — verify escape fixtures fail without opening files outside the plugin dir.

Lane gate: lane 3 gate plus the io fixtures.

## 5. Close-out (main session)

- [x] 5.1 `openspec validate host-tools-parity --strict` clean; archive the change and sync specs.
- [x] 5.2 Record the confined-io divergence as a decision concept per the OKF pipeline (main-session duty; subagents do not touch `knowledge/`).
