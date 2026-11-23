# Design: cache-only-read

## D1. Read purity

`Service.Detail` always builds its payload from `saved_manga` plus `title_meta`. A missing row produces a
placeholder payload with the slug as name and `saved: false`; it never calls `fetch`.

## D2. Chapter reads

`Service.StoredChapters` is the read-path helper. `Service.Chapters` remains the mutation helper used only by
the explicit Fetch/Force action and the update scheduler. HTTP GET handlers stop falling back to it.

## D3. Enrichment split

`Candidates` becomes a pure `api_cache` read. `RefreshCandidates` performs the network calls and writes the
cache. `POST /api/enrichment/refresh/:plugin/:slug` is the user action that calls it; the scheduler may call it
directly.

## D4. Candidate application

`applyCandidate` resolves the title from the saved row and reads stored candidates, so applying a candidate
can never trigger a fresh website fetch.

## D5. UI

The Enrichment section gains a Refresh candidates button posting to the new route. Cold detail pages show
placeholders and the existing Fetch action fills them.

## D6. Tests

Cover: cold `Detail` performs no fetch; stale `api_cache` is ignored on read; `Candidates` never calls a source;
`RefreshCandidates` populates the store; the new route and button exist.
