package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/pavelc4/pixtify/internal/repository/postgres/tag"
)

var (
	ErrTagNameTooShort  = errors.New("tag name must be at least 2 characters")
	ErrTagNameTooLong   = errors.New("tag name must be at most 50 characters")
	ErrTagAlreadyExists = errors.New("tag with this name already exists")
	ErrTagNotFound      = errors.New("tag not found")
)

// TagService handles tag business logic
type TagService struct {
	tagRepo *tag.Repository
}

// NewTagService creates a new tag service
func NewTagService(tagRepo *tag.Repository) *TagService {
	return &TagService{
		tagRepo: tagRepo,
	}
}

// CreateTag creates a new tag with auto-generated slug
func (s *TagService) CreateTag(ctx context.Context, name string) (*tag.Tag, error) {
	// Validate name
	name = strings.TrimSpace(name)
	if len(name) < 2 {
		return nil, ErrTagNameTooShort
	}
	if len(name) > 50 {
		return nil, ErrTagNameTooLong
	}

	// Generate slug
	slug := s.generateSlug(name)

	// Check if tag with this slug already exists
	existing, err := s.tagRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing tag: %w", err)
	}
	if existing != nil {
		return nil, ErrTagAlreadyExists
	}

	// Create tag
	newTag, err := s.tagRepo.Create(ctx, name, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	return newTag, nil
}

// ListTags returns all tags with pagination
func (s *TagService) ListTags(ctx context.Context, limit, offset int) ([]*tag.Tag, int, error) {
	// Validate pagination
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	tags, total, err := s.tagRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tags: %w", err)
	}

	return tags, total, nil
}

// DeleteTag deletes a tag by ID (moderator only)
func (s *TagService) DeleteTag(ctx context.Context, id uuid.UUID) error {
	// Check if tag exists
	existing, err := s.tagRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check tag: %w", err)
	}
	if existing == nil {
		return ErrTagNotFound
	}

	// Delete tag (cascade deletes wallpaper_tags)
	err = s.tagRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	return nil
}

// GetTagBySlug retrieves a tag by its slug
func (s *TagService) GetTagBySlug(ctx context.Context, slug string) (*tag.Tag, error) {
	tagObj, err := s.tagRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag by slug: %w", err)
	}
	if tagObj == nil {
		return nil, ErrTagNotFound
	}

	return tagObj, nil
}

func (s *TagService) generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Remove special characters except spaces and hyphens
	reg := regexp.MustCompile(`[^a-z0-9\s-]+`)
	slug = reg.ReplaceAllString(slug, "")

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Replace multiple hyphens with single hyphen
	reg = regexp.MustCompile(`-+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	return slug
}
