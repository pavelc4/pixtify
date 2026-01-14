package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pavelc4/pixtify/internal/repository/postgres/collection"
	"github.com/pavelc4/pixtify/internal/repository/postgres/wallpaper"
)

type CollectionService struct {
	collectionRepo *collection.Repository
	wallpaperRepo  *wallpaper.Repository
}

func NewCollectionService(collectionRepo *collection.Repository, wallpaperRepo *wallpaper.Repository) *CollectionService {
	return &CollectionService{
		collectionRepo: collectionRepo,
		wallpaperRepo:  wallpaperRepo,
	}
}

// CreateCollection creates a new collection for a user
func (s *CollectionService) CreateCollection(ctx context.Context, userIDStr, name, description string, isPublic bool) (*collection.Collection, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	if name == "" {
		return nil, fmt.Errorf("collection name is required")
	}

	c := &collection.Collection{
		UserID:   userID,
		Name:     name,
		IsPublic: isPublic,
	}
	if description != "" {
		c.Description = &description
	}

	if err := s.collectionRepo.Create(ctx, c); err != nil {
		return nil, err
	}

	return c, nil
}

// GetUserCollections retrieves all collections for a user
func (s *CollectionService) GetUserCollections(ctx context.Context, userIDStr string, page, limit int) ([]*collection.Collection, int, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid user ID")
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	offset := (page - 1) * limit

	return s.collectionRepo.GetUserCollections(ctx, userID, limit, offset)
}

// GetCollectionDetails retrieves collection details with privacy check
func (s *CollectionService) GetCollectionDetails(ctx context.Context, collectionIDStr, requestUserIDStr string) (*collection.Collection, error) {
	collectionID, err := uuid.Parse(collectionIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid collection ID")
	}

	c, err := s.collectionRepo.GetByID(ctx, collectionID)
	if err != nil {
		return nil, fmt.Errorf("collection not found")
	}

	// Check privacy
	if !c.IsPublic {
		requestUserID, err := uuid.Parse(requestUserIDStr)
		if err != nil || requestUserID != c.UserID {
			return nil, fmt.Errorf("collection is private")
		}
	}

	return c, nil
}

// AddWallpaperToCollection adds a wallpaper to a collection with ownership verification
func (s *CollectionService) AddWallpaperToCollection(ctx context.Context, userIDStr, collectionIDStr, wallpaperIDStr string) error {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user ID")
	}

	collectionID, err := uuid.Parse(collectionIDStr)
	if err != nil {
		return fmt.Errorf("invalid collection ID")
	}

	wallpaperID, err := uuid.Parse(wallpaperIDStr)
	if err != nil {
		return fmt.Errorf("invalid wallpaper ID")
	}

	// Verify collection ownership
	c, err := s.collectionRepo.GetByID(ctx, collectionID)
	if err != nil {
		return fmt.Errorf("collection not found")
	}
	if c.UserID != userID {
		return fmt.Errorf("you don't own this collection")
	}

	// Verify wallpaper exists
	_, err = s.wallpaperRepo.GetByID(ctx, wallpaperID)
	if err != nil {
		return fmt.Errorf("wallpaper not found")
	}

	return s.collectionRepo.AddWallpaper(ctx, collectionID, wallpaperID)
}

// RemoveWallpaperFromCollection removes a wallpaper from a collection with ownership verification
func (s *CollectionService) RemoveWallpaperFromCollection(ctx context.Context, userIDStr, collectionIDStr, wallpaperIDStr string) error {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user ID")
	}

	collectionID, err := uuid.Parse(collectionIDStr)
	if err != nil {
		return fmt.Errorf("invalid collection ID")
	}

	wallpaperID, err := uuid.Parse(wallpaperIDStr)
	if err != nil {
		return fmt.Errorf("invalid wallpaper ID")
	}

	// Verify collection ownership
	c, err := s.collectionRepo.GetByID(ctx, collectionID)
	if err != nil {
		return fmt.Errorf("collection not found")
	}
	if c.UserID != userID {
		return fmt.Errorf("you don't own this collection")
	}

	return s.collectionRepo.RemoveWallpaper(ctx, collectionID, wallpaperID)
}

// GetCollectionWallpapers retrieves wallpapers in a collection with privacy check
func (s *CollectionService) GetCollectionWallpapers(ctx context.Context, collectionIDStr, requestUserIDStr string, page, limit int) ([]*wallpaper.Wallpaper, int, error) {
	collectionID, err := uuid.Parse(collectionIDStr)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid collection ID")
	}

	// Verify collection access
	c, err := s.collectionRepo.GetByID(ctx, collectionID)
	if err != nil {
		return nil, 0, fmt.Errorf("collection not found")
	}

	// Check privacy
	if !c.IsPublic {
		requestUserID, err := uuid.Parse(requestUserIDStr)
		if err != nil || requestUserID != c.UserID {
			return nil, 0, fmt.Errorf("collection is private")
		}
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Get wallpaper IDs
	wallpaperIDs, total, err := s.collectionRepo.GetCollectionWallpapers(ctx, collectionID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	if len(wallpaperIDs) == 0 {
		return []*wallpaper.Wallpaper{}, total, nil
	}

	// Bulk fetch wallpapers in a single query (avoids N+1 problem)
	wallpapers, err := s.wallpaperRepo.GetByIDs(ctx, wallpaperIDs)
	if err != nil {
		return nil, 0, err
	}

	return wallpapers, total, nil
}

// DeleteCollection deletes a collection with ownership verification
func (s *CollectionService) DeleteCollection(ctx context.Context, userIDStr, collectionIDStr string) error {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user ID")
	}

	collectionID, err := uuid.Parse(collectionIDStr)
	if err != nil {
		return fmt.Errorf("invalid collection ID")
	}

	// Verify collection ownership
	c, err := s.collectionRepo.GetByID(ctx, collectionID)
	if err != nil {
		return fmt.Errorf("collection not found")
	}
	if c.UserID != userID {
		return fmt.Errorf("you don't own this collection")
	}

	return s.collectionRepo.Delete(ctx, collectionID)
}
