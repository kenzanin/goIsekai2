# Decisions
* [goIsekai2 is a greenfield Go rewrite against the cppIsekai baseline](goisekai-v2-bootstrap.md) - Decisions: cppIsekai is the read-only behavioral baseline (its knowledge bundle + openspec specs), goIsekai V1 is a read-only scraping corpus, bundles never merge, module goisekai2.
* [Adopt gookit/event behind a single-actor bus; separate stream plane](event-dispatch.md) - User ruling 2026-11-21: gookit/event v1.2.0 as dispatch engine, owned by one actor goroutine, async API banned, typed payload facade, drop-oldest stream hub for SSE; hand-written hub kept as documented fallback.
