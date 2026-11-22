# Tasks: add-reader-parity

## 1. Backend contracts

- [ ] 1.1 Export `reader.Service.Chapters` and pass all chapter options into `components.Reader` — verify the reader boot JSON contains every `{id, number, title}` entry.
- [ ] 1.2 Add `internal/store/progress.go` with `PutReadProgress`, `ReadProgress` and `ReadProgresses` on `plugin_state` under `host` — verify a progress value round-trips and a missing value returns zero.
- [ ] 1.3 Update `reader.Service.PutPosition` to store progress and set `ChapterRead` when the page reaches the chapter's stored page count — verify a last-page position flips the bit.
- [ ] 1.4 Extend `POST /chapter/flags` with repeated `ch_id` values and a JSON `chapter_ids` array — verify one request marks or clears multiple chapters.
- [ ] 1.5 Populate `ChapterRow.Read` and pass all chapter options to `components.Chapters` — verify a row renders `Read/Pages` and bulk buttons exist.

## 2. UI parity

- [ ] 2.1 Overlay top bar: chapter select from boot data and page select from `pages[]` — verify both navigate in place and update position.
- [ ] 2.2 Overlay bottom bar: numbered page buttons 1..N; collapse past 24 pages to the big current-page control — verify current button is highlighted and small chapters show the row.
- [ ] 2.3 Hidden-nav indicator: dim `p N / M` — verify it stays non-interactive while the overlay is hidden.
- [ ] 2.4 Keyboard: ignore keys targeting `INPUT`, `SELECT` or `TEXTAREA` — verify arrows inside the page jump input do not turn pages.
- [ ] 2.5 Detail chapter list: read-progress counters plus mark-all, clear-read and mark-previous with a confirm naming the range — verify one flags request is issued.

## 3. Gates and E2E

- [ ] 3.1 Regenerate templ assets with `make generate` — verify generated files are committed and `make test lint` stays green.
- [ ] 3.2 Run the full gate set — verify `make test lint`, `go vet ./...`, empty `gofmt -l .`, `CGO_ENABLED=0 go build ./...`, `openspec validate add-reader-parity --strict`, okf and AAG lint all exit 0.
- [ ] 3.3 Live check on the demo daemon — verify the reader overlay shows both selects, numbered buttons, the dim indicator, and the detail page shows progress counters and bulk controls.

## Workflow follow-up

- Merge `change/add-reader-parity` into `main` with `--no-ff`.
- Apply the MODIFIED delta to `openspec/specs/reader/spec.md`.
- Archive the change as `2026-11-22-add-reader-parity`.
