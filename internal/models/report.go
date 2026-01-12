package models

import (
	"time"

	"github.com/google/uuid"
)

type Report struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	WallpaperID      uuid.UUID  `json:"wallpaper_id" db:"wallpaper_id"`
	ReporterID       uuid.UUID  `json:"reporter_id" db:"reporter_id"`
	WallpaperTitle   string     `json:"wallpaper_title" db:"wallpaper_title"`
	WallpaperURL     string     `json:"wallpaper_url" db:"wallpaper_url"`
	ReporterUsername string     `json:"reporter_username" db:"reporter_username"`
	Reason           string     `json:"reason" db:"reason"`
	Status           string     `json:"status" db:"status"`
	ResolvedBy       *uuid.UUID `json:"resolved_by,omitempty" db:"resolved_by"`
	ResolvedAt       *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

type ReportRequest struct {
	WallpaperID uuid.UUID `json:"wallpaper_id" validate:"required,uuid"`
	Reason      string    `json:"reason" validate:"required,min=10,max=500"`
}
