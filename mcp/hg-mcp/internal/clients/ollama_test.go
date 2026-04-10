package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

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

func TestRecommendedOllamaPullCommandsUseBackingModelsForAliases(t *testing.T) {
	t.Setenv("OLLAMA_CHAT_MODEL", "code-primary")
	t.Setenv("OLLAMA_FAST_MODEL", "code-fast")
	t.Setenv("OLLAMA_CODE_MODEL", "code-primary")
	t.Setenv("OLLAMA_HEAVY_CODE_MODEL", "code-heavy")
	t.Setenv("OLLAMA_HIGH_CONTEXT_CODE_MODEL", "code-long")
	t.Setenv("OLLAMA_EMBED_MODEL", "nomic-embed-text:v1.5")

	got := RecommendedOllamaPullCommands()
	want := []string{
		"ollama pull devstral-small-2",
		"ollama pull qwen2.5-coder:7b",
		"ollama pull devstral-2",
		"ollama pull qwen3-coder-next",
		"ollama pull nomic-embed-text:v1.5",
	}

	if len(got) != len(want) {
		t.Fatalf("RecommendedOllamaPullCommands() length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("RecommendedOllamaPullCommands()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestGetReadinessReportsMissingAlias(t *testing.T) {
	t.Setenv("OLLAMA_CHAT_MODEL", "code-primary")
	t.Setenv("OLLAMA_FAST_MODEL", "code-fast")
	t.Setenv("OLLAMA_EMBED_MODEL", "nomic-embed-text:v1.5")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/version":
			_, _ = w.Write([]byte(`{"version":"0.20.4"}`))
		case "/api/tags":
			_, _ = w.Write([]byte(`{"models":[{"name":"devstral-small-2"},{"name":"code-fast"},{"name":"qwen2.5-coder:7b"},{"name":"nomic-embed-text:v1.5"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := &OllamaClient{baseURL: srv.URL, httpClient: srv.Client()}
	readiness, err := client.GetReadiness(context.Background(), false)
	if err != nil {
		t.Fatalf("GetReadiness() error = %v", err)
	}
	if readiness.Ready {
		t.Fatal("expected readiness to fail when code-primary alias is missing")
	}
	if len(readiness.MissingModels) != 1 || readiness.MissingModels[0] != "code-primary" {
		t.Fatalf("missing_models = %#v, want [code-primary]", readiness.MissingModels)
	}
	if readiness.AliasChecks[2].Status != "missing_alias" {
		t.Fatalf("code-primary alias status = %q, want missing_alias", readiness.AliasChecks[2].Status)
	}
}

func TestGenerateRecordsGenAISpan(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		otel.SetTracerProvider(prev)
		_ = tp.Shutdown(context.Background())
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Fatalf("path = %q, want /api/generate", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model":             "code-primary",
			"response":          "ok",
			"done":              true,
			"prompt_eval_count": 17,
			"eval_count":        9,
		})
	}))
	defer srv.Close()

	client := &OllamaClient{baseURL: srv.URL, httpClient: srv.Client()}
	resp, err := client.Generate(context.Background(), &GenerateRequest{Model: "code-primary", Prompt: "hello"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if resp == nil || resp.Model != "code-primary" {
		t.Fatalf("Generate() = %#v, want model code-primary", resp)
	}

	assertOllamaSpanAttrs(t, exporter.GetSpans(), map[string]any{
		"gen_ai.system":              "ollama",
		"gen_ai.operation.name":      "generate",
		"gen_ai.request.model":       "code-primary",
		"gen_ai.usage.input_tokens":  int64(17),
		"gen_ai.usage.output_tokens": int64(9),
		"server.address":             srv.URL,
	})
}

func TestChatRecordsGenAISpan(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		otel.SetTracerProvider(prev)
		_ = tp.Shutdown(context.Background())
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Fatalf("path = %q, want /api/chat", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model": "code-long",
			"message": map[string]any{
				"role":    "assistant",
				"content": "ok",
			},
			"done":              true,
			"prompt_eval_count": 23,
			"eval_count":        11,
		})
	}))
	defer srv.Close()

	client := &OllamaClient{baseURL: srv.URL, httpClient: srv.Client()}
	resp, err := client.Chat(context.Background(), &ChatRequest{
		Model: "code-long",
		Messages: []ChatMessage{{
			Role:    "user",
			Content: "hello",
		}},
	})
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}
	if resp == nil || resp.Model != "code-long" {
		t.Fatalf("Chat() = %#v, want model code-long", resp)
	}

	assertOllamaSpanAttrs(t, exporter.GetSpans(), map[string]any{
		"gen_ai.system":              "ollama",
		"gen_ai.operation.name":      "chat",
		"gen_ai.request.model":       "code-long",
		"gen_ai.usage.input_tokens":  int64(23),
		"gen_ai.usage.output_tokens": int64(11),
		"server.address":             srv.URL,
	})
}

func TestEmbedRecordsGenAISpan(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		otel.SetTracerProvider(prev)
		_ = tp.Shutdown(context.Background())
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embeddings" {
			t.Fatalf("path = %q, want /api/embeddings", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"embedding": []float64{0.1, 0.2},
		})
	}))
	defer srv.Close()

	client := &OllamaClient{baseURL: srv.URL, httpClient: srv.Client()}
	resp, err := client.Embed(context.Background(), "nomic-embed-text:v1.5", "hello")
	if err != nil {
		t.Fatalf("Embed() error = %v", err)
	}
	if resp == nil || len(resp.Embedding) != 2 {
		t.Fatalf("Embed() = %#v, want two embedding values", resp)
	}

	assertOllamaSpanAttrs(t, exporter.GetSpans(), map[string]any{
		"gen_ai.system":         "ollama",
		"gen_ai.operation.name": "embed",
		"gen_ai.request.model":  "nomic-embed-text:v1.5",
		"server.address":        srv.URL,
	})
}

func assertOllamaSpanAttrs(t *testing.T, spans tracetest.SpanStubs, want map[string]any) {
	t.Helper()
	if len(spans) == 0 {
		t.Fatal("expected at least one exported span")
	}

	attrMap := make(map[string]any, len(spans[0].Attributes))
	for _, attr := range spans[0].Attributes {
		attrMap[string(attr.Key)] = attr.Value.AsInterface()
	}

	for key, expected := range want {
		if got := attrMap[key]; got != expected {
			t.Fatalf("%s = %v, want %v", key, got, expected)
		}
	}
}
