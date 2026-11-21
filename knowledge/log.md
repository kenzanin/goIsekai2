## 2026-11-21
* **Update**: Linked `architecture/event-driven-core.md` to `architecture/state-machines.md` (domain events originate from persisted FSM transitions).
* **Update**: Linked `architecture/state-machines.md` to `decisions/state-machine-engine.md` (implements the engine choice and wrapper rules).
* **Update**: Linked `architecture/state-machines.md` to `architecture/event-driven-core.md` (plugin and job transitions are the source of domain events; persist-then-publish ordering).
* **Creation**: Documented concept `decisions/state-machine-engine.md` (Use qmuntal/stateless as the FSM engine behind typed domain wrappers).
* **Creation**: Documented concept `architecture/state-machines.md` (State machines: plugin and job lifecycle on qmuntal/stateless; cache entry and daemon deliberately not).
* **Update**: Updated concept `decisions/event-dispatch.md`.
* **Update**: Linked `architecture/event-driven-core.md` to `decisions/event-dispatch.md`.
* **Creation**: Documented concept `decisions/event-dispatch.md` (Adopt gookit/event behind a single-actor bus; separate stream plane).
* **Creation**: Documented concept `architecture/event-driven-core.md` (Event-driven core: domain events on gookit/event, stream plane separate).
* **Creation**: Documented concept `references/reference-sources.md` (Reference source map (cppIsekai baseline, goIsekai V1 corpus)).
* **Creation**: Documented concept `decisions/goisekai-v2-bootstrap.md` (goIsekai2 is a greenfield Go rewrite against the cppIsekai baseline).
* **Creation**: Initialized OKF v0.2 knowledge bundle.
