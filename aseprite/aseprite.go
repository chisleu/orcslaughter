package aseprite

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"io"
	"os"
)

// File represents an Aseprite file
type File struct {
	Header *Header
	Frames []*Frame
	Tags   []*Tag
}

// Header represents the Aseprite file header
type Header struct {
	FileSize    uint32
	MagicNumber uint16
	Frames      uint16
	Width       uint16
	Height      uint16
	ColorDepth  uint16
	Flags       uint32
	Speed       uint16
	Transparent uint8
	Colors      uint16
	PixelWidth  uint8
	PixelHeight uint8
	GridX       int16
	GridY       int16
	GridWidth   uint16
	GridHeight  uint16
}

// Frame represents a single frame in the animation
type Frame struct {
	Header *FrameHeader
	Chunks []*Chunk
}

// FrameHeader represents the frame header
type FrameHeader struct {
	BytesInFrame uint32
	MagicNumber  uint16
	OldChunks    uint16
	Duration     uint16
	NewChunks    uint32
}

// Chunk represents a data chunk within a frame
type Chunk struct {
	Size uint32
	Type uint16
	Data []byte
}

// Cel represents a cel (layer content at a specific frame)
type Cel struct {
	LayerIndex uint16
	X          int16
	Y          int16
	Opacity    uint8
	Type       uint16
	ZIndex     int16
	Width      uint16
	Height     uint16
	Pixels     []byte
}

// Tag represents an animation tag
type Tag struct {
	Name      string
	FromFrame uint16
	ToFrame   uint16
	Direction uint8
	Repeat    uint16
	Color     [3]uint8 // RGB color (deprecated but kept for compatibility)
}

// Direction constants for animation tags
const (
	DirectionForward     = 0
	DirectionReverse     = 1
	DirectionPingPong    = 2
	DirectionPingPongRev = 3
)

// LoadFile loads an Aseprite file from disk
func LoadFile(filename string) (*File, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return ParseFile(data)
}

// ParseFile parses Aseprite file data
func ParseFile(data []byte) (*File, error) {
	reader := bytes.NewReader(data)

	// Read header
	header, err := readHeader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Validate magic number
	if header.MagicNumber != 0xA5E0 {
		return nil, fmt.Errorf("invalid magic number: %x", header.MagicNumber)
	}

	file := &File{
		Header: header,
		Frames: make([]*Frame, header.Frames),
		Tags:   []*Tag{},
	}

	// Read frames
	for i := uint16(0); i < header.Frames; i++ {
		frame, err := readFrame(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read frame %d: %w", i, err)
		}
		file.Frames[i] = frame

		// Process chunks to find tags
		for _, chunk := range frame.Chunks {
			if chunk.Type == 0x2018 { // Tags chunk
				tags, err := parseTagsChunk(chunk.Data)
				if err == nil {
					file.Tags = append(file.Tags, tags...)
				}
			}
		}
	}

	return file, nil
}

// GetFrameImage extracts an image from a specific frame
func (f *File) GetFrameImage(frameIndex int) (image.Image, error) {
	if frameIndex >= len(f.Frames) {
		return nil, fmt.Errorf("frame index %d out of range", frameIndex)
	}

	frame := f.Frames[frameIndex]
	img := image.NewRGBA(image.Rect(0, 0, int(f.Header.Width), int(f.Header.Height)))

	// Process chunks to find cel data
	for _, chunk := range frame.Chunks {
		if chunk.Type == 0x2005 { // Cel chunk
			cel, err := parseCelChunk(chunk.Data)
			if err != nil {
				continue // Skip invalid cels
			}

			// Draw cel to image
			err = drawCelToImage(img, cel, f.Header.ColorDepth)
			if err != nil {
				continue // Skip cels that can't be drawn
			}
		}
	}

	return img, nil
}

func readHeader(reader io.Reader) (*Header, error) {
	header := &Header{}

	if err := binary.Read(reader, binary.LittleEndian, &header.FileSize); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &header.MagicNumber); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &header.Frames); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &header.Width); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &header.Height); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &header.ColorDepth); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &header.Flags); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &header.Speed); err != nil {
		return nil, err
	}

	// Skip reserved fields
	reader.Read(make([]byte, 8)) // Skip 2 DWORDs

	if err := binary.Read(reader, binary.LittleEndian, &header.Transparent); err != nil {
		return nil, err
	}

	// Skip more reserved bytes
	reader.Read(make([]byte, 3))

	if err := binary.Read(reader, binary.LittleEndian, &header.Colors); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &header.PixelWidth); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &header.PixelHeight); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &header.GridX); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &header.GridY); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &header.GridWidth); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &header.GridHeight); err != nil {
		return nil, err
	}

	// Skip remaining header bytes
	reader.Read(make([]byte, 84))

	return header, nil
}

func readFrame(reader io.Reader) (*Frame, error) {
	frameHeader := &FrameHeader{}

	if err := binary.Read(reader, binary.LittleEndian, &frameHeader.BytesInFrame); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &frameHeader.MagicNumber); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &frameHeader.OldChunks); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &frameHeader.Duration); err != nil {
		return nil, err
	}

	// Skip reserved bytes
	reader.Read(make([]byte, 2))

	if err := binary.Read(reader, binary.LittleEndian, &frameHeader.NewChunks); err != nil {
		return nil, err
	}

	// Determine number of chunks
	numChunks := frameHeader.NewChunks
	if numChunks == 0 {
		numChunks = uint32(frameHeader.OldChunks)
	}

	frame := &Frame{
		Header: frameHeader,
		Chunks: make([]*Chunk, numChunks),
	}

	// Read chunks
	for i := uint32(0); i < numChunks; i++ {
		chunk, err := readChunk(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read chunk %d: %w", i, err)
		}
		frame.Chunks[i] = chunk
	}

	return frame, nil
}

func readChunk(reader io.Reader) (*Chunk, error) {
	chunk := &Chunk{}

	if err := binary.Read(reader, binary.LittleEndian, &chunk.Size); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &chunk.Type); err != nil {
		return nil, err
	}

	// Read chunk data (size includes the size and type fields)
	dataSize := chunk.Size - 6
	chunk.Data = make([]byte, dataSize)
	if _, err := io.ReadFull(reader, chunk.Data); err != nil {
		return nil, err
	}

	return chunk, nil
}

func parseCelChunk(data []byte) (*Cel, error) {
	reader := bytes.NewReader(data)
	cel := &Cel{}

	if err := binary.Read(reader, binary.LittleEndian, &cel.LayerIndex); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &cel.X); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &cel.Y); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &cel.Opacity); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &cel.Type); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &cel.ZIndex); err != nil {
		return nil, err
	}

	// Skip reserved bytes
	reader.Read(make([]byte, 5))

	// Handle different cel types
	switch cel.Type {
	case 2: // Compressed Image
		if err := binary.Read(reader, binary.LittleEndian, &cel.Width); err != nil {
			return nil, err
		}
		if err := binary.Read(reader, binary.LittleEndian, &cel.Height); err != nil {
			return nil, err
		}

		// Read compressed pixel data
		compressedData := make([]byte, len(data)-int(reader.Size())+int(reader.Len()))
		if _, err := io.ReadFull(reader, compressedData); err != nil {
			return nil, err
		}

		// Decompress with zlib
		zlibReader, err := zlib.NewReader(bytes.NewReader(compressedData))
		if err != nil {
			return nil, err
		}
		defer zlibReader.Close()

		cel.Pixels, err = io.ReadAll(zlibReader)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported cel type: %d", cel.Type)
	}

	return cel, nil
}

func drawCelToImage(img *image.RGBA, cel *Cel, colorDepth uint16) error {
	if len(cel.Pixels) == 0 {
		return fmt.Errorf("no pixel data")
	}

	bytesPerPixel := int(colorDepth / 8)
	if bytesPerPixel == 0 {
		bytesPerPixel = 1 // For indexed color
	}

	for y := 0; y < int(cel.Height); y++ {
		for x := 0; x < int(cel.Width); x++ {
			pixelIndex := (y*int(cel.Width) + x) * bytesPerPixel
			if pixelIndex+bytesPerPixel > len(cel.Pixels) {
				continue
			}

			var c color.RGBA
			switch colorDepth {
			case 32: // RGBA
				if pixelIndex+3 < len(cel.Pixels) {
					c = color.RGBA{
						R: cel.Pixels[pixelIndex],
						G: cel.Pixels[pixelIndex+1],
						B: cel.Pixels[pixelIndex+2],
						A: cel.Pixels[pixelIndex+3],
					}
				}
			case 16: // Grayscale
				if pixelIndex+1 < len(cel.Pixels) {
					gray := cel.Pixels[pixelIndex]
					alpha := cel.Pixels[pixelIndex+1]
					c = color.RGBA{R: gray, G: gray, B: gray, A: alpha}
				}
			case 8: // Indexed - for now, treat as grayscale
				if pixelIndex < len(cel.Pixels) {
					gray := cel.Pixels[pixelIndex]
					c = color.RGBA{R: gray, G: gray, B: gray, A: 255}
				}
			}

			// Apply cel opacity
			c.A = uint8((uint16(c.A) * uint16(cel.Opacity)) / 255)

			// Set pixel in image
			imgX := int(cel.X) + x
			imgY := int(cel.Y) + y
			if imgX >= 0 && imgY >= 0 && imgX < img.Bounds().Dx() && imgY < img.Bounds().Dy() {
				img.Set(imgX, imgY, c)
			}
		}
	}

	return nil
}

func parseTagsChunk(data []byte) ([]*Tag, error) {
	reader := bytes.NewReader(data)
	var tags []*Tag

	// Read number of tags
	var numTags uint16
	if err := binary.Read(reader, binary.LittleEndian, &numTags); err != nil {
		return nil, err
	}

	// Skip reserved bytes
	reader.Read(make([]byte, 8))

	// Read each tag
	for i := uint16(0); i < numTags; i++ {
		tag := &Tag{}

		if err := binary.Read(reader, binary.LittleEndian, &tag.FromFrame); err != nil {
			return nil, err
		}
		if err := binary.Read(reader, binary.LittleEndian, &tag.ToFrame); err != nil {
			return nil, err
		}
		if err := binary.Read(reader, binary.LittleEndian, &tag.Direction); err != nil {
			return nil, err
		}
		if err := binary.Read(reader, binary.LittleEndian, &tag.Repeat); err != nil {
			return nil, err
		}

		// Skip reserved bytes
		reader.Read(make([]byte, 6))

		// Read deprecated color (3 bytes RGB)
		if err := binary.Read(reader, binary.LittleEndian, &tag.Color); err != nil {
			return nil, err
		}

		// Skip extra byte
		reader.Read(make([]byte, 1))

		// Read tag name (STRING format: WORD length + bytes)
		var nameLength uint16
		if err := binary.Read(reader, binary.LittleEndian, &nameLength); err != nil {
			return nil, err
		}

		nameBytes := make([]byte, nameLength)
		if _, err := io.ReadFull(reader, nameBytes); err != nil {
			return nil, err
		}
		tag.Name = string(nameBytes)

		tags = append(tags, tag)
	}

	return tags, nil
}
