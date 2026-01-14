package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pavelc4/pixtify/internal/repository/postgres/like"
	"github.com/pavelc4/pixtify/internal/repository/postgres/wallpaper"
)

type LikeService struct {
	likeRepo      *like.Repository
	wallpaperRepo *wallpaper.Repository
}

func NewLikeService(likeRepo *like.Repository, wallpaperRepo *wallpaper.Repository) *LikeService {
	return &LikeService{
		likeRepo:      likeRepo,
		wallpaperRepo: wallpaperRepo,
	}
}

// ToggleLike adds or removes a like and updates wallpaper like_count
func (s *LikeService) ToggleLike(ctx context.Context, userIDStr, wallpaperIDStr string) (bool, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return false, fmt.Errorf("invalid user ID")
	}

	wallpaperID, err := uuid.Parse(wallpaperIDStr)
	if err != nil {
		return false, fmt.Errorf("invalid wallpaper ID")
	}

	// Check if wallpaper exists
	_, err = s.wallpaperRepo.GetByID(ctx, wallpaperID)
	if err != nil {
		return false, fmt.Errorf("wallpaper not found")
	}

	// Use transaction to toggle like and update count atomically
	return s.likeRepo.ToggleLikeWithTx(ctx, userID, wallpaperID)
}

// CheckLikeStatus checks if user has liked a wallpaper
func (s *LikeService) CheckLikeStatus(ctx context.Context, userIDStr, wallpaperIDStr string) (bool, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return false, fmt.Errorf("invalid user ID")
	}

	wallpaperID, err := uuid.Parse(wallpaperIDStr)
	if err != nil {
		return false, fmt.Errorf("invalid wallpaper ID")
	}

	return s.likeRepo.IsLiked(ctx, userID, wallpaperID)
}

// GetUserLikedWallpapers returns wallpapers liked by user
func (s *LikeService) GetUserLikedWallpapers(ctx context.Context, userIDStr string, page, limit int) ([]*wallpaper.Wallpaper, int, error) {
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

	// Get wallpaper IDs
	wallpaperIDs, total, err := s.likeRepo.GetUserLikes(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	if len(wallpaperIDs) == 0 {
		return []*wallpaper.Wallpaper{}, total, nil
	}

	// TODO: Implement bulk fetch method in wallpaper repository to avoid N+1 queries
	var wallpapers []*wallpaper.Wallpaper
	for _, id := range wallpaperIDs {
		wp, err := s.wallpaperRepo.GetByID(ctx, id)
		if err != nil {
			continue
		}
		wallpapers = append(wallpapers, wp)
	}

	return wallpapers, total, nil
}
