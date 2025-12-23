# PROJECT_NOTES.md

## 1. Context & Background
The `aftrs-shell` repository is the umbrella project for managing, installing, and keeping up-to-date a curated set of shell-based tools and plugins.  
All tools are forked into the [`aftrs-shell`](https://github.com/aftrs-shell) GitHub organization, with the goal of:
- Installing and configuring them together via a single installer script.
- Keeping forks in sync with their upstream sources automatically.
- Providing a consistent developer environment with autocomplete, theming, and helper utilities.

This repo also serves as the home for the `aftrs` CLI, which orchestrates updates and manages plugin completions.

---

## 2. Current Progress
- **Repository Setup**
  - Initialized `aftrs-shell/aftrs-shell` as the main aggregation point.
  - Enhanced `install.sh` to:
    - Clone/update org repos into `~/Docs` by default (`AFTRS_SHELL_HOME` override).
    - Scan local tools under `/home/hg/Docs/aftrs-shell` and run their `install.sh` when present.
    - Aggregate zsh/bash/fish completions and wire them into shell startup.
    - Seed configs from `aftrs_init/config` via `aftrs_initial_unpack.sh`.
    - Install and enable an `oh-my-posh` powerline theme.
    - Optional pretty logs with `gum` (auto-disabled on host "wizardstower").
  - Added `aftrs_initial_unpack.sh` and initial theme seed at `aftrs_init/config/oh-my-posh/aftrs-powerline.omp.json`.
  - Added tests: `test/install_test.sh` and `test/install_features_test.sh` (dry-run, unpack, flags).
  - Normalized line endings with `.gitattributes` and ensured scripts are executable.

- **Fork Syncing**
  - Reusable GitHub Actions workflow (`sync_fork_template.yml`) to pull from upstream forks on a schedule.

- **Documentation**
  - Updated `README.md` to document flags, environment variables, completions, theme, and config seeding.
  - Added generator for an all-repos workspace under `~/Docs`.

---

## 3. Pending / In-Flight Work
- Add excludes/throttling to completions scan to avoid very large trees (e.g., `zsh-more-completions`).
- Optional: wrap installer features into an `aftrs` CLI shim with subcommands.
- Extend CI to smoke-test completion generation paths (zellij, etc.).
- Ensure fork-sync workflow is enabled in forks.
- Bootstrap profiles (minimal/full) and idempotent re-runs.

---

## 4. Future Roadmap & Ideas
- **CLI Enhancements**
  - Support multiple orgs in a single update run.
  - Add `--include-sources` flag to clone non-fork repos.
  - Add `aftrs doctor` command for environment diagnostics.
  - Support interactive mode for selectively updating tools.
  - Add plugin registry config for curated optional installs.

- **Automation**
  - Twice-daily upstream sync for all forks.
  - Auto-pull latest completions during `aftrs update`.

- **Developer Experience**
  - Add tab-completion for the `aftrs` CLI itself.
  - Add theming and shell prompt integration.

---

## 5. Repo Integration Notes
- **Install**
  ```bash
  chmod +x ./aftrs
  sudo cp ./aftrs /usr/local/bin/aftrs
  ```

  - WSL/Cursor terminal working directory: workspace now sets the integrated terminal to start in `${env:HOME}/Docs`. Optionally add a bash fallback snippet to `~/.bashrc` to auto-`cd` to `~/Docs` when the terminal starts inside a Windows `AppData` path under `/mnt/c/Users/.../AppData/...`.
  - `scripts/generate_docs_workspace.sh` creates `~/Docs/Docs.code-workspace` including every `.git` repo under `~/Docs`.
