package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Clear env vars that might interfere
	for _, key := range []string{"GATEWARDEN_PORT", "GATEWARDEN_DB_DSN", "GATEWARDEN_JWT_SECRET", "GATEWARDEN_LOG_LEVEL", "GATEWARDEN_ALLOWED_ORIGINS"} {
		os.Unsetenv(key)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Port != "8085" {
		t.Errorf("Port = %q, want \"8085\"", cfg.Port)
	}
	if cfg.JWTSecret != "dev-secret-change-in-production" {
		t.Errorf("JWTSecret = %q, want \"dev-secret-change-in-production\"", cfg.JWTSecret)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want \"info\"", cfg.LogLevel)
	}
	if len(cfg.AllowedOrigins) == 0 {
		t.Error("AllowedOrigins should have at least one default origin")
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	os.Setenv("GATEWARDEN_PORT", "9090")
	os.Setenv("GATEWARDEN_JWT_SECRET", "my-production-secret")
	os.Setenv("GATEWARDEN_LOG_LEVEL", "debug")
	os.Setenv("GATEWARDEN_ALLOWED_ORIGINS", "https://example.com")
	defer func() {
		os.Unsetenv("GATEWARDEN_PORT")
		os.Unsetenv("GATEWARDEN_JWT_SECRET")
		os.Unsetenv("GATEWARDEN_LOG_LEVEL")
		os.Unsetenv("GATEWARDEN_ALLOWED_ORIGINS")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want \"9090\"", cfg.Port)
	}
	if cfg.JWTSecret != "my-production-secret" {
		t.Errorf("JWTSecret = %q, want \"my-production-secret\"", cfg.JWTSecret)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want \"debug\"", cfg.LogLevel)
	}
	if len(cfg.AllowedOrigins) != 1 || cfg.AllowedOrigins[0] != "https://example.com" {
		t.Errorf("AllowedOrigins = %v, want [\"https://example.com\"]", cfg.AllowedOrigins)
	}
}

func TestLoadMultipleOrigins(t *testing.T) {
	os.Setenv("GATEWARDEN_ALLOWED_ORIGINS", "https://app.example.com,https://admin.example.com")
	defer os.Unsetenv("GATEWARDEN_ALLOWED_ORIGINS")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if len(cfg.AllowedOrigins) != 2 {
		t.Fatalf("AllowedOrigins = %v, want 2 origins", cfg.AllowedOrigins)
	}
	if cfg.AllowedOrigins[0] != "https://app.example.com" {
		t.Errorf("AllowedOrigins[0] = %q, want \"https://app.example.com\"", cfg.AllowedOrigins[0])
	}
	if cfg.AllowedOrigins[1] != "https://admin.example.com" {
		t.Errorf("AllowedOrigins[1] = %q, want \"https://admin.example.com\"", cfg.AllowedOrigins[1])
	}
}

func TestParseOrigins(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"", []string{"http://localhost:5174", "http://localhost:8085"}},
		{"*", []string{"*"}},
		{"https://example.com", []string{"https://example.com"}},
		{"https://a.com,https://b.com", []string{"https://a.com", "https://b.com"}},
		{"https://a.com, https://b.com", []string{"https://a.com", "https://b.com"}},
	}
	for _, tt := range tests {
		got := parseOrigins(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("parseOrigins(%q) = %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("parseOrigins(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_VAR_EXISTS", "hello")
	defer os.Unsetenv("TEST_VAR_EXISTS")

	got := getEnv("TEST_VAR_EXISTS", "fallback")
	if got != "hello" {
		t.Errorf("getEnv() = %q, want \"hello\"", got)
	}

	got2 := getEnv("TEST_VAR_DOES_NOT_EXIST", "fallback")
	if got2 != "fallback" {
		t.Errorf("getEnv() = %q, want \"fallback\"", got2)
	}
}
