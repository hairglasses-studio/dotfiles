# router

> Smart routing - natural language to appropriate tool

**1 tools**

## Tools

- [`aftrs_ask`](#aftrs-ask)

---

## aftrs_ask

Natural language query router. Ask anything and it routes to the appropriate tool. Examples: 'what's the TouchDesigner FPS?', 'are there any NDI sources?', 'start the show', 'what's wrong with the lighting?'

**Complexity:** simple

**Tags:** `router`, `ask`, `natural language`, `query`

**Use Cases:**
- Ask questions naturally
- Route to correct tool

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `explain` | boolean |  | If true, explains the routing instead of executing |
| `query` | string | Yes | Your question or command in natural language |

### Example

```json
{
  "explain": false,
  "query": "example"
}
```

---

