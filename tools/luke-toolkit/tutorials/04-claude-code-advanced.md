# Advanced Claude Code Techniques

Level up your Claude Code skills with these advanced workflows.

## Multi-File Operations

### Search Across Your Codebase

```
"Find all files that import the database module"
"Show me everywhere we handle errors"
"List all TODO comments in the project"
```

### Refactoring

```
"Rename the function 'processData' to 'transformData' everywhere"
"Move the helper functions from utils.py to helpers.py"
"Update all imports after I renamed the module"
```

### Batch Edits

```
"Add error handling to all API calls"
"Update the copyright year in all file headers"
"Convert all var declarations to const/let"
```

---

## Effective Searching

### Basic Search Patterns

```
"Find files named config"
# Claude uses: Glob for **/*config*

"Search for 'TODO' in all Python files"
# Claude uses: Grep with pattern and file type filter

"Find where 'UserService' is defined"
# Claude uses: Grep for class/function definition patterns
```

### Advanced Patterns

```
"Find all functions that take more than 3 parameters"
"Show me files modified in the last week"
"Find unused imports"
```

### Combining Searches

```
"Find all error handling code in the API routes, then show me patterns"
```

---

## Git Integration

### Checking Status

```
"What files have I changed?"
# Claude runs: git status

"Show me the diff for main.py"
# Claude runs: git diff main.py

"What commits did I make today?"
# Claude runs: git log --since="today" --oneline
```

### Making Commits

```
"Commit these changes with a good message"
# Claude will:
# 1. Run git status to see changes
# 2. Run git diff to understand what changed
# 3. Create a descriptive commit message
# 4. Run git add and git commit
```

### Working with Branches

```
"Create a new branch for the login feature"
# Claude runs: git checkout -b feature/login

"What's different between my branch and main?"
# Claude runs: git diff main...HEAD

"Merge main into my branch"
# Claude runs: git merge main
```

### Creating Pull Requests

```
"Create a PR for this branch"
# Claude will:
# 1. Push the branch
# 2. Use gh to create the PR
# 3. Write a good PR description
```

---

## Debugging with Claude

### Investigating Errors

```
"I'm getting this error: [paste error]"
# Claude will:
# 1. Read the relevant files
# 2. Understand the error context
# 3. Suggest or implement a fix
```

### Adding Debug Logging

```
"Add logging to track what's happening in processOrder()"
# Claude adds appropriate log statements

"Help me understand why this loop is slow"
# Claude analyzes and suggests optimizations
```

### Stack Trace Analysis

```
"Explain this stack trace and fix the issue:
[paste stack trace]"
```

### Testing Fixes

```
"Run the tests to see if my fix worked"
"Add a test case for the bug I just fixed"
```

---

## Session Management

### Slash Commands

| Command | What it does |
|---------|--------------|
| `/help` | Show available commands |
| `/clear` | Clear conversation, fresh start |
| `/compact` | Summarize context and continue |
| `/cost` | Show token usage |
| `/status` | Show current working state |
| `/doctor` | Check Claude Code health |

### When to Use Each

**Use `/compact` when:**
- Conversation is getting long
- Claude seems to forget earlier context
- Want to continue but save tokens

**Use `/clear` when:**
- Starting a completely new task
- Want to reset assumptions
- Conversation went off track

### Memory Tips

```
"Remember that we're using PostgreSQL not MySQL"
"Keep in mind this is a TypeScript project"
"Note: the API runs on port 8080"
```

---

## Working with Large Codebases

### Orientation

```
"Give me an overview of this codebase structure"
"What's the main entry point?"
"How is the project organized?"
```

### Focused Work

```
"Let's focus on the authentication module"
"Only look at files in src/api/"
"Ignore the test files for now"
```

### Understanding Dependencies

```
"What does this module depend on?"
"Show me the import chain for UserService"
"What would break if I changed this function?"
```

---

## Best Practices for Prompting

### Be Specific

Instead of:
> "Fix the bug"

Try:
> "The login form shows 'undefined' after submission. Check the handleSubmit function in LoginForm.tsx"

### Provide Context

Instead of:
> "Add a feature"

Try:
> "Add a 'remember me' checkbox to the login form that saves the username in localStorage"

### Ask for Explanation First

```
"Explain what this code does before changing it"
"What would be the impact of this change?"
"Are there any risks with this approach?"
```

### Incremental Changes

```
"Let's do this step by step:
1. First, add the new field
2. Then update the form
3. Finally, add validation"
```

### Verify Understanding

```
"Before you make changes, tell me what you're going to do"
"Summarize the plan"
"What files will this affect?"
```

---

## Quick Reference

### Common Workflows

| Task | What to say |
|------|-------------|
| Start new feature | "Create a new branch and set up the basic structure for [feature]" |
| Fix a bug | "I'm seeing [error]. Investigate and fix it" |
| Refactor | "Refactor [component] to use [pattern]" |
| Add tests | "Add tests for [function/module]" |
| Review code | "Review this code and suggest improvements" |
| Document | "Add documentation for [module]" |
| Deploy | "What steps to deploy this?" |

### Useful Phrases

```
"Walk me through..."
"Explain before changing..."
"What's the best way to..."
"Can you show me an example of..."
"What could go wrong with..."
"How do I test this?"
```

---

## Next Steps

- [MCP Servers](05-mcp-servers.md) - Extend Claude with custom tools
- [Building MCP Modules](06-building-mcp-modules.md) - Create your own tools
