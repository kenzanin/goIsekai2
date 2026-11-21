---
type: Fact
title: "tmp probe test green: go test ./tmp -count=1 exit 0, TestOK passes against OK() sentinel"
description: "go test ./tmp -count=1 exits 0: TestOK asserts OK()==\"OK\" and the existing sentinel passes."
tags: [tmp, verification, testing]
generated: { by: agent/mcp, at: "2026-11-21T17:21:42Z" }
status: stable
---

Verified 2026-11-22: `go test ./tmp -count=1` exits 0 with output `ok  	goisekai2/tmp	0.003s`, after adding `tmp/probe_test.go` (package tmp, `TestOK` asserting `OK() == "OK"`). The existing `OK()` sentinel in package tmp satisfies the assertion unchanged. Consequence: the `go test ./...` gate also covers `internal/`-adjacent scratch package `tmp`; the probe test is disposable scaffolding and can be deleted with the rest of `tmp/` once real packages land.

# Related Concepts
- [tmp/probe.go builds: package tmp with OK() sentinel, go build ./tmp exit 0](tmp-probe-go-build.md): Same scratch package: the build probe proved compilation, this test probe proves the test gate too.
