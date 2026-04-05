# allthelinks — Implementation Plan

> Single-file personal bookmark dashboard replacing browser new-tab pages.
> Zero dependencies. Works from `file://` and as a browser extension new-tab override.

## Context

The repo at `~/hairglasses-studio/allthelinks` has one commit (README.md placeholder). The source design file is `/home/hg/Downloads/index.html` — a dark-themed eBay mini PC watchlist with 10 cards, Snazzy-on-black palette, monospace typography, colored tag pills, and hover effects. We must preserve and extend this design language into a template-driven bookmark dashboard with localStorage persistence, search, collapsible groups, CRUD editing, and JSON import/export.

---

## Architecture

- **Single file**: `index.html` with embedded `<style>` and `<script>` blocks
- **Persistence**: localStorage (key: `allthelinks_data`), with graceful degradation when blocked (Firefox file://)
- **Default data**: Embedded as `DEFAULT_DATA` JS constant (can't fetch files from file:// due to CORS)
- **Rendering**: Full innerHTML replacement from JS data — no framework, no virtual DOM
- **Event handling**: Delegated click handlers on `#groups` container, survives re-renders

## Data Schema

### Top-level
```jsonc
{ "settings": { "title", "columns", "showClock", "showSearch" }, "groups": [...] }
```

### Group
```jsonc
{ "id": "g_xxx", "title": "...", "icon": "emoji", "collapsed": false, "links": [...] }
```

### Link (uses `urls[]` array — confirmed with user)
```jsonc
{
  "id": "l_xxx",
  "label": "Beelink SER7 — Ryzen 7 7840HS",
  "urls": [
    { "href": "https://...", "text": "Buy It Now" },
    { "href": "https://...", "text": "Sold prices" }
  ],
  "tags": [
    { "text": "Radeon 780M · RDNA 3", "color": "#bb86fc" }
  ],
  "note": "Best overall."
}
```

Tag border colors computed at render time via `dimColor()` (multiply RGB channels by 0.4).

---

## Critical Files

| File | Purpose |
|------|---------|
| `/home/hg/Downloads/index.html` | Source design reference (10 eBay cards to extract) |
| `index.html` | The entire application (~900-1100 lines) |
| `schema.json` | Reference schema with default eBay data (not loaded by app) |
| `README.md` | Usage, browser setup instructions, keyboard shortcuts |
| `RESEARCH.md` | Browser new-tab mechanisms, localStorage findings |
| `PLAN.md` | Technical plan (copy of this document) |
| `LICENSE` | MIT, hairglasses-studio |

---

## CSS Design

Preserve all original CSS custom properties verbatim. Add new variables:
- `--surface-hover: #222`, `--danger: #ef5350`, `--radius: 6px`
- `--font`, `--modal-backdrop: rgba(0,0,0,0.6)`, `--max-width: 960px`

Key layout decisions:
- Header: flexbox, title + clock left, gear + add-group right
- Search: full-width input with focus border transition
- Groups: CSS grid, `cols-2` class toggles 2-column, `@media (max-width: 700px)` forces 1-column
- Group actions & link actions: `opacity: 0`, shown on parent `:hover`
- Collapse: `display: none` on `.group-body` via `.collapsed` class
- Modal: fixed overlay with centered card, max-width 520px

---

## JavaScript Architecture

### Functions needed

**Helpers**: `uid(prefix)`, `esc(str)` (XSS prevention), `dimColor(hex)`, `debounce(fn, ms)`, `parseTags(str)`, `serializeTags(tags)`

**State**: `loadData()`, `saveData(data)`, `testStorage()` — probe write to detect Firefox file:// blocking

**Rendering**: `renderAll()` → `renderHeader()`, `renderSearch()`, `renderGroups()` → `renderGroup(group)` → `renderLink(link)`

**Search**: `onSearch()` — debounced 150ms, toggles `.hidden` on link cards and groups, shows "no results" message

**Collapse**: `toggleCollapse(groupId)` — toggles `collapsed` flag, saves, re-renders groups

**CRUD**: `addGroup()`, `editGroup()`, `deleteGroup()`, `addLink()`, `editLink()`, `deleteLink()`

**Modal**: `openModal(title, bodyHtml, footerHtml)`, `closeModal()`, `openSettingsModal()`, `openGroupModal(groupId)`, `openLinkModal(groupId, linkId)` — single modal overlay reused for all 3 types

**Settings**: `saveSettings()`, `exportJSON()` (clipboard), `importJSON()` (validate + load)

**Clock**: `updateClock()`, `startClock()` — interval every 30s

**Keyboard**: `/` focuses search, `Escape` clears search or closes modal

**Init**: `testStorage()` → `loadData()` → show warning if no storage → `renderAll()` → set up event delegation → `startClock()`

### Event delegation
Single `onclick` on `#groups` container dispatches to collapse, add-link, edit-group, delete-group, edit-link, delete-link based on button class. Set up once during init.

### Tag input format
Users type `#bb86fc:GPU tag, #64b5f6:RAM tag` — parsed by `parseTags()`. Tags without a color prefix default to accent green `#6be5a0`.

---

## Edge Cases

- **Corrupted localStorage**: shape validation in `loadData()`, fallback to `DEFAULT_DATA`
- **Firefox file://**: `testStorage()` probe fails → yellow warning bar, app works in memory-only mode
- **QuotaExceededError**: caught in `saveData()`, user alerted to export and prune
- **Empty states**: no groups, no links in group, no search results — each shows a helpful message
- **XSS**: all user strings escaped via `esc()` before innerHTML insertion

---

## Non-Goals (v0.1.0)

- No sync, accounts, or server components
- No drag-and-drop reordering
- No weather, RSS, or external integrations
- No build step

---

## Commit Sequence

Build on existing initial commit (confirmed with user).

### Commit 1: `feat: add initial eBay mini PC link index`
- Copy `/home/hg/Downloads/index.html` into repo as `index.html`
- Tag `v0.0.1`

### Commit 2: `docs: add RESEARCH, PLAN, LICENSE, schema, and README`
- Create `RESEARCH.md` — browser new-tab override mechanisms, localStorage behavior on file://, dashboard UX patterns
- Create `PLAN.md` — full technical plan
- Create `LICENSE` — MIT, hairglasses-studio
- Create `schema.json` — reference schema with all 10 eBay cards
- Update `README.md` — project description, features, browser setup instructions

### Commit 3: `feat: template-driven bookmark dashboard with search and collapsible groups`
- Rewrite `index.html` as complete dashboard
- Embedded CSS: all original variables + new layout (header, search, grid, groups, links, tags, modal, responsive)
- Embedded JS: DEFAULT_DATA (10 eBay cards), helpers, state management, render pipeline, search, collapse, clock, keyboard shortcuts, event delegation
- Seeded with eBay mini PC group from DEFAULT_DATA

### Commit 4: `feat: CRUD editing and JSON import/export`
- Modal system: settings, group edit/add, link edit/add
- CRUD operations: add/edit/delete groups and links
- Settings modal: title, columns, clock, search toggles + JSON textarea
- Export to clipboard, import with validation
- Tag input parsing (`#color:text` format)
- All changes persist to localStorage immediately

### Commit 5: `feat: keyboard shortcuts, empty states, and polish`
- Firefox file:// warning bar
- Empty state messages (no groups, no links, no search results)
- Responsive tuning
- Update README with keyboard shortcuts documentation
- Tag `v0.1.0`

---

## Verification

1. Open `index.html` as `file:///` URL in Chrome — confirm default eBay group renders, search filters, groups collapse, CRUD works, data persists across reloads
2. Open in Firefox as `file:///` — confirm warning bar appears, app works in memory-only mode, no JS errors
3. Test at narrow viewport (< 700px) — confirm single column layout
4. Test JSON export → modify → import round-trip
5. Test keyboard shortcuts: `/` focuses search, `Escape` clears/closes
6. Verify no external network requests (DevTools Network tab should be empty)
