# AFTRS Wiki MCP Server

A Model Context Protocol server for managing the AFTRS Wiki documentation system, providing comprehensive tools for reading, writing, searching, and analyzing the organizational knowledge base.

## Features

### Document Management
- **list_wiki_categories**: Browse all wiki categories with document counts
- **list_documents**: List documents in specific categories or across the entire wiki
- **read_document**: Read documents with multiple format options (raw, parsed, HTML)
- **create_document**: Create new wiki documents with frontmatter support
- **update_document**: Update existing documents with metadata merging

### Content Discovery
- **search_wiki**: Full-text search across all wiki documents with category filtering
- **cross_reference_search**: Find incoming and outgoing references between documents
- **extract_todos_and_issues**: Extract TODO items and issues from documentation

### Analytics & Reporting
- **get_wiki_statistics**: Comprehensive statistics about wiki content
- **get_project_inventory**: Parse project inventory TSV files

### Project Integration
- **Infrastructure**: Network diagrams, API management, Tailscale configuration
- **Projects**: AFTRS CLI, CR8 CLI, Console Hax, Agent CTL documentation
- **Patterns**: AI integration patterns, CLI design principles
- **Roadmap**: Feature tracking and issue management

## Installation

```bash
cd aftrs-cline/mcp-server-projects/aftrs-wiki
npm install
npm run build
```

## Configuration

Add to your Cline MCP configuration:

```json
{
  "mcpServers": {
    "aftrs-wiki": {
      "command": "node",
      "args": ["path/to/aftrs-wiki/dist/index.js"],
      "env": {
        "WIKI_BASE_PATH": "/home/hg/Docs/aftrs-void/aftrs_wiki"
      }
    }
  }
}
```

## Wiki Structure

The server manages the following documentation categories:

- **infrastructure/**: Network architecture, API management, security
- **projects/**: Individual project documentation and inventories
- **patterns/**: Design patterns and development guidelines
- **roadmap/**: Features and issues tracking
- **miscellaneous/**: General documentation

## Usage Examples

### Browse wiki categories
```typescript
await use_mcp_tool({
  server_name: "aftrs-wiki",
  tool_name: "list_wiki_categories",
  arguments: {}
});
```

### Search documentation
```typescript
await use_mcp_tool({
  server_name: "aftrs-wiki",
  tool_name: "search_wiki",
  arguments: {
    query: "tailscale configuration",
    category: "infrastructure"
  }
});
```

### Read project documentation
```typescript
await use_mcp_tool({
  server_name: "aftrs-wiki",
  tool_name: "read_document",
  arguments: {
    path: "projects/aftrs_cli.md",
    format: "parsed"
  }
});
```

### Extract project inventory
```typescript
await use_mcp_tool({
  server_name: "aftrs-wiki",
  tool_name: "get_project_inventory",
  arguments: {
    project: "console-hax"
  }
});
```

### Find document cross-references
```typescript
await use_mcp_tool({
  server_name: "aftrs-wiki",
  tool_name: "cross_reference_search",
  arguments: {
    document: "projects/aftrs_cli.md",
    searchType: "both"
  }
});
```

### Create new documentation
```typescript
await use_mcp_tool({
  server_name: "aftrs-wiki",
  tool_name: "create_document",
  arguments: {
    path: "projects/new_project.md",
    content: "# New Project\n\nProject documentation...",
    metadata: {
      author: "AFTRS Team",
      created: "2025-09-23",
      status: "draft"
    }
  }
});
```

## Advanced Features

### TODO and Issue Extraction
Extract actionable items from documentation:
```typescript
await use_mcp_tool({
  server_name: "aftrs-wiki",
  tool_name: "extract_todos_and_issues",
  arguments: {
    category: "roadmap",
    includeCompleted: false
  }
});
```

### Wiki Statistics
Get comprehensive analytics:
```typescript
await use_mcp_tool({
  server_name: "aftrs-wiki",
  tool_name: "get_wiki_statistics",
  arguments: {
    detailed: true
  }
});
```

## Architecture Integration

This server provides centralized access to AFTRS documentation:

- **Knowledge Management**: Single source of truth for project documentation
- **Cross-Project Discovery**: Find relationships between different AFTRS components
- **Development Workflows**: Extract TODOs and track project status
- **Inventory Management**: Parse and analyze project asset inventories
- **Pattern Recognition**: Identify common design patterns across projects

## File Format Support

- **Markdown**: Primary documentation format with frontmatter support
- **TSV**: Project inventory files (parsed as structured data)
- **YAML**: Frontmatter metadata extraction and manipulation
- **HTML**: Optional output format for rendered content

## Security Considerations

- Read-only access by default (write operations require explicit tool calls)
- Path traversal protection (all paths relative to wiki base)
- Metadata validation for document creation/updates
- Safe file parsing with error handling

## Development

```bash
# Install dependencies
npm install

# Run in development mode
npm run dev

# Build for production
npm run build

# Type checking
npm run type-check
```

## Troubleshooting

### Wiki Path Configuration
Ensure `WIKI_BASE_PATH` points to the correct directory:
```bash
ls -la /home/hg/Docs/aftrs-void/aftrs_wiki/
```

### File Permissions
Verify read/write permissions for wiki directories:
```bash
chmod -R 755 /home/hg/Docs/aftrs-void/aftrs_wiki/
```

### Markdown Parsing
Invalid frontmatter can cause parsing errors. Ensure YAML syntax is correct:
```yaml
---
title: "Document Title"
author: "Author Name"
date: "2025-09-23"
---
