package service

import (
	"bytes"
	"context"
	"fmt"
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
	DeviceType  string // "mobile" or "desktop"
	ImageData   []byte
	ContentType string
	Tags        []string // keywords for search
}

func (s *WallpaperService) CreateWallpaper(ctx context.Context, input CreateWallpaperInput) (*wallpaper.Wallpaper, error) {
	// Validate Image
	info, err := s.processor.ValidateImage(input.ImageData, input.ContentType)
	if err != nil {
		return nil, fmt.Errorf("invalid image: %w", err)
	}

	userUUID, err := uuid.Parse(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	// Validate device type
	deviceType := input.DeviceType
	if deviceType != "mobile" && deviceType != "desktop" {
		deviceType = "desktop" // default
	}

	wallpaperID := uuid.New()
	slug := slugify(input.Title)
	if slug == "" {
		slug = wallpaperID.String()[:8]
	}

	ext := ".jpg"
	if info.Format == "png" {
		ext = ".png"
	} else if info.Format == "webp" {
		ext = ".webp"
	}

	// Upload Original Image (no compression, full quality)
	imageKey := fmt.Sprintf("%s/%s%s", wallpaperID, slug, ext)
	imageURL, err := s.storage.Upload(ctx, s.bucketOrigin, imageKey, bytes.NewReader(input.ImageData), int64(len(input.ImageData)), input.ContentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload image: %w", err)
	}

	// Generate & Upload Thumbnail (compressed for fast loading)
	thumbData, err := s.processor.GenerateThumbnail(input.ImageData, 400, 400)
	if err != nil {
		return nil, fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	thumbKey := fmt.Sprintf("%s/%s_thumb.jpg", wallpaperID, slug)
	thumbURL, err := s.storage.Upload(ctx, s.bucketThumb, thumbKey, bytes.NewReader(thumbData), int64(len(thumbData)), "image/jpeg")
	if err != nil {
		return nil, fmt.Errorf("failed to upload thumbnail: %w", err)
	}

	// Save Metadata
	wp := &wallpaper.Wallpaper{
		ID:            wallpaperID,
		UserID:        userUUID,
		Title:         input.Title,
		Description:   &input.Description,
		OriginalURL:   imageURL,
		ImageURL:      imageURL,
		ThumbnailURL:  thumbURL,
		DeviceType:    deviceType,
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

	// Save Tags (keywords for search)
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

// slugify converts title to URL-safe slug
func slugify(title string) string {
	// Simple slugify: lowercase, replace spaces with hyphens, remove special chars
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")

	// Keep only alphanumeric and hyphens
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	// Trim leading/trailing hyphens and collapse multiple hyphens
	slug = result.String()
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")

	// Limit length
	if len(slug) > 50 {
		slug = slug[:50]
	}

	return slug
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

// UpdateWallpaper updates wallpaper metadata (owner only)
func (s *WallpaperService) UpdateWallpaper(ctx context.Context, wallpaperIDStr, userIDStr string, title, description *string) error {
	wallpaperID, err := uuid.Parse(wallpaperIDStr)
	if err != nil {
		return fmt.Errorf("invalid wallpaper ID")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user ID")
	}

	// Check if wallpaper exists and get owner
	wp, err := s.repo.GetByID(ctx, wallpaperID)
	if err != nil {
		return fmt.Errorf("wallpaper not found")
	}

	// Check ownership
	if wp.UserID != userID {
		return fmt.Errorf("you don't own this wallpaper")
	}

	return s.repo.Update(ctx, wallpaperID, title, description)
}

// DeleteWallpaper soft deletes wallpaper (owner or moderator)
func (s *WallpaperService) DeleteWallpaper(ctx context.Context, wallpaperIDStr, userIDStr, userRole string) error {
	wallpaperID, err := uuid.Parse(wallpaperIDStr)
	if err != nil {
		return fmt.Errorf("invalid wallpaper ID")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user ID")
	}

	// Check if wallpaper exists
	wp, err := s.repo.GetByID(ctx, wallpaperID)
	if err != nil {
		return fmt.Errorf("wallpaper not found")
	}

	// Check permissions: owner OR moderator/admin
	isOwner := wp.UserID == userID
	isModerator := userRole == "moderator" || userRole == "owner"

	if !isOwner && !isModerator {
		return fmt.Errorf("you don't have permission to delete this wallpaper")
	}

	return s.repo.SoftDelete(ctx, wallpaperID)
}

// SetFeaturedStatus toggles featured status (moderator only)
func (s *WallpaperService) SetFeaturedStatus(ctx context.Context, wallpaperIDStr string, isFeatured bool) error {
	wallpaperID, err := uuid.Parse(wallpaperIDStr)
	if err != nil {
		return fmt.Errorf("invalid wallpaper ID")
	}

	// Verify wallpaper exists before updating
	_, err = s.repo.GetByID(ctx, wallpaperID)
	if err != nil {
		return fmt.Errorf("wallpaper not found")
	}

	return s.repo.SetFeaturedStatus(ctx, wallpaperID, isFeatured)
}

// ListFeaturedWallpapers retrieves all featured wallpapers with pagination
func (s *WallpaperService) ListFeaturedWallpapers(ctx context.Context, page, limit int) ([]*wallpaper.Wallpaper, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	offset := (page - 1) * limit
	return s.repo.ListFeatured(ctx, limit, offset)
}
