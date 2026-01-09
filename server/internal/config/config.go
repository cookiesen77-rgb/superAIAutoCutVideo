package config

import (
	"errors"
	"os"
	"strings"
	"time"
)

type Config struct {
	Addr              string
	DatabaseURL       string
	UserDatabaseURL   string
	JWTSecret         string
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
	AWSRegion         string
	S3Bucket          string
	CorsAllowedOrigin []string
}

func Load() (Config, error) {
	cfg := Config{
		Addr:            getenv("SERVER_ADDR", ":8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		UserDatabaseURL: getenv("USER_DATABASE_URL", os.Getenv("SUPABASE_DATABASE_URL")),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		AccessTokenTTL:  durationEnv("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL: durationEnv("REFRESH_TOKEN_TTL", 30*24*time.Hour),
		AWSRegion:       getenv("AWS_REGION", "us-east-1"),
		S3Bucket:        os.Getenv("S3_BUCKET"),
	}

	origins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if origins == "" {
		cfg.CorsAllowedOrigin = []string{"http://localhost:1420", "null"}
	} else {
		cfg.CorsAllowedOrigin = splitCSV(origins)
	}

	if cfg.DatabaseURL == "" {
		return cfg, errors.New("DATABASE_URL is required")
	}
	if cfg.UserDatabaseURL == "" {
		return cfg, errors.New("USER_DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return cfg, errors.New("JWT_SECRET is required")
	}
	if cfg.S3Bucket == "" {
		return cfg, errors.New("S3_BUCKET is required")
	}

	return cfg, nil
}

func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func getenv(key, fallback string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return fallback
	}
	return val
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return d
}
