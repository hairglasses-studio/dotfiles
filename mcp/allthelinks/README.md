# allthelinks

A personal bookmark dashboard that replaces your browser's new-tab page. Single HTML file, zero dependencies, dark monospace aesthetic.

## Features

- **Collapsible groups** with persisted state
- **Real-time search** filter across all groups and links (`/` to focus)
- **Tag pills** with custom colors
- **One-liner notes** below each link
- **Clock display** (HH:MM, updates every 30s)
- **JSON import/export** via settings modal
- **Inline CRUD** — add, edit, and delete groups and links
- **Keyboard navigation** — `/` for search, `Escape` to clear

## Quick Start

Open `index.html` in your browser. That's it.

```
firefox index.html
# or
google-chrome-stable index.html
```

The dashboard ships with a default bookmark group (Mini PC eBay Watch List) as seed data. All changes are saved to `localStorage` automatically.

## Browser New-Tab Setup

### Chrome (unpacked extension)

1. Create a directory (e.g., `~/.config/allthelinks-ext/`)
2. Copy `index.html` into that directory
3. Create `manifest.json` in the same directory:
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
4. Go to `chrome://extensions`, enable **Developer mode**
5. Click **Load unpacked** and select the directory
6. New tabs now open your dashboard

### Firefox (temporary add-on)

1. Create a directory with `index.html` and the same `manifest.json` as above
2. Go to `about:debugging#/runtime/this-firefox`
3. Click **Load Temporary Add-on** and select `manifest.json`
4. New tabs now open your dashboard

Note: Temporary add-ons in Firefox are removed on restart. For persistent use, consider signing the extension via [addons.mozilla.org](https://addons.mozilla.org).

### Direct file:// usage

Set your browser's homepage to `file:///path/to/allthelinks/index.html`. Works in Chrome. In Firefox, `localStorage` is blocked on `file://` — the dashboard works but changes won't persist across reloads (a warning bar will appear).

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `/` | Focus search input |
| `Escape` | Clear search / close modal |

## Data Schema

Bookmark data is stored as JSON in `localStorage` (key: `allthelinks_data`). See `schema.json` for the full structure. Top-level shape:

```jsonc
{
  "settings": { "title", "columns", "showClock", "showSearch" },
  "groups": [
    {
      "id": "g_xxx", "title": "...", "icon": "emoji",
      "collapsed": false,
      "links": [
        {
          "id": "l_xxx", "label": "...",
          "urls": [{ "href": "https://...", "text": "Link text" }],
          "tags": [{ "text": "Tag", "color": "#bb86fc" }],
          "note": "Optional note"
        }
      ]
    }
  ]
}
```

## Import/Export

1. Click the gear icon in the header
2. The settings modal shows the complete JSON state in a textarea
3. **Export**: click "Copy JSON" to copy to clipboard
4. **Import**: paste new JSON into the textarea and click "Save"

## Design

- Palette: `#0f0f0f` background, `#1a1a1a` surface, `#6be5a0` accent green, `#82b1ff` links
- Typography: `SF Mono`, `Cascadia Code`, `JetBrains Mono`, monospace
- No images, no icon fonts, only Unicode/emoji
- No CDN, no npm, no build step
- Responsive: 1 column on narrow viewports, 2 columns on wide

## License

MIT — see [LICENSE](LICENSE)
