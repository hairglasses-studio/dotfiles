<!--
  AFTRS‑Shell meta repository
  ===========================

  This repository acts as an umbrella for a collection of shell
  plugins and tools maintained in the `aftrs-shell` GitHub
  organisation.  Use the included scripts to clone and update
  all of the individual projects, and to keep your forks in sync
  with their upstream sources.
-->

# AFTRS‑Shell

AFTRS‑Shell aims to provide a unified way to install a set of
shell utilities and plugins forked from various upstream
projects.  Each tool lives in its own repository under the
[`aftrs-shell`](https://github.com/aftrs-shell) organisation.
This meta‑repository contains automation to install those
projects locally and to keep them up to date.

## Prerequisites

- [GitHub CLI `gh`](https://cli.github.com/) must be installed
  and authenticated.  Use `gh auth login` to sign into GitHub.
- Optional: tools such as `jq` may be useful for advanced
  scripting, but the provided `install.sh` does not require them.

## Installing all tools

Run the provided `install.sh` script to:

- Clone/update every repository in the organisation (except this meta‑repo) into `~/Docs` by default
  (override with `AFTRS_SHELL_HOME`).
- Scan your local tools at `/home/hg/Docs/aftrs-shell` (override with `AFTRS_LOCAL_ROOT`) and run their
  `install.sh` if present.
- Aggregate autocompletions (zsh/bash/fish) and wire them into your shell startup.
- Install and enable an `oh-my-posh` powerline theme.

For example:

```bash
git clone https://github.com/aftrs-shell/aftrs-shell.git
cd aftrs-shell
./install.sh
```

To update later, simply rerun the script.

### Flags

- `--skip-clone`: skip org cloning and only operate on `AFTRS_LOCAL_ROOT`
- `--no-install`: do not run per-repo `install.sh` scripts
- `--no-completions`: skip completion aggregation
- `-n, --dry-run`: show what would be done

Environment variables:

- `AFTRS_SHELL_HOME`: destination for cloned repos (default: `~/Docs`)
- `AFTRS_LOCAL_ROOT`: local tools root to scan (default: `/home/hg/Docs/aftrs-shell`)
- `AFTRS_COMPLETIONS_DIR`: base dir for collected completions (default: `~/.config/aftrs/completions`)

Optional pretty output uses `gum` if available, and is automatically disabled on host "wizardstower".

## Completions and Prompt

- Completions are collected into `~/.config/aftrs/completions` and linked to:
  - zsh: `~/.local/share/zsh/site-functions` (with `fpath` automatically updated)
  - bash: `~/.local/share/bash-completion/completions` (with `bash-completion` sourced from `~/.bashrc`)
  - fish: `~/.config/fish/completions`
- A powerline-style `oh-my-posh` theme is installed to `~/.config/oh-my-posh/aftrs-powerline.omp.json`
  and enabled for bash and zsh. Use a Nerd Font (e.g., JetBrainsMono Nerd Font) for best results.

To switch theme, point the init line in your `~/.zshrc`/`~/.bashrc` at a different `*.omp.json`.

## Workspace covering all repos in `~/Docs`

- Generate a multi-root workspace that includes every `.git` repo under `~/Docs`:

  ```bash
  ./scripts/generate_docs_workspace.sh
  ```

  This writes `~/Docs/Docs.code-workspace`. Open it in Cursor/VS Code to get all current repos in one workspace. Re-run the script anytime you add/remove repos.

## Config seed/unpack

- Config templates live under `aftrs_init/config` inside this repo.
- The script `./aftrs_initial_unpack.sh` (called by `install.sh`) copies them into `~/.config/...`.
- You can safely customize the files in `~/.config`; re-running unpack will overwrite them, so back up if needed.

## Keeping forks in sync

For forked repositories that track upstream projects, you can
enable automatic synchronisation using a GitHub Actions workflow.
The file [`sync_fork_template.yml`](./sync_fork_template.yml)
provides a reusable example.  Copy it into the `.github/workflows`
directory of each fork and replace the `{{UPSTREAM_REPOSITORY}}`,
`{{UPSTREAM_BRANCH}}` and `{{TARGET_BRANCH}}` placeholders with
the appropriate values.  The workflow will run twice daily and
attempt to merge upstream changes into your fork.

For example, in a fork of [`zshzoo/cd-ls`](https://github.com/zshzoo/cd-ls)
you would set:

```yaml
upstream_repository: "zshzoo/cd-ls"
upstream_branch: "main"
target_branch: "main"
```

This pattern can be repeated for each fork.  Alternatively,
consider using a dedicated fork‑sync action such as
[`aormsby/fork-sync-with-upstream-action`](https://github.com/marketplace/actions/fork-sync-with-upstream-action),
which encapsulates this logic into a single step.

## Why not submodules?

The individual tools are maintained as separate repositories to
allow independent versioning and contribute changes back to
upstream projects.  Submodules tightly couple the meta‑repo to
specific commits of each dependency and require more manual
maintenance.  Using a script to clone and update tools provides
flexibility and ensures you always have the latest versions.

### Workspace
- Workspace now includes all Git repos under /home/hg/Docs/aftrs-shell as VS Code folders.

## Forked Repositories Table (2025-09-14)

The following table lists all currently forked public repos in the `aftrs-shell` GitHub org. These will be archived to avoid confusion with internal codebase and documentation.

| Forked Repo Name |
|-----------------|
| FTXUI |
| superfile |
| lnav |
| bars-and-widgets-list |
| wttr.in |
| tldr |
| nixpkgs |
| shellcheck |
| tinted-shell |
| cowsay.zsh |
| fzf |
| newsboat |
| helix |
| tmuxinator |
| zsh-misc-completions |
| fx |
| trippy |
| starship |
| streamlink |
| conda-init-zsh-plugin |
| textual |
| ratatui |
| zellij |
| zsh-sshinfo |
| mpv |
| dotfiles.github.com |
| ohmyzsh |
| z.lua |
| genact |
| forgit |
| atuin |
| awesome-zsh-plugins |
| oh-my-posh |
| lf |
| jumper |
| zsh-banner |
| wtf |
| bubbletea |
| yazi |
| prezto |
| ez-compinit |
| zsh-expand |
| btop |
| gitui |
| wiki |
| nnn |
| ripgrep |
| redo |
| awesome-sysadmin |
| zsh-completions |
| robo |
| zsh-command-time |
| irssi |
| navi |
| apparix |
| selector |
| broot |
| rich |
| dpd-plugin |
| zsh-more-completions |
| goaccess |
| dockly |
| jq-zsh-plugin |
| zinit |
| visidata |
| posting |
| zsh-system-update |
| eza |
| zsh-colorize-functions |
| asciinema |
| zsh-abbr |
| zsh-diff-so-fancy |
| awesome |
| fast-syntax-highlighting |
| git-extra-commands |
| basher |
| codex-zsh-plugin |
| NerdFetch |
| fzf-tab |
| zsh-you-should-use |
| zsh-allclear |
| pterm |
| zsh-autocomplete |
| awesome-console-services |
| shtab |
| YouPlot |
| s3cmd |
| zsh-env-secrets |
| zsh-ai-completions |
| zcolors |
| oh-my-zsh-git-plugin-cheatsheet |
| zoxide |
| yt-dlp-omz-plugin |
| zsh-ssh |
| shell-fns |

**Consolidation Plan:**
- Archive all forked public repos in `aftrs-shell` using GitHub CLI, keeping only originals.
- Move unique internal repos from `aftrs-shell` to `aftrs-void` for future development and documentation.
- Update aftrs-wiki and push changes to preserve findings and prevent confusion between public forks and internal codebase.
