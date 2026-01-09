package config

import (
	"fmt"
	"log"
	"os"
)

type Config struct {
	Port     string
	Env      string
	Database DatabaseConfig
	OAuth    OAuthConfig
}

type OAuthConfig struct {
	GithubClientID     string
	GithubClientSecret string
	GithubRedirectURL  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func Load() *Config {
	cfg := &Config{
		Port: getEnv("PORT"),
		Env:  getEnv("ENV"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST"),
			Port:     getEnv("DB_PORT"),
			User:     getEnv("DB_USER"),
			Password: getEnv("DB_PASSWORD"),
			Name:     getEnv("DB_NAME"),
			SSLMode:  getEnv("DB_SSLMODE"),
		},
		OAuth: OAuthConfig{
			GithubClientID:     getEnv("GITHUB_CLIENT_ID"),
			GithubClientSecret: getEnv("GITHUB_CLIENT_SECRET"),
			GithubRedirectURL:  getEnv("GITHUB_REDIRECT_URL"),
		},
	}
	validate(cfg)

	return cfg
}

func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.Name,
		c.SSLMode,
	)
}

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Error: Required environment variable %s is not set", key)
	}
	return value
}

func validate(cfg *Config) {
	if cfg.Database.Host == "" {
		log.Fatal("Error: DB_HOST cannot be empty")
	}
	if cfg.Database.Port == "" {
		log.Fatal("Error: DB_PORT cannot be empty")
	}
	if cfg.Database.User == "" {
		log.Fatal("Error: DB_USER cannot be empty")
	}
	if cfg.Database.Name == "" {
		log.Fatal("Error: DB_NAME cannot be empty")
	}
}
