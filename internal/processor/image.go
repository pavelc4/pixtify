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
		maxSizeBytes: 100 * 1024 * 1024, // 100MB
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
		return nil, fmt.Errorf("image too large (max 100MB)")
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

func (p *ImageProcessor) ResizeForMobile(data []byte, maxWidth int, format string) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	origWidth := bounds.Dx()

	// If image is already smaller, return nil to indicate no resize needed
	if origWidth <= maxWidth {
		return nil, nil
	}

	resized := imaging.Resize(img, maxWidth, 0, imaging.Lanczos)

	buf := new(bytes.Buffer)

	switch format {
	case "jpeg":
		if err := imaging.Encode(buf, resized, imaging.JPEG, imaging.JPEGQuality(100)); err != nil {
			return nil, err
		}
	case "png":
		if err := imaging.Encode(buf, resized, imaging.PNG); err != nil {
			return nil, err
		}
	default:
		// Fallback to JPEG with max quality
		if err := imaging.Encode(buf, resized, imaging.JPEG, imaging.JPEGQuality(100)); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}
