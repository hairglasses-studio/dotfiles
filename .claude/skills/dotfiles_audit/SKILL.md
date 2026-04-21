---
name: dotfiles_audit
description: Run a comprehensive dotfiles health and stale-reference audit. Use when cleaning up after removing tools/terminals/bars, after palette changes, after MCP consolidation, or whenever you want to verify the repo is internally consistent. Also use proactively before commits that touch many files, after major refactors, or when the user says "audit", "check for stale", "verify cleanup", "anything left?", or "is the repo clean?".
---

# Dotfiles Audit

Run a comprehensive audit of the dotfiles repo to find broken, stale, or inconsistent references. Report findings organized by severity.

## When to use

- After removing a tool, terminal, bar, or palette (ghostty, eww, foot, Snazzy, etc.)
- After MCP server consolidation or tool handler changes
- After palette unification or rename
- Before pushing a large cleanup commit
- When the user asks if the repo is clean or consistent

## Audit procedure

Run all checks in parallel where possible. Organize results into three severity tiers.

### Tier 1 — BROKEN (build failures, dead code paths, missing files)

1. **Go build**: `cd mcp/dotfiles-mcp && go build ./... && go vet ./...`
2. **Shell syntax**: `bash -n scripts/*.sh scripts/lib/*.sh`
3. **Chezmoi verify**: `chezmoi verify`
4. **Contract drift**: `cd mcp/dotfiles-mcp && make contract-check` (if target exists, otherwise skip)
5. **Systemd failures**: `systemctl --user --failed` and `systemctl --system --failed`
6. **Broken symlinks**: `find ~/.config -xtype l 2>/dev/null` (filter browser/app noise)
7. **Dead resource handlers**: grep Go code for resource URIs referencing deleted directories
8. **Orphan systemd units from recent removal commits** — catches the "dotfiles removed config, package + systemd unit still enabled" class. Reference case: makima.service restart-looped 777+ times on 2026-04-16 after commit `664d0f3` dropped `~/.config/makima/` without disabling the system-level unit. For each `.service` / `.socket` / `.timer` / `.rules` / `.desktop` file removed in the last 14 days of commits, verify it's not still live on the running system:
   ```bash
   cd "$(git rev-parse --show-toplevel)"
   removed=$(git log --since='14 days ago' --diff-filter=D --name-only --pretty=format: \
     | grep -E '\.(service|socket|timer|rules|desktop)$' \
     | awk -F/ '{print $NF}' | sed 's/\.[^.]*$//' | sort -u)
   for name in $removed; do
     # is the system unit still active/enabled/unmasked?
     if systemctl cat "$name.service" >/dev/null 2>&1; then
       state=$(systemctl is-enabled "$name.service" 2>/dev/null)
       case "$state" in
         masked|not-found) : ;;  # fine
         *) echo "ORPHAN: $name.service state=$state — mask or remove";;
       esac
     fi
     # same for user units
     if sudo -u hg systemctl --user cat "$name.service" >/dev/null 2>&1; then
       state=$(sudo -u hg systemctl --user is-enabled "$name.service" 2>/dev/null)
       case "$state" in masked|not-found) : ;; *) echo "ORPHAN (user): $name state=$state";; esac
     fi
     # quick restart-loop heuristic — any unit whose restart-counter climbed recently
     journalctl -u "$name.service" --since '1 hour ago' --no-pager 2>/dev/null \
       | grep -c 'restart counter is at' | awk -v n="$name" '$1>20 {print "RESTART-LOOP: " n "=" $1 " restarts in last hour"}'
   done
   ```
   For each ORPHAN: `sudo systemctl disable --now <unit>` then `sudo systemctl mask <unit>` (masking is durable across updates; disable-alone can be re-enabled by a package update). If the whole package is obsolete, `pacman -Rns <pkg>` instead. **Document the decision in the removal commit message** (leaving a trail for audits of audits).

### Tier 2 — STALE (wrong names, old hex values, phantom tool references)

Run this grep across the entire repo (excluding .git, vendor, go.sum, .well-known, snapshots, CHANGELOG):

```bash
grep -rn \
  "ghostty\|juhradial\|eww\|p10k\|powerlevel10k\|Snazzy\|Voltage After Dark\|06070a\|57c7ff\|ff6ac1\|6ae4ff\|ff6ad5\|foot-server\|sway-mcp\|input-mcp\|shader-mcp\|process-mcp\|systemd-mcp\|tmux-mcp" \
  --include="*.sh" --include="*.go" --include="*.md" --include="*.yml" \
  --include="*.toml" --include="*.yaml" --include="*.css" --include="*.conf" \
  --include="*.json" . 2>/dev/null \
  | grep -v ".git/\|node_modules\|CHANGELOG\|vendor/\|go.sum\|\.well-known\|snapshots/"
```

For each hit, classify as:
- **Intentional** — shader-transpile uniform renames, functional color values in active theme files (glow, k9s, lazygit, tmux, hyprlock, wallpaper scripts, starship palette), historical changelog entries
- **Stale** — comments, descriptions, skill allowed-tools, agent instructions, feature flags, CI path filters, docs

The stale terms list should be updated whenever a component is removed. Current removed components: ghostty, juhradial (as standalone), eww, foot, p10k/powerlevel10k, CRTty, sway-mcp, input-mcp, shader-mcp, process-mcp, systemd-mcp, tmux-mcp (as standalone repos).

Also check skills and rules specifically:
```bash
grep -rn "ghostty\|eww\|Snazzy\|foot-server" .claude/skills/ .claude/rules/ .claude/agents/ .agents/skills/
```

### Tier 3 — INCOMPLETE (inaccurate docs, missing entries, wrong counts)

1. **Shader count**: `find kitty/shaders/darkwindow -name '*.glsl' | wc -l` vs README badge value
2. **MCP tool count**: compare CLAUDE.md header count vs `grep -c 'TypedHandler\[' mcp/dotfiles-mcp/mod_*.go mcp/dotfiles-mcp/main.go | awk -F: '{sum+=$2} END{print sum}'`
3. **Chezmoi coverage**: compare `chezmoi managed | wc -l` vs `bash install.sh --print-link-specs | wc -l` — gaps mean new files aren't in chezmoi
4. **README sections**: verify kanshi, glshell, wluma appear in directory table if those dirs exist
5. **dotfiles.toml**: check for `= false` flags referencing removed components (vestigial noise)

## Output format

```
## BROKEN (N items)
1. [description] — [file:line] — [fix needed]

## STALE (N items)  
1. [description] — [file:line] — [classification: stale comment / phantom tool / old hex]

## INCOMPLETE (N items)
1. [description] — [expected vs actual] — [fix needed]

## CLEAN
- [list of checks that passed]

## Intentional Exceptions
- [list of hits classified as intentional, with rationale]
```

After reporting, offer to fix all items automatically. Group fixes by phase (Go code changes → script changes → doc changes → config changes) and commit each phase separately.
