# Spec Delta

## MODIFIED Requirements

### Requirement: Grid contract

Saved entries SHALL render cover cards whose image source is a `/cache/img` URL, never a raw CDN URL, so browser hotlink blocks cannot break the library grid. All other card behavior stays unchanged.

#### Scenario: Saved entry appears
- **WHEN** a manga is saved and the home page is opened
- **THEN** its cover card renders with a `/cache/img` source and its stat fields in the SSR context

#### Scenario: Deep-link plugin filter
- **WHEN** the home page opens with `?plugin=<id>`
- **THEN** the grid starts filtered to that plugin client-side

#### Scenario: Library versus search cache
- **WHEN** `title` holds seed rows but nothing is saved
- **THEN** the empty state renders instead of a grid

### Requirement: Search form

Plugin search hits SHALL carry an already-proxied cover URL and the SSR list SHALL render it as a thumbnail with an initial-tile fallback. The template SHALL NOT construct remote URLs.

#### Scenario: Disabled plugin is not offered
- **WHEN** a plugin is disabled
- **THEN** the source picker omits it and the page falls back to another enabled source or the empty prompt

#### Scenario: Linkable search
- **WHEN** a search runs with plugin, query, genre and page
- **THEN** reloading that URL reproduces the same request and results

#### Scenario: Inline states
- **WHEN** the query yields no results, or the plugin call fails
- **THEN** the grid shows the empty or error state respectively, without navigating away

#### Scenario: Search hit with cover
- **WHEN** a plugin returns an item with a cover URL
- **THEN** the result row renders a thumbnail served through `/cache/img`

#### Scenario: Search hit without cover
- **WHEN** a plugin returns an item with no cover URL
- **THEN** the row renders an initial tile and makes no image request
