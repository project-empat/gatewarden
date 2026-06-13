package config

import "os"

type Config struct {
	Port            string
	DatabaseDSN     string
	JWTSecret       string
	LogLevel        string
	AllowedOrigins  string
}

func Load() (*Config, error) {
	return &Config{
		Port:           getEnv("GATEWARDEN_PORT", "8085"),
		DatabaseDSN:    getEnv("GATEWARDEN_DB_DSN", "postgres://gatewarden:gatewarden@localhost:5432/gatewarden?sslmode=disable&connect_timeout=10"),
		JWTSecret:      getEnv("GATEWARDEN_JWT_SECRET", "dev-secret-change-in-production"),
		LogLevel:       getEnv("GATEWARDEN_LOG_LEVEL", "info"),
		AllowedOrigins: getEnv("GATEWARDEN_ALLOWED_ORIGINS", "*"),
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
