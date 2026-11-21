# Architecture
* [Event-driven core: domain events on gookit/event, stream plane separate](event-driven-core.md) - Two-plane event architecture: gookit/event managed by a single actor goroutine for domain facts after commit, plus a separate bounded drop-oldest stream hub for SSE; typed facade, event catalog, concurrency contract and gookit internals to defend against.
