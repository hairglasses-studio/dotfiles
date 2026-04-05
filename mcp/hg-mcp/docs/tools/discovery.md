# discovery

> Tool discovery and catalog features

**9 tools**

## Tools

- [`aftrs_tool_catalog`](#aftrs-tool-catalog)
- [`aftrs_tool_search`](#aftrs-tool-search)
- [`aftrs_tool_stats`](#aftrs-tool-stats)
- [`aftrs_tools_alias`](#aftrs-tools-alias)
- [`aftrs_tools_by_system`](#aftrs-tools-by-system)
- [`aftrs_tools_favorites`](#aftrs-tools-favorites)
- [`aftrs_tools_recent`](#aftrs-tools-recent)
- [`aftrs_tools_related`](#aftrs-tools-related)
- [`aftrs_tools_workflow`](#aftrs-tools-workflow)

---

## aftrs_tool_catalog

Browse the complete tool catalog organized by category.

**Complexity:** simple

**Tags:** `discovery`, `catalog`, `browse`

**Use Cases:**
- Browse available tools
- Explore categories

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `category` | string |  | Filter by category (e.g., 'discord', 'touchdesigner') |
| `format` | string |  | Output format: 'compact' (names only) or 'full' (with descriptions). Default: compact |

### Example

```json
{
  "category": "example",
  "format": "example"
}
```

---

## aftrs_tool_search

Search for tools by keyword. Returns matching tools with descriptions.

**Complexity:** simple

**Tags:** `discovery`, `search`, `help`

**Use Cases:**
- Find tools by keyword
- Discover available functionality

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `limit` | number |  | Maximum results to return (default: 10, max: 50) |
| `query` | string | Yes | Search query (e.g., 'discord', 'status', 'send message') |

### Example

```json
{
  "limit": 0,
  "query": "example"
}
```

---

## aftrs_tool_stats

Get statistics about the tool registry: total tools, categories, and modules.

**Complexity:** simple

**Tags:** `discovery`, `stats`

**Use Cases:**
- Understand tool distribution
- Get registry overview

---

## aftrs_tools_alias

Create short aliases for frequently used tools.

**Complexity:** simple

**Tags:** `discovery`, `alias`, `shortcuts`

**Use Cases:**
- Create tool shortcuts
- Simplify tool names

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string |  | Action: 'list' (default), 'set', 'remove', or 'resolve' |
| `alias` | string |  | Short alias name (required for set/remove/resolve) |
| `tool_name` | string |  | Full tool name (required for 'set' action) |

### Example

```json
{
  "action": "example",
  "alias": "example",
  "tool_name": "example"
}
```

---

## aftrs_tools_by_system

List all tools for a specific system (resolume, ableton, obs, etc.).

**Complexity:** simple

**Tags:** `discovery`, `system`, `filter`

**Use Cases:**
- Find all tools for a system
- Explore system capabilities

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `system` | string |  | System name (e.g., 'resolume', 'ableton', 'obs'). Leave empty to list systems. |

### Example

```json
{
  "system": "example"
}
```

---

## aftrs_tools_favorites

Manage favorite tools for quick access.

**Complexity:** simple

**Tags:** `discovery`, `favorites`, `quick-access`

**Use Cases:**
- Mark frequently used tools
- Quick access to favorites

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string |  | Action: 'list' (default), 'add', or 'remove' |
| `tool_name` | string |  | Tool name (required for add/remove) |

### Example

```json
{
  "action": "example",
  "tool_name": "example"
}
```

---

## aftrs_tools_recent

List recently used tools in this session.

**Complexity:** simple

**Tags:** `discovery`, `recent`, `history`

**Use Cases:**
- See recent tool usage
- Quick access to used tools

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `limit` | number |  | Maximum tools to return (default: 10) |

### Example

```json
{
  "limit": 0
}
```

---

## aftrs_tools_related

Find tools related to a given tool by shared category and tags.

**Complexity:** simple

**Tags:** `discovery`, `related`, `navigation`

**Use Cases:**
- Find similar tools
- Discover related functionality

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `limit` | number |  | Maximum related tools to return (default: 10) |
| `tool_name` | string | Yes | Name of the tool to find relations for |

### Example

```json
{
  "limit": 0,
  "tool_name": "example"
}
```

---

## aftrs_tools_workflow

Get recommended tool sequences for common goals and workflows.

**Complexity:** simple

**Tags:** `discovery`, `workflow`, `sequence`

**Use Cases:**
- Get recommended tool order
- Learn common patterns

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `goal` | string |  | Goal or workflow name (e.g., 'bpm_sync', 'troubleshoot_audio'). Leave empty to list all. |

### Example

```json
{
  "goal": "example"
}
```

---

