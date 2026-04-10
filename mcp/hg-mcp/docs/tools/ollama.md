# ollama

> Local LLM inference using Ollama for text generation, chat, and native model capabilities

**10 tools**

## Tools

- [`aftrs_ollama_status`](#aftrs_ollama_status)
- [`aftrs_ollama_models`](#aftrs_ollama_models)
- [`aftrs_ollama_loaded`](#aftrs_ollama_loaded)
- [`aftrs_ollama_show`](#aftrs_ollama_show)
- [`aftrs_ollama_generate`](#aftrs_ollama_generate)
- [`aftrs_ollama_structured`](#aftrs_ollama_structured)
- [`aftrs_ollama_chat`](#aftrs_ollama_chat)
- [`aftrs_ollama_tool_chat`](#aftrs_ollama_tool_chat)
- [`aftrs_ollama_health`](#aftrs_ollama_health)
- [`aftrs_ollama_readiness`](#aftrs_ollama_readiness)

## aftrs_ollama_status

Get Ollama service status including version, installed models, and the first loaded model when one is resident.

## aftrs_ollama_models

List installed Ollama models from `/api/tags`.

## aftrs_ollama_loaded

List loaded Ollama models from `/api/ps`, including residency metadata such as `expires_at`, `context_length`, and `size_vram`.

## aftrs_ollama_show

Show detailed metadata for one model via `/api/show`.

Parameters:

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `model` | string | Yes | Model name or alias to inspect |

## aftrs_ollama_generate

Generate text completion using a local LLM model.

Parameters:

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `prompt` | string | Yes | Prompt to complete |
| `model` | string |  | Model to use, default `code-primary` |
| `system` | string |  | Optional system prompt |
| `temperature` | number |  | Sampling temperature, default `0.8` |
| `max_tokens` | integer |  | Maximum tokens to generate |

## aftrs_ollama_structured

Generate schema-constrained JSON using Ollama's native `format` support.

Parameters:

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `prompt` | string | Yes | Prompt to answer with structured JSON |
| `schema` | object | Yes | JSON schema object passed through as `format` |
| `model` | string |  | Model to use, default `code-primary` |
| `system` | string |  | Optional system prompt |
| `temperature` | number |  | Sampling temperature, default `0.0` |
| `max_tokens` | integer |  | Maximum tokens to generate |

## aftrs_ollama_chat

Chat with a local LLM model using conversation history.

Parameters:

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `message` | string | Yes | Message to send |
| `model` | string |  | Model to use, default `code-primary` |
| `system` | string |  | Optional system prompt |
| `history` | array |  | Previous conversation messages |

## aftrs_ollama_tool_chat

Run a native Ollama chat request with explicit tool definitions and inspect returned tool calls without executing them server-side.

Parameters:

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `message` | string | Yes | Message to send |
| `tools` | array | Yes | Tool definitions to expose to the model |
| `model` | string |  | Model to use, default `code-primary` |
| `system` | string |  | Optional system prompt |
| `history` | array |  | Previous conversation messages |

## aftrs_ollama_health

Check Ollama service health and receive troubleshooting recommendations.

## aftrs_ollama_readiness

Return the live Ollama readiness report, including required models, managed `code-*` alias state, and recommended pull commands.

Parameters:

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `require_heavy` | boolean |  | Also require the heavy local coding lane |
