package config

import "os"

type Config struct {
	Port string
	Env  string
}

func Load() *Config {
	return &Config{
		Port: getEnv("Port", "8080"),
		Env:  getEnv("ENV", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
