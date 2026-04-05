package samples

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
)

var samplesDir = func() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("Music", "Samples", "american-psycho")
	}
	return filepath.Join(home, "Music", "Samples", "american-psycho")
}()

var (
	hipToBeSquare = filepath.Join(samplesDir, "hip-to-be-square-intro.aiff")
	businessCard  = filepath.Join(samplesDir, "business-card.aiff")
	doYouLikeHuey = filepath.Join(samplesDir, "do-you-like-huey.aiff")
)

func requireFFmpeg(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not in PATH")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not in PATH")
	}
}

func requireSamples(t *testing.T) {
	t.Helper()
	for _, f := range []string{hipToBeSquare, businessCard, doYouLikeHuey} {
		if _, err := os.Stat(f); err != nil {
			t.Skipf("sample not found: %s", f)
		}
	}
}

func makeReq(args map[string]interface{}) mcp.CallToolRequest {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args
	return req
}

func getText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("result has no content")
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	return tc.Text
}

// --- Probe ---

func TestProbeHipToBeSquare(t *testing.T) {
	requireFFmpeg(t)
	requireSamples(t)

	result, err := handleProbe(context.Background(), makeReq(map[string]interface{}{"file_path": hipToBeSquare}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handler returned error: %s", getText(t, result))
	}

	text := getText(t, result)
	t.Logf("Probe output:\n%s", text)

	for _, want := range []string{"9.", "192000", "pcm_s16be"} {
		if !strings.Contains(text, want) {
			t.Errorf("probe output missing %q", want)
		}
	}
}

func TestProbeBusinessCard(t *testing.T) {
	requireFFmpeg(t)
	requireSamples(t)

	result, err := handleProbe(context.Background(), makeReq(map[string]interface{}{"file_path": businessCard}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handler returned error: %s", getText(t, result))
	}

	text := getText(t, result)
	t.Logf("Probe output:\n%s", text)

	if !strings.Contains(text, "9.") {
		t.Error("expected duration ~9.3s")
	}
}

func TestProbeLongestSample(t *testing.T) {
	requireFFmpeg(t)
	requireSamples(t)

	result, err := handleProbe(context.Background(), makeReq(map[string]interface{}{"file_path": doYouLikeHuey}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handler returned error: %s", getText(t, result))
	}

	text := getText(t, result)
	t.Logf("Probe output:\n%s", text)

	if !strings.Contains(text, "14.") {
		t.Error("expected duration ~14.9s")
	}
}

func TestProbeMissingFile(t *testing.T) {
	requireFFmpeg(t)

	result, err := handleProbe(context.Background(), makeReq(map[string]interface{}{"file_path": "/nonexistent/file.aiff"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for nonexistent file")
	}
}

// --- Silence Detection ---

func TestDetectSilence(t *testing.T) {
	requireFFmpeg(t)
	requireSamples(t)

	result, err := handleDetectSilence(context.Background(), makeReq(map[string]interface{}{"audio_path": hipToBeSquare}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handler returned error: %s", getText(t, result))
	}

	text := getText(t, result)
	t.Logf("Silence output:\n%s", text)

	if !strings.Contains(text, "Silence") {
		t.Error("output missing silence heading")
	}
}

// --- BPM Detection ---

func TestDetectBPM(t *testing.T) {
	requireFFmpeg(t)
	requireSamples(t)

	result, err := handleDetectBPM(context.Background(), makeReq(map[string]interface{}{"audio_path": doYouLikeHuey}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handler returned error: %s", getText(t, result))
	}

	text := getText(t, result)
	t.Logf("BPM output:\n%s", text)

	if !strings.Contains(text, "BPM") {
		t.Error("output missing BPM heading")
	}
}

// --- Waveform ---

func TestWaveformText(t *testing.T) {
	requireFFmpeg(t)
	requireSamples(t)

	result, err := handleWaveform(context.Background(), makeReq(map[string]interface{}{"audio_path": businessCard}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handler returned error: %s", getText(t, result))
	}

	text := getText(t, result)
	t.Logf("Waveform output:\n%s", text)

	if !strings.Contains(text, "Waveform") {
		t.Error("output missing Waveform heading")
	}

	hasBlock := false
	for _, ch := range []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"} {
		if strings.Contains(text, ch) {
			hasBlock = true
			break
		}
	}
	if !hasBlock {
		t.Error("output missing block characters")
	}
}

func TestWaveformJSON(t *testing.T) {
	requireFFmpeg(t)
	requireSamples(t)

	result, err := handleWaveform(context.Background(), makeReq(map[string]interface{}{
		"audio_path": businessCard,
		"format":     "json",
		"resolution": 40.0,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handler returned error: %s", getText(t, result))
	}

	text := getText(t, result)
	t.Logf("Waveform JSON:\n%.500s", text)

	var parsed struct {
		Peaks      []float64 `json:"peaks"`
		Duration   float64   `json:"duration"`
		Resolution int       `json:"resolution"`
	}
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(parsed.Peaks) != 40 {
		t.Errorf("expected 40 peaks, got %d", len(parsed.Peaks))
	}
	if parsed.Duration < 9.0 || parsed.Duration > 10.0 {
		t.Errorf("expected duration ~9.3s, got %.1f", parsed.Duration)
	}
}

// --- Reverse ---

func TestReverse(t *testing.T) {
	requireFFmpeg(t)
	requireSamples(t)

	outPath := filepath.Join(t.TempDir(), "reversed.aiff")
	result, err := handleReverse(context.Background(), makeReq(map[string]interface{}{
		"audio_path":  hipToBeSquare,
		"output_path": outPath,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handler returned error: %s", getText(t, result))
	}

	text := getText(t, result)
	t.Logf("Reverse output:\n%s", text)

	if !strings.Contains(text, "Reversed") {
		t.Error("output missing Reversed heading")
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Errorf("output file not created: %v", err)
	}
}

// --- Pitch Shift ---

func TestPitchShiftUp(t *testing.T) {
	requireFFmpeg(t)
	requireSamples(t)

	outPath := filepath.Join(t.TempDir(), "pitched-up.aiff")
	result, err := handlePitchShift(context.Background(), makeReq(map[string]interface{}{
		"audio_path":  businessCard,
		"semitones":   3.0,
		"output_path": outPath,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handler returned error: %s", getText(t, result))
	}

	text := getText(t, result)
	t.Logf("Pitch shift up:\n%s", text)

	if !strings.Contains(text, "+3") {
		t.Error("output missing +3 semitone indicator")
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Errorf("output file not created: %v", err)
	}
}

func TestPitchShiftDown(t *testing.T) {
	requireFFmpeg(t)
	requireSamples(t)

	outPath := filepath.Join(t.TempDir(), "pitched-down.aiff")
	result, err := handlePitchShift(context.Background(), makeReq(map[string]interface{}{
		"audio_path":  businessCard,
		"semitones":   -5.0,
		"output_path": outPath,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handler returned error: %s", getText(t, result))
	}

	text := getText(t, result)
	t.Logf("Pitch shift down:\n%s", text)

	if !strings.Contains(text, "-5") {
		t.Error("output missing -5 semitone indicator")
	}
}

func TestPitchShiftOutOfRange(t *testing.T) {
	requireFFmpeg(t)
	requireSamples(t)

	result, err := handlePitchShift(context.Background(), makeReq(map[string]interface{}{
		"audio_path": businessCard,
		"semitones":  15.0,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for out-of-range semitones")
	}
}

// --- Time Stretch ---

func TestTimeStretchHalf(t *testing.T) {
	requireFFmpeg(t)
	requireSamples(t)

	outPath := filepath.Join(t.TempDir(), "half-speed.aiff")
	result, err := handleTimeStretch(context.Background(), makeReq(map[string]interface{}{
		"audio_path":  hipToBeSquare,
		"speed":       0.5,
		"output_path": outPath,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handler returned error: %s", getText(t, result))
	}

	t.Logf("Time stretch 0.5x:\n%s", getText(t, result))

	if _, err := os.Stat(outPath); err != nil {
		t.Errorf("output file not created: %v", err)
	}
}

func TestTimeStretchDouble(t *testing.T) {
	requireFFmpeg(t)
	requireSamples(t)

	outPath := filepath.Join(t.TempDir(), "double-speed.aiff")
	result, err := handleTimeStretch(context.Background(), makeReq(map[string]interface{}{
		"audio_path":  hipToBeSquare,
		"speed":       2.0,
		"output_path": outPath,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handler returned error: %s", getText(t, result))
	}

	t.Logf("Time stretch 2.0x:\n%s", getText(t, result))

	if _, err := os.Stat(outPath); err != nil {
		t.Errorf("output file not created: %v", err)
	}
}

// --- Loop ---

func TestLoop(t *testing.T) {
	requireFFmpeg(t)
	requireSamples(t)

	outPath := filepath.Join(t.TempDir(), "looped.aiff")
	result, err := handleLoop(context.Background(), makeReq(map[string]interface{}{
		"audio_path":  businessCard,
		"repeats":     2.0,
		"output_path": outPath,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handler returned error: %s", getText(t, result))
	}

	text := getText(t, result)
	t.Logf("Loop output:\n%s", text)

	if !strings.Contains(text, "Loop") {
		t.Error("output missing Loop heading")
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Errorf("output file not created: %v", err)
	}
}

// --- Convert ---

func TestConvertToWav(t *testing.T) {
	requireFFmpeg(t)
	requireSamples(t)

	outPath := filepath.Join(t.TempDir(), "converted.wav")
	result, err := handleConvert(context.Background(), makeReq(map[string]interface{}{
		"audio_path":  businessCard,
		"format":      "wav",
		"output_path": outPath,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handler returned error: %s", getText(t, result))
	}

	text := getText(t, result)
	t.Logf("Convert WAV:\n%s", text)

	if !strings.Contains(strings.ToUpper(text), "WAV") {
		t.Error("output missing WAV format indicator")
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Errorf("output file not created: %v", err)
	}
}

func TestConvertToMp3(t *testing.T) {
	requireFFmpeg(t)
	requireSamples(t)

	outPath := filepath.Join(t.TempDir(), "converted.mp3")
	result, err := handleConvert(context.Background(), makeReq(map[string]interface{}{
		"audio_path":  businessCard,
		"format":      "mp3",
		"output_path": outPath,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handler returned error: %s", getText(t, result))
	}

	text := getText(t, result)
	t.Logf("Convert MP3:\n%s", text)

	if !strings.Contains(strings.ToUpper(text), "MP3") {
		t.Error("output missing MP3 format indicator")
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Errorf("output file not created: %v", err)
	}
}

func TestConvertToFlac(t *testing.T) {
	requireFFmpeg(t)
	requireSamples(t)

	outPath := filepath.Join(t.TempDir(), "converted.flac")
	result, err := handleConvert(context.Background(), makeReq(map[string]interface{}{
		"audio_path":  businessCard,
		"format":      "flac",
		"output_path": outPath,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handler returned error: %s", getText(t, result))
	}

	text := getText(t, result)
	t.Logf("Convert FLAC:\n%s", text)

	if !strings.Contains(strings.ToUpper(text), "FLAC") {
		t.Error("output missing FLAC format indicator")
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Errorf("output file not created: %v", err)
	}
}

// --- Clip & Process ---

func TestClip(t *testing.T) {
	requireFFmpeg(t)
	requireSamples(t)

	outPath := filepath.Join(t.TempDir(), "clipped.aiff")
	result, err := handleClip(context.Background(), makeReq(map[string]interface{}{
		"audio_path":  doYouLikeHuey,
		"start":       "3.0",
		"end":         "7.0",
		"output_path": outPath,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handler returned error: %s", getText(t, result))
	}

	text := getText(t, result)
	t.Logf("Clip output:\n%s", text)

	if _, err := os.Stat(outPath); err != nil {
		t.Errorf("output file not created: %v", err)
	}
}

func TestProcess(t *testing.T) {
	requireFFmpeg(t)
	requireSamples(t)

	// Clip first so we have a disposable file to process
	clipped := filepath.Join(t.TempDir(), "to-process.aiff")
	clipResult, err := handleClip(context.Background(), makeReq(map[string]interface{}{
		"audio_path":  doYouLikeHuey,
		"start":       "3.0",
		"end":         "7.0",
		"output_path": clipped,
	}))
	if err != nil {
		t.Fatalf("clip error: %v", err)
	}
	if clipResult.IsError {
		t.Fatalf("clip returned error: %s", getText(t, clipResult))
	}

	outPath := filepath.Join(t.TempDir(), "processed.aiff")
	result, err := handleProcess(context.Background(), makeReq(map[string]interface{}{
		"audio_path":  clipped,
		"output_path": outPath,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handler returned error: %s", getText(t, result))
	}

	text := getText(t, result)
	t.Logf("Process output:\n%s", text)

	if !strings.Contains(text, "Audio Processed") {
		t.Error("output missing 'Audio Processed' heading")
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Errorf("output file not created: %v", err)
	}
}

// --- Rekordbox XML ---

func TestRekordboxXML(t *testing.T) {
	requireFFmpeg(t)
	requireSamples(t)

	tmpDir := t.TempDir()

	// Clip a sample into the temp dir so we have audio to generate XML from
	clipped := filepath.Join(tmpDir, "sample.aiff")
	clipResult, err := handleClip(context.Background(), makeReq(map[string]interface{}{
		"audio_path":  businessCard,
		"start":       "0",
		"end":         "5.0",
		"output_path": clipped,
	}))
	if err != nil {
		t.Fatalf("clip error: %v", err)
	}
	if clipResult.IsError {
		t.Fatalf("clip returned error: %s", getText(t, clipResult))
	}

	result, err := handleRekordboxXML(context.Background(), makeReq(map[string]interface{}{
		"directory": tmpDir,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handler returned error: %s", getText(t, result))
	}

	text := getText(t, result)
	t.Logf("Rekordbox XML output:\n%.500s", text)

	if !strings.Contains(text, "Rekordbox") {
		t.Error("output missing Rekordbox heading")
	}

	// Find and parse the generated XML file
	xmlPath := filepath.Join(tmpDir, filepath.Base(tmpDir)+".xml")
	xmlData, err := os.ReadFile(xmlPath)
	if err != nil {
		// Try the default naming pattern
		entries, _ := os.ReadDir(tmpDir)
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".xml") {
				xmlPath = filepath.Join(tmpDir, e.Name())
				xmlData, err = os.ReadFile(xmlPath)
				break
			}
		}
	}
	if err != nil {
		t.Fatalf("could not read generated XML: %v", err)
	}
	var parsed interface{}
	if xmlErr := xml.Unmarshal(xmlData, &parsed); xmlErr != nil {
		t.Errorf("generated XML is not valid: %v", xmlErr)
	}
}

// --- List ---

func TestList(t *testing.T) {
	requireFFmpeg(t)
	requireSamples(t)

	result, err := handleList(context.Background(), makeReq(map[string]interface{}{
		"directory": filepath.Dir(samplesDir),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("handler returned error: %s", getText(t, result))
	}

	text := getText(t, result)
	t.Logf("List output:\n%s", text)

	if !strings.Contains(text, "american-psycho") {
		t.Error("output missing 'american-psycho' pack")
	}
}

// --- Module ---

func TestModuleInfo(t *testing.T) {
	testutil.AssertModuleValid(t, &Module{})
}
