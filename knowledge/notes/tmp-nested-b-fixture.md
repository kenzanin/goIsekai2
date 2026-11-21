---
type: Fact
title: "tmp/nested/b.txt fixture: exact BETA payload, no newline; find exits 0"
description: tmp/nested/b.txt holds the exact 4-byte payload BETA (no trailing newline) and find tmp -type f exits 0.
tags: [tmp, verification, fixture]
generated: { by: agent/mcp, at: "2026-11-21T17:19:08Z" }
status: stable
---

Verified by running, in repo root:

```
$ mkdir -p tmp/nested && printf BETA > tmp/nested/b.txt && find tmp -type f; echo "exit=$?"
tmp/dummy.txt
tmp/nested/b.txt
exit=0
```

Evidence: `printf` (not `echo`) writes BETA with no trailing newline; `find tmp -type f` exits 0 and lists both scratch files. Consequence: `tmp/nested/b.txt` is a 4-byte canonical fixture; any byte-exact comparison against it must not expect `\n`. Scratch files under `tmp/` are reproducible one-liners, safe to delete.

# Related Concepts
- [tmp/dummy.txt sentinel probe: exact KIT_OK payload, ls exits 0](tmp-dummy-probe.md): Sibling scratch-probe fact: same tmp/ fixture pattern, byte-exact payload plus exit-code proof.
