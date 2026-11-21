---
type: Fact
title: "Zero config.Value renders \"0\", not \"\" — absent-knob paths must early-return"
description: "config.Value{}.String() returns \"0\" because KindInt is the zero Kind, so any code returning a zero Value for a missing knob leaks \"0\"; the tools ConfigGet had to add an explicit not-found branch."
tags: [internal/config, internal/plugin/tools]
generated: { by: agent/mcp, at: "2026-11-21T18:30:53Z" }
status: stable
---

Found while writing `ToolSet.ConfigGet`: the naive `res, ok := Snapshot.Get(key); return res.Value.String(), ok` failed its own test:

```
--- FAIL: TestConfigGet (0.01s)
    tools_test.go:55: ConfigGet("nope.nope") = "0",false; want "",false
```

Cause: `Value.String()` switches on `Kind`; the zero `Kind` is `KindInt` (iota 0), so a zero `Value` formats as `strconv.FormatInt(0, 1) = "0"` (internal/config/registry.go:87).

Consequence: every future `(string, bool)`-style accessor over `Snapshot.Get` needs an explicit `if !ok { return "", false }` — do not trust the zero `Value`. The plain `Snapshot.String` is safe (it returns `""` on absent). Fix is in internal/plugin/tools/tools.go `ConfigGet`; test `TestConfigGet` pins it.
