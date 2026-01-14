package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	userRepo "github.com/pavelc4/pixtify/internal/repository/postgres/user"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

type UserService struct {
	repo *userRepo.Repository
}

func NewUserService(repo *userRepo.Repository) *UserService {
	return &UserService{repo: repo}
}

type RegisterInput struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name,omitempty"`
}

func (s *UserService) Register(ctx context.Context, input RegisterInput) (*userRepo.User, error) {
	existing, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, userRepo.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if existing != nil {
		return nil, ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &userRepo.User{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		Role:         "user",
	}

	user.FullName = stringPtr(input.FullName)

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (s *UserService) Login(ctx context.Context, email, password string) (*userRepo.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, userRepo.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*userRepo.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userRepo.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*userRepo.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, userID string) (*userRepo.User, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}
	return s.GetProfile(ctx, uid)
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, fullName, bio, avatarURL *string) (*userRepo.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userRepo.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if fullName != nil {
		user.FullName = fullName
	}
	if bio != nil {
		user.Bio = bio
	}
	if avatarURL != nil {
		user.AvatarURL = avatarURL
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userRepo.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := s.repo.Delete(ctx, user.ID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (s *UserService) ListUsers(ctx context.Context, page, limit int) ([]*userRepo.User, int, error) {
	// Calculate offset
	offset := (page - 1) * limit

	users, total, err := s.repo.ListWithPagination(ctx, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

func (s *UserService) BanUser(ctx context.Context, userID, bannedBy uuid.UUID) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userRepo.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := s.repo.BanUser(ctx, user.ID, bannedBy); err != nil {
		return fmt.Errorf("failed to ban user: %w", err)
	}

	return nil
}

func (s *UserService) UnbanUser(ctx context.Context, userID uuid.UUID) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userRepo.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := s.repo.UnbanUser(ctx, user.ID); err != nil {
		return fmt.Errorf("failed to unban user: %w", err)
	}

	return nil
}

func (s *UserService) GetUserStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userRepo.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get aggregated stats from database
	aggregatedStats, err := s.repo.GetUserStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	stats := map[string]interface{}{
		"user_id":         user.ID,
		"username":        user.Username,
		"role":            user.Role,
		"is_verified":     user.IsVerified,
		"is_banned":       user.IsBanned,
		"created_at":      user.CreatedAt,
		"wallpaper_count": aggregatedStats.WallpaperCount,
		"likes_received":  aggregatedStats.LikesReceived,
		"reports_count":   aggregatedStats.ReportsCount,
	}

	if user.BannedAt != nil {
		stats["banned_at"] = user.BannedAt
	}

	return stats, nil
}
