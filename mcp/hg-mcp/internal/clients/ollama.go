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

type ollamaStatusSnapshotOptions struct {
	includeRunningModels bool
}

// OllamaModelDetails represents shared model metadata returned by tags/show/ps.
type OllamaModelDetails struct {
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

// OllamaModel represents an available model
type OllamaModel struct {
	Name       string             `json:"name"`
	ModifiedAt time.Time          `json:"modified_at"`
	Size       int64              `json:"size"`
	Digest     string             `json:"digest"`
	Details    OllamaModelDetails `json:"details"`
}

// OllamaStatus represents service status
type OllamaStatus struct {
	Running     bool          `json:"running"`
	Version     string        `json:"version,omitempty"`
	Models      []OllamaModel `json:"models,omitempty"`
	LoadedModel string        `json:"loaded_model,omitempty"`
	BaseURL     string        `json:"base_url"`
}

// OllamaRunningModel represents a loaded model returned by /api/ps.
type OllamaRunningModel struct {
	Name          string             `json:"name"`
	Model         string             `json:"model"`
	Size          int64              `json:"size"`
	Digest        string             `json:"digest"`
	Details       OllamaModelDetails `json:"details"`
	ExpiresAt     string             `json:"expires_at,omitempty"`
	SizeVRAM      int64              `json:"size_vram,omitempty"`
	ContextLength int                `json:"context_length,omitempty"`
}

// OllamaToolFunction describes a tool definition or tool call function.
type OllamaToolFunction struct {
	Index       int                    `json:"index,omitempty"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Arguments   map[string]interface{} `json:"arguments,omitempty"`
}

// OllamaTool describes a tool schema or a tool call.
type OllamaTool struct {
	Type     string             `json:"type"`
	Function OllamaToolFunction `json:"function"`
}

// GenerateRequest represents a generation request
type GenerateRequest struct {
	Model     string                 `json:"model"`
	Prompt    string                 `json:"prompt"`
	System    string                 `json:"system,omitempty"`
	Stream    bool                   `json:"stream"`
	KeepAlive string                 `json:"keep_alive,omitempty"`
	Format    interface{}            `json:"format,omitempty"`
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
	Role      string       `json:"role"`
	Content   string       `json:"content"`
	ToolName  string       `json:"tool_name,omitempty"`
	ToolCalls []OllamaTool `json:"tool_calls,omitempty"`
}

// ChatRequest represents a chat request
type ChatRequest struct {
	Model     string                 `json:"model"`
	Messages  []ChatMessage          `json:"messages"`
	Stream    bool                   `json:"stream"`
	KeepAlive string                 `json:"keep_alive,omitempty"`
	Format    interface{}            `json:"format,omitempty"`
	Tools     []OllamaTool           `json:"tools,omitempty"`
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
	BaseURL         string   `json:"base_url,omitempty"`
	Version         string   `json:"version,omitempty"`
	RequiredModels  []string `json:"required_models,omitempty"`
	ReadyModels     []string `json:"ready_models,omitempty"`
	MissingModels   []string `json:"missing_models,omitempty"`
	PullCommands    []string `json:"pull_commands,omitempty"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// OllamaAliasStatus captures whether a managed alias resolves cleanly.
type OllamaAliasStatus struct {
	Alias  string `json:"alias"`
	Source string `json:"source"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

// OllamaReadiness captures machine-readable daemon/model readiness.
type OllamaReadiness struct {
	BaseURL        string              `json:"base_url"`
	Version        string              `json:"version,omitempty"`
	Reachable      bool                `json:"reachable"`
	Ready          bool                `json:"ready"`
	RequireHeavy   bool                `json:"require_heavy"`
	RequiredModels []string            `json:"required_models"`
	ReadyModels    []string            `json:"ready_models,omitempty"`
	MissingModels  []string            `json:"missing_models,omitempty"`
	AliasChecks    []OllamaAliasStatus `json:"alias_checks,omitempty"`
	PullCommands   []string            `json:"pull_commands,omitempty"`
	Error          string              `json:"error,omitempty"`
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
	return "code-primary"
}

func defaultOllamaModel(envKey, fallback string) string {
	if model := strings.TrimSpace(os.Getenv(envKey)); model != "" {
		return model
	}
	return fallback
}

// DefaultOllamaFastModel returns the default local fast coding model.
func DefaultOllamaFastModel() string {
	return defaultOllamaModel("OLLAMA_FAST_MODEL", "code-fast")
}

// DefaultOllamaCodeModel returns the default local benchmark-ranked coding model.
func DefaultOllamaCodeModel() string {
	return defaultOllamaModel("OLLAMA_CODE_MODEL", "code-primary")
}

// DefaultOllamaHeavyCodeModel returns the deeper offload-heavy local coding model.
func DefaultOllamaHeavyCodeModel() string {
	return defaultOllamaModel("OLLAMA_HEAVY_CODE_MODEL", "code-heavy")
}

// DefaultOllamaHighContextCodeModel returns the larger-context local coding model.
func DefaultOllamaHighContextCodeModel() string {
	return defaultOllamaModel("OLLAMA_HIGH_CONTEXT_CODE_MODEL", "code-long")
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
	seen := make(map[string]struct{}, len(models))
	for _, model := range models {
		command := fmt.Sprintf("ollama pull %s", pullableOllamaModelName(model))
		if _, ok := seen[command]; ok {
			continue
		}
		seen[command] = struct{}{}
		commands = append(commands, command)
	}
	return commands
}

func pullableOllamaModelName(model string) string {
	if source := ollamaAliasSourceModel(model); source != "" {
		return source
	}
	return strings.TrimSpace(model)
}

func ollamaAliasSourceModel(model string) string {
	switch strings.TrimSpace(model) {
	case "code-fast", "code-compact":
		return "qwen2.5-coder:7b"
	case "code-primary", "code-reasoner":
		return "devstral-small-2"
	case "code-long":
		return "qwen3-coder-next"
	case "code-heavy":
		return "devstral-2"
	default:
		return ""
	}
}

func ollamaModelInstalled(model string, installed map[string]struct{}) bool {
	model = strings.TrimSpace(model)
	if model == "" {
		return false
	}
	if _, ok := installed[model]; ok {
		return true
	}
	if strings.HasSuffix(model, ":latest") {
		_, ok := installed[strings.TrimSuffix(model, ":latest")]
		return ok
	}
	if !strings.Contains(model, ":") {
		_, ok := installed[model+":latest"]
		return ok
	}
	return false
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient() (*OllamaClient, error) {
	return &OllamaClient{
		baseURL:    DefaultOllamaBaseURL(),
		httpClient: httpclient.WithTimeout(5 * time.Minute),
	}, nil
}

func (c *OllamaClient) fetchStatusSnapshot(ctx context.Context, opts ollamaStatusSnapshotOptions) (*OllamaStatus, error) {
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

	if opts.includeRunningModels {
		runningModels, err := c.ListRunningModels(ctx)
		if err == nil && len(runningModels) > 0 {
			status.LoadedModel = runningModels[0].Name
		}
	}

	return status, nil
}

// GetStatus returns the service status
func (c *OllamaClient) GetStatus(ctx context.Context) (*OllamaStatus, error) {
	return c.fetchStatusSnapshot(ctx, ollamaStatusSnapshotOptions{includeRunningModels: true})
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

// ListRunningModels returns models currently loaded in memory.
func (c *OllamaClient) ListRunningModels(ctx context.Context) ([]OllamaRunningModel, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/ps", nil)
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
		Models []OllamaRunningModel `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Models, nil
}

// Generate generates text completion
func (c *OllamaClient) Generate(ctx context.Context, req *GenerateRequest) (result *GenerateResponse, err error) {
	if req.Model == "" {
		req.Model = DefaultOllamaChatModel()
	}
	req.Stream = false
	if req.KeepAlive == "" {
		req.KeepAlive = DefaultOllamaKeepAlive()
	}
	ctx, span := startOllamaLLMSpan(ctx, "generate", req.Model, c.baseURL)
	defer func() {
		model := req.Model
		inputTokens := 0
		outputTokens := 0
		if result != nil {
			if result.Model != "" {
				model = result.Model
			}
			inputTokens = result.PromptEvalCount
			outputTokens = result.EvalCount
		}
		finishOllamaLLMSpan(span, model, inputTokens, outputTokens, err)
	}()

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

	var decoded GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result = &decoded
	return result, nil
}

// Chat performs a chat completion
func (c *OllamaClient) Chat(ctx context.Context, req *ChatRequest) (result *ChatResponse, err error) {
	if req.Model == "" {
		req.Model = DefaultOllamaChatModel()
	}
	req.Stream = false
	if req.KeepAlive == "" {
		req.KeepAlive = DefaultOllamaKeepAlive()
	}
	ctx, span := startOllamaLLMSpan(ctx, "chat", req.Model, c.baseURL)
	defer func() {
		model := req.Model
		inputTokens := 0
		outputTokens := 0
		if result != nil {
			if result.Model != "" {
				model = result.Model
			}
			inputTokens = result.PromptEvalCount
			outputTokens = result.EvalCount
		}
		finishOllamaLLMSpan(span, model, inputTokens, outputTokens, err)
	}()

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

	var decoded ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result = &decoded
	return result, nil
}

// Embed generates embeddings for text
func (c *OllamaClient) Embed(ctx context.Context, model, prompt string) (result *EmbeddingResponse, err error) {
	if model == "" {
		model = DefaultOllamaEmbedModel()
	}
	ctx, span := startOllamaLLMSpan(ctx, "embed", model, c.baseURL)
	defer func() {
		finishOllamaLLMSpan(span, model, 0, 0, err)
	}()

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

	var decoded EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result = &decoded
	return result, nil
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

func (c *OllamaClient) buildReadinessFromStatus(status *OllamaStatus, requireHeavy bool) *OllamaReadiness {
	readiness := &OllamaReadiness{
		BaseURL:      c.baseURL,
		RequireHeavy: requireHeavy,
		PullCommands: RecommendedOllamaPullCommands(),
	}
	if status == nil {
		readiness.Error = "Ollama status unavailable"
		return readiness
	}

	readiness.Reachable = status.Running
	readiness.Version = status.Version
	if !status.Running {
		readiness.Error = "Ollama service not running"
		return readiness
	}

	required := []string{
		DefaultOllamaChatModel(),
		DefaultOllamaFastModel(),
		DefaultOllamaEmbedModel(),
	}
	if requireHeavy {
		required = append(required, DefaultOllamaHeavyCodeModel())
	}
	readiness.RequiredModels = append(readiness.RequiredModels, required...)

	installed := make(map[string]struct{}, len(status.Models)*2)
	for _, model := range status.Models {
		if model.Name != "" {
			installed[strings.TrimSpace(model.Name)] = struct{}{}
		}
	}
	for _, model := range required {
		if ollamaModelInstalled(model, installed) {
			readiness.ReadyModels = append(readiness.ReadyModels, model)
		} else {
			readiness.MissingModels = append(readiness.MissingModels, model)
		}
	}

	for _, pair := range []struct {
		alias  string
		source string
	}{
		{alias: "code-fast", source: "qwen2.5-coder:7b"},
		{alias: "code-compact", source: "qwen2.5-coder:7b"},
		{alias: "code-primary", source: "devstral-small-2"},
		{alias: "code-reasoner", source: "devstral-small-2"},
		{alias: "code-long", source: "qwen3-coder-next"},
		{alias: "code-heavy", source: "devstral-2"},
	} {
		check := OllamaAliasStatus{Alias: pair.alias, Source: pair.source, Status: "skipped"}
		if ollamaModelInstalled(pair.source, installed) {
			if !ollamaModelInstalled(pair.alias, installed) {
				check.Status = "missing_alias"
				check.Detail = "backing model is installed but the managed alias is missing"
			} else {
				check.Status = "ok"
			}
		} else if ollamaModelInstalled(pair.alias, installed) {
			check.Status = "alias_only"
			check.Detail = "alias is installed but the backing model is absent"
		}
		readiness.AliasChecks = append(readiness.AliasChecks, check)
	}

	readiness.Ready = readiness.Reachable && len(readiness.MissingModels) == 0
	for _, check := range readiness.AliasChecks {
		if check.Status == "missing_alias" || check.Status == "digest_mismatch" {
			readiness.Ready = false
			break
		}
	}
	if !readiness.Ready && readiness.Error == "" {
		switch {
		case len(readiness.MissingModels) > 0:
			readiness.Error = fmt.Sprintf("missing required models: %s", strings.Join(readiness.MissingModels, ", "))
		default:
			for _, check := range readiness.AliasChecks {
				if check.Status == "missing_alias" || check.Status == "digest_mismatch" {
					readiness.Error = fmt.Sprintf("managed alias problem for %s: %s", check.Alias, check.Detail)
					break
				}
			}
		}
	}

	return readiness
}

// GetHealth returns health status
func (c *OllamaClient) GetHealth(ctx context.Context) (*OllamaHealth, error) {
	health := &OllamaHealth{
		Score:        100,
		Status:       "healthy",
		BaseURL:      c.baseURL,
		PullCommands: RecommendedOllamaPullCommands(),
	}

	status, err := c.fetchStatusSnapshot(ctx, ollamaStatusSnapshotOptions{})
	if err != nil {
		return nil, err
	}
	readiness := c.buildReadinessFromStatus(status, false)
	health.Version = readiness.Version
	health.RequiredModels = append([]string(nil), readiness.RequiredModels...)
	health.ReadyModels = append([]string(nil), readiness.ReadyModels...)
	health.MissingModels = append([]string(nil), readiness.MissingModels...)

	if !status.Running || !readiness.Reachable {
		health.ServiceRunning = false
		health.Score -= 50
		health.Issues = append(health.Issues, "Ollama service not running")
		health.Recommendations = append(health.Recommendations,
			"Start Ollama: ollama serve")
	} else {
		health.ServiceRunning = true
		health.ModelsAvailable = len(status.Models)
	}
	if len(status.Models) == 0 {
		health.Score -= 30
		health.Issues = append(health.Issues, "No models installed")
		health.Recommendations = append(health.Recommendations,
			fmt.Sprintf("Pull a model set: %s", strings.Join(RecommendedOllamaPullCommands(), " && ")))
	}
	if len(readiness.MissingModels) > 0 {
		health.Score -= 25
		health.Issues = append(health.Issues, fmt.Sprintf("Missing required models: %s", strings.Join(readiness.MissingModels, ", ")))
		health.Recommendations = append(health.Recommendations,
			fmt.Sprintf("Pull the missing model set: %s", strings.Join(RecommendedOllamaPullCommands(), " && ")))
	}
	for _, check := range readiness.AliasChecks {
		switch check.Status {
		case "missing_alias", "digest_mismatch":
			health.Score -= 10
			health.Issues = append(health.Issues, fmt.Sprintf("Managed alias problem for %s: %s", check.Alias, check.Detail))
			health.Recommendations = append(health.Recommendations,
				"Rebuild the managed aliases: ~/hairglasses-studio/dotfiles/scripts/hg-ollama-sync-aliases.sh")
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

// GetReadiness returns the live daemon/model readiness snapshot.
func (c *OllamaClient) GetReadiness(ctx context.Context, requireHeavy bool) (*OllamaReadiness, error) {
	status, err := c.fetchStatusSnapshot(ctx, ollamaStatusSnapshotOptions{})
	if err != nil {
		readiness := &OllamaReadiness{
			BaseURL:      c.baseURL,
			RequireHeavy: requireHeavy,
			PullCommands: RecommendedOllamaPullCommands(),
			Error:        err.Error(),
		}
		return readiness, nil
	}
	return c.buildReadinessFromStatus(status, requireHeavy), nil
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
