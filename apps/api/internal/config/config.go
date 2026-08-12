package config

import "os"

type Config struct {
	Port           string
	DatabaseDSN    string
	JWTSecret      string
	LogLevel       string
	AllowedOrigins []string
}

func Load() (*Config, error) {
	origins := getEnv("GATEWARDEN_ALLOWED_ORIGINS", "")
	return &Config{
		Port:           getEnv("GATEWARDEN_PORT", "8085"),
		DatabaseDSN:    getEnv("GATEWARDEN_DB_DSN", "postgres://gatewarden:gatewarden@localhost:5432/gatewarden?sslmode=disable&connect_timeout=10"),
		JWTSecret:      getEnv("GATEWARDEN_JWT_SECRET", "dev-secret-change-in-production"),
		LogLevel:       getEnv("GATEWARDEN_LOG_LEVEL", "info"),
		AllowedOrigins: parseOrigins(origins),
	}, nil
}

// parseOrigins splits a comma-separated origin string into a slice.
// Empty input or "*" returns a single wildcard origin.
// For production, specify explicit origins: https://app.example.com,https://admin.example.com
func parseOrigins(s string) []string {
	if s == "" {
		return []string{"http://localhost:5174", "http://localhost:8085"}
	}
	if s == "*" {
		return []string{"*"}
	}
	// Split by comma, trim spaces
	parts := make([]string, 0)
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			part := s[start:i]
			// Trim leading spaces
			for len(part) > 0 && part[0] == ' ' {
				part = part[1:]
			}
			// Trim trailing spaces
			for len(part) > 0 && part[len(part)-1] == ' ' {
				part = part[:len(part)-1]
			}
			if part != "" {
				parts = append(parts, part)
			}
			start = i + 1
		}
	}
	return parts
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
