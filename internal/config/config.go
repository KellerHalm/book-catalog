package config

import (
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	staticDir     = "./web/static"
	uploadDir     = "./data/uploads"
	maxUploadSize = 5 * 1024 * 1024
	tokenTTL      = 24 * time.Hour
)

type Config struct {
	Port      string
	DBPath    string
	JWTSecret string
	Origins   []string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		Port:      env("PORT", "8080"),
		DBPath:    env("DB_PATH", "./data/books.db"),
		JWTSecret: env("JWT_SECRET", "book-catalog-dev-secret"),
		Origins:   splitOrigins(env("ALLOWED_ORIGINS", "")),
	}
}

func splitOrigins(raw string) []string {
	origins := []string{
		"http://localhost:8080",
		"http://127.0.0.1:8080",
		"http://localhost:5173",
		"http://127.0.0.1:5173",
	}
	if raw == "" {
		return origins
	}
	for _, origin := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(origin); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

func (s *Config) StaticDir() string    { return staticDir }
func (s *Config) UploadDir() string    { return uploadDir }
func (s *Config) MaxUploadSize() int64 { return maxUploadSize }
func (s *Config) TokenTTL() time.Duration {
	return tokenTTL
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
