/effort max

# Community Dotfiles Scraping Index — 10 Rounds

You are a research agent building a comprehensive search index of the Linux ricing community's dotfiles repos. You will execute 10 sequential rounds of GitHub scraping, each targeting a different domain. Every round discovers new repos, classifies them, extracts metadata, writes structured entry files, and commits a checkpoint.

## My Stack (for relevance scoring)

I run: Hyprland, eww bar, foot terminal, Ghostty (138 GLSL shaders), zsh+starship, neovim, mako notifications, wofi launcher, catppuccin/Snazzy custom palette, Arch/Manjaro. Anything using these tools or cyberpunk/neon aesthetics is HIGH relevance. General Linux ricing with different tools is MEDIUM. X11-only or macOS-only is LOW.

## Step 0: Setup

Create this directory tree under the current working directory:

    mkdir -p links/{hyprland,sway,bars,terminals,themes,nixos,components,collections}

Read `docs/RESOURCES.md` to seed your awareness of already-known awesome-lists (80+ entries). Do NOT re-index repos we already track there — focus on discovering their CONTENTS (the dotfiles repos listed inside those awesome-lists).

Initialize `links/_registry.json`:

    {
      "version": 1,
      "last_updated": "2026-04-05T00:00:00Z",
      "total_entries": 0,
      "entries": {}
    }

## Entry File Format

Each repo gets one `.md` file: `links/{category}/{github_username}.md` (or `{username}-{repo-suffix}.md` for users with multiple repos).

YAML frontmatter — every field required, use `null` for unknown:

    ---
    name: "dots-hyprland"
    github: "https://github.com/end-4/dots-hyprland"
    author: "end-4"
    stars: 13712
    last_updated: "2026-03-15"
    discovered_in_round: 1
    also_found_in_rounds: []
    distro: ["arch", "fedora"]
    wm: "hyprland"
    bar: "ags"
    terminal: "foot"
    shell: "zsh"
    editor: "neovim"
    launcher: "fuzzel"
    notifications: "swaync"
    lockscreen: null
    color_scheme: "custom"
    gtk_theme: null
    icon_theme: null
    has_screenshots: true
    has_install_script: true
    has_readme: true
    screenshot_urls: []
    notable_techniques: ["Material You theming", "Quickshell QML widgets"]
    tags: ["polished", "feature-rich"]
    relevance: "high"
    ---

    Brief description of what makes this dotfiles notable.

    **Stack**: Hyprland + AGS + foot + zsh | Quickshell QML UI, Material You dynamic theming

    **Notable patterns**: Material You color extraction, Quickshell widget framework, comprehensive install script

    **Links**: [GitHub](url) | [Reddit](url)

Field constraints:
- `wm`: hyprland | sway | i3 | bspwm | awesome | dwm | qtile | xmonad | null
- `bar`: eww | waybar | polybar | ags | lemonbar | null
- `terminal`: ghostty | kitty | alacritty | foot | wezterm | st | null
- `shell`: zsh | fish | bash | nushell | null
- `editor`: neovim | vim | emacs | helix | null
- `launcher`: rofi | wofi | fuzzel | dmenu | tofi | null
- `notifications`: mako | dunst | swaync | fnott | null
- `lockscreen`: hyprlock | swaylock | i3lock | null
- `relevance`: high | medium | low (see "My Stack" above for criteria)
- `stars`: integer, 0 means "not checked"
- `last_updated`: repo's last push date (YYYY-MM-DD), not discovery date

## Dedup Protocol

`links/_registry.json` is the single source of truth. Structure:

    {
      "entries": {
        "https://github.com/user/repo": {
          "file": "hyprland/username.md",
          "stars": 1234,
          "discovered_in_round": 1,
          "also_found_in_rounds": [3, 7]
        }
      }
    }

At the START of every round: read `_registry.json`. Before writing any entry, check if the URL exists. If it does, append the current round to `also_found_in_rounds` and skip — do NOT create a duplicate file.

## Classification Rules

Classify each repo by its PRIMARY window manager:
- Full dotfiles with Hyprland as WM -> `hyprland/`
- Full dotfiles with Sway or i3 as WM -> `sway/`
- Standalone bar configs (eww, waybar, AGS repos that aren't full dotfiles) -> `bars/`
- Terminal-specific repos (shader collections, terminal themes) -> `terminals/`
- Color scheme framework ports (catppuccin, nord, dracula repos) -> `themes/`
- NixOS/home-manager declarative configs -> `nixos/`
- Standalone component repos (lockscreen themes, launcher configs, notification themes) -> `components/`
- Awesome-lists and curated collections -> `collections/`

When a repo spans multiple categories, put it in the WM directory and note other components in the body.

## Round Execution Protocol

For EACH round (1-10):

1. **Load**: Read `links/_registry.json`
2. **Search**: Run the round's queries (details below)
3. **Dedup**: Filter against registry
4. **Classify**: Assign category per rules above
5. **Extract**: For each new repo with 10+ stars, read the README via `mcp__github__get_file_contents` (params: `owner`, `repo`, `path: "README.md"`). Budget: max 20 file reads per round. For repos under 10 stars, fill metadata from the search result description only.
6. **Write**: Create entry `.md` files with full frontmatter
7. **Update**: Write `_registry.json` with new entries
8. **Commit**: `git add links/ && git commit -m "links: round N - {domain} ({X} new, {Y} total)"`
9. **Report**: Print a 3-line summary: new repos found, total indexed, any errors

**On error**: If GitHub returns 403/429 (rate limit), wait 30 seconds and retry once. If still failing, switch to WebSearch for that query. If a round completes with partial results, commit what you have and note the gap in the commit message.

**Parallel agents**: You MAY launch up to 3 agents per round to run different search queries simultaneously. Merge their results before the dedup step.

## MCP Tool Reference

GitHub search — sort/filter goes IN the query string using GitHub search syntax:
- `mcp__github__search_repositories(query="hyprland dotfiles sort:stars", perPage=30)`
- `mcp__github__search_code(q="filename:eww.yuck deflisten", per_page=30)`
- `mcp__github__get_file_contents(owner="end-4", repo="dots-hyprland", path="README.md")`

WebFetch — for awesome-list README parsing and Reddit threads.
WebSearch — for Reddit and broader web discovery.

## Round 1: Hyprland Dotfiles

Queries:
1. `mcp__github__search_repositories(query="hyprland dotfiles sort:stars", perPage=30)`
2. `mcp__github__search_repositories(query="hyprland rice sort:stars", perPage=30)`
3. `mcp__github__search_repositories(query="dots-hyprland sort:stars", perPage=30)`
4. WebFetch `https://raw.githubusercontent.com/hyprland-community/awesome-hyprland/main/README.md` — extract all GitHub repo links from "Dotfiles" and "Rices" sections
5. `mcp__github__search_repositories(query="hyprland config eww waybar sort:stars", perPage=30)`

Category: `hyprland/`

## Round 2: Sway & i3

Queries:
1. `mcp__github__search_repositories(query="sway dotfiles sort:stars", perPage=30)`
2. `mcp__github__search_repositories(query="sway rice wayland sort:stars", perPage=30)`
3. `mcp__github__search_repositories(query="i3 dotfiles linux sort:stars", perPage=30)`
4. WebFetch `https://raw.githubusercontent.com/rcalixte/awesome-wayland/master/README.md` — extract dotfiles/rices section links
5. `mcp__github__search_repositories(query="wayland tiling dotfiles sort:stars", perPage=30)`

Category: `sway/`

## Round 3: Status Bars (eww, waybar, AGS)

Queries:
1. `mcp__github__search_repositories(query="eww bar linux sort:stars", perPage=30)`
2. `mcp__github__search_repositories(query="waybar config custom modules sort:stars", perPage=30)`
3. `mcp__github__search_repositories(query="ags hyprland shell sort:stars", perPage=30)`
4. `mcp__github__search_code(q="filename:eww.yuck deflisten defwidget", per_page=30)`
5. `mcp__github__search_repositories(query="eww widgets rice sort:stars", perPage=30)`

Category: `bars/`

## Round 4: Arch & Manjaro Rices

Queries:
1. `mcp__github__search_repositories(query="arch linux dotfiles rice sort:stars", perPage=30)`
2. `mcp__github__search_repositories(query="manjaro dotfiles sort:stars", perPage=30)`
3. `mcp__github__search_repositories(query="arch linux rice 2024 2025 sort:stars", perPage=30)`
4. WebSearch `site:reddit.com/r/unixporn arch linux rice 2024 2025` — extract GitHub links from results
5. `mcp__github__search_repositories(query="archlinux hyprland rice sort:stars", perPage=30)`

Classification: by WM (most go into `hyprland/` or `sway/`). Distro captured in `distro` field only.

## Round 5: NixOS Dotfiles

Queries:
1. `mcp__github__search_repositories(query="nixos dotfiles flake sort:stars", perPage=30)`
2. `mcp__github__search_repositories(query="home-manager dotfiles sort:stars", perPage=30)`
3. `mcp__github__search_repositories(query="nixos hyprland rice sort:stars", perPage=30)`
4. WebFetch `https://raw.githubusercontent.com/nix-community/awesome-nix/main/README.md` — extract dotfiles section
5. `mcp__github__search_repositories(query="nix-darwin dotfiles sort:stars", perPage=30)`

Category: `nixos/` (Nix configs are structurally different — declarative expressions, not imperative dotfiles)

## Round 6: Terminal Configs

Queries:
1. `mcp__github__search_repositories(query="ghostty config shaders sort:stars", perPage=30)`
2. `mcp__github__search_repositories(query="kitty config theme sort:stars", perPage=30)`
3. `mcp__github__search_repositories(query="alacritty config dotfiles sort:stars", perPage=30)`
4. WebFetch `https://raw.githubusercontent.com/fearlessgeekmedia/Awesome-Ghostty/main/README.md` — extract all repo links
5. `mcp__github__search_repositories(query="wezterm config lua sort:stars", perPage=30)`

Category: `terminals/`

## Round 7: Unixporn Hall of Fame

Queries:
1. WebFetch `https://raw.githubusercontent.com/fosslife/awesome-ricing/master/README.md` — extract ALL GitHub links
2. WebFetch `https://raw.githubusercontent.com/zemmsoares/awesome-rices/main/README.md` — extract ALL entries
3. WebSearch `site:reddit.com/r/unixporn "dotfiles" top 2024 2025` — extract GitHub links
4. WebSearch `unixporn best rice 2024 2025 github dotfiles`
5. `mcp__github__search_repositories(query="unixporn dotfiles sort:stars", perPage=30)`

Cross-classify by WM. Create `collections/` entries for the awesome-lists themselves (awesome-ricing, awesome-rices, etc.). High dedup expected.

## Round 8: Theme Frameworks

Queries:
1. `mcp__github__search_repositories(query="catppuccin dotfiles rice sort:stars", perPage=30)`
2. `mcp__github__search_repositories(query="nord theme dotfiles linux sort:stars", perPage=30)`
3. `mcp__github__search_repositories(query="gruvbox dotfiles rice sort:stars", perPage=30)`
4. `mcp__github__search_repositories(query="dracula theme dotfiles linux sort:stars", perPage=30)`
5. `mcp__github__search_repositories(query="tokyo-night rose-pine dotfiles sort:stars", perPage=30)`

Theme framework repos go to `themes/`. Full dotfiles rices go to their WM directory with `color_scheme` field set.

## Round 9: Components

Queries:
1. `mcp__github__search_repositories(query="swaylock lockscreen config theme sort:stars", perPage=30)`
2. `mcp__github__search_repositories(query="rofi theme rice sort:stars", perPage=30)`
3. `mcp__github__search_repositories(query="wofi fuzzel launcher config sort:stars", perPage=30)`
4. `mcp__github__search_repositories(query="mako dunst swaync config sort:stars", perPage=30)`
5. `mcp__github__search_repositories(query="wlogout swayidle config sort:stars", perPage=30)`

Category: `components/`

## Round 10: Consolidation

No new searches. Instead:

1. Read every `.md` in `links/*/`. Verify frontmatter is valid. Fix malformed entries.
2. Verify `_registry.json` entry count matches file count on disk. Report discrepancies.
3. Generate `links/INDEX.md` with:
   - Stats: total repos, count per category, count per WM, count per bar, count per terminal
   - Full markdown table: Name | Author | Stars | WM | Bar | Terminal | Color Scheme | Relevance
   - "Priority Reference" section listing all `relevance: high` repos
   - "Most Starred" section with top 20 by stars
4. Generate `links/SCHEMA.md` documenting entry format and directory structure
5. Final commit: `git add links/ && git commit -m "links: round 10 - INDEX.md, SCHEMA.md ({N} total repos)"`

## Completion

Print final summary:
- Total unique repos indexed
- Breakdown by category
- Top 10 most-starred
- Top 10 highest-relevance
- Rounds with issues (if any)
- Suggestion for which repos are worth forking into dotfile-museum

## Important

- Working directory: `~/hairglasses-studio/dotfiles/`
- Git: hairglasses / mitch@hairglasses.studio, no GPG signing
- Stars threshold: 10+ for full metadata extraction, under 10 only if notably unique
- For repos already in `_registry.json` from a prior round, just update `also_found_in_rounds` — do NOT re-read their files
- If a search returns fewer than 10 results, try alternate query terms before moving on
- Prefer `mcp__github__get_file_contents` over WebFetch for reading individual repo READMEs
- `_registry.json` is your crash recovery — always keep it current before committing
