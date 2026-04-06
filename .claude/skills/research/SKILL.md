---
name: research
description: Search the docs knowledge base for existing research on a topic. Shows matches with relevance scores and identifies gaps.
allowed-tools: mcp__dotfiles__ops_research_check
---

Search the docs repo knowledge base for existing research before starting work on a topic. `$ARGUMENTS` is the search query (keywords or topic).

## Workflow

1. Call `ops_research_check(query="$ARGUMENTS")`
2. Display results:

### If results found:
```
## Research: "{query}"

**{total_docs} docs indexed** | {len(results)} matches found

| Relevance | Title | Path | Tags |
|-----------|-------|------|------|
| {relevance} | {title} | {path} | {tags} |

{If excerpt available, show top 3 excerpts}
```

### If gaps detected:
```
### Knowledge Gaps

{suggestion}

Missing coverage for: {gaps as comma-separated list}

Consider creating new research at the suggested path.
```

### If no query provided:
Show usage hint:
```
Usage: /research <topic>

Examples:
  /research mcp protocol
  /research circuit breaker patterns
  /research cost optimization
```
