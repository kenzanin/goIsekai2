---
type: Fact
title: "tmp/a.txt fixture: exact ALPHA payload, no newline; wc -c exits 0"
description: tmp/a.txt holds the exact 5-byte payload ALPHA (no trailing newline) and wc -c on it exits 0.
tags: [tmp, verification, fixture]
generated: { by: agent/mcp, at: "2026-11-21T17:19:28Z" }
status: stable
---

Verified by running:

```
$ printf 'ALPHA' > tmp/a.txt; wc -c tmp/a.txt; echo exit=$?
5 tmp/a.txt
exit=0
```

`printf` (not `echo`) guarantees no trailing newline, so the count is exactly 5. `wc -c` exits 0 on success.

Consequence: any script or test relying on tmp/a.txt must expect a 5-byte file; a trailing-newline variant would count 6.

# Related Concepts
- [tmp/nested/b.txt fixture: exact BETA payload, no newline; find exits 0](tmp-nested-b-fixture.md): Sibling tmp/ fixture probes with identical verification pattern (exact payload, no trailing newline, exit 0).
