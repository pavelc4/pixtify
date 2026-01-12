package processor

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/disintegration/imaging"
)

type ImageProcessor struct {
	maxSizeBytes int64
	allowedTypes []string
}

func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{
		maxSizeBytes: 10 * 1024 * 1024, // 10MB
		allowedTypes: []string{"image/jpeg", "image/png", "image/webp"},
	}
}

type ImageInfo struct {
	Width  int
	Height int
	Format string
}

// ValidateImage checks file type, size, and returns dimensions
func (p *ImageProcessor) ValidateImage(data []byte, contentType string) (*ImageInfo, error) {
	// Check size
	if int64(len(data)) > p.maxSizeBytes {
		return nil, fmt.Errorf("image too large (max 10MB)")
	}

	// Check type
	allowed := false
	for _, t := range p.allowedTypes {
		if t == contentType {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, fmt.Errorf("unsupported image type: %s", contentType)
	}

	// Decode config to check dimensions without decoding whole image
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("invalid image format: %w", err)
	}

	return &ImageInfo{
		Width:  cfg.Width,
		Height: cfg.Height,
		Format: format,
	}, nil
}

// GenerateThumbnail creates resized version
func (p *ImageProcessor) GenerateThumbnail(data []byte, width, height int) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	// Resize keeping aspect ratio
	thumbnail := imaging.Fit(img, width, height, imaging.Lanczos)

	buf := new(bytes.Buffer)
	// Always encode thumbnails as JPEG for efficiency
	if err := imaging.Encode(buf, thumbnail, imaging.JPEG, imaging.JPEGQuality(85)); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
