# AGENTS.md - DMAA Protocol v0.1 - Instructions for AI Agents in `goIsekai2`

> Powered by [OKF Agent Memory](https://github.com/okf-memory/okf-agent-memory) -- Open Knowledge Format (OKF) v0.2 persistent project memory for AI agents.

## 0. Project & Domain Codex (Software Engineering)

- TONE: style == direct_concise, zero_pleasantries
- LANGUAGE: committed files => en_US; user_chat => Indonesian (casual)
- FORMAT: diagrams => ASSERT(syntax == mermaid, ELSE=STOP("All diagrams MUST use Mermaid syntax; ASCII/text art prohibited."))
- GOAL: maintain(high_factual_integrity, domain_conventions)
- REFERENCE: cppIsekai is the behavioral baseline; deviations must be recorded as a decision concept, never silent.

<!-- BEGIN OKF AGENT MEMORY -->
## 1. Behavioral Invariants & Constraints (RFC 2119)
- MUST run `okf_search(query=keywords, limit=3)` before architecture, dependency, or substantial-change proposals.
- NEVER scan `knowledge/` via `list_dir`, `grep_search`, `find`, or raw file readers.
- NEVER forge human verification (`verified:` is human-only; declare `generated: { by: "<actor>", at: "<iso-time>" }`).
- PREFER native `okf_*` MCP tools OVER CLI fallback commands.
- MUST persist durable decisions/facts via `okf_create`/`okf_update` in the session they are made (session-only knowledge is lost).

## 2. Guard Clauses & Scope Governance
- ON edit(@path/):
    IF first_visit(@path/) => okf_search(for_path=@path/)
    IF governance == "hold" => STOP("Subsystem frozen by governance. Request explicit human confirmation.")
    IF governance == "constraint" => MUST adhere to all listed invariants
    IF governance == "context" => proceed with awareness
- ON user_query(architecture | requirements | conventions | domain_facts):
    okf_search(query=keywords, limit=3) => evaluate summary description
    IF relevant => okf_show(concept_id) ONLY on demand
    IF updating_existing_concept => PREFER okf_update OVER okf_create

## 3. Completion Pipeline (Sequential Assertion Gates)
1. IF decisions_made => MUST okf_create(category/*, type="decision", title=..., desc=...)
2. IF requirements_discovered => MUST okf_update(concept_id)
3. IF concepts_mutated => MUST sync(knowledge/log.md, knowledge/index.md)
4. ASSERT(okf_validate(strict=true, drift=true) == {errors: 0, warnings: 0}, ELSE=fix_before_exit)
<!-- END OKF AGENT MEMORY -->

---

## Project

goIsekai2 -- second-generation Go rewrite of the Isekai manga viewer (local-first
manga library, plugin-per-site scrapers, disk cache, reader UI, tray daemon).

**Status: greenfield.** No application code yet; only the memory + spec harness exists.

**Reference sources (read-only, never merged into this bundle):**

- `~/tmp/cpp/cppIsekai` -- the behavioral baseline (C++23, finished feature set).
  Its `knowledge/` bundle and `openspec/specs/` are the porting source of truth.
- `~/tmp/goIsekai` -- first Go generation (V1): stack trials, ~40 Lua plugins,
  ~40 JS plugins, wasm/native/yaegi experiments, real HAR captures under `tmp/`.
  Treat as a read-only corpus, not a dependency.

`knowledge/references/reference-sources` holds the detailed map with exact paths.

## Commands

```bash
make build            # go build -o bin/goisekai2 ./cmd/goisekai2
make test             # go test ./...
make validate         # okf validate knowledge --strict --drift
make search q=plugin  # okf search over the bundle
go build ./...        # compiler gate
```

OpenSpec (planning home = repo root):

```bash
openspec list --json                            # active changes
openspec validate --specs                       # validate main specs
openspec validate --change <name>               # validate one change
openspec archive <name> -y                      # archive after implementation
```

Change artifacts per spec-driven schema: `proposal.md`, `design.md` (decisions D1..Dn),
`specs/<capability>/spec.md` (ADDED/MODIFIED/REMOVED deltas), `tasks.md` (checkboxes).

## Architecture (target, not yet built)

Mirror cppIsekai's layer split; each layer becomes an OpenSpec capability spec before
code lands. Planned packages under `internal/`:

- `config/` -- typed config registry (CLI flag + env + ini + default, provenance dump)
- `store/` -- single-connection SQLite store; plugins never touch DB directly
- `plugin/` -- plugin-per-site runtime, JSON-text data contract, host tools
  (http fetch, storage, log, crypto, text)
- `cache/` -- disk LRU + per-plugin tree (`<plugin>/covers/`, `<plugin>/<slug>/ch.../p...`)
- `http/` -- daemon routes, SSR templates, image cache route, config save
- `reader/` -- window math, read-ahead prefetch, position state
- `events/` -- typed facade over `gookit/event`; ONE actor goroutine owns the manager; domain facts
  published after commit only. See `knowledge/architecture/event-driven-core.md`.
- `stream/` -- bounded drop-oldest fan-out for SSE (log, progress, grid refresh)
- `cmd/goisekai2/` -- entry point

## Conventions

- Go, stdlib-first. No new dependency without a recorded decision naming the
  alternative considered; prefer reusing what V1 already proved.
- `go vet ./...` + `gofmt` clean; tests next to the package (`*_test.go`), one
  runnable check per non-trivial unit. No test framework beyond stdlib `testing`.
- Plugin ABI, cache layout, config knobs and status vocabularies must match
  cppIsekai semantics exactly; any intended divergence needs a decision concept
  plus an OpenSpec MODIFIED delta.
- Committed artifacts (docs, knowledge, openspec, reports) are English only.
- Bus rules: event names are constants in `internal/events`, payloads are typed structs, no goroutine
  touches `*event.Manager` except the actor, the async API is banned, and commands are never events.
- Commit subjects: short, lowercase imperative (`store: add migration runner`).
- Branch per change named after the change id; merge to `main` with `--no-ff`.
