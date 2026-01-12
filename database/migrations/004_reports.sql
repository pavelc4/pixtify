CREATE TABLE IF NOT EXISTS reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallpaper_id UUID,  -- Nullable, no FK constraint (wallpapers table doesn't exist yet)
    reporter_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallpaper_title VARCHAR(255),
    wallpaper_url TEXT,
    reporter_username VARCHAR(100),
    reason TEXT NOT NULL CHECK (length(reason) >= 10),
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'resolved', 'dismissed')),
    resolved_by UUID REFERENCES users(id),
    resolved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_reports_status ON reports(status);
CREATE INDEX idx_reports_wallpaper ON reports(wallpaper_id);
CREATE INDEX idx_reports_reporter ON reports(reporter_id);
CREATE INDEX idx_reports_created_at ON reports(created_at DESC);
