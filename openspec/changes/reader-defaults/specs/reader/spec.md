# Spec Delta

## MODIFIED Requirements

### Requirement: Reading modes

With no persisted choice, the reader SHALL open in `paged` mode with `ltr` direction and `width` zoom. A stored mode, direction or zoom SHALL override the default. An explicit user choice SHALL beat the plugin's long-strip hint.

#### Scenario: Fresh reader defaults
- **WHEN** a manga has no stored reader position
- **THEN** it opens in paged mode with fit-width zoom

#### Scenario: Spread pairing with an odd page count
- **WHEN** pages 1..5 are read in `spread-h`
- **THEN** page 1 renders alone and pages (2,3) and (4,5) pair, with the lower-numbered page on the right in `rtl`

#### Scenario: Mode persists per manga
- **WHEN** a manga is switched to `spread-v` and reopened later
- **THEN** the reader restores `spread-v` and the last position

#### Scenario: User override wins
- **WHEN** the plugin flags long-strip but the user chose `paged`
- **THEN** the reader opens in `paged`

### Requirement: Position persistence and resume

`POST /reader/position` SHALL upsert `{plugin, slug, chapter, page, mode, direction, zoom}` per manga, keeping stored values for absent fields, and SHALL trigger prefetch. Reopening SHALL resume the persisted chapter, page, mode, direction and zoom.

#### Scenario: Resume after restart
- **WHEN** a manga is left at chapter 12 page 3 and the daemon restarts
- **THEN** reopening resumes at chapter 12 page 3

#### Scenario: Partial update keeps mode
- **WHEN** a position POST carries only chapter and page
- **THEN** the stored mode, direction and zoom are unchanged

#### Scenario: Zoom persists
- **WHEN** the user selects original zoom and reopens the manga
- **THEN** the reader restores original zoom

### Requirement: Full-screen surface and overlay

The overlay SHALL have a single top bar containing the back-to-detail control at the far right, chapter and page selects, adjacent chapter buttons, and mode/direction/zoom selectors. No bottom overlay bar SHALL be rendered. The dim `p N / M` pill stays while the overlay is hidden.

#### Scenario: Centre tap toggles
- **WHEN** the nav is hidden and the user taps the centre third
- **THEN** the overlay appears; tapping again hides it

#### Scenario: Chapter select switches in place
- **WHEN** the user picks a chapter in the top bar select
- **THEN** the reader requests that chapter's exact id, renders it without a reload, and updates the URL and position

#### Scenario: Page jump
- **WHEN** a chapter has 12 pages and the user picks page 7 in the bar
- **THEN** the reader shows page 7 and records the position

#### Scenario: Numbered page buttons
- **WHEN** the overlay is shown for a 12-page chapter
- **THEN** no numbered page button row is rendered; the top page select remains

#### Scenario: Large chapter compacts navigation
- **WHEN** a chapter has 80 pages and the nav is shown
- **THEN** no bottom jump control is rendered; the top page select still lists all 80

#### Scenario: No bottom bar
- **WHEN** the overlay is shown for any chapter length
- **THEN** no bottom overlay navigation panel is rendered

#### Scenario: Webtoon centre tap
- **WHEN** the reader is in webtoon mode and the user taps the centre third
- **THEN** the overlay toggles and no page change occurs
