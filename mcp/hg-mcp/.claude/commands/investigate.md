# Investigate Studio

Check studio system health and identify issues.

## Usage
```
/investigate [system]
```

Systems: `td` (TouchDesigner), `stream` (streaming), `unraid`, `all`

## Steps

1. Check requested system health:
   - **td**: TouchDesigner status, FPS, errors, GPU memory
   - **stream**: NDI sources, OBS status, stream health
   - **unraid**: Array status, VMs, Docker containers
   - **all**: All systems

2. Identify any issues or warnings

3. Provide recommendations

## Output Format

```markdown
# Studio Investigation: [System]

## Status: [Healthy/Degraded/Critical]

### [System Name]
| Metric | Value | Status |
|--------|-------|--------|
| metric | value | ok/warn/error |

### Issues Found
- Issue 1: description
- Issue 2: description

### Recommendations
1. Recommendation 1
2. Recommendation 2

### Quick Fixes
\`\`\`bash
# Command to fix issue 1
command here
\`\`\`
```

## Notes

- Use consolidated tools when available (e.g., `aftrs_studio_health_full`)
- For TouchDesigner, use `mcp__touchdesigner__*` tools
- Color code: green=healthy, yellow=degraded, red=critical
