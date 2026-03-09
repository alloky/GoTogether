package config

import "os"

type Config struct {
	TelegramBotToken string
	TelegramAPIURL   string
	BackendURL       string
	JWTSecret        string
}

func Load() *Config {
	return &Config{
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramAPIURL:   getEnv("TELEGRAM_API_URL", ""),
		BackendURL:       getEnv("BACKEND_URL", "http://127.0.0.1:8080"),
		JWTSecret:        getEnv("JWT_SECRET", "dev-secret-change-me"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
