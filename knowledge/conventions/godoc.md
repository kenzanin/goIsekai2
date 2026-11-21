---
type: Process
title: Godoc comment rule for all Go source
description: "Mandatory godoc conventions for goIsekai2 Go code: package comments, name-first doc comments on every exported symbol, behavior over implementation, no banner comments, examples as tests, and the review gate that enforces it."
tags: [convention, godoc, comments, review, go]
generated: { by: agent/cli, at: "2026-11-21T13:27:50Z" }
status: stable
---

# Godoc comment rule

Every Go source file in this repo is documented the godoc way. This is a review gate, not a style
preference: a diff that adds an undocumented exported symbol is rejected.

## Rules

1. **Package comment.** Every package has exactly one doc comment directly above `package <name>`, as
   `// Package config resolves ...`. It says what the package is for and what it is *not* for. Test
   packages (`package config_test`) need none.
2. **Exported identifiers.** Every exported type, func, method, const and var carries a doc comment
   whose first word is the identifier itself:
   `// Snapshot is an immutable resolved configuration.` Not `// This struct holds...`.
3. **Behaviour, not implementation.** Document the contract: what it does, preconditions, what it
   returns on error, who owns the returned value, and whether it is safe for concurrent use.
   `// Resolve reads the ini, then the environment, then explicit flags.` Not `// Loop over entries.`
4. **Do not restate the signature.** `// Set sets the value` adds nothing. Say what callers cannot see
   from the type: ordering, units, nil semantics, ownership, idempotence, side effects.
5. **Unexported code.** Comment when the intent is not obvious from the name — invariants ("callers
   must hold s.mu"), why an alternative was rejected, or a non-obvious workaround. Do not narrate
   trivial two-liners.
6. **No banner comments.** No `// ===== helpers =====`, no ASCII rules. If a file needs sections, it
   needs files.
7. **Sentences.** Sentence case, full sentence, ends with a period. Multi-paragraph doc comments are
   fine; the first line is the summary shown by `go doc`.
8. **Examples instead of prose.** Behaviour worth illustrating becomes an `Example`/`ExampleFoo` test,
   not a code block in a comment.
9. **Markers.** `// TODO(name): ...` names an owner. A deliberate simplification is marked
   `// ponytail: ceiling X, upgrade when Y`. Generated files start with the standard
   `// Code generated ... DO NOT EDIT.` line and nothing else is required.
10. **Non-obvious Go traps** (transaction nesting, atomic ordering, why a mutex is held across a
    call) get a comment *at the trap*, not in a separate document.

## Why

The baseline cppIsekai documents behaviour per module in `knowledge/`; the Go port keeps the same
information next to the code so `go doc ./internal/config` answers "what does this do" without
opening the file, and so a reviewer sees the contract in the diff. Ported facts (baseline file,
deliberate divergence) go in the doc comment of the symbol that carries the behaviour.

## Gate

`gofmt -l .` (empty) plus review: no undocumented exported symbol in the diff. `go vet` does not
check doc comments, so absence shows up in review, not in CI — which is why it is written down here.

# Related Concepts
- [Architecture overview](../architecture/overview.md)
- [Stack review: library vs hand-written](../architecture/stack.md)
- [Capability porting order](../process/porting-order.md) — the godoc rule applies to every lane a
  worker implements.
- [Architecture overview: layers, import rules and frozen external contracts](../architecture/overview.md): applies to every Go package described by the layer map
