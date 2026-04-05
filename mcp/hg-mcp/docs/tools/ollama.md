# ollama

> Local LLM inference using Ollama for text generation and chat

**5 tools**

## Tools

- [`aftrs_ollama_chat`](#aftrs-ollama-chat)
- [`aftrs_ollama_generate`](#aftrs-ollama-generate)
- [`aftrs_ollama_health`](#aftrs-ollama-health)
- [`aftrs_ollama_models`](#aftrs-ollama-models)
- [`aftrs_ollama_status`](#aftrs-ollama-status)

---

## aftrs_ollama_chat

Chat with a local LLM model using conversation history

**Complexity:** moderate

**Tags:** `ollama`, `chat`, `conversation`, `ai`

**Use Cases:**
- Chat with AI
- Multi-turn conversation
- Interactive Q&A

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `history` | array |  | Previous conversation messages |
| `message` | string | Yes | The message to send |
| `model` | string |  | Model to use (default: llama2) |
| `system` | string |  | Optional system prompt to set assistant behavior |

### Example

```json
{
  "history": [],
  "message": "example",
  "model": "example",
  "system": "example"
}
```

---

## aftrs_ollama_generate

Generate text completion using a local LLM model

**Complexity:** moderate

**Tags:** `ollama`, `generate`, `completion`, `ai`

**Use Cases:**
- Generate text
- Complete prompts
- Creative writing

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `max_tokens` | integer |  | Maximum tokens to generate |
| `model` | string |  | Model to use (default: llama2) |
| `prompt` | string | Yes | The prompt to generate completion for |
| `system` | string |  | Optional system prompt to set context |
| `temperature` | number |  | Sampling temperature (0.0-2.0, default: 0.8) |

### Example

```json
{
  "max_tokens": 0,
  "model": "example",
  "prompt": "example",
  "system": "example",
  "temperature": 0
}
```

---

## aftrs_ollama_health

Check Ollama service health and get troubleshooting recommendations

**Complexity:** simple

**Tags:** `ollama`, `health`, `diagnostics`, `troubleshooting`

**Use Cases:**
- Diagnose issues
- Check service health
- Get troubleshooting tips

---

## aftrs_ollama_models

List available Ollama models installed locally

**Complexity:** simple

**Tags:** `ollama`, `models`, `list`, `ai`

**Use Cases:**
- List installed models
- Check model sizes
- View available options

---

## aftrs_ollama_status

Get Ollama service status including version and loaded models

**Complexity:** simple

**Tags:** `ollama`, `llm`, `status`, `ai`

**Use Cases:**
- Check service status
- View loaded models
- Verify connectivity

---

