# Tasks: add-updates

## 1. Scheduler core (parallel lanes 1A/1B)

- [ ] 1A.1 Implement `internal/updates/doc.go + scheduler.go`: one goroutine, 60s tick, jittered due check, config re-read, panic recovery and clean shutdown — verify a cancelled context exits within one tick.
- [ ] 1A.2 Implement candidate selection with per-plugin cap, never-attempted-first rotation and backoff filter — verify table tests for read_ch, disabled plugin and future backoff.
- [ ] 1B.1 Implement chapter-only refresh with growth bit0, bit6 preservation and attempt stamping — verify 40→42 sets bit0 and keeps dim.
- [ ] 1B.2 Implement exponential backoff capped at 72h and manual bypass — verify 3rd failure = 8h and 10th = 72h.

## 2. Maintenance and wiring (parallel lanes 2A/2B)

- [ ] 2A.1 Implement pacing, `update/last_run` watermark and run summary logging — verify delay runs only between titles.
- [ ] 2A.2 Implement API cache prune, VACUUM backup rotation and duplicate recluster as independent steps — verify a backup failure leaves the watermark unchanged but the batch runs.
- [ ] 2B.1 Wire scheduler start/stop into `cmd/goisekai2` around listen/shutdown — verify `CGO_ENABLED=0 go build ./...` and an integration run starts and stops cleanly.
- [ ] 2B.2 Add observability tests for run summary and named failures — verify logs carry plugin, slug and upstream error.

## 3. Gates

- [ ] 3.1 Run the full gate set — verify `make test lint`, `go vet ./...`, `gofmt -l .` empty, `openspec validate add-updates --strict` and `CGO_ENABLED=0 go build ./...` all exit 0.

## Workflow follow-up

- Merge `change/add-updates` into `main` with `--no-ff`.
- Archive the change with `openspec archive add-updates -y` after the gates pass.
