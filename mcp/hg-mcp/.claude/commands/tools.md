# Find Tools

Search for available MCP tools by keyword.

## Usage
```
/tools <keyword>
```

## Steps

1. If a keyword is provided, search for tools matching that keyword
2. List matching tools with their descriptions
3. Show example usage for each tool

## Output Format

```
## Tools matching "<keyword>"

| Tool | Description |
|------|-------------|
| tool_name | What it does |

### Example Usage

tool_name:
  - Example 1: description
  - Example 2: description
```

## Notes

- Use `aftrs_tool_search` when implemented
- For now, reference docs/MCP_TOOLS.md
- TouchDesigner tools available via `mcp__touchdesigner__*`
