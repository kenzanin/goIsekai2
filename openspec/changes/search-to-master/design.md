# Design: search-to-master

## D1. Search result is the fetch action

Search cards link to `/view/manga/:plugin/:slug?fetch=1`. Only that parameter triggers `Service.Prepare`; plain reads from the library or a typed URL never fetch.

## D2. JSON cache is staging

`fetchRaw` calls the plugin and writes the raw envelope to `api_cache` under the existing `(plugin, method, slug)` hash. `Prepare` calls it for `detail` and `chapters` and does not call `materialize`.

## D3. Save promotes

`saveLibrary` calls `Service.PromoteCached` before `library.Save`. Promotion reads the cached detail envelope, canonicalizes it and writes metadata into `saved_manga` while preserving `added_at`; the library Save then sets membership and publishes the event.

## D4. Fallback

If no cached envelope exists, Save keeps its existing minimal-payload behavior so direct API callers still work.

## D5. Tests

Cover: `?fetch=1` populates cache without a master row; Save without cache still works; Save with cache copies metadata; library read after Save makes no plugin call.
