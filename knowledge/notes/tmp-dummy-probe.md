---
type: Fact
title: "tmp/dummy.txt sentinel probe: exact KIT_OK payload, ls exits 0"
description: tmp/dummy.txt was created with the exact 6-byte payload KIT_OK and ls -l on it exits 0.
tags: [tmp, scratch, verification, kit]
generated: { by: agent/mcp, at: "2026-11-21T17:08:03Z" }
status: stable
---

Verified by running:

```
$ mkdir -p tmp && printf 'KIT_OK' > tmp/dummy.txt && ls -l tmp/dummy.txt; echo "exit=$?"
-rw-r--r-- 1 kenzanin kenzanin 6 Nov 22 00:07 tmp/dummy.txt
exit=0
```

6 bytes = the payload `KIT_OK` with no trailing newline; `ls -l` exited 0.

Consequence: `tmp/` exists as a writable scratch location in the repo and the `KIT_OK` sentinel file round-trips byte-exact through shell redirection — safe to use as a completion marker in worker/lane checks without quoting or newline surprises.

# Related Concepts
- [KIT headless worker harness (.kit.yml + scripts/kit-worker + kit-worker skill)](../references/kit-worker.md): KIT_OK sentinel probe used to verify worker lane completion markers.
