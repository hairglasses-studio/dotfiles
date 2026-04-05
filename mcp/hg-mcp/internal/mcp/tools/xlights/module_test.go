package xlights

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
)

func init() {
	clients.TestOverrideXLightsClient = clients.NewTestXLightsClient()
}

func TestModuleInfo(t *testing.T) {
	testutil.AssertModuleValid(t, &Module{})
}

func TestHandleInfoMissingPath(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleInfo(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing path")
	}
}

func TestHandleFramesMissingPath(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleFrames(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing path")
	}
}

// buildTestFSEQ creates a minimal valid uncompressed FSEQ v2 file.
func buildTestFSEQ(channelCount uint32, frameCount uint32, stepTimeMs uint8) []byte {
	var buf bytes.Buffer

	// Magic
	buf.WriteString("PSEQ")
	// ChannelDataOffset (header is 32 bytes for v2 minimal)
	binary.Write(&buf, binary.LittleEndian, uint16(32))
	// MinorVersion
	buf.WriteByte(0)
	// MajorVersion
	buf.WriteByte(2)
	// HeaderLength
	binary.Write(&buf, binary.LittleEndian, uint16(32))
	// ChannelCount
	binary.Write(&buf, binary.LittleEndian, channelCount)
	// FrameCount
	binary.Write(&buf, binary.LittleEndian, frameCount)
	// StepTimeMs
	buf.WriteByte(stepTimeMs)
	// Flags
	buf.WriteByte(0)
	// CompressionType (0 = none)
	buf.WriteByte(0)
	// CompressionBlocks
	buf.WriteByte(0)
	// SparseRangeCount
	buf.WriteByte(0)
	// Reserved
	buf.WriteByte(0)
	// UUID (16 bytes of zeros)
	buf.Write(make([]byte, 16))

	// Frame data: each frame is channelCount bytes
	for i := uint32(0); i < frameCount; i++ {
		frame := make([]byte, channelCount)
		for j := uint32(0); j < channelCount; j++ {
			frame[j] = byte((i + j) % 256)
		}
		buf.Write(frame)
	}

	return buf.Bytes()
}

func TestParseHeader(t *testing.T) {
	data := buildTestFSEQ(512, 100, 50)
	path := filepath.Join(t.TempDir(), "test.fseq")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	client := clients.NewTestXLightsClient()
	header, err := client.ParseHeader(path)
	if err != nil {
		t.Fatalf("ParseHeader failed: %v", err)
	}

	if header.Magic != "PSEQ" {
		t.Errorf("Magic = %q, want \"PSEQ\"", header.Magic)
	}
	if header.MajorVersion != 2 {
		t.Errorf("MajorVersion = %d, want 2", header.MajorVersion)
	}
	if header.ChannelCount != 512 {
		t.Errorf("ChannelCount = %d, want 512", header.ChannelCount)
	}
	if header.FrameCount != 100 {
		t.Errorf("FrameCount = %d, want 100", header.FrameCount)
	}
	if header.StepTimeMs != 50 {
		t.Errorf("StepTimeMs = %d, want 50", header.StepTimeMs)
	}
	if header.FPS != 20.0 {
		t.Errorf("FPS = %f, want 20.0", header.FPS)
	}
	if header.DurationSeconds != 5.0 {
		t.Errorf("DurationSeconds = %f, want 5.0", header.DurationSeconds)
	}
	if header.CompressionTypeName != "none" {
		t.Errorf("CompressionTypeName = %q, want \"none\"", header.CompressionTypeName)
	}

	t.Logf("Header: channels=%d frames=%d fps=%.0f duration=%.1fs",
		header.ChannelCount, header.FrameCount, header.FPS, header.DurationSeconds)
}

func TestReadFrames(t *testing.T) {
	data := buildTestFSEQ(8, 10, 50)
	path := filepath.Join(t.TempDir(), "test.fseq")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	client := clients.NewTestXLightsClient()
	frames, err := client.ReadFrames(path, 0, 3)
	if err != nil {
		t.Fatalf("ReadFrames failed: %v", err)
	}

	if len(frames) != 3 {
		t.Fatalf("got %d frames, want 3", len(frames))
	}

	if frames[0].FrameIndex != 0 {
		t.Errorf("frame 0 index = %d, want 0", frames[0].FrameIndex)
	}
	if frames[0].ChannelCount != 8 {
		t.Errorf("frame 0 channel count = %d, want 8", frames[0].ChannelCount)
	}
	if len(frames[0].Preview) != 8 {
		t.Errorf("frame 0 preview length = %d, want 8", len(frames[0].Preview))
	}

	t.Logf("Frame 0 preview: %v", frames[0].Preview)
	t.Logf("Frame 1 preview: %v", frames[1].Preview)
}
