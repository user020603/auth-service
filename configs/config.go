package configs

import (
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort      string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	RedisAddr       string
	RedisPassword   string
	JWTSecret       string
	JWTExpiresIn    int
	RefreshTokenTTL int
	LogLevel        string
	LogFile         string
}

var (
	configInstance *Config
	once           sync.Once
)

func LoadConfig() *Config {
	once.Do(func() {
		_ = godotenv.Load()

		jwtExpiresIn, err := strconv.Atoi(getEnv("JWT_EXPIRES_IN", "3600"))
		if err != nil {
			jwtExpiresIn = 3600
		}
		refreshTokenTTL, err := strconv.Atoi(getEnv("REFRESH_TOKEN_TTL", "604800"))
		if err != nil {
			refreshTokenTTL = 604800
		}

		configInstance = &Config{
			ServerPort:      getEnv("SERVER_PORT", "8000"),
			DBHost:          getEnv("DB_HOST", "localhost"),
			DBPort:          getEnv("DB_PORT", "5432"),
			DBUser:          getEnv("DB_USER", "postgres"),
			DBPassword:      getEnv("DB_PASSWORD", "password"),
			DBName:          getEnv("DB_NAME", "authdb"),
			RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
			RedisPassword:   getEnv("REDIS_PASSWORD", ""),
			JWTSecret:       getEnv("JWT_SECRET", "supersecretkey"),
			JWTExpiresIn:    jwtExpiresIn,
			RefreshTokenTTL: refreshTokenTTL,
			LogLevel:        getEnv("LOG_LEVEL", "info"),
			LogFile:         getEnv("LOG_FILE", "../logs/auth.log"),
		}
	})
	return configInstance
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
