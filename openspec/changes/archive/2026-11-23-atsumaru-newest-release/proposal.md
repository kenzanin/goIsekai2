# Proposal: atsumaru-newest-release

## Why

The user reported that chapter 1 of `AJDs` rendered a different page set from the upstream reader. Both point to Chapter 1 but use different release ids: upstream opens `2V2CqE`, our plugin returned the older duplicate `qRWvEm`. The old baseline rule "keep the first duplicate because pages are routed by number" is obsolete now that V2 routes pages by `ch_id`.

## What Changes

- `plugins/lua/atsumaru/main.lua` keeps the newest release (largest `createdAt`/`release_at`) per chapter number instead of the first.
- No host, store or reader code changes.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `sites` — duplicate chapter releases select the newest upstream release.

## Impact

- Chapter lists for titles with multiple scanlation groups change to the current upstream release; page order then matches the site.

## Baseline

V2-specific divergence: cppIsekai routed pages by number and could not address duplicate releases.
