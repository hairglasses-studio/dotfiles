# Getting Started with GitHub

This guide walks you through setting up Git on your Mac and cloning the luke-toolkit repository.

---

## Part 1: Install Git

### Open Terminal

1. Press `Cmd + Space` to open Spotlight
2. Type `Terminal` and press Enter
3. A window with a command line appears

### Install Xcode Command Line Tools

This installs Git and other developer tools:

```bash
xcode-select --install
```

A popup will appear. Click **Install** and wait (takes 5-10 minutes).

### Verify Git Installed

```bash
git --version
```

You should see something like `git version 2.39.0`. If you see an error, the install didn't complete - try running `xcode-select --install` again.

---

## Part 2: Configure Git

Tell Git who you are (used for commit history):

```bash
git config --global user.name "lukelasleyfilm"
git config --global user.email "lukelasleyfilm@gmail.com"
```

Verify it worked:

```bash
git config --list
```

You should see your name and email in the output.

---

## Part 3: Set Up GitHub Authentication

### Create SSH Key

This lets you push/pull without entering passwords:

```bash
ssh-keygen -t ed25519 -C "lukelasleyfilm@gmail.com"
```

When prompted:
- **File location**: Press Enter (use default)
- **Passphrase**: Press Enter twice (no passphrase, or set one if you prefer)

### Start SSH Agent

```bash
eval "$(ssh-agent -s)"
```

### Add Key to Agent

```bash
ssh-add ~/.ssh/id_ed25519
```

### Copy Public Key

```bash
cat ~/.ssh/id_ed25519.pub | pbcopy
```

This copies your public key to the clipboard.

### Add Key to GitHub

1. Go to [github.com](https://github.com) and sign in
2. Click your profile picture (top right) → **Settings**
3. Click **SSH and GPG keys** (left sidebar)
4. Click **New SSH key**
5. Title: `Luke's MacBook` (or similar)
6. Key: Press `Cmd + V` to paste
7. Click **Add SSH key**

### Test Connection

```bash
ssh -T git@github.com
```

Type `yes` if prompted. You should see:
```
Hi username! You've successfully authenticated...
```

---

## Part 4: Clone Luke's Toolkit

### Create Project Folder

```bash
mkdir -p ~/hairglasses-studio && cd ~/hairglasses-studio
```

### Clone the Repository

```bash
git clone git@github.com:hairglasses-studio/luke-toolkit.git
```

### Enter the Folder

```bash
cd luke-toolkit
```

### Verify It Worked

```bash
ls -la
```

You should see the README.md, tutorials folder, and other files.

---

## Part 5: Basic Git Workflow

Now that you have the repo, here's how to work with it.

### Check Status

See what files have changed:

```bash
git status
```

Colors:
- **Red**: Changed but not staged
- **Green**: Ready to commit
- **Nothing to commit**: Everything saved

### Pull Latest Changes

Before starting work, get the latest updates:

```bash
git pull
```

### Make Changes and Save

After editing files:

```bash
git add .
git commit -m "Description of what I changed"
git push
```

---

## Key Concepts

| Term | What it means |
|------|---------------|
| **Repository (repo)** | A project folder tracked by Git |
| **Clone** | Download a copy to your computer |
| **Commit** | Save a snapshot of changes |
| **Push** | Upload your commits to GitHub |
| **Pull** | Download others' changes |
| **Branch** | Work on something without affecting main |

---

## Quick Reference

### Daily Workflow

```bash
cd ~/hairglasses-studio/luke-toolkit    # Go to repo
git pull                           # Get latest changes
# ... make your edits ...
git status                         # See what changed
git add .                          # Stage all changes
git commit -m "What I did"         # Save with message
git push                           # Upload to GitHub
```

### See What Changed

```bash
git diff                           # Show file changes
git log --oneline                  # Show commit history
```

### Undo Changes

```bash
git checkout -- filename           # Undo changes to one file
git reset --hard HEAD              # Undo ALL changes (careful!)
```

### Working on Features

```bash
git checkout -b my-feature         # Create new branch
# ... make changes ...
git add . && git commit -m "Add feature"
git push -u origin my-feature      # Push branch to GitHub
```

---

## Cheat Sheet

| What you want | Command |
|---------------|---------|
| Go to repo | `cd ~/hairglasses-studio/luke-toolkit` |
| Get latest | `git pull` |
| Check status | `git status` |
| Save changes | `git add . && git commit -m "message"` |
| Upload | `git push` |
| See changes | `git diff` |
| See history | `git log --oneline` |
| New branch | `git checkout -b name` |
| Switch branch | `git checkout name` |

---

## Troubleshooting

### "Permission denied (publickey)"

Your SSH key isn't set up correctly. Go back to Part 3.

### "Repository not found"

Check the URL is correct:
```bash
git remote -v
```

Should show `git@github.com:hairglasses-studio/luke-toolkit.git`

### "Please tell me who you are"

Run the config commands from Part 2:
```bash
git config --global user.name "lukelasleyfilm"
git config --global user.email "lukelasleyfilm@gmail.com"
```

### "Failed to push - rejected"

Someone else pushed changes. Pull first, then push:
```bash
git pull
git push
```

---

## Next Steps

- [Claude Code Introduction](02-claude-code.md) - Let AI help you code
- [YouTube Resources](13-youtube-resources.md) - Video tutorials on Git
