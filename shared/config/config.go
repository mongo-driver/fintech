package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type ServiceConfig struct {
	ServiceName string
	HTTPPort    string
	GRPCPort    string

	PostgresURL  string
	RedisAddr    string
	RedisPass    string
	RedisDB      int
	KafkaBrokers []string
	KafkaTopic   string

	JWTSecret string
	TokenTTL  time.Duration

	AuthGRPCAddr   string
	UserGRPCAddr   string
	WalletGRPCAddr string

	RateLimitRPS   float64
	RateLimitBurst int
	RequestTimeout time.Duration
}

func Load(serviceName string) ServiceConfig {
	return ServiceConfig{
		ServiceName: serviceName,
		HTTPPort:    getenv("HTTP_PORT", "8080"),
		GRPCPort:    getenv("GRPC_PORT", "9080"),

		PostgresURL:  getenv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/fintech?sslmode=disable"),
		RedisAddr:    getenv("REDIS_ADDR", "localhost:6379"),
		RedisPass:    getenv("REDIS_PASSWORD", ""),
		RedisDB:      getenvInt("REDIS_DB", 0),
		KafkaBrokers: splitCSV(getenv("KAFKA_BROKERS", "localhost:9092")),
		KafkaTopic:   getenv("KAFKA_TOPIC_NOTIFICATIONS", "notifications.v1"),

		JWTSecret: getenv("JWT_SECRET", "change-me-in-prod"),
		TokenTTL:  getenvDuration("JWT_TTL", 24*time.Hour),

		AuthGRPCAddr:   getenv("AUTH_GRPC_ADDR", "localhost:9081"),
		UserGRPCAddr:   getenv("USER_GRPC_ADDR", "localhost:9082"),
		WalletGRPCAddr: getenv("WALLET_GRPC_ADDR", "localhost:9083"),

		RateLimitRPS:   getenvFloat("RATE_LIMIT_RPS", 200),
		RateLimitBurst: getenvInt("RATE_LIMIT_BURST", 400),
		RequestTimeout: getenvDuration("REQUEST_TIMEOUT", 5*time.Second),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}

func getenvFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return f
}

func getenvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
