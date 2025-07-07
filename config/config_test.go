package config

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func resetConfig() {
	once = sync.Once{}
	configInstance = nil
}

func TestLoadConfig_Defaults(t *testing.T) {
	resetConfig()
	os.Clearenv()

	cfg := LoadConfig()

	assert.Equal(t, "8000", cfg.ServerPort)
	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, "5432", cfg.DBPort)
	assert.Equal(t, "postgres", cfg.DBUser)
	assert.Equal(t, "password", cfg.DBPassword)
	assert.Equal(t, "authdb", cfg.DBName)
	assert.Equal(t, "localhost:6379", cfg.RedisAddr)
	assert.Equal(t, "", cfg.RedisPassword)
	assert.Equal(t, 3600, cfg.JWTExpiresIn)
	assert.Equal(t, 604800, cfg.RefreshTokenTTL)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "../logs/auth.log", cfg.LogFile)
}

func TestLoadConfig_FromEnvironment(t *testing.T) {
	resetConfig()

	os.Setenv("SERVER_PORT", "9000")
	os.Setenv("DB_HOST", "db.example.com")
	os.Setenv("DB_PORT", "3306")
	os.Setenv("DB_USER", "admin")
	os.Setenv("DB_PASSWORD", "securepassword")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("REDIS_ADDR", "redis.example.com:6380")
	os.Setenv("REDIS_PASSWORD", "redispassword")
	os.Setenv("JWT_SECRET", "mysecretkey")
	os.Setenv("JWT_EXPIRES_IN", "7200")
	os.Setenv("REFRESH_TOKEN_TTL", "1209600")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FILE", "/var/logs/auth.log")

	cfg := LoadConfig()

	assert.Equal(t, "9000", cfg.ServerPort)
	assert.Equal(t, "db.example.com", cfg.DBHost)
	assert.Equal(t, "3306", cfg.DBPort)
	assert.Equal(t, "admin", cfg.DBUser)
	assert.Equal(t, "securepassword", cfg.DBPassword)
	assert.Equal(t, "testdb", cfg.DBName)
	assert.Equal(t, "redis.example.com:6380", cfg.RedisAddr)
	assert.Equal(t, "redispassword", cfg.RedisPassword)
	assert.Equal(t, "mysecretkey", cfg.JWTSecret)
	assert.Equal(t, 7200, cfg.JWTExpiresIn)
	assert.Equal(t, 1209600, cfg.RefreshTokenTTL)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "/var/logs/auth.log", cfg.LogFile)
}

func TestLoadConfig_InvalidIntegers(t *testing.T) {
	resetConfig()

	os.Setenv("JWT_EXPIRES_IN", "not_a_number")
	os.Setenv("REFRESH_TOKEN_TTL", "invalid")

	cfg := LoadConfig()

	assert.Equal(t, 3600, cfg.JWTExpiresIn)
	assert.Equal(t, 604800, cfg.RefreshTokenTTL)
}

func TestLoadConfig_SingletonBehavior(t *testing.T) {
	resetConfig()

	os.Setenv("SERVER_PORT", "8001")
	cfg1 := LoadConfig()
	cfg2 := LoadConfig()

	assert.Same(t, cfg1, cfg2, "LoadConfig should return the same instance on subsequent calls")
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_ENV_KEY", "value")
	assert.Equal(t, "value", getEnv("TEST_ENV_KEY", "default"))

	os.Unsetenv("TEST_ENV_KEY")
	assert.Equal(t, "default", getEnv("TEST_ENV_KEY", "default"))
}
