package infrastructure

import (
	"context"
	"time"

	"thanhnt208/vcs-sms/auth-service/config"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(cfg *config.Config) (IRedis, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := pingRedis(ctx, rdb); err != nil {
		return nil, err
	}

	return &Redis{client: rdb}, nil
}

func NewRedisWithClient(client *redis.Client) IRedis {
	return &Redis{client: client}
}

func pingRedis(ctx context.Context, client *redis.Client) error {
	return client.Ping(ctx).Err()
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
