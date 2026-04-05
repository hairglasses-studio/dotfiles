# notion

> Notion workspace integration for pages, databases, and content management

**8 tools**

## Tools

- [`aftrs_notion_append`](#aftrs-notion-append)
- [`aftrs_notion_create_page`](#aftrs-notion-create-page)
- [`aftrs_notion_database`](#aftrs-notion-database)
- [`aftrs_notion_database_query`](#aftrs-notion-database-query)
- [`aftrs_notion_page`](#aftrs-notion-page)
- [`aftrs_notion_page_content`](#aftrs-notion-page-content)
- [`aftrs_notion_search`](#aftrs-notion-search)
- [`aftrs_notion_status`](#aftrs-notion-status)

---

## aftrs_notion_append

Append content to an existing Notion page

**Complexity:** simple

**Tags:** `notion`, `append`, `content`, `update`

**Use Cases:**
- Add content
- Update page
- Append text

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `block_type` | string |  | Block type: 'paragraph', 'heading', 'bullet', 'todo' (default: paragraph) |
| `content` | string | Yes | Text content to append |
| `page_id` | string | Yes | The Notion page ID to append to |

### Example

```json
{
  "block_type": "example",
  "content": "example",
  "page_id": "example"
}
```

---

## aftrs_notion_create_page

Create a new Notion page in a parent page or database

**Complexity:** moderate

**Tags:** `notion`, `create`, `page`, `new`

**Use Cases:**
- Create page
- Add database entry
- New document

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `content` | string |  | Initial page content (plain text, will be added as paragraph) |
| `is_database` | boolean |  | Whether parent is a database (default: false) |
| `parent_id` | string | Yes | Parent page or database ID |
| `title` | string | Yes | Page title |

### Example

```json
{
  "content": "example",
  "is_database": false,
  "parent_id": "example",
  "title": "example"
}
```

---

## aftrs_notion_database

Get a Notion database schema and metadata

**Complexity:** simple

**Tags:** `notion`, `database`, `schema`, `properties`

**Use Cases:**
- Get database schema
- View properties
- Check structure

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `database_id` | string | Yes | The Notion database ID |

### Example

```json
{
  "database_id": "example"
}
```

---

## aftrs_notion_database_query

Query a Notion database with optional filters and sorting

**Complexity:** moderate

**Tags:** `notion`, `database`, `query`, `filter`, `sort`

**Use Cases:**
- Query database
- Filter records
- List entries

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `database_id` | string | Yes | The Notion database ID |
| `filter_property` | string |  | Property name to filter on |
| `filter_value` | string |  | Value to filter for (equals match) |
| `limit` | integer |  | Maximum results (default: 100) |
| `sort_direction` | string |  | Sort direction: 'ascending' or 'descending' |
| `sort_property` | string |  | Property name to sort by |

### Example

```json
{
  "database_id": "example",
  "filter_property": "example",
  "filter_value": "example",
  "limit": 0,
  "sort_direction": "example",
  "sort_property": "example"
}
```

---

## aftrs_notion_page

Get a Notion page by ID with its properties

**Complexity:** simple

**Tags:** `notion`, `page`, `get`, `properties`

**Use Cases:**
- Get page details
- View properties
- Check metadata

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `page_id` | string | Yes | The Notion page ID (UUID format) |

### Example

```json
{
  "page_id": "example"
}
```

---

## aftrs_notion_page_content

Get the content blocks of a Notion page

**Complexity:** simple

**Tags:** `notion`, `page`, `content`, `blocks`

**Use Cases:**
- Read page content
- Get blocks
- Extract text

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `limit` | integer |  | Maximum blocks to return (default: 100) |
| `page_id` | string | Yes | The Notion page ID |

### Example

```json
{
  "limit": 0,
  "page_id": "example"
}
```

---

## aftrs_notion_search

Search across all Notion pages and databases

**Complexity:** simple

**Tags:** `notion`, `search`, `pages`, `databases`

**Use Cases:**
- Find pages
- Search content
- Locate databases

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `filter` | string |  | Filter results: 'page' or 'database' (default: all) |
| `limit` | integer |  | Maximum results (default: 25, max: 100) |
| `query` | string | Yes | Search query text |

### Example

```json
{
  "filter": "example",
  "limit": 0,
  "query": "example"
}
```

---

## aftrs_notion_status

Get Notion workspace connection status and current user info

**Complexity:** simple

**Tags:** `notion`, `status`, `user`, `workspace`

**Use Cases:**
- Check connection
- View user info
- Verify access

---

