package config

import "os"

type Config struct {
	DatabaseURL    string
	JWTSecret      string
	Port           string
	CORSOrigin     string
	MigrationsPath string
	OTELEndpoint   string
	SentryDSN      string
	BotLinkSecret  string
	SMTPHost       string
	SMTPPort       string
	SMTPUser       string
	SMTPPassword   string
	SMTPFrom       string
}

func Load() *Config {
	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://gotogether:gotogether@localhost:5432/gotogether?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "dev-secret-change-me"),
		Port:           getEnv("PORT", "8080"),
		CORSOrigin:     getEnv("CORS_ORIGIN", "http://localhost:3000"),
		MigrationsPath: getEnv("MIGRATIONS_PATH", "/migrations"),
		OTELEndpoint:   getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		SentryDSN:      getEnv("SENTRY_DSN", ""),
		BotLinkSecret:  getEnv("BOT_LINK_SECRET", "dev-link-secret"),
		SMTPHost:       getEnv("SMTP_HOST", ""),
		SMTPPort:       getEnv("SMTP_PORT", "587"),
		SMTPUser:       getEnv("SMTP_USER", ""),
		SMTPPassword:   getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:       getEnv("SMTP_FROM", "noreply@gotogether.local"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
