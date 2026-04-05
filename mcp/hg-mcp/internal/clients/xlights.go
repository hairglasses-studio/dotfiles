// Package clients provides API clients for external services.
package clients

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// FSEQHeader represents the parsed header of an FSEQ v2 file.
type FSEQHeader struct {
	Magic               string  `json:"magic"`
	ChannelDataOffset   uint16  `json:"channel_data_offset"`
	MinorVersion        uint8   `json:"minor_version"`
	MajorVersion        uint8   `json:"major_version"`
	HeaderLength        uint16  `json:"header_length"`
	ChannelCount        uint32  `json:"channel_count"`
	FrameCount          uint32  `json:"frame_count"`
	StepTimeMs          uint8   `json:"step_time_ms"`
	Flags               uint8   `json:"flags"`
	CompressionType     uint8   `json:"compression_type"`
	CompressionBlocks   uint8   `json:"compression_blocks"`
	SparseRangeCount    uint8   `json:"sparse_range_count"`
	UUID                string  `json:"uuid"`
	CompressionTypeName string  `json:"compression_type_name"`
	DurationSeconds     float64 `json:"duration_seconds"`
	FPS                 float64 `json:"fps"`
}

// FSEQFrameData represents extracted frame data from an FSEQ file.
type FSEQFrameData struct {
	FrameIndex   int    `json:"frame_index"`
	ChannelCount int    `json:"channel_count"`
	Values       []byte `json:"-"`
	Preview      []int  `json:"preview"` // First N channel values for preview
}

// XLightsClient provides FSEQ file parsing capabilities.
// This is a file parser, not a network client.
type XLightsClient struct{}

// NewXLightsClient creates a new xLights FSEQ parser.
func NewXLightsClient() (*XLightsClient, error) {
	return &XLightsClient{}, nil
}

// NewTestXLightsClient creates a test client.
func NewTestXLightsClient() *XLightsClient {
	return &XLightsClient{}
}

// TestOverrideXLightsClient, when non-nil, is used instead of the singleton.
var TestOverrideXLightsClient *XLightsClient

// GetXLightsClient returns the xLights client (no singleton needed for file parser).
func GetXLightsClient() (*XLightsClient, error) {
	if TestOverrideXLightsClient != nil {
		return TestOverrideXLightsClient, nil
	}
	return NewXLightsClient()
}

// ParseHeader reads and parses the FSEQ v2 header from a file.
func (c *XLightsClient) ParseHeader(path string) (*FSEQHeader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	return c.parseHeaderFromReader(f)
}

// parseHeaderFromReader reads the FSEQ header from any reader.
func (c *XLightsClient) parseHeaderFromReader(r io.Reader) (*FSEQHeader, error) {
	// Read magic (4 bytes)
	magic := make([]byte, 4)
	if _, err := io.ReadFull(r, magic); err != nil {
		return nil, fmt.Errorf("failed to read magic: %w", err)
	}
	if string(magic) != "PSEQ" {
		return nil, fmt.Errorf("invalid FSEQ file: magic is %q, expected \"PSEQ\"", string(magic))
	}

	header := &FSEQHeader{
		Magic: string(magic),
	}

	// Read fixed header fields (little-endian)
	if err := binary.Read(r, binary.LittleEndian, &header.ChannelDataOffset); err != nil {
		return nil, fmt.Errorf("failed to read channel data offset: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &header.MinorVersion); err != nil {
		return nil, fmt.Errorf("failed to read minor version: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &header.MajorVersion); err != nil {
		return nil, fmt.Errorf("failed to read major version: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &header.HeaderLength); err != nil {
		return nil, fmt.Errorf("failed to read header length: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &header.ChannelCount); err != nil {
		return nil, fmt.Errorf("failed to read channel count: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &header.FrameCount); err != nil {
		return nil, fmt.Errorf("failed to read frame count: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &header.StepTimeMs); err != nil {
		return nil, fmt.Errorf("failed to read step time: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &header.Flags); err != nil {
		return nil, fmt.Errorf("failed to read flags: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &header.CompressionType); err != nil {
		return nil, fmt.Errorf("failed to read compression type: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &header.CompressionBlocks); err != nil {
		return nil, fmt.Errorf("failed to read compression blocks: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &header.SparseRangeCount); err != nil {
		return nil, fmt.Errorf("failed to read sparse range count: %w", err)
	}

	// Skip reserved byte
	reserved := make([]byte, 1)
	if _, err := io.ReadFull(r, reserved); err != nil {
		return nil, fmt.Errorf("failed to read reserved: %w", err)
	}

	// Read UUID (16 bytes)
	uuid := make([]byte, 16)
	if _, err := io.ReadFull(r, uuid); err != nil {
		return nil, fmt.Errorf("failed to read UUID: %w", err)
	}
	header.UUID = fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])

	// Compute derived fields
	switch header.CompressionType {
	case 0:
		header.CompressionTypeName = "none"
	case 1:
		header.CompressionTypeName = "zstd"
	case 2:
		header.CompressionTypeName = "zlib"
	default:
		header.CompressionTypeName = fmt.Sprintf("unknown(%d)", header.CompressionType)
	}

	if header.StepTimeMs > 0 {
		header.FPS = 1000.0 / float64(header.StepTimeMs)
		header.DurationSeconds = float64(header.FrameCount) * float64(header.StepTimeMs) / 1000.0
	}

	return header, nil
}

// ReadFrames reads frame data from an uncompressed FSEQ file.
func (c *XLightsClient) ReadFrames(path string, startFrame, count int) ([]FSEQFrameData, error) {
	if startFrame < 0 {
		return nil, fmt.Errorf("start frame must be non-negative")
	}
	if count <= 0 {
		return nil, fmt.Errorf("count must be positive")
	}

	header, err := c.ParseHeader(path)
	if err != nil {
		return nil, err
	}

	if header.CompressionType != 0 {
		return nil, fmt.Errorf("compressed FSEQ files not supported (type: %s)", header.CompressionTypeName)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Seek to frame data
	frameSize := int64(header.ChannelCount)
	offset := int64(header.ChannelDataOffset) + int64(startFrame)*frameSize
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek to frame %d: %w", startFrame, err)
	}

	endFrame := startFrame + count
	if endFrame > int(header.FrameCount) {
		endFrame = int(header.FrameCount)
	}

	frames := make([]FSEQFrameData, 0, endFrame-startFrame)
	for i := startFrame; i < endFrame; i++ {
		data := make([]byte, header.ChannelCount)
		if _, err := io.ReadFull(f, data); err != nil {
			break
		}

		previewLen := 16
		if int(header.ChannelCount) < previewLen {
			previewLen = int(header.ChannelCount)
		}
		preview := make([]int, previewLen)
		for j := 0; j < previewLen; j++ {
			preview[j] = int(data[j])
		}

		frames = append(frames, FSEQFrameData{
			FrameIndex:   i,
			ChannelCount: int(header.ChannelCount),
			Values:       data,
			Preview:      preview,
		})
	}

	return frames, nil
}
