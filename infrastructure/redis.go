package infrastructure

import (
	"context"
	"thanhnt208/vcs-sms/auth-service/configs"
	"thanhnt208/vcs-sms/auth-service/pkg/logger"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
	logger logger.ILogger
}

func NewRedis(cfg *configs.Config, logger logger.ILogger) (IRedis, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Fatal("Failed to connect to Redis", "error", err)
		return nil, err
	}
	logger.Info("Connected to Redis", "addr", cfg.RedisAddr)

	return &Redis{
		client: rdb,
		logger: logger,
	}, nil
}

func (r *Redis) GetClient() *redis.Client {
	return r.client
}

func (r *Redis) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *Redis) Close() error {
	return r.client.Close()
}
