# Notes
* [unknwon.dev/ini/v2 round-trip behaviour](go-ini-v2-behaviour.md) - v2 preserves comments and key order semantically, NewKey updates an existing key in place, and an empty value reads back as unset.
* [Config resolution gotchas found while building add-config](config-resolution-gotchas.md) - Registry bounds use 0 as unbounded, empty ini/env values count as unset, and gofmt is not on PATH so the Makefile must resolve it through the GOROOT.
* [FTS5 works on the pure-Go GORM stack, but V2 specs omit it](fts5-purego-probe.md) - modernc.org/sqlite v1.62.0 exposes SQLite 3.54.1 with FTS5; CREATE VIRTUAL TABLE and MATCH succeed under CGO_ENABLED=0, while the goIsekai2 specs still lack an FTS5 requirement.
