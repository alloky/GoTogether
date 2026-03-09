package config

import "os"

type Config struct {
	DatabaseURL    string
	JWTSecret      string
	Port           string
	CORSOrigin     string
	MigrationsPath string
}

func Load() *Config {
	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://gotogether:gotogether@localhost:5432/gotogether?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "dev-secret-change-me"),
		Port:           getEnv("PORT", "8080"),
		CORSOrigin:     getEnv("CORS_ORIGIN", "http://localhost:3000"),
		MigrationsPath: getEnv("MIGRATIONS_PATH", "/migrations"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
