# Collaborative Git Workflow

How to work together on code projects in the hairglasses-studio organization.

---

## The Collaboration Model

When working on shared projects:

1. **Main branch** is always stable/working
2. **Feature branches** for new work
3. **Pull requests** for code review
4. **Merge** after approval

```
main ─────●─────●─────●─────●─────●───
           \         /     \     /
feature-a   ●───●───●       \   /
                             \ /
feature-b           ●───●─────●
```

---

## Part 1: Starting Work on a Shared Project

### Clone the Repository

```bash
cd ~/hairglasses-studio
git clone git@github.com:hairglasses-studio/project-name.git
cd project-name
```

### Always Pull Before Starting

```bash
git checkout main
git pull
```

### Create a Feature Branch

Never work directly on `main`. Create a branch:

```bash
git checkout -b luke/add-audio-reactive
```

Branch naming conventions:
- `luke/feature-name` - Your features
- `fix/bug-description` - Bug fixes
- `experiment/idea-name` - Experiments

---

## Part 2: Working on Your Branch

### Make Changes and Commit Often

```bash
# Check what's changed
git status

# Stage and commit
git add .
git commit -m "Add audio input component"

# Keep committing as you work
git add .
git commit -m "Add FFT analysis to audio input"
```

### Good Commit Messages

**Do:**
```
Add audio reactive component with FFT analysis
Fix null error when no audio device connected
Update OSC port to match TD settings
```

**Don't:**
```
stuff
fixed it
WIP
asdfasdf
```

### Push Your Branch

```bash
git push -u origin luke/add-audio-reactive
```

The `-u` flag links your local branch to the remote. After that, just `git push`.

---

## Part 3: Staying in Sync

### Update Your Branch with Main

While working, main may have changed. Stay in sync:

```bash
# Get latest changes from remote
git fetch origin

# Merge main into your branch
git merge origin/main
```

Or use rebase for cleaner history:

```bash
git fetch origin
git rebase origin/main
```

### Resolve Conflicts

If Git shows conflicts:

1. Open the conflicted files
2. Look for conflict markers:
   ```
   <<<<<<< HEAD
   your changes
   =======
   their changes
   >>>>>>> origin/main
   ```
3. Edit to keep what you want
4. Remove the markers
5. Stage and commit:
   ```bash
   git add .
   git commit -m "Resolve merge conflicts"
   ```

---

## Part 4: Pull Requests

### Create a Pull Request

When your feature is ready:

```bash
# Push latest changes
git push

# Create PR via GitHub CLI
gh pr create --title "Add audio reactive component" --body "$(cat <<'EOF'
## Summary
- Added audio input component with FFT analysis
- Integrated with existing visual pipeline
- Tested with various audio sources

## Test Plan
- [ ] Test with microphone input
- [ ] Test with line input
- [ ] Verify FFT values match expected ranges
EOF
)"
```

Or create via GitHub website:
1. Go to the repository on GitHub
2. Click "Compare & pull request"
3. Fill in description
4. Click "Create pull request"

### PR Description Template

```markdown
## Summary
What does this PR do? (1-3 bullet points)

## Changes
- List specific changes made
- Include file names if helpful

## Test Plan
How can this be tested?

## Screenshots/Videos
(If visual changes, add media)
```

---

## Part 5: Code Review

### Reviewing Others' Code

```bash
# Fetch and checkout their branch
gh pr checkout 123

# Or manually
git fetch origin
git checkout feature-branch-name
```

Test the changes locally, then comment on the PR.

### Responding to Review Comments

If changes are requested:

```bash
# Make the requested changes
git add .
git commit -m "Address review feedback: improve error handling"
git push
```

The PR updates automatically.

### Approving and Merging

Once approved:
1. Click "Merge pull request" on GitHub
2. Or use CLI: `gh pr merge --squash`
3. Delete the branch after merging

---

## Part 6: After Merging

### Clean Up Local Branches

```bash
# Switch to main
git checkout main

# Pull the merged changes
git pull

# Delete your old branch locally
git branch -d luke/add-audio-reactive

# Prune remote tracking branches
git fetch --prune
```

### Start Fresh for Next Feature

```bash
git checkout main
git pull
git checkout -b luke/next-feature
```

---

## Common Scenarios

### Someone Else Changed the Same File

```bash
# You try to push but get rejected
git push
# ! [rejected] - remote contains work you don't have

# Pull their changes
git pull --rebase

# Resolve any conflicts, then push
git push
```

### Need to Undo a Commit

```bash
# Undo last commit but keep changes
git reset --soft HEAD~1

# Undo last commit and discard changes
git reset --hard HEAD~1
```

### Wrong Branch

```bash
# Accidentally committed to main? Move to new branch:
git branch luke/my-feature
git reset --hard origin/main
git checkout luke/my-feature
```

### See What Others Are Working On

```bash
# List all branches
git branch -a

# See remote branches
git branch -r

# Check out someone's branch
git checkout origin/mitch/new-feature
```

---

## Project Communication

### Before Starting Work

1. Check GitHub Issues for tasks
2. Comment that you're working on something
3. Ask in Discord/Slack if unsure

### While Working

1. Push regularly (at least daily)
2. Create draft PRs for visibility: `gh pr create --draft`
3. Ask for help early if stuck

### Handoffs

If you need to pause work:

1. Push all changes
2. Update PR description with status
3. Let the team know

---

## Quick Reference

### Daily Workflow

```bash
git checkout main && git pull                    # Start fresh
git checkout -b luke/feature-name                # New branch
# ... do work ...
git add . && git commit -m "Description"         # Commit often
git push -u origin luke/feature-name             # Push branch
gh pr create                                      # Open PR
```

### Staying in Sync

```bash
git fetch origin                                  # Get remote changes
git merge origin/main                             # Merge into your branch
git push                                          # Push updated branch
```

### Common Commands

| Task | Command |
|------|---------|
| New branch | `git checkout -b name` |
| Switch branch | `git checkout name` |
| See all branches | `git branch -a` |
| Push new branch | `git push -u origin name` |
| Create PR | `gh pr create` |
| View PR | `gh pr view` |
| Checkout PR | `gh pr checkout 123` |
| Merge PR | `gh pr merge` |

---

## hairglasses-studio Conventions

### Branch Naming

```
luke/feature-description     # Your features
mitch/feature-description    # Mitch's features
fix/issue-description        # Bug fixes
experiment/idea              # Experiments
release/v1.0.0              # Releases
```

### Commit Style

- Present tense: "Add feature" not "Added feature"
- Capitalize first letter
- No period at end
- Keep under 50 characters

### PR Requirements

- Clear description of changes
- Test instructions
- Screenshots for visual changes
- Link to related issues

---

## Next Steps

- [OSC & MIDI Communication](16-osc-midi-communication.md) - Connect your apps
- [Real-time A/V Sync](17-realtime-av-sync.md) - Sync visuals to audio
