// Package clients provides API clients for external services.
package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	httpclient "github.com/hairglasses-studio/mcpkit/client"
)

// WhisperClient provides audio transcription using Whisper
type WhisperClient struct {
	mode        string // "local" or "api"
	model       string
	apiKey      string
	apiBaseURL  string
	whisperPath string
	httpClient  *http.Client
}

// TranscriptionResult represents transcription output
type TranscriptionResult struct {
	Text     string              `json:"text"`
	Language string              `json:"language,omitempty"`
	Duration float64             `json:"duration,omitempty"`
	Segments []TranscriptSegment `json:"segments,omitempty"`
}

// TranscriptSegment represents a segment of the transcription
type TranscriptSegment struct {
	ID       int     `json:"id"`
	Start    float64 `json:"start"`
	End      float64 `json:"end"`
	Text     string  `json:"text"`
	NoSpeech float64 `json:"no_speech_prob,omitempty"`
}

// TranslationResult represents translation output
type TranslationResult struct {
	Text           string  `json:"text"`
	SourceLanguage string  `json:"source_language,omitempty"`
	TargetLanguage string  `json:"target_language"`
	Duration       float64 `json:"duration,omitempty"`
}

// WhisperStatus represents service status
type WhisperStatus struct {
	Mode            string   `json:"mode"`
	Model           string   `json:"model"`
	WhisperPath     string   `json:"whisper_path,omitempty"`
	APIConfigured   bool     `json:"api_configured,omitempty"`
	AvailableModels []string `json:"available_models"`
}

// WhisperHealth represents health status
type WhisperHealth struct {
	Score            int      `json:"score"`
	Status           string   `json:"status"`
	WhisperInstalled bool     `json:"whisper_installed"`
	APIConfigured    bool     `json:"api_configured"`
	Mode             string   `json:"mode"`
	Issues           []string `json:"issues,omitempty"`
	Recommendations  []string `json:"recommendations,omitempty"`
}

// NewWhisperClient creates a new Whisper client
func NewWhisperClient() (*WhisperClient, error) {
	// Check for OpenAI API key (for API mode)
	apiKey := os.Getenv("OPENAI_API_KEY")

	// Check for local whisper installation
	whisperPath := os.Getenv("WHISPER_PATH")
	if whisperPath == "" {
		// Try to find in PATH
		if path, err := exec.LookPath("whisper"); err == nil {
			whisperPath = path
		}
	}

	// Determine mode
	mode := os.Getenv("WHISPER_MODE")
	if mode == "" {
		if apiKey != "" {
			mode = "api"
		} else if whisperPath != "" {
			mode = "local"
		} else {
			mode = "local" // Default, will fail gracefully if not installed
		}
	}

	// Model selection
	model := os.Getenv("WHISPER_MODEL")
	if model == "" {
		model = "base" // Default model
	}

	return &WhisperClient{
		mode:        mode,
		model:       model,
		apiKey:      apiKey,
		apiBaseURL:  "https://api.openai.com/v1",
		whisperPath: whisperPath,
		httpClient: httpclient.WithTimeout(10 * time.Minute),
	}, nil
}

// GetStatus returns service status
func (c *WhisperClient) GetStatus(ctx context.Context) (*WhisperStatus, error) {
	status := &WhisperStatus{
		Mode:          c.mode,
		Model:         c.model,
		WhisperPath:   c.whisperPath,
		APIConfigured: c.apiKey != "",
		AvailableModels: []string{
			"tiny", "tiny.en",
			"base", "base.en",
			"small", "small.en",
			"medium", "medium.en",
			"large", "large-v2", "large-v3",
		},
	}

	return status, nil
}

// Transcribe transcribes audio to text
func (c *WhisperClient) Transcribe(ctx context.Context, filePath string, language string) (*TranscriptionResult, error) {
	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	if c.mode == "api" {
		return c.transcribeAPI(ctx, filePath, language)
	}
	return c.transcribeLocal(ctx, filePath, language)
}

// transcribeAPI uses OpenAI Whisper API
func (c *WhisperClient) transcribeAPI(ctx context.Context, filePath string, language string) (*TranscriptionResult, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}

	// Add model
	if err := writer.WriteField("model", "whisper-1"); err != nil {
		return nil, err
	}

	// Add language if specified
	if language != "" {
		if err := writer.WriteField("language", language); err != nil {
			return nil, err
		}
	}

	// Add response format
	if err := writer.WriteField("response_format", "verbose_json"); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", c.apiBaseURL+"/audio/transcriptions", &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	// Parse response
	var apiResp struct {
		Text     string  `json:"text"`
		Language string  `json:"language"`
		Duration float64 `json:"duration"`
		Segments []struct {
			ID           int     `json:"id"`
			Start        float64 `json:"start"`
			End          float64 `json:"end"`
			Text         string  `json:"text"`
			NoSpeechProb float64 `json:"no_speech_prob"`
		} `json:"segments"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result := &TranscriptionResult{
		Text:     apiResp.Text,
		Language: apiResp.Language,
		Duration: apiResp.Duration,
	}

	for _, seg := range apiResp.Segments {
		result.Segments = append(result.Segments, TranscriptSegment{
			ID:       seg.ID,
			Start:    seg.Start,
			End:      seg.End,
			Text:     seg.Text,
			NoSpeech: seg.NoSpeechProb,
		})
	}

	return result, nil
}

// transcribeLocal uses local whisper installation
func (c *WhisperClient) transcribeLocal(ctx context.Context, filePath string, language string) (*TranscriptionResult, error) {
	if c.whisperPath == "" {
		return nil, fmt.Errorf("whisper not installed - install with: pip install openai-whisper")
	}

	// Create temp output directory
	tmpDir, err := os.MkdirTemp("", "whisper-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Build command
	args := []string{
		filePath,
		"--model", c.model,
		"--output_dir", tmpDir,
		"--output_format", "json",
	}

	if language != "" {
		args = append(args, "--language", language)
	}

	cmd := exec.CommandContext(ctx, c.whisperPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("whisper failed: %w - output: %s", err, string(output))
	}

	// Find and read output JSON
	baseName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	jsonPath := filepath.Join(tmpDir, baseName+".json")

	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read output: %w", err)
	}

	var whisperOutput struct {
		Text     string `json:"text"`
		Language string `json:"language"`
		Segments []struct {
			ID           int     `json:"id"`
			Start        float64 `json:"start"`
			End          float64 `json:"end"`
			Text         string  `json:"text"`
			NoSpeechProb float64 `json:"no_speech_prob"`
		} `json:"segments"`
	}

	if err := json.Unmarshal(jsonData, &whisperOutput); err != nil {
		return nil, fmt.Errorf("failed to parse output: %w", err)
	}

	result := &TranscriptionResult{
		Text:     whisperOutput.Text,
		Language: whisperOutput.Language,
	}

	for _, seg := range whisperOutput.Segments {
		result.Segments = append(result.Segments, TranscriptSegment{
			ID:       seg.ID,
			Start:    seg.Start,
			End:      seg.End,
			Text:     seg.Text,
			NoSpeech: seg.NoSpeechProb,
		})
	}

	// Calculate duration from last segment
	if len(result.Segments) > 0 {
		result.Duration = result.Segments[len(result.Segments)-1].End
	}

	return result, nil
}

// Translate transcribes and translates audio to English
func (c *WhisperClient) Translate(ctx context.Context, filePath string) (*TranslationResult, error) {
	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	if c.mode == "api" {
		return c.translateAPI(ctx, filePath)
	}
	return c.translateLocal(ctx, filePath)
}

// translateAPI uses OpenAI Whisper translation API
func (c *WhisperClient) translateAPI(ctx context.Context, filePath string) (*TranslationResult, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}

	// Add model
	if err := writer.WriteField("model", "whisper-1"); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", c.apiBaseURL+"/audio/translations", &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var apiResp struct {
		Text string `json:"text"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &TranslationResult{
		Text:           apiResp.Text,
		TargetLanguage: "en",
	}, nil
}

// translateLocal uses local whisper with translation task
func (c *WhisperClient) translateLocal(ctx context.Context, filePath string) (*TranslationResult, error) {
	if c.whisperPath == "" {
		return nil, fmt.Errorf("whisper not installed")
	}

	// Create temp output directory
	tmpDir, err := os.MkdirTemp("", "whisper-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Build command with translation task
	args := []string{
		filePath,
		"--model", c.model,
		"--output_dir", tmpDir,
		"--output_format", "json",
		"--task", "translate",
	}

	cmd := exec.CommandContext(ctx, c.whisperPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("whisper failed: %w - output: %s", err, string(output))
	}

	// Find and read output JSON
	baseName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	jsonPath := filepath.Join(tmpDir, baseName+".json")

	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read output: %w", err)
	}

	var whisperOutput struct {
		Text     string `json:"text"`
		Language string `json:"language"`
	}

	if err := json.Unmarshal(jsonData, &whisperOutput); err != nil {
		return nil, fmt.Errorf("failed to parse output: %w", err)
	}

	return &TranslationResult{
		Text:           whisperOutput.Text,
		SourceLanguage: whisperOutput.Language,
		TargetLanguage: "en",
	}, nil
}

// GetModels returns available models
func (c *WhisperClient) GetModels(ctx context.Context) ([]map[string]interface{}, error) {
	models := []map[string]interface{}{
		{"name": "tiny", "size": "39M", "english_only": false, "description": "Fastest, least accurate"},
		{"name": "tiny.en", "size": "39M", "english_only": true, "description": "English-only tiny model"},
		{"name": "base", "size": "74M", "english_only": false, "description": "Good balance of speed and accuracy"},
		{"name": "base.en", "size": "74M", "english_only": true, "description": "English-only base model"},
		{"name": "small", "size": "244M", "english_only": false, "description": "Better accuracy, slower"},
		{"name": "small.en", "size": "244M", "english_only": true, "description": "English-only small model"},
		{"name": "medium", "size": "769M", "english_only": false, "description": "High accuracy"},
		{"name": "medium.en", "size": "769M", "english_only": true, "description": "English-only medium model"},
		{"name": "large", "size": "1550M", "english_only": false, "description": "Original large model"},
		{"name": "large-v2", "size": "1550M", "english_only": false, "description": "Improved large model"},
		{"name": "large-v3", "size": "1550M", "english_only": false, "description": "Latest large model, best accuracy"},
	}

	return models, nil
}

// SetModel sets the model to use
func (c *WhisperClient) SetModel(model string) {
	c.model = model
}

// GetHealth returns health status
func (c *WhisperClient) GetHealth(ctx context.Context) (*WhisperHealth, error) {
	health := &WhisperHealth{
		Score:  100,
		Status: "healthy",
		Mode:   c.mode,
	}

	// Check whisper installation
	if c.whisperPath == "" {
		health.WhisperInstalled = false
		if c.mode == "local" {
			health.Score -= 50
			health.Issues = append(health.Issues, "whisper not installed")
			health.Recommendations = append(health.Recommendations,
				"Install whisper: pip install openai-whisper")
		}
	} else {
		health.WhisperInstalled = true

		// Test whisper
		cmd := exec.CommandContext(ctx, c.whisperPath, "--help")
		if err := cmd.Run(); err != nil {
			health.Score -= 25
			health.Issues = append(health.Issues, "whisper exists but failed to run")
		}
	}

	// Check API configuration
	if c.apiKey != "" {
		health.APIConfigured = true
	} else if c.mode == "api" {
		health.Score -= 50
		health.Issues = append(health.Issues, "OPENAI_API_KEY not set but API mode selected")
		health.Recommendations = append(health.Recommendations,
			"Set OPENAI_API_KEY environment variable")
	}

	// Recommend API if local not available
	if !health.WhisperInstalled && !health.APIConfigured {
		health.Score -= 20
		health.Recommendations = append(health.Recommendations,
			"Either install local whisper or set OPENAI_API_KEY for API mode")
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
