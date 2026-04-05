# graph

> Knowledge graph tools for semantic search and context discovery

**6 tools**

## Tools

- [`aftrs_context_from_graph`](#aftrs-context-from-graph)
- [`aftrs_graph_insights`](#aftrs-graph-insights)
- [`aftrs_graph_rebuild`](#aftrs-graph-rebuild)
- [`aftrs_graph_search`](#aftrs-graph-search)
- [`aftrs_resolution_path`](#aftrs-resolution-path)
- [`aftrs_similar_shows`](#aftrs-similar-shows)

---

## aftrs_context_from_graph

Get related context for a document via graph connections.

**Complexity:** simple

**Tags:** `graph`, `context`, `related`, `connections`

**Use Cases:**
- Get context for a show
- Find related runbooks

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `document` | string | Yes | Document path or name |
| `max_results` | number |  | Maximum results (default: 10) |

### Example

```json
{
  "document": "example",
  "max_results": 0
}
```

---

## aftrs_graph_insights

Get analytics about the knowledge graph: node counts, clusters, orphans, most connected documents.

**Complexity:** simple

**Tags:** `graph`, `analytics`, `insights`, `stats`

**Use Cases:**
- Understand vault structure
- Find orphan documents

---

## aftrs_graph_rebuild

Rebuild the knowledge graph from vault documents. Extracts [[wiki links]] and #tags to build semantic connections.

**Complexity:** moderate

**Tags:** `graph`, `rebuild`, `index`, `vault`

**Use Cases:**
- Update graph after vault changes
- Initial graph build

---

## aftrs_graph_search

Graph-enhanced search: finds documents and related content through semantic connections.

**Complexity:** simple

**Tags:** `graph`, `search`, `semantic`, `query`

**Use Cases:**
- Find related documents
- Discover connections

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `max_hops` | number |  | Maximum graph hops (default: 2) |
| `query` | string | Yes | Search query |

### Example

```json
{
  "max_hops": 0,
  "query": "example"
}
```

---

## aftrs_resolution_path

Find how similar issues were resolved before based on learnings and runbooks.

**Complexity:** simple

**Tags:** `graph`, `resolution`, `troubleshooting`, `history`

**Use Cases:**
- Find past solutions
- Troubleshoot issues

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `issue` | string | Yes | Issue or problem description |

### Example

```json
{
  "issue": "example"
}
```

---

## aftrs_similar_shows

Find shows similar to given criteria (tags, equipment, venue).

**Complexity:** simple

**Tags:** `graph`, `shows`, `similar`, `history`

**Use Cases:**
- Find past shows like this
- Reference similar events

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `criteria` | string | Yes | Search criteria (tags, equipment, venue, etc.) |

### Example

```json
{
  "criteria": "example"
}
```

---

