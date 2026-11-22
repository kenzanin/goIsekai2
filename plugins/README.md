# Bundled plugin corpus

These plugins are ported from the read-only cppIsekai baseline. Sources are copied
verbatim; only `main.json` is normalized for goIsekai2 (`entry`, `logo`, `api: 1`).

## Selection evidence

Top five manga plugins by `saved_manga` rows in
`~/tmp/cpp/cppIsekai/data/library.db` (chapters and reader positions are tie-breakers):

| Plugin | Saved | Chapters | Positions | Runtime |
|---|---:|---:|---:|---|
| mangafire | 39 | 3240 | 11 | lua |
| atsumaru | 15 | 3251 | 9 | lua |
| 1manga | 13 | 9968 | 11 | lua |
| mangatown | 7 | 489 | 6 | js |
| mangahub | 7 | 337 | 4 | lua |

Two enrichment sources come from `plugins_dev/lua/enrichment-mangadex` and
`plugins_dev/lua/enrichment-mangaupdates`.

## Layout

```
plugins/lua/<id>/          manga scrapers, Lua runtime
plugins/js/<id>/           manga scrapers, QuickJS runtime
plugins/enrichment/<id>/   enrichment-only sources (separate registry)
```

## Install

```bash
./scripts/install-plugins /path/to/data
```

This copies into `data/plugins` and `data/enrichment` without deleting unrelated
folders. The daemon config key `plugin.plugins` must point at `data/plugins`.

## Port status

- Ported now: mangafire, atsumaru, mangatown, mangahub, mangadex-enrichment,
  mangaupdates-enrichment.
- Deferred: 1manga uses multi-file `require`, which the Lua runtime does not expose
  yet. It will land with the plugin-side module resolution lane.
- Enrichment sources are installed but only become active when the `add-detail`
  change lands the separate enrichment loader.

## Verification

With the installed corpus and a daemon on port 18091, `GET /plugins` returned 200
and listed `Atsumaru`, `MangaFire`, `MangaHub` and `MangaTown`. Enrichment sources
are intentionally absent from that scraper list.
