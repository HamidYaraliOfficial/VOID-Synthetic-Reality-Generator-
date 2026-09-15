// Package config centralizes VOID's runtime configuration, loaded from
// environment variables (12-factor style) with sane local-dev defaults.
// Secrets (DB passwords, JWT signing keys, API keys) are read from the
// environment / secret manager only - never hardcoded and never shipped to
// the frontend bundle.
package config

import (
	"os"
	"strconv"
)

type Config struct {
	Env      string
	HTTPAddr string
	DataDir  string

	JWTSecret     string
	TokenTTLHours int

	PostgresDSN string
	RedisAddr   string
	NATSUrl     string
	S3Endpoint  string
	S3Bucket    string

	DefaultWorkerCount int
	MaxEventQueue      int

	CORSAllowedOrigin string
}

func Load() Config {
	return Config{
		Env:      getEnv("VOID_ENV", "development"),
		HTTPAddr: getEnv("VOID_HTTP_ADDR", ":8080"),
		DataDir:  getEnv("VOID_DATA_DIR", "./data"),

		JWTSecret:     getEnv("VOID_JWT_SECRET", "dev-insecure-secret-change-me"),
		TokenTTLHours: getEnvInt("VOID_TOKEN_TTL_HOURS", 24),

		PostgresDSN: getEnv("VOID_POSTGRES_DSN", ""),
		RedisAddr:   getEnv("VOID_REDIS_ADDR", ""),
		NATSUrl:     getEnv("VOID_NATS_URL", ""),
		S3Endpoint:  getEnv("VOID_S3_ENDPOINT", ""),
		S3Bucket:    getEnv("VOID_S3_BUCKET", "void-snapshots"),

		DefaultWorkerCount: getEnvInt("VOID_WORKER_COUNT", 8),
		MaxEventQueue:      getEnvInt("VOID_MAX_EVENT_QUEUE", 2_000_000),

		CORSAllowedOrigin: getEnv("VOID_CORS_ORIGIN", "http://localhost:3000"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
