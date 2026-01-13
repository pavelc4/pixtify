package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	Env          string
	CookieSecret string
	Database     DatabaseConfig
	OAuth        OAuthConfig
	JWT          JWTConfig
	Storage      StorageConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type OAuthConfig struct {
	GithubClientID     string
	GithubClientSecret string
	GithubRedirectURL  string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessExpiry  string
	RefreshExpiry string
}

type StorageConfig struct {
	Provider         string
	Endpoint         string
	AccessKey        string
	SecretKey        string
	BucketOriginals  string
	BucketThumbnails string
	UseSSL           bool
	CDNURL           string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}
	cfg := &Config{
		Port:         getEnv("PORT", "8080"),
		Env:          getEnv("ENV", "development"),
		CookieSecret: getEnv("COOKIE_SECRET", ""),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "pixtify"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		OAuth: OAuthConfig{
			GithubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
			GithubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
			GithubRedirectURL:  getEnv("GITHUB_REDIRECT_URL", ""),
			GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", ""),
		},
		JWT: JWTConfig{
			AccessSecret:  getEnv("JWT_SECRET", ""),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", ""),
			AccessExpiry:  getEnv("JWT_ACCESS_EXPIRY", "15m"),
			RefreshExpiry: getEnv("JWT_REFRESH_EXPIRY", "168h"),
		},
		Storage: StorageConfig{
			Provider:         getEnv("STORAGE_PROVIDER", "minio"),
			Endpoint:         getEnv("STORAGE_ENDPOINT", "localhost:9000"),
			AccessKey:        getEnv("STORAGE_ACCESS_KEY", "minioadmin"),
			SecretKey:        getEnv("STORAGE_SECRET_KEY", "minioadmin"),
			BucketOriginals:  getEnv("STORAGE_BUCKET_ORIGINALS", "pixtify-originals"),
			BucketThumbnails: getEnv("STORAGE_BUCKET_THUMBNAILS", "pixtify-thumbnails"),
			UseSSL:           getEnv("STORAGE_USE_SSL", "false") == "true",
			CDNURL:           getEnv("CDN_BASE_URL", ""),
		},
	}

	validate(cfg)
	return cfg
}

func validate(cfg *Config) {
	required := map[string]string{
		"DB_PASSWORD":        cfg.Database.Password,
		"JWT_SECRET":         cfg.JWT.AccessSecret,
		"JWT_REFRESH_SECRET": cfg.JWT.RefreshSecret,
		"COOKIE_SECRET":      cfg.CookieSecret,
	}

	var missing []string
	for key, value := range required {
		if value == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		log.Fatalf("Missing required environment variables: %v", missing)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (d *DatabaseConfig) GetDSN() string {
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		return databaseURL
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}
