package clients

import "testing"

func TestDefaultOllamaBaseURLUsesBaseURLFirst(t *testing.T) {
	t.Setenv("OLLAMA_BASE_URL", "http://127.0.0.1:11434/")
	t.Setenv("OLLAMA_HOST", "legacy-host")
	t.Setenv("OLLAMA_PORT", "9999")

	if got := DefaultOllamaBaseURL(); got != "http://127.0.0.1:11434" {
		t.Fatalf("DefaultOllamaBaseURL() = %q, want %q", got, "http://127.0.0.1:11434")
	}
}

func TestDefaultOllamaBaseURLFallsBackToLegacyHostPort(t *testing.T) {
	t.Setenv("OLLAMA_BASE_URL", "")
	t.Setenv("OLLAMA_HOST", "legacy-host")
	t.Setenv("OLLAMA_PORT", "9999")

	if got := DefaultOllamaBaseURL(); got != "http://legacy-host:9999" {
		t.Fatalf("DefaultOllamaBaseURL() = %q, want %q", got, "http://legacy-host:9999")
	}
}

func TestDefaultOllamaBaseURLAcceptsLegacyHostWithScheme(t *testing.T) {
	t.Setenv("OLLAMA_BASE_URL", "")
	t.Setenv("OLLAMA_HOST", "https://legacy-host:9999/")
	t.Setenv("OLLAMA_PORT", "11434")

	if got := DefaultOllamaBaseURL(); got != "https://legacy-host:9999" {
		t.Fatalf("DefaultOllamaBaseURL() = %q, want %q", got, "https://legacy-host:9999")
	}
}

func TestDefaultOllamaModelsUseEnvOverrides(t *testing.T) {
	t.Setenv("OLLAMA_CHAT_MODEL", "chat-model")
	t.Setenv("OLLAMA_FAST_MODEL", "fast-model")
	t.Setenv("OLLAMA_CODE_MODEL", "code-model")
	t.Setenv("OLLAMA_HEAVY_CODE_MODEL", "heavy-model")
	t.Setenv("OLLAMA_HIGH_CONTEXT_CODE_MODEL", "context-model")
	t.Setenv("OLLAMA_EMBED_MODEL", "embed-model")
	t.Setenv("OLLAMA_KEEP_ALIVE", "30m")

	if got := DefaultOllamaChatModel(); got != "chat-model" {
		t.Fatalf("DefaultOllamaChatModel() = %q, want %q", got, "chat-model")
	}
	if got := DefaultOllamaFastModel(); got != "fast-model" {
		t.Fatalf("DefaultOllamaFastModel() = %q, want %q", got, "fast-model")
	}
	if got := DefaultOllamaCodeModel(); got != "code-model" {
		t.Fatalf("DefaultOllamaCodeModel() = %q, want %q", got, "code-model")
	}
	if got := DefaultOllamaHeavyCodeModel(); got != "heavy-model" {
		t.Fatalf("DefaultOllamaHeavyCodeModel() = %q, want %q", got, "heavy-model")
	}
	if got := DefaultOllamaHighContextCodeModel(); got != "context-model" {
		t.Fatalf("DefaultOllamaHighContextCodeModel() = %q, want %q", got, "context-model")
	}
	if got := DefaultOllamaEmbedModel(); got != "embed-model" {
		t.Fatalf("DefaultOllamaEmbedModel() = %q, want %q", got, "embed-model")
	}
	if got := DefaultOllamaKeepAlive(); got != "30m" {
		t.Fatalf("DefaultOllamaKeepAlive() = %q, want %q", got, "30m")
	}
}
