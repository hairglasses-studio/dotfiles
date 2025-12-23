# aftrs-lighting

### 🔁 Enable Git Hooks (auto-render diagrams on commit)
```bash
# Linux/macOS
./scripts/setup_hooks.sh
# Windows PowerShell
./scripts/setup_hooks.ps1
```

#### Linux pre-commit hook
A bash-based pre-commit hook is installed when you run `./scripts/setup_hooks.sh`. It detects Docker/Podman/mmdc and only re-renders diagrams that changed (or are stale). If no renderer is found, the commit proceeds with a warning.

#### Pre-push guard (Linux/macOS)
A **pre-push** hook blocks pushes if Mermaid sources changed but `diagrams/_rendered/*.svg` are stale.
It auto-renders via Docker/Podman/mmdc; if outputs change, the push fails with instructions to commit the updated SVGs.

> Hook path: run `./scripts/setup_hooks.sh` (or `./scripts/setup_hooks.ps1` on Windows) to enable `.githooks/`.

#### CI enforcement
A GitHub Action **Verify Mermaid Diagrams Freshness** runs on every pull request and **fails** if `diagrams/_rendered/*.svg` are stale relative to `.mmd` sources. It uses `@mermaid-js/mermaid-cli` to render in CI.

#### Update CI badges
After pushing to GitHub and setting a remote, run:
```bash
./scripts/update_badges_from_git_remote.sh
# or on Windows
./scripts/update_badges_from_git_remote.ps1
```
This replaces `OWNER/REPO` in badge URLs with your actual repository slug.

### 🚀 Bootstrap
One-time setup after cloning:
```bash
# Linux/macOS
./scripts/bootstrap.sh
```
```powershell
# Windows PowerShell
./scripts/bootstrap.ps1
```
This will: enable git hooks, update README badges based on your git remote, and render all Mermaid diagrams.

## 🧭 Getting Started

### Requirements
- **Git**
- **One of:** Docker (Desktop) • Podman • Node `@mermaid-js/mermaid-cli`
- **Optional:** Python 3 + `pyyaml` (for `tools/validate_patch.py`)

### Quick Start
**Linux/macOS**
```bash
git clone <your-repo-url>
cd aftrs-lighting
./scripts/bootstrap.sh        # enables hooks, patches badges, renders diagrams
# or
make bootstrap
```
**Windows**
```powershell
git clone <your-repo-url>
cd aftrs-lighting
.\scripts\bootstrap.ps1     # or just run .\setup.bat
```

### Everyday Commands
```bash
make diagrams      # re-render all Mermaid diagrams
make validate      # run DMX patch validator (Python + PyYAML)
# enable hooks manually (already included in bootstrap):
./scripts/setup_hooks.sh
```

### Hooks & CI at a Glance
- **pre-commit**: renders changed/stale `.mmd` → updates `diagrams/_rendered/*.svg` → auto-stages them.
- **pre-push**: blocks push if `.mmd` changed and SVGs are stale or no renderer is available.
- **CI (PRs)**: verifies diagram freshness (`verify-diagrams.yml`) and fails PR if stale.
- **CI (push)**: renders and commits SVGs on `main` (fork-safe; no commit on PRs).

### Badge Update
After you set the Git remote and push once, fix README badges:
```bash
./scripts/update_badges_from_git_remote.sh
# Windows:
./scripts/update_badges_from_git_remote.ps1
```

### ✅ Pre-show sanity check
Run a one-shot health check before a session:
```bash
make dev-check          # Linux/macOS
# or
./scripts/dev-check.sh
```
```powershell
.\scripts\dev-check.ps1  # Windows
```
This validates DMX patching, verifies Mermaid freshness, and pings Art-Net node IPs from `configs/artnet_plan.yaml`.
Exit codes: 0 (OK), 1 (warnings), 2 (errors).
