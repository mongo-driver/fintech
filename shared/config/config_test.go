package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Setenv("HTTP_PORT", "9999")
	t.Setenv("KAFKA_BROKERS", "a:1,b:2")
	cfg := Load("test")
	require.Equal(t, "test", cfg.ServiceName)
	require.Equal(t, "9999", cfg.HTTPPort)
	require.Len(t, cfg.KafkaBrokers, 2)
	require.Equal(t, "a:1", cfg.KafkaBrokers[0])
}

func TestLoadDefaultsAndFallbacks(t *testing.T) {
	t.Setenv("HTTP_PORT", "")
	t.Setenv("REDIS_DB", "bad-int")
	t.Setenv("RATE_LIMIT_RPS", "bad-float")
	t.Setenv("JWT_TTL", "bad-duration")
	cfg := Load("defaults")
	require.Equal(t, "8080", cfg.HTTPPort)
	require.Equal(t, 0, cfg.RedisDB)
	require.Equal(t, 200.0, cfg.RateLimitRPS)
	require.Equal(t, "defaults", cfg.ServiceName)
}
