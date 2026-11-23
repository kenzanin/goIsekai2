# Design: atsumaru-newest-release

## D1. Dedup key

Chapter number stays the dedup key because the library UI groups by number. Duplicate entries already carry distinct composite ids (`mangaId:chapterId`), so no information is lost when one is chosen.

## D2. Winner

The release with the greatest `release_at` (converted from upstream `createdAt`) wins. Ties keep the earlier entry, so the mapping is deterministic.

## D3. Ordering

The existing ascending numeric sort is unchanged. Replacing the entry in place preserves deterministic output.
