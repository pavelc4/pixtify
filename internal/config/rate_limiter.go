package config

import "time"

type RateLimitConfig struct {
	LoginMax       int
	LoginWindow    time.Duration
	RegisterMax    int
	RegisterWindow time.Duration
	OAuthMax       int
	OAuthWindow    time.Duration

	APIMax    int
	APIWindow time.Duration

	AdminMax    int
	AdminWindow time.Duration
}

func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{

		LoginMax:       5,
		LoginWindow:    1 * time.Minute,
		RegisterMax:    3,
		RegisterWindow: 5 * time.Minute,
		OAuthMax:       10,
		OAuthWindow:    1 * time.Minute,

		APIMax:    100,
		APIWindow: 1 * time.Minute,

		AdminMax:    200,
		AdminWindow: 1 * time.Minute,
	}
}
