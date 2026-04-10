package clients

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDashboardClientQuickStatusShowsDegradedOllama(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/version":
			_, _ = w.Write([]byte(`{"version":"0.11.0"}`))
		case "/api/tags":
			_, _ = w.Write([]byte(`{"models":[{"name":"devstral-small-2"},{"name":"code-fast"},{"name":"qwen2.5-coder:7b"},{"name":"nomic-embed-text:v1.5"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("OLLAMA_BASE_URL", server.URL)
	t.Setenv("OLLAMA_CHAT_MODEL", "code-primary")
	t.Setenv("OLLAMA_FAST_MODEL", "code-fast")
	t.Setenv("OLLAMA_EMBED_MODEL", "nomic-embed-text:v1.5")

	client := &DashboardClient{
		systems: map[string]*SystemConfig{
			"ollama": {
				Name:        "ollama",
				Category:    "ai",
				Host:        "localhost",
				Port:        11434,
				Protocol:    "http",
				HealthPath:  "/api/tags",
				Description: "Ollama local LLM",
			},
		},
	}

	status := client.GetQuickStatus(context.Background())
	if !strings.Contains(status, "1 degraded") {
		t.Fatalf("quick status should report degraded ollama, got %q", status)
	}
	if !strings.Contains(status, "⚠ollama") {
		t.Fatalf("quick status should flag ollama as degraded, got %q", status)
	}
}

func TestCheckSystemDetailedIncludesOllamaReadiness(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/version":
			_, _ = w.Write([]byte(`{"version":"0.11.0"}`))
		case "/api/tags":
			_, _ = w.Write([]byte(`{"models":[{"name":"devstral-small-2"},{"name":"code-fast"},{"name":"qwen2.5-coder:7b"},{"name":"nomic-embed-text:v1.5"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("OLLAMA_BASE_URL", server.URL)
	t.Setenv("OLLAMA_CHAT_MODEL", "code-primary")
	t.Setenv("OLLAMA_FAST_MODEL", "code-fast")
	t.Setenv("OLLAMA_EMBED_MODEL", "nomic-embed-text:v1.5")

	client := &DashboardClient{}
	status := client.checkSystemDetailed(context.Background(), &SystemConfig{
		Name:     "ollama",
		Category: "ai",
	})
	if status.Status != "degraded" {
		t.Fatalf("status = %q, want degraded", status.Status)
	}
	if status.Readiness == nil {
		t.Fatal("expected readiness payload")
	}
	if len(status.Readiness.MissingModels) != 1 || status.Readiness.MissingModels[0] != "code-primary" {
		t.Fatalf("unexpected missing models %+v", status.Readiness.MissingModels)
	}
	if !strings.Contains(status.Message, "code-primary") {
		t.Fatalf("expected status message to mention missing alias/model, got %q", status.Message)
	}
}
