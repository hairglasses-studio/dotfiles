/effort max

# Community Dotfiles & Ricing Index — 10 Scraping Agent Rounds

You are a research agent building the definitive search index of the Linux ricing community's dotfiles, tools, and resources. Execute 10 sequential rounds of GitHub scraping. Each round: discover repos, classify, extract metadata, write structured entries, commit checkpoint.

## Phase 0: Context Loading

Before scraping, load prior research to avoid duplication. Read these 3 essential files (skip the rest — their key findings are summarized inline below):

1. Read `docs/RESOURCES.md` — 21 awesome-* repos already tracked across 14 categories. Do NOT re-index these awesome-lists themselves — scrape their CONTENTS.
2. Read `ghostty/shaders/SOURCES.md` — 132 shaders attributed to 27 contributors. Cross-reference in Round 4.
3. Read `~/hairglasses-studio/docs/indexes/SEARCH-GUIDE.md` — topic coverage map, 2,177 indexed URLs.

**Prior research summary (already completed by earlier agents — build on, don't duplicate):**
- `docs/research/terminal/terminal-tui-rice-ecosystem.md` (727 lines) — full terminal emulator survey, Ghostty/WezTerm/Kitty/foot compared, eww/waybar/AGS ecosystem, migration to Quickshell/AGS v3/Astal noted
- `docs/research/terminal/glsl-shader-ecosystem.md` (692 lines) — Shadertoy, shader platforms, performance tuning
- `docs/research/terminal/dotfile-manager-comparison-2026.md` — 13 tools evaluated, dotter migration failed, custom install.sh confirmed best
- `docs/strategy/open-sourcing/competitive-analysis/dotfiles.md` — end-4 (13.7k stars), hyprdots (8.5k), ML4W (4.6k). Stars driven by screenshots + install scripts
- `docs/research/competitive/awesome-list-submissions.md` — 6 primary submission targets with PR templates

**Research gaps this session fills:**
1. No awesome-ghostty deep survey → Round 4
2. No awesome-hyprland deep survey → Round 1
3. No r/unixporn top repos audit → Round 6
4. No GitHub dotfiles trending → Rounds 1-9 (sort:stars + sort:updated)
5. No shader source attribution audit → Round 4 cross-ref
6. No Arch/Manjaro ricing tools survey → Round 6
7. No ricing tools taxonomy → Round 7 (awesome-ricing categories)
8. No dotfile framework comparison refresh → Round 7
9. No component config survey → Round 9
10. No consolidated research links for future sessions → Round 10

## My Stack (for relevance scoring)

Hyprland, eww bar, foot + Ghostty (138 GLSL shaders, manifest pipeline), zsh+starship, neovim, mako, wofi, hyprlock, catppuccin/Snazzy palette, Arch/Manjaro. 7 MCP servers (dotfiles-mcp, hg-mcp, process/systemd/tmux-mcp, mapitall, mapping). Cross-compositor shell library (Hyprland/Sway/AeroSpace).

- **HIGH**: Uses my stack tools, cyberpunk/neon aesthetics, eww widgets, Ghostty shaders, Arch Hyprland rices, desktop-as-MCP
- **MEDIUM**: General Wayland ricing, different tools but interesting patterns, NixOS Hyprland
- **LOW**: X11-only, macOS-only, Windows, no transferable techniques

## Output Location & Setup

    mkdir -p links/{hyprland,sway,bars,terminals,themes,nixos,components,collections,tools,shaders}

Initialize `links/_registry.json`:

    {
      "version": 1,
      "last_updated": "2026-04-05T00:00:00Z",
      "total_entries": 0,
      "entries": {}
    }

## Entry File Format

Each repo gets `links/{category}/{github_username}.md` (or `{username}-{repo-suffix}.md` for multiple repos per user).

YAML frontmatter — every field required, `null` for unknown:

    ---
    name: "dots-hyprland"
    github: "https://github.com/end-4/dots-hyprland"
    author: "end-4"
    stars: 13712
    last_updated: "2026-03-15"
    discovered_in_round: 1
    also_found_in_rounds: []
    is_fork: false
    type: "dotfiles"
    distro: ["arch", "fedora"]
    wm: "hyprland"
    bar: "ags"
    terminal: "foot"
    shell: "zsh"
    editor: "neovim"
    launcher: "fuzzel"
    notifications: "swaync"
    lockscreen: "hyprlock"
    color_scheme: "custom"
    has_screenshots: true
    has_install_script: true
    notable_techniques: ["Material You theming", "Quickshell QML"]
    tags: ["polished", "feature-rich"]
    relevance: "high"
    category_in_awesome: "Rices"
    source_awesome_list: "awesome-hyprland"
    ---

Body format (NOT indented — write at column 0):

```
Brief description of what makes this notable.

**Stack**: Hyprland + AGS + foot + zsh | Quickshell QML UI, Material You dynamic theming

**Notable patterns**: Material You color extraction, Quickshell widget framework

**Links**: [GitHub](url) | [Reddit](url) | [awesome-hyprland](url)
```

**Field constraints:**
- `type`: dotfiles | tool | collection | theme | shader | framework
- `wm`: hyprland | sway | i3 | bspwm | awesome | dwm | qtile | xmonad | null
- `bar`: eww | waybar | polybar | ags | lemonbar | null
- `terminal`: ghostty | kitty | alacritty | foot | wezterm | st | null
- `shell`: zsh | fish | bash | nushell | null
- `editor`: neovim | vim | emacs | helix | null
- `launcher`: rofi | wofi | fuzzel | dmenu | tofi | null
- `notifications`: mako | dunst | swaync | fnott | null
- `lockscreen`: hyprlock | swaylock | i3lock | null
- `color_scheme`: catppuccin-* | nord | gruvbox | dracula | tokyo-night | rose-pine | custom | null
- `relevance`: high | medium | low (see "My Stack" above)
- `is_fork`: true if GitHub marks it as a fork — skip forks of repos already indexed
- `stars`: integer, 0 = unchecked
- `last_updated`: repo's last push date (YYYY-MM-DD)

## Category Directories

- `hyprland/` — Full dotfiles with Hyprland as primary WM
- `sway/` — Full dotfiles with Sway or i3
- `bars/` — Standalone bar configs (eww, waybar, AGS, polybar)
- `terminals/` — Terminal configs, themes, plugins (Ghostty, kitty, alacritty, wezterm)
- `shaders/` — GLSL shader collections, Ghostty shaders, terminal visual effects
- `themes/` — Color scheme frameworks and ports (catppuccin, nord, dracula)
- `nixos/` — NixOS/home-manager declarative configs
- `components/` — Standalone lockscreen, launcher, notification, wallpaper, SDDM, GRUB, Plymouth configs
- `tools/` — Ricing tools, utilities, dotfile managers, install frameworks
- `collections/` — Awesome-lists and curated indexes (preserve their category structures)

Classify by primary purpose. Note secondary categories in body.

## Dedup Protocol

`links/_registry.json` keyed by GitHub URL. Load at START of every round. If URL exists, update `also_found_in_rounds` only — never duplicate files. Skip repos that are GitHub forks of repos already in the registry.

## Round Execution Protocol

For EACH round (1-10):

1. **Load**: Read `links/_registry.json`
2. **Search**: Run the round's queries using the Agent tool for parallelism (launch up to 3 agents per round for independent queries — each agent runs its queries and returns a JSON list of `{url, name, stars, description, is_fork}`)
3. **Dedup**: Filter against registry. Skip forks of already-indexed repos.
4. **Classify**: Assign category directory
5. **Extract**: For repos with 10+ stars, read README via `mcp__github__get_file_contents(owner="X", repo="Y", path="README.md")`. Budget: max 25 file reads per round. Under 10 stars: metadata from search description only.
6. **Write**: Create `.md` entry files with full frontmatter + body at column 0
7. **Update**: Write `_registry.json` with new entries, update `total_entries`
8. **Commit**: `git add links/ && git commit -m "links: round N - {domain} ({X} new, {Y} total)"`
9. **Report**: Print 5-line summary: new repos, total, errors, top 3 starred, any rate limit issues

**On error**: GitHub 403/429 → wait 60s, retry once. Still failing → switch to WebSearch. Partial results → commit what you have, note gap in commit message.

**Fork handling**: GitHub search results include `fork: true` in API response. Skip forks unless they have significantly more stars than the parent (indicates the fork became the active project).

## MCP Tool Reference

GitHub search — use `perPage=100` for maximum coverage per query:

    mcp__github__search_repositories(query="hyprland dotfiles sort:stars", perPage=100)
    mcp__github__search_code(q="filename:eww.yuck deflisten", per_page=100)
    mcp__github__get_file_contents(owner="end-4", repo="dots-hyprland", path="README.md")

For repos with 200+ search results, also fetch `page=2`:

    mcp__github__search_repositories(query="hyprland dotfiles sort:stars", perPage=100, page=2)

WebFetch — for awesome-list READMEs. Try `main` branch first; if 404, retry with `master`:

    WebFetch("https://raw.githubusercontent.com/org/repo/main/README.md", ...)
    # if 404: WebFetch("https://raw.githubusercontent.com/org/repo/master/README.md", ...)

WebSearch — for Reddit and broader web discovery.

## Round 1: Hyprland Dotfiles & Rices

Queries:
1. `search_repositories(query="hyprland dotfiles sort:stars", perPage=100)`
2. `search_repositories(query="hyprland rice sort:stars", perPage=100)`
3. `search_repositories(query="dots-hyprland sort:stars", perPage=100)`
4. WebFetch `https://raw.githubusercontent.com/hyprland-community/awesome-hyprland/main/README.md` — extract ALL repo links. PRESERVE category names (Rices, Runners, Tools, Plugins, etc.) in `category_in_awesome` field.
5. `search_repositories(query="hyprland config eww waybar sort:stars", perPage=100)`

Create `collections/awesome-hyprland.md` with the full category tree from the awesome-list.

## Round 2: Sway, i3 & Wayland Rices

Queries:
1. `search_repositories(query="sway dotfiles sort:stars", perPage=100)`
2. `search_repositories(query="sway rice wayland sort:stars", perPage=100)`
3. `search_repositories(query="i3 dotfiles linux sort:stars", perPage=100)`
4. WebFetch `https://raw.githubusercontent.com/rcalixte/awesome-wayland/master/README.md` — extract all links with categories
5. `search_repositories(query="wayland tiling dotfiles sort:stars", perPage=100)`

Create `collections/awesome-wayland.md`.

## Round 3: Status Bars & Widget Systems

Queries:
1. `search_repositories(query="eww bar linux sort:stars", perPage=100)`
2. `search_repositories(query="waybar config custom modules sort:stars", perPage=100)`
3. `search_repositories(query="ags hyprland shell sort:stars", perPage=100)`
4. `search_code(q="filename:eww.yuck deflisten defwidget", per_page=100)`
5. `search_repositories(query="quickshell astal widget sort:stars", perPage=100)`

Include Quickshell, AGS v3/Astal, and any emerging bar/widget frameworks.

## Round 4: Ghostty, Shaders & Terminal Customization

Queries:
1. `search_repositories(query="ghostty config shaders sort:stars", perPage=100)`
2. `search_repositories(query="ghostty theme sort:stars", perPage=100)`
3. WebFetch `https://raw.githubusercontent.com/fearlessgeekmedia/Awesome-Ghostty/main/README.md` — extract ALL links with categories
4. `search_repositories(query="terminal shader glsl crt sort:stars", perPage=100)`
5. `search_repositories(query="kitty alacritty wezterm config theme sort:stars", perPage=100)`

Create `collections/awesome-ghostty.md`. After indexing, cross-reference against `ghostty/shaders/SOURCES.md`: list any shader repos discovered here that are NOT in SOURCES.md (these are new shader sources we should investigate).

## Round 5: NixOS & Declarative Dotfiles

Queries:
1. `search_repositories(query="nixos dotfiles flake sort:stars", perPage=100)`
2. `search_repositories(query="home-manager dotfiles sort:stars", perPage=100)`
3. `search_repositories(query="nixos hyprland rice sort:stars", perPage=100)`
4. WebFetch `https://raw.githubusercontent.com/nix-community/awesome-nix/main/README.md` — extract dotfiles/configs sections
5. `search_repositories(query="nix-darwin dotfiles sort:stars", perPage=100)`

Create `collections/awesome-nix.md`.

## Round 6: Arch/Manjaro Ricing & Unixporn

Queries:
1. `search_repositories(query="awesome archlinux sort:stars", perPage=100)`
2. `search_repositories(query="arch linux rice tools sort:stars", perPage=100)`
3. WebFetch `https://raw.githubusercontent.com/PandaFoss/Awesome-Arch/master/README.md` — extract ALL categories and links
4. WebSearch `site:reddit.com/r/unixporn arch manjaro hyprland rice 2024 2025 dotfiles` — extract GitHub repo links from top results
5. `search_repositories(query="unixporn dotfiles sort:stars", perPage=100)`

Create `collections/awesome-arch.md` with full category structure. Tool repos → `tools/`, full dotfiles → WM directories. This round covers gaps #3 (unixporn audit) and #6 (Arch tools).

## Round 7: Ricing Tools & Dotfile Managers

Queries:
1. WebFetch `https://raw.githubusercontent.com/fosslife/awesome-ricing/master/README.md` — extract ALL links with category structure (WMs, Bars, Launchers, Notifications, Compositors, Color Schemes, etc.). This is the community's canonical ricing taxonomy.
2. WebFetch `https://raw.githubusercontent.com/webpro/awesome-dotfiles/master/README.md` — extract all tools and frameworks
3. `search_repositories(query="dotfile manager stow chezmoi yadm sort:stars", perPage=100)`
4. `search_repositories(query="linux ricing tool pywal wpgtk sort:stars", perPage=100)`
5. WebFetch `https://raw.githubusercontent.com/zemmsoares/awesome-rices/main/README.md` — extract ALL entries

Create `collections/awesome-ricing.md`, `collections/awesome-dotfiles.md`, and `collections/awesome-rices.md` — each with FULL category trees preserved. This round is critical for the taxonomy.

## Round 8: Color Schemes & Theme Frameworks

Queries:
1. `search_repositories(query="catppuccin dotfiles rice sort:stars", perPage=100)`
2. `search_repositories(query="nord gruvbox theme dotfiles sort:stars", perPage=100)`
3. `search_repositories(query="dracula theme dotfiles linux sort:stars", perPage=100)`
4. `search_repositories(query="tokyo-night rose-pine dotfiles sort:stars", perPage=100)`
5. `search_repositories(query="pywal base16 flavor theme generator sort:stars", perPage=100)`

Theme framework repos → `themes/`. Full dotfiles rices → their WM directory with `color_scheme` set.

## Round 9: Components & Standalone Configs

Queries:
1. `search_repositories(query="hyprlock swaylock config theme sort:stars", perPage=100)`
2. `search_repositories(query="rofi theme rice collection sort:stars", perPage=100)`
3. `search_repositories(query="mako dunst swaync notification theme sort:stars", perPage=100)`
4. `search_repositories(query="sddm theme plymouth grub theme sort:stars", perPage=100)`
5. `search_repositories(query="cursor theme icon theme linux sort:stars", perPage=100)`

All → `components/`. Include lockscreens, launchers, notification themes, display managers, boot themes, cursor/icon themes.

## Round 10: Consolidation, INDEX & Research Links

No new searches. Instead:

1. **Validate**: Read all `.md` files in `links/*/`. Fix malformed frontmatter.
2. **Verify**: `_registry.json` entry count matches file count on disk. Report discrepancies.
3. **Generate `links/INDEX.md`**:
   - Summary stats: total repos, count by category, by WM, by bar, by terminal, by color scheme, by type
   - Full table: Name | Author | Stars | Type | WM | Bar | Terminal | Color | Relevance
   - "Priority Reference" section: all `relevance: high` repos sorted by stars
   - "Most Starred" section: top 20 by stars
   - "Awesome-List Category Index": every category from every awesome-list scraped, with links to entries
4. **Generate `links/SCHEMA.md`**: Document entry format, directories, how to add entries manually
5. **Generate `links/RESEARCH-LINKS.md`**: Reference for future Claude Code sessions:
   - All awesome-list URLs scraped with their category structures
   - All Reddit threads mined with titles/dates
   - Direct links to best community resources discovered
   - Cross-references to existing research in `~/hairglasses-studio/docs/research/terminal/` (8 files, 5,391 lines)
   - Quick-reference: "looking for eww patterns? start here..." / "best Hyprland rices? see..." / "shader sources? check..."
   - Links to the 5 prior research files this index builds upon
6. **Docs repo write-back**: Write `~/hairglasses-studio/docs/research/terminal/community-dotfiles-index-2026-04-05.md` summarizing:
   - Stats: what was indexed (totals, categories, coverage)
   - Top discoveries: repos the org should watch
   - Gaps remaining: what a future session should target
   - All new URLs for TAGGED-URL-INDEX integration
7. **Dotfiles commit**: `git add links/ && git commit -m "links: round 10 - INDEX, SCHEMA, RESEARCH-LINKS ({N} total repos)"`
8. **Docs commit** (use absolute path, don't cd):
   `git -C ~/hairglasses-studio/docs add research/terminal/community-dotfiles-index-2026-04-05.md && git -C ~/hairglasses-studio/docs commit -m "research: community dotfiles index ({N} repos)"`

## Completion

Print final summary:
- Total unique repos indexed (target: 300-500)
- Breakdown by category directory
- Top 10 most-starred discoveries
- Top 10 highest-relevance for our stack
- Awesome-list coverage: which lists fully scraped, which partially
- New shader sources discovered (not in SOURCES.md)
- Research gaps remaining for next session
- Fork candidates for dotfile-museum (high-relevance repos worth local archiving via `dotfiles_gh_onboard_repos`)

## Important Notes

- Working directory: `~/hairglasses-studio/dotfiles/`
- Git: hairglasses / mitch@hairglasses.studio, no GPG signing
- Build upon existing research — DO NOT re-read the 5,391 lines in docs/research/terminal/ (summaries above are sufficient)
- Stars threshold: 10+ for full metadata, under 10 only if notably unique
- Fork handling: skip forks unless they surpassed the parent in stars
- For repos already in `_registry.json`, just update `also_found_in_rounds`
- `_registry.json` is crash recovery — always current before committing
- When scraping awesome-lists, PRESERVE their category taxonomy — this is valuable structured data
- The `collections/` entries should document each awesome-list's full category tree
- Use `perPage=100` not 30 — 3x more results per API call
- Try `main` branch first for raw GitHub URLs, fall back to `master` on 404
