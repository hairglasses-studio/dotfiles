# Contributing to Your Own Repo with Claude

Learn how to use Claude Code to build, maintain, and improve your own projects.

## The Workflow

1. **Clone or create** your repo
2. **Navigate** to the project folder
3. **Start Claude Code** in that folder
4. **Describe** what you want to build
5. **Review** Claude's changes
6. **Commit and push** when satisfied

---

## Starting a Session

```bash
# Navigate to your project
cd ~/projects/my-project

# Start Claude Code
claude
```

Claude automatically sees:
- All files in the directory
- Git status and history
- README and documentation

---

## Part 1: Adding Features

### Describe What You Want

Be specific about what you want to add:

```
"Add a command that lists all MP3 files in the current directory sorted by date"
```

Claude will:
1. Understand your existing code structure
2. Write code that fits your style
3. Create or modify appropriate files
4. Show you the changes

### Review Before Committing

After Claude makes changes, review them:

```
"Show me what you changed"
```

Or check git diff:

```
"Run git diff to show me the changes"
```

### Iterate

If something isn't quite right:

```
"The output should also show file sizes in MB"
```

Claude remembers the context and makes adjustments.

---

## Part 2: Bug Fixes

### Describe the Problem

```
"When I run the setlist command with no arguments, it crashes with a null error"
```

### Let Claude Investigate

```
"Look at the setlist function and find what causes the crash when no arguments are given"
```

### Apply the Fix

Claude will find the issue and propose a fix. Review it:

```
"Show me the fix before applying it"
```

---

## Part 3: Refactoring

### Clean Up Code

```
"This script is getting long. Split it into separate functions for each command"
```

### Improve Structure

```
"Move the configuration into a separate file and load it at startup"
```

### Add Documentation

```
"Add comments explaining what each function does"
```

---

## Part 4: Writing Tests

### Create Test Files

```
"Create a test file for the convert function that tests MP4 to MP3 conversion"
```

### Run Tests

```
"Run the tests and show me which ones fail"
```

### Fix Failing Tests

```
"Fix the failing test - the expected output format was wrong"
```

---

## Part 5: Git Operations

### Check Status

```
"Show me what files have changed"
```

### Create a Commit

```
"Commit these changes with a message describing the new feature"
```

### Create a Branch

```
"Create a new branch called 'add-playlist-export'"
```

### Push Changes

```
"Push this branch to GitHub"
```

### Create a Pull Request

```
"Create a pull request for this branch"
```

---

## Effective Prompting Patterns

### Be Specific

**Vague** (less effective):
```
"Make it better"
```

**Specific** (much better):
```
"Add error handling when the input file doesn't exist, with a clear message telling the user what went wrong"
```

### Provide Context

```
"In the convert.sh script, the video_to_audio function should also accept .mov files"
```

### Ask for Explanations

```
"Explain what this ffmpeg command does before you modify it"
```

### Request Options

```
"What are three different ways I could implement playlist export? Explain the pros and cons"
```

---

## Building a Complete Feature

### Example: Adding Playlist Export

**Step 1: Plan**
```
"I want to add a command that exports my setlist as an M3U playlist file. What approach would you recommend?"
```

**Step 2: Create**
```
"Create the playlist export command with the approach we discussed"
```

**Step 3: Test**
```
"Test the new command with my current setlist"
```

**Step 4: Refine**
```
"The M3U file needs to use relative paths instead of absolute paths"
```

**Step 5: Document**
```
"Update the README to document the new playlist export command"
```

**Step 6: Commit**
```
"Commit all the changes for the playlist export feature"
```

---

## Working with Existing Code

### Understand Before Changing

```
"Explain how the session logging currently works"
```

### Find Related Code

```
"Find all places where the BPM is used"
```

### Trace the Flow

```
"Walk me through what happens when I run 'setlist add'"
```

---

## Common Tasks

| Task | Prompt |
|------|--------|
| Add a feature | "Add a command that..." |
| Fix a bug | "This crashes when... Fix it" |
| Refactor | "Split this into smaller functions" |
| Add tests | "Create tests for the X function" |
| Update docs | "Update the README with the new options" |
| Check status | "Show me git status" |
| Commit | "Commit with message describing X" |
| Create PR | "Create a pull request for this branch" |

---

## Safety Tips

### Always Review Changes

Before committing, check what changed:

```
"Show me git diff"
```

### Use Branches for Big Changes

```
"Create a branch called 'experiment-new-feature' before making changes"
```

### Commit Frequently

Small, focused commits are easier to understand and revert:

```
"Commit just the changes to the convert function"
```

### Keep Backups

For important files you're modifying:

```
"Copy setlist.sh to setlist.sh.backup before making changes"
```

---

## Getting Unstuck

### When Claude Makes a Mistake

```
"That's not quite right. The function should return the count, not print it"
```

### When You Need to Start Over

```
"Undo those changes and let's try a different approach"
```

### When Something Breaks

```
"That change broke the script. Show me git diff and let's revert"
```

### When You're Lost

```
"I'm confused. Explain what the current state of the code is"
```

---

## Example Session

Here's a complete example of adding a feature:

```
Luke: I want to add a command to my setlist script that calculates total playing time

Claude: I'll look at your setlist.sh script and add a command to calculate total duration.
[Shows the code changes]

Luke: That works but I want it to show hours:minutes:seconds format

Claude: I'll update the output formatting to show HH:MM:SS format.
[Shows the updated code]

Luke: Perfect. Commit this with a good message

Claude: I'll commit these changes.
[Creates commit: "Add total duration command with HH:MM:SS format"]

Luke: Push it to GitHub

Claude: [Pushes to origin/main]
Done! The changes are now on GitHub.
```

---

## Project Maintenance

### Regular Updates

```
"Check if any of my dependencies have security updates"
```

### Code Quality

```
"Are there any obvious improvements to make in this codebase?"
```

### Documentation

```
"Is the README up to date with all current commands?"
```

---

## Next Steps

- [YouTube Resources](13-youtube-resources.md) - Video tutorials for learning
- [Project Templates](11-project-templates.md) - Starting new projects

---

## Quick Reference

```bash
# Start a session
cd ~/projects/my-project && claude

# Common prompts
"Show me what files are in this project"
"Explain how [feature] works"
"Add a command that [does something]"
"Fix the bug where [problem description]"
"Commit these changes"
"Push to GitHub"
"Create a pull request"
```
