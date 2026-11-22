---
type: Reference
title: PI worker harness
description: Headless pi lanes run in herdr tabs with persistent resumable sessions and explicit timeouts.
tags: [pi, worker, herdr]
generated: { by: agent/mcp, at: "2026-11-22T01:37:03Z" }
status: stable
---

`scripts/pi-worker` launches a pi lane in a new herdr tab using `herdr tab create`, `pane run`, a `__PI_EXIT=` sentinel and `pane wait-output`. It uses the 9router `free` model from `~/.pi/agent/auth.json` and appends `.pi/system-prompt.md`. Sessions are always enabled: `--session-dir /tmp/opencode/pi-sessions` plus `--sid=<id>`; reusing the same `--sid` resumes the session. `--timeout=<ms>` (or `PI_WAIT_MS`, default 3600000) is passed to the herdr wait, so long lanes are allowed. Extensions and skills stay enabled. Example: `./scripts/pi-worker --sid=add-detail-a --timeout=3600000 "<prompt>"`. Verified headless: prompt delivered verbatim, file written, exit 0.

# Related Concepts
- [KIT headless worker harness (.kit.yml + scripts/kit-worker + kit-worker skill)](kit-worker.md): Alternative headless worker harness using pi instead of KIT
