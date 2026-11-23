# Spec Delta

## MODIFIED Requirements

### Requirement: Detail page rendering

`GET /view/manga/:plugin/:slug` SHALL render the detail shell and hydrate metadata and the chapter list from the JSON APIs, at most 20 chapter rows per request via `?page=` with the `/chapters` pager (prev/next and n/total, hidden on a single page); the full list stays at `/api/chapters` and `/chapters/:plugin/:slug`. Missing optional fields SHALL render as per-field placeholders. Every chapter row SHALL offer a read link and a cbz link, and search result cards SHALL link into this route.

#### Scenario: Page renders from plugin data
- **WHEN** a user opens a known title
- **THEN** the page shows cover, title, status, tags and the chapter list, each row with read and cbz entries

#### Scenario: Missing fields degrade per field
- **WHEN** the plugin omits optional metadata
- **THEN** those fields show placeholders and the rest of the page still renders

#### Scenario: Long chapter list paginates
- **WHEN** the chapter list holds 45 entries
- **THEN** page 1 renders 20 rows and the pager reads 1/3

#### Scenario: Short list hides pager
- **WHEN** the chapter list holds 12 entries
- **THEN** all 12 rows render and no pager is emitted

## ADDED Requirements

### Requirement: Enrichment candidates are cached

Enrichment candidate lookups SHALL be cached by `(kind, title)` in the shared `api_cache` with a finite TTL. A cache hit SHALL render without calling any enrichment plugin; a miss SHALL call sources once and store their merged candidates.

#### Scenario: Repeated detail renders do not fetch enrichment
- **WHEN** the same detail page renders twice inside the TTL
- **THEN** the enrichment source is invoked only on the first render

#### Scenario: Expired cache refreshes
- **WHEN** the cached entry has expired
- **THEN** the sources are called again and the new result replaces the entry
