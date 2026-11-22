# Tasks: add-behavior-specs

## 1. Write the deltas

- [ ] 1.1 MODIFY `library` save/unsave, placeholder-name and toggle-hydration requirements — verify they describe `added_at` membership, metadata preservation, remove-keeps-row and the `saved` payload field.
- [ ] 1.2 ADD the plugin-backed `/api/search` requirement to `library` — verify the endpoint path, query parameters and response envelope are explicit.
- [ ] 1.3 MODIFY `detail` materialization and add canonicalization/image-proxy requirements — verify `description→synopsis`, array flattening, rematerialization, `saved`, and `/cache/img` proxying are covered.
- [ ] 1.4 ADD `observability` capability — verify `/about`, `/log`, `/ws/log`, snapshot replay and debug-only instrumentation are specified.
- [ ] 1.5 ADD `ui-shell` capability — verify the canonical primary/utility nav order is specified.

## 2. Gates and archive

- [ ] 2.1 Run `openspec validate add-behavior-specs --strict` and `openspec validate --specs` — verify both pass.
- [ ] 2.2 Confirm no runtime source file changed — verify `git diff --name-only` outside `openspec/` is empty.
- [ ] 2.3 Merge, apply the deltas and archive — verify capability map and porting order mention the two new V2-only capabilities.

## Workflow follow-up

- Merge `change/add-behavior-specs` into `main` with `--no-ff`.
- Archive as `2026-11-22-add-behavior-specs`.
