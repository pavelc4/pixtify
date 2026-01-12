package service

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/pavelc4/pixtify/internal/processor"
	"github.com/pavelc4/pixtify/internal/repository/postgres/wallpaper"
	"github.com/pavelc4/pixtify/internal/storage"
)

type WallpaperService struct {
	repo         *wallpaper.Repository
	storage      storage.Service
	processor    *processor.ImageProcessor
	bucketOrigin string
	bucketThumb  string
}

// SplitTags
func SplitTags(tags string) []string {
	if tags == "" {
		return []string{}
	}
	return strings.Split(tags, ",")
}

func NewWallpaperService(repo *wallpaper.Repository, storage storage.Service, processor *processor.ImageProcessor) *WallpaperService {
	return &WallpaperService{
		repo:         repo,
		storage:      storage,
		processor:    processor,
		bucketOrigin: "pixtify-originals",
		bucketThumb:  "pixtify-thumbnails",
	}
}

type CreateWallpaperInput struct {
	UserID      string
	Title       string
	Description string
	ImageData   []byte
	ContentType string
	Tags        []string
}

func (s *WallpaperService) CreateWallpaper(ctx context.Context, input CreateWallpaperInput) (*wallpaper.Wallpaper, error) {
	//  Validate Image
	info, err := s.processor.ValidateImage(input.ImageData, input.ContentType)
	if err != nil {
		return nil, fmt.Errorf("invalid image: %w", err)
	}

	userUUID, err := uuid.Parse(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	wallpaperID := uuid.New()
	ext := filepath.Ext("image." + info.Format)
	if info.Format == "jpeg" {
		ext = ".jpg"
	}
	if info.Format == "png" {
		ext = ".png"
	}
	if info.Format == "webp" {
		ext = ".webp"
	}

	//  Upload Original
	originalKey := fmt.Sprintf("%s/original%s", wallpaperID, ext)
	originalURL, err := s.storage.Upload(ctx, s.bucketOrigin, originalKey, bytes.NewReader(input.ImageData), int64(len(input.ImageData)), input.ContentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload original: %w", err)
	}

	//  Generate & Upload Thumbnails
	thumbData, err := s.processor.GenerateThumbnail(input.ImageData, 400, 400)
	if err != nil {
		return nil, fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	thumbKey := fmt.Sprintf("%s/thumbnail.jpg", wallpaperID)
	thumbURL, err := s.storage.Upload(ctx, s.bucketThumb, thumbKey, bytes.NewReader(thumbData), int64(len(thumbData)), "image/jpeg")
	if err != nil {
		return nil, fmt.Errorf("failed to upload thumbnail: %w", err)
	}

	//  Save Metadata
	wp := &wallpaper.Wallpaper{
		ID:            wallpaperID,
		UserID:        userUUID,
		Title:         input.Title,
		Description:   &input.Description,
		OriginalURL:   originalURL,
		LargeURL:      originalURL, // Todo: implement distinct sizes
		MediumURL:     originalURL, // Todo: implement distinct sizes
		ThumbnailURL:  thumbURL,
		Width:         info.Width,
		Height:        info.Height,
		FileSizeBytes: int64(len(input.ImageData)),
		MimeType:      input.ContentType,
		Status:        "active",
		IsFeatured:    false,
	}

	if err := s.repo.Create(ctx, wp); err != nil {
		return nil, fmt.Errorf("failed to create wallpaper record: %w", err)
	}

	// 5. Save Tags
	if len(input.Tags) > 0 {
		var cleanTags []string
		for _, t := range input.Tags {
			if t != "" {
				cleanTags = append(cleanTags, t)
			}
		}
		if len(cleanTags) > 0 {
			_ = s.repo.AddTags(ctx, wallpaperID, cleanTags)
			wp.Tags = cleanTags
		}
	}

	return wp, nil
}

func (s *WallpaperService) ListWallpapers(ctx context.Context, page, limit int) ([]*wallpaper.Wallpaper, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	offset := (page - 1) * limit
	return s.repo.List(ctx, limit, offset)
}

func (s *WallpaperService) GetWallpaper(ctx context.Context, idStr string) (*wallpaper.Wallpaper, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid wallpaper ID")
	}
	return s.repo.GetByID(ctx, id)
}
