# Design: add-lua55

## D1. One VM per call, same exported surface

`LuaRuntime` keeps `SetTools`, `SetDir`, `NewLuaRuntime`, `NewLuaRuntimeWithTools` and `Call`. `Call` creates
a fresh `lua.VM`, defers `Close`, wires tools and the require shim, evaluates the source, then calls the global
`handle` with the request decoded to a Go map. The JSON envelope is unchanged.

## D2. Host bindings as real Lua tables

`modernc.org/lua` does not create parent tables for dotted `RegisterFunc` names. Build `host`, `host.log`,
`host.storage`, `host.text`, `host.plugin`, `host.http`, `host.call` and `json` with `NewTable` + `Value.Set`.
Go functions keep the current signatures; errors are raised as Lua errors or returned in the frozen
`{ok,data,error}` envelope exactly as before.

## D3. Value conversion

Lua numbers arrive as `int64` or `float64`; both convert to Go numbers. Lua tables become `[]any` for dense
numeric keys (preserving the array fix from `579692f`) and `map[string]any` otherwise. Go maps/slices passed
into Lua are converted recursively so `json.decode` and `host.http.get_json` still produce Lua tables.

## D4. Require shim

Override the global `require` with a Go function. Resolve only `<pluginDir>/<name>.lua` with `filepath.Clean`
and a check that the result stays under the plugin directory; cache modules per VM. Evaluate with `EvalFile`
and convert the returned `*lua.Value` to Go values before returning it, so `require("genres").map` keeps
working. Reject `ffi`, `bit` and outside paths with explicit errors.

## D5. Budget and cancellation

Replace `L.SetContext(ctx)` with `SetEvalTimeout` derived from the context deadline plus a small watcher
goroutine that calls `VM.Interrupt` when `ctx.Done()` closes. The watcher exits when a `done` channel closes so
no goroutine leaks. Overruns still produce an `ok:false` budget error.

## D6. Sandbox

Remove `os`, `io`, `debug`, `dofile`, `loadfile` and the `package` table from the VM. Standard libraries that
the corpus actually uses (`base`, `string`, `table`, `math`, `utf8`) stay. File access is possible only through
the confined `require` shim.

## D7. Validation and conformance

`NewLuaRuntime` checks syntax with a throwaway VM and confirms a `handle` global exists. Tests cover a Lua 5.5
syntax case, a JSON round-trip, a plugin-local module, a sandbox escape, and a runaway loop. End-to-end
verification runs the shipped `atsumaru`, `mangafire`, `mangatown` and `mangahub` plugins through
search/detail/chapters/pages against the live daemon.

## D8. Dependency pinning

Pin `modernc.org/lua v0.1.0` and let MVS select the `modernc.org/cc*`/`libc` versions it requires; verify those
are pure Go and that `CGO_ENABLED=0 go build ./...` still passes. Remove `github.com/yuin/gopher-lua`.
