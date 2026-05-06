package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server ServerConfig
	DB     DBConfig
	Redis  RedisConfig
	Auth   AuthConfig
}

type ServerConfig struct {
	Host           string
	Port           string
	AllowedOrigins []string
}

type DBConfig struct {
	URL string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type AuthConfig struct {
	Enabled bool
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host:           getEnv("SERVER_HOST", "127.0.0.1"),
			Port:           getEnv("SERVER_PORT", "5687"),
			AllowedOrigins: parseCSV(getEnv("ALLOWED_ORIGINS", "*")),
		},
		DB: DBConfig{
			URL: getEnv("DATABASE_URL", "postgres://alteon:alteon@localhost:5432/alteon?sslmode=disable"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       parseInt(os.Getenv("REDIS_DB"), 0),
		},
		Auth: AuthConfig{
			Enabled: getEnv("AUTH_DISABLED", "") != "true",
		},
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func parseCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
