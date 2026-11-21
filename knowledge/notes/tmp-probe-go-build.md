---
type: Fact
title: "tmp/probe.go builds: package tmp with OK() sentinel, go build ./tmp exit 0"
description: "tmp/probe.go (package tmp, func OK() string) compiles: go build ./tmp exits 0."
tags: [tmp, verification]
generated: { by: agent/mcp, at: "2026-11-21T17:20:23Z" }
status: stable
---

## Evidence

`tmp/probe.go` created with exactly:

```go
package tmp

func OK() string { return "OK" }
```

```
$ go build ./tmp; echo "exit=$?"
exit=0
```

## Consequence

The `tmp` package is a valid standalone Go package; `go build ./tmp` succeeds with CGO-free stdlib toolchain. Safe to keep or delete as scratch; nothing imports it yet.

# Related Concepts
- [tmp/dummy.txt sentinel probe: exact KIT_OK payload, ls exits 0](tmp-dummy-probe.md): Sibling scratch verification in tmp/: this probe extends the fixture chain with a compilable Go package sentinel.
