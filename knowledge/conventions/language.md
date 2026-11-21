---
type: Process
title: "English for all artifacts, Indonesian only in chat"
description: Every artifact goIsekai2 produces is written in English; Indonesian is allowed only in chat replies to the user.
tags: [convention, language, documentation]
generated: { by: agent/mcp, at: "2026-11-21T13:46:41Z" }
status: stable
---

# Language rule

Everything this project produces is written in **English**. Indonesian is allowed in exactly one
place: chat replies to the user.

## Applies to

- Go source: identifiers, comments, godoc, package and variable names.
- Tests: test names, table-entry labels, assertion messages, fixtures.
- OpenSpec: proposals, designs, specs, deltas, tasks, change ids.
- OKF: concept bodies, titles, descriptions, tags, log entries.
- Git: commit subjects, bodies, branch labels, tag messages.
- Reports and handoffs: any file written for a reader, human or agent.
- KIT worker briefs and the worker's own answer. A lane that answers in Indonesian is off-contract.
- Config files, YAML, templates: keys, comments, and inline documentation.
- UI copy and markup that ship in the repository.

## Does not apply to

- Chat replies to the user. Indonesian, English, or a mix is fine there, casual register.
- Upstream data kept verbatim because it is part of the contract: site status vocabulary
  (`ongoing`, upstream typos such as `ongoin`/`onhaiatus`), scraped titles, plugin metadata,
  HAR captures, and error strings quoted from a third-party tool. Never translate these; they are
  values, not prose.

## Why

The baseline (cppIsekai), the specs, and the okf bundle are English. Mixed-language artifacts cost
review fidelity: a spec written in one language and code in another hides mismatches until runtime,
and it makes diffs unreadable to anyone else who picks the bundle up. The rule is mechanical on
purpose so it never needs a judgement call.

## Related Concepts
- [Godoc comment rule](godoc.md)
- [KIT headless worker](../references/kit-worker.md)
