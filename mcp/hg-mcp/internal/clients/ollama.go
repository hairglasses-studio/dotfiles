// Package clients provides API clients for external services.
package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	httpclient "github.com/hairglasses-studio/mcpkit/client"
)

// OllamaClient provides access to local Ollama LLM
type OllamaClient struct {
	baseURL    string
	httpClient *http.Client
}

// OllamaModel represents an available model
type OllamaModel struct {
	Name       string    `json:"name"`
	ModifiedAt time.Time `json:"modified_at"`
	Size       int64     `json:"size"`
	Digest     string    `json:"digest"`
	Details    struct {
		Format            string   `json:"format"`
		Family            string   `json:"family"`
		Families          []string `json:"families"`
		ParameterSize     string   `json:"parameter_size"`
		QuantizationLevel string   `json:"quantization_level"`
	} `json:"details"`
}

// OllamaStatus represents service status
type OllamaStatus struct {
	Running     bool          `json:"running"`
	Version     string        `json:"version,omitempty"`
	Models      []OllamaModel `json:"models,omitempty"`
	LoadedModel string        `json:"loaded_model,omitempty"`
	BaseURL     string        `json:"base_url"`
}

// GenerateRequest represents a generation request
type GenerateRequest struct {
	Model     string                 `json:"model"`
	Prompt    string                 `json:"prompt"`
	System    string                 `json:"system,omitempty"`
	Stream    bool                   `json:"stream"`
	KeepAlive string                 `json:"keep_alive,omitempty"`
	Options   map[string]interface{} `json:"options,omitempty"`
}

// GenerateResponse represents a generation response
type GenerateResponse struct {
	Model              string `json:"model"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	Context            []int  `json:"context,omitempty"`
	TotalDuration      int64  `json:"total_duration,omitempty"`
	LoadDuration       int64  `json:"load_duration,omitempty"`
	PromptEvalCount    int    `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64  `json:"prompt_eval_duration,omitempty"`
	EvalCount          int    `json:"eval_count,omitempty"`
	EvalDuration       int64  `json:"eval_duration,omitempty"`
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents a chat request
type ChatRequest struct {
	Model     string                 `json:"model"`
	Messages  []ChatMessage          `json:"messages"`
	Stream    bool                   `json:"stream"`
	KeepAlive string                 `json:"keep_alive,omitempty"`
	Options   map[string]interface{} `json:"options,omitempty"`
}

// ChatResponse represents a chat response
type ChatResponse struct {
	Model              string      `json:"model"`
	Message            ChatMessage `json:"message"`
	Done               bool        `json:"done"`
	CreatedAt          time.Time   `json:"created_at"`
	TotalDuration      int64       `json:"total_duration,omitempty"`
	LoadDuration       int64       `json:"load_duration,omitempty"`
	PromptEvalCount    int         `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64       `json:"prompt_eval_duration,omitempty"`
	EvalCount          int         `json:"eval_count,omitempty"`
	EvalDuration       int64       `json:"eval_duration,omitempty"`
}

// EmbeddingRequest represents an embedding request
type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// EmbeddingResponse represents an embedding response
type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

// OllamaHealth represents health status
type OllamaHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	ServiceRunning  bool     `json:"service_running"`
	ModelsAvailable int      `json:"models_available"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// DefaultOllamaBaseURL returns the configured Ollama base URL.
func DefaultOllamaBaseURL() string {
	baseURL := strings.TrimSpace(os.Getenv("OLLAMA_BASE_URL"))
	if baseURL != "" {
		return strings.TrimRight(baseURL, "/")
	}

	host := strings.TrimSpace(os.Getenv("OLLAMA_HOST"))
	if host == "" {
		return "http://127.0.0.1:11434"
	}

	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		return strings.TrimRight(host, "/")
	}

	port := strings.TrimSpace(os.Getenv("OLLAMA_PORT"))
	if port == "" {
		port = "11434"
	}

	return fmt.Sprintf("http://%s:%s", host, port)
}

// DefaultOllamaChatModel returns the default local chat model.
func DefaultOllamaChatModel() string {
	if model := strings.TrimSpace(os.Getenv("OLLAMA_CHAT_MODEL")); model != "" {
		return model
	}
	return "qwen3:8b"
}

func defaultOllamaModel(envKey, fallback string) string {
	if model := strings.TrimSpace(os.Getenv(envKey)); model != "" {
		return model
	}
	return fallback
}

// DefaultOllamaFastModel returns the default local fast coding model.
func DefaultOllamaFastModel() string {
	return defaultOllamaModel("OLLAMA_FAST_MODEL", "qwen2.5-coder:7b")
}

// DefaultOllamaCodeModel returns the default local benchmark-ranked coding model.
func DefaultOllamaCodeModel() string {
	return defaultOllamaModel("OLLAMA_CODE_MODEL", "devstral-small-2")
}

// DefaultOllamaHeavyCodeModel returns the deeper offload-heavy local coding model.
func DefaultOllamaHeavyCodeModel() string {
	return defaultOllamaModel("OLLAMA_HEAVY_CODE_MODEL", "devstral-2")
}

// DefaultOllamaHighContextCodeModel returns the larger-context local coding model.
func DefaultOllamaHighContextCodeModel() string {
	return defaultOllamaModel("OLLAMA_HIGH_CONTEXT_CODE_MODEL", "qwen3-coder-next")
}

// DefaultOllamaEmbedModel returns the default local embedding model.
func DefaultOllamaEmbedModel() string {
	if model := strings.TrimSpace(os.Getenv("OLLAMA_EMBED_MODEL")); model != "" {
		return model
	}
	return "nomic-embed-text:v1.5"
}

// DefaultOllamaKeepAlive returns the workstation-standard model residency duration.
func DefaultOllamaKeepAlive() string {
	if value := strings.TrimSpace(os.Getenv("OLLAMA_KEEP_ALIVE")); value != "" {
		return value
	}
	return "15m"
}

// RecommendedOllamaModels returns the workstation-standard local model set.
func RecommendedOllamaModels() []string {
	models := []string{
		DefaultOllamaChatModel(),
		DefaultOllamaFastModel(),
		DefaultOllamaCodeModel(),
		DefaultOllamaHeavyCodeModel(),
		DefaultOllamaHighContextCodeModel(),
		DefaultOllamaEmbedModel(),
	}

	seen := make(map[string]struct{}, len(models))
	deduped := make([]string, 0, len(models))
	for _, model := range models {
		if _, ok := seen[model]; ok {
			continue
		}
		seen[model] = struct{}{}
		deduped = append(deduped, model)
	}

	return deduped
}

// RecommendedOllamaPullCommands returns pull commands for the standard model set.
func RecommendedOllamaPullCommands() []string {
	models := RecommendedOllamaModels()
	commands := make([]string, 0, len(models))
	for _, model := range models {
		commands = append(commands, fmt.Sprintf("ollama pull %s", model))
	}
	return commands
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient() (*OllamaClient, error) {
	return &OllamaClient{
		baseURL:    DefaultOllamaBaseURL(),
		httpClient: httpclient.WithTimeout(5 * time.Minute),
	}, nil
}

// GetStatus returns the service status
func (c *OllamaClient) GetStatus(ctx context.Context) (*OllamaStatus, error) {
	status := &OllamaStatus{
		BaseURL: c.baseURL,
	}

	// Check if service is running
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/version", nil)
	if err != nil {
		return status, nil
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		status.Running = false
		return status, nil
	}
	defer resp.Body.Close()

	status.Running = true

	// Parse version
	var versionResp struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&versionResp); err == nil {
		status.Version = versionResp.Version
	}

	// Get models
	models, err := c.ListModels(ctx)
	if err == nil {
		status.Models = models
	}

	return status, nil
}

// ListModels returns available models
func (c *OllamaClient) ListModels(ctx context.Context) ([]OllamaModel, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API error: %s", string(body))
	}

	var result struct {
		Models []OllamaModel `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Models, nil
}

// Generate generates text completion
func (c *OllamaClient) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	if req.Model == "" {
		req.Model = DefaultOllamaChatModel()
	}
	req.Stream = false
	if req.KeepAlive == "" {
		req.KeepAlive = DefaultOllamaKeepAlive()
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API error: %s", string(respBody))
	}

	var result GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// Chat performs a chat completion
func (c *OllamaClient) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = DefaultOllamaChatModel()
	}
	req.Stream = false
	if req.KeepAlive == "" {
		req.KeepAlive = DefaultOllamaKeepAlive()
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API error: %s", string(respBody))
	}

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// Embed generates embeddings for text
func (c *OllamaClient) Embed(ctx context.Context, model, prompt string) (*EmbeddingResponse, error) {
	if model == "" {
		model = DefaultOllamaEmbedModel()
	}

	req := EmbeddingRequest{
		Model:  model,
		Prompt: prompt,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API error: %s", string(respBody))
	}

	var result EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// PullModel downloads a model
func (c *OllamaClient) PullModel(ctx context.Context, modelName string) error {
	req := map[string]interface{}{
		"name":   modelName,
		"stream": false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/pull", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Ollama API error: %s", string(respBody))
	}

	return nil
}

// DeleteModel removes a model
func (c *OllamaClient) DeleteModel(ctx context.Context, modelName string) error {
	req := map[string]string{
		"name": modelName,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", c.baseURL+"/api/delete", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Ollama API error: %s", string(respBody))
	}

	return nil
}

// GetHealth returns health status
func (c *OllamaClient) GetHealth(ctx context.Context) (*OllamaHealth, error) {
	health := &OllamaHealth{
		Score:  100,
		Status: "healthy",
	}

	// Check if service is running
	status, err := c.GetStatus(ctx)
	if err != nil || !status.Running {
		health.ServiceRunning = false
		health.Score -= 50
		health.Issues = append(health.Issues, "Ollama service not running")
		health.Recommendations = append(health.Recommendations,
			"Start Ollama: ollama serve")
	} else {
		health.ServiceRunning = true
		health.ModelsAvailable = len(status.Models)

		if len(status.Models) == 0 {
			health.Score -= 30
			health.Issues = append(health.Issues, "No models installed")
			health.Recommendations = append(health.Recommendations,
				fmt.Sprintf("Pull a model set: %s", strings.Join(RecommendedOllamaPullCommands(), " && ")))
		}
	}

	switch {
	case health.Score >= 80:
		health.Status = "healthy"
	case health.Score >= 50:
		health.Status = "degraded"
	default:
		health.Status = "critical"
	}

	return health, nil
}

// GetModelInfo returns detailed info about a model
func (c *OllamaClient) GetModelInfo(ctx context.Context, modelName string) (map[string]interface{}, error) {
	req := map[string]string{
		"name": modelName,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/show", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API error: %s", string(respBody))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}
