ALTER TABLE wallpapers ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_wallpapers_deleted_at ON wallpapers(deleted_at) WHERE deleted_at IS NULL;
