# Committing Existing Projects to GitHub

Upload your existing work (TouchDesigner projects, scripts, plugins) to a new GitHub repository.

## Overview

This guide covers:
1. Preparing your project folder
2. Creating a repository on GitHub
3. Connecting your local folder to GitHub
4. Pushing your files

---

## Part 1: Prepare Your Project

### Clean Up First

Before committing, remove unnecessary files:

```bash
# Navigate to your project
cd ~/path/to/your/project

# See what's there
ls -la
```

### Files to Keep

- Source code (`.py`, `.tox`, `.toe`, `.js`, etc.)
- Configuration files
- Documentation
- Assets (within reason)

### Files to Exclude

- Backup files (`*.bak`, `*.toe.*`)
- Cache folders (`__pycache__/`, `Backup/`)
- System files (`.DS_Store`)
- Large media files (videos, large audio)
- Personal settings

### Create .gitignore

Create a `.gitignore` file to exclude unwanted files:

```bash
# Create the file
touch .gitignore
```

For **TouchDesigner** projects:

```gitignore
# TouchDesigner
*.toe.bak
*.toe.*.toe
Backup/

# Python
__pycache__/
*.pyc

# OS
.DS_Store
Thumbs.db

# Media (if large)
*.mov
*.mp4
*.wav
```

---

## Part 2: Create Repository on GitHub

### Using GitHub Website

1. Go to [github.com](https://github.com)
2. Click the **+** in the top right
3. Select **New repository**
4. Fill in:
   - **Repository name**: `my-td-plugins` (or similar)
   - **Description**: What this project does
   - **Visibility**: Public or Private
   - **DON'T** check "Add a README" (we'll push our own)
5. Click **Create repository**

### Using Claude Code

```
"Create a new GitHub repository called 'luke-td-plugins' with description 'TouchDesigner visual plugins'"
```

---

## Part 3: Initialize Git Locally

### Navigate to Your Project

```bash
cd ~/path/to/your/td-plugins
```

### Initialize Git

```bash
# Initialize the repository
git init

# See the status
git status
```

### Create README

Every project needs a README:

```bash
echo "# My TD Plugins\n\nTouchDesigner visual effects and components." > README.md
```

---

## Part 4: First Commit

### Stage All Files

```bash
# Add everything (respects .gitignore)
git add .

# Check what's staged
git status
```

### Create Initial Commit

```bash
git commit -m "Initial commit: add TD plugins"
```

---

## Part 5: Connect to GitHub

### Add Remote

Replace `username` with your GitHub username and `repo-name` with your repository name:

```bash
git remote add origin https://github.com/username/repo-name.git
```

For the AFTRS studio:

```bash
git remote add origin https://github.com/hairglasses/luke-td-plugins.git
```

### Push to GitHub

```bash
# Set main as default branch and push
git branch -M main
git push -u origin main
```

---

## Complete One-Liner

Here's the entire process as a single command sequence:

```bash
cd ~/your-project && git init && echo "# Project Name\n\nDescription here." > README.md && touch .gitignore && git add . && git commit -m "Initial commit" && git remote add origin https://github.com/hairglasses/repo-name.git && git branch -M main && git push -u origin main
```

---

## Using Claude Code

### The Easy Way

Navigate to your project and ask Claude:

```
"Help me commit this TouchDesigner project to a new GitHub repo called 'luke-td-plugins'"
```

Claude will:
1. Check your files
2. Create appropriate `.gitignore`
3. Initialize git
4. Create README
5. Create the GitHub repo
6. Push everything

---

## Example: TD Plugins to AFTRS

### Step 1: Prepare

```bash
cd ~/Documents/TouchDesigner/MyPlugins
ls -la
```

### Step 2: Create .gitignore

```bash
cat > .gitignore << 'EOF'
*.toe.bak
*.toe.*.toe
Backup/
__pycache__/
*.pyc
.DS_Store
EOF
```

### Step 3: Create README

```bash
cat > README.md << 'EOF'
# Luke's TD Plugins

TouchDesigner visual effects and components for live performance.

## Contents

- `visual-noise/` - Noise-based generators
- `audio-reactive/` - Sound-responsive effects
- `utilities/` - Helper components

## Usage

Drag `.tox` files into your TouchDesigner project.

## Author

Luke @ AFTRS
EOF
```

### Step 4: Initialize and Push

```bash
git init && git add . && git commit -m "Initial commit: TD plugins collection" && git remote add origin https://github.com/hairglasses/luke-td-plugins.git && git branch -M main && git push -u origin main
```

---

## TouchDesigner-Specific Tips

### What to Commit

| Include | Exclude |
|---------|---------|
| `.toe` project files | `.toe.bak` backup files |
| `.tox` components | `Backup/` folder |
| Python scripts `.py` | Large media files |
| Documentation | Personal paths in settings |
| Textures (if small) | Generated cache |

### Organize Before Committing

```
my-td-plugins/
├── README.md
├── .gitignore
├── components/
│   ├── visual-noise.tox
│   └── audio-reactive.tox
├── effects/
│   └── glitch-effect.tox
├── scripts/
│   └── osc-handler.py
└── docs/
    └── usage.md
```

### Large Files Warning

If you have large files (>100MB), GitHub will reject them. Either:
- Remove them from the commit
- Use [Git LFS](https://git-lfs.github.com/) for large files
- Store media elsewhere (Google Drive, etc.)

---

## Adding More Files Later

After the initial commit, adding new files is easy:

```bash
# Navigate to repo
cd ~/path/to/repo

# Add new files
git add new-effect.tox

# Commit
git commit -m "Add new-effect.tox"

# Push
git push
```

---

## Updating Existing Files

When you modify files:

```bash
# See what changed
git status

# Add changes
git add .

# Commit with description
git commit -m "Update visual-noise with new parameters"

# Push
git push
```

---

## Quick Reference

| Task | Command |
|------|---------|
| Initialize repo | `git init` |
| Add all files | `git add .` |
| Create commit | `git commit -m "message"` |
| Add remote | `git remote add origin URL` |
| Push to GitHub | `git push -u origin main` |
| Check status | `git status` |
| View history | `git log --oneline` |

---

## Troubleshooting

### "Permission denied"

```bash
# Check if SSH is set up
ssh -T git@github.com

# Or use HTTPS instead of SSH
git remote set-url origin https://github.com/username/repo.git
```

### "Repository not found"

- Check the URL is correct
- Make sure the repo exists on GitHub
- Verify you have access (if private)

### "File too large"

```bash
# Remove large file from staging
git reset HEAD large-file.mov

# Add to .gitignore
echo "large-file.mov" >> .gitignore
```

### "Updates were rejected"

```bash
# Pull first, then push
git pull origin main --rebase
git push
```

---

## Next Steps

- [Contributing with Claude](12-contributing-with-claude.md) - Continue developing with Claude
- [GitHub Workflows](07-github-workflows.md) - Automate testing
- [Project Templates](11-project-templates.md) - Best practices
