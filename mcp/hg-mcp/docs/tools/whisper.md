# whisper

> Audio transcription and translation using OpenAI Whisper

**4 tools**

## Tools

- [`aftrs_whisper_health`](#aftrs-whisper-health)
- [`aftrs_whisper_models`](#aftrs-whisper-models)
- [`aftrs_whisper_transcribe`](#aftrs-whisper-transcribe)
- [`aftrs_whisper_translate`](#aftrs-whisper-translate)

---

## aftrs_whisper_health

Check Whisper service health and get troubleshooting recommendations

**Complexity:** simple

**Tags:** `whisper`, `health`, `diagnostics`, `troubleshooting`

**Use Cases:**
- Diagnose issues
- Check installation
- Verify API configuration

---

## aftrs_whisper_models

List available Whisper models with size and capability info

**Complexity:** simple

**Tags:** `whisper`, `models`, `config`

**Use Cases:**
- View available models
- Compare model sizes
- Choose accuracy level

---

## aftrs_whisper_transcribe

Transcribe audio file to text using Whisper

**Complexity:** moderate

**Tags:** `whisper`, `transcribe`, `speech-to-text`, `audio`

**Use Cases:**
- Transcribe audio
- Convert speech to text
- Create captions

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `file_path` | string | Yes | Path to the audio file to transcribe |
| `language` | string |  | Language code (e.g., 'en', 'es', 'fr') - auto-detected if not specified |

### Example

```json
{
  "file_path": "example",
  "language": "example"
}
```

---

## aftrs_whisper_translate

Transcribe and translate audio to English

**Complexity:** moderate

**Tags:** `whisper`, `translate`, `speech-to-text`, `audio`

**Use Cases:**
- Translate foreign audio
- Create English subtitles
- Multi-language support

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `file_path` | string | Yes | Path to the audio file to translate |

### Example

```json
{
  "file_path": "example"
}
```

---

