# Tasks: add-plugin-host

## 1. Discovery and manifest (parallel lanes 1A/1B)

- [x] 1A.1 Add `yuin/gopher-lua`, `modernc.org/quickjs`, `tetratelabs/wazero` and `qmuntal/stateless` at the pinned versions — verify `go list -m` reports each version and `CGO_ENABLED=0 go build ./...` exits 0.
- [x] 1A.2 Implement `internal/plugin/manifest.go`: `Manifest{Name,Version,Runtime,Capabilities,API,Entry,Logo}` with `Load(path)` validation — verify a test rejects missing fields, `api != 1` and unknown runtime, and names the folder path in the error.
- [x] 1B.1 Implement `internal/plugin/discovery.go`: scan one folder per plugin, folder name is id, skip broken manifests with a warning, first duplicate in scan order wins — verify tests for valid, broken and duplicate folders.
- [x] 1B.2 Implement `internal/plugin/registry.go` holding the discovered manifest list with lookup by id — verify a test finds a registered plugin and returns not-found for an unknown id.

## 2. Runtime boundary (parallel lanes 2A/2B)

- [x] 2A.1 Define `internal/plugin/runtime.go`: `Runtime.Call(ctx, envelope []byte) ([]byte, error)` and `Envelope{API,Method,Params}` with the frozen method vocabulary — verify a test rejects an unknown method without invoking the runtime.
- [x] 2A.2 Implement `internal/plugin/lua.go` with gopher-lua, only base/string/table/math/json/read-only io, no `os`/`debug`, `require("ffi")` fails clearly, instruction budget — verify a runaway loop test returns a budget error.
- [x] 2B.1 Implement `internal/plugin/js.go` with modernc.org/quickjs, `ExecutePendingJobs`, interrupt deadline and no FS/network/process globals — verify a Promise return is an error and a loop hits the deadline.
- [x] 2B.2 Implement `internal/plugin/wasm.go` with wazero WASI preview1, exported `handle` and fuel limit — verify a module without `handle` is rejected and a loop hits the fuel limit.

## 3. Lifecycle and serialization (parallel lanes 3A/3B)

- [x] 3A.1 Implement `internal/plugin/lifecycle.go` with qmuntal/stateless states `Disabled`, `Enabled{Idle,Busy}`, `Failed` and a typed illegal-trigger error plus counter — verify a conformance test over all state/trigger pairs and the graph golden.
- [x] 3A.2 Persist each transition through the store, commit, then publish `plugin.lifecycle` after the commit — verify a recorder test asserts file/DB change before the event and that a failed write publishes nothing.
- [x] 3B.1 Implement the single-call-per-instance FIFO with per-caller deadline — verify a queued caller whose deadline expires returns a timeout and the plugin is invoked once.
- [x] 3B.2 Implement health as `{ok,error,at}` in memory, never a lifecycle transition — verify a failing call shows the error while state stays `Enabled`, and a later success clears it.

## 4. Host wiring and page (parallel lanes 4A/4B)

- [x] 4A.1 Wire discovery + reconciliation into daemon boot after store migrations and before `net.Listen` — verify a broken plugin does not block listening and a missing directory logs an error without pruning.
- [x] 4A.2 Add `POST /plugins/reload` and the per-plugin reload action using the boot pipeline — verify edited code runs on the next call and a broken edit marks that plugin `Failed` only.
- [x] 4B.1 Add the `HostTools` interface and a test double; no runtime receives the store or bus — verify a compile-time test keeps both handles out of the runtime package.
- [x] 4B.2 Add the templ `/plugins` page with status/runtime/capability/title-count rows, enable/disable toggle, reload action and empty-state guidance — verify httptest asserts row rendering, empty state and `enabled` JSON parity.

## 5. Purge and gates

- [x] 5.1 Purge data rows of a plugin whose folder disappeared in one transaction, plugin row last; short-circuit when the directory is absent — verify the missing-directory case deletes nothing and the removed-folder case deletes all seven row types.
- [x] 5.2 Run the full gate set — verify `make test lint`, `go vet ./...`, `gofmt -l .` empty, `openspec validate add-plugin-host --strict` and `CGO_ENABLED=0 go build ./...` all exit 0.

## Workflow follow-up

- Merge `change/add-plugin-host` into `main` with `--no-ff`.
- Archive the change with `openspec archive add-plugin-host -y` after the gates pass.
