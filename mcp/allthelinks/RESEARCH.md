# allthelinks — Research Notes

## 1. Browser New-Tab Override Mechanisms

### Chrome

Chrome requires a **Manifest V3 WebExtension** to override the new-tab page. The minimal `manifest.json`:

```json
{
  "manifest_version": 3,
  "name": "allthelinks",
  "version": "0.1.0",
  "chrome_url_overrides": {
    "newtab": "index.html"
  }
}
```

Key constraints:
- The HTML file **must be bundled** inside the extension directory — cannot reference a `file://` URL
- Only one extension can override the new-tab page at a time
- Does not work in incognito windows
- Manifest V3 is required (V2 is fully deprecated as of 2024)

### Firefox

Firefox uses the same `chrome_url_overrides` WebExtension API with identical syntax. Firefox supports both Manifest V2 and V3 (V3 preferred for future-proofing).

The legacy `browser.newtab.url` preference was removed in Firefox 41 for security reasons. There is no `about:config` workaround — an extension is required.

### Cross-Browser Approach

A single extension works for both browsers:
1. Create a directory containing `manifest.json` + `index.html`
2. Chrome: load as unpacked extension via `chrome://extensions`
3. Firefox: load as temporary add-on via `about:debugging#/runtime/this-firefox`

### file:// Protocol Alternative

The page also works when opened directly as a `file:///path/to/index.html` URL. This is simpler for development and casual use, but has localStorage limitations (see Section 3).

## 2. New-Tab Dashboard UX Patterns

Researched: Bonjourr, nightTab, Tabliss, and Speed Dial extensions.

### Organization
- **Collapsible groups** (nightTab): semantic categories that can be folded to reduce visual clutter
- **Multi-page layouts** (Bonjourr): multiple pages for large collections — overkill for v0.1.0
- Keep initial view to 5-6 groups maximum for fast scanning

### Search and Filtering
- **Instant search**: real-time filter across all visible bookmarks as user types
- **Match highlighting**: optional visual highlight of matching text
- **Live count**: show number of matching results (nice-to-have)

### Inline Editing
- **Edit mode toggle** (nightTab): click a button to reveal edit controls, rather than always-visible buttons — keeps the default view clean
- **Hover-reveal actions**: edit/delete buttons appear on hover (simpler approach for v0.1.0)
- **Modal forms**: dedicated modal for editing link properties (URL, title, tags)

### Import/Export
- **JSON backup**: export full config as human-readable JSON file
- **Clipboard copy**: one-click copy to clipboard for quick sharing
- **Validate before import**: parse and validate JSON before overwriting state

### Keyboard Shortcuts
- `/` to focus search (universal convention from Vim, GitHub, etc.)
- `Escape` to clear search or close modal
- `Tab` / arrow keys for navigation within groups (future enhancement)

## 3. Local Storage Options

### localStorage

**Recommended for this project.**

- **Size**: 5-10 MiB per origin (5 MiB in most browsers). More than sufficient for hundreds of bookmarks.
- **API**: Simple synchronous `getItem`/`setItem` with `JSON.stringify`/`JSON.parse`.
- **Chrome file://**: Works. Each file gets isolated storage.
- **Firefox file://**: **Blocked.** Throws `SecurityError` due to same-origin policy ambiguity. This is an intentional security restriction (Mozilla bug #507361).
- **Extension origins**: Works in both Chrome (`chrome-extension://`) and Firefox (`moz-extension://`).

Error handling considerations:
- `QuotaExceededError`: catch and alert user to export and prune data
- `SecurityError` (Firefox file://): catch and fall back to in-memory operation with warning

### IndexedDB

Overkill for this use case. Asynchronous API adds complexity with no benefit for <500 bookmarks. Also blocked from `file://` protocol in standard browsers.

### Flat JSON File

Cannot read/write local files from JavaScript without user interaction (security restriction). The File API requires explicit user action (`<input type="file">` or drag-and-drop). Not viable as primary storage, but useful for import/export.

### Recommendation

Use **localStorage** as the primary persistence layer:
- Works seamlessly in Chrome (both file:// and extension)
- Works in Firefox as extension
- Graceful degradation for Firefox file:// (in-memory mode with warning bar)
- JSON import/export via settings modal provides manual backup/transfer
