package infrastructure

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type IDatabase interface {
	GetDB() *gorm.DB
	Close() error
	Ping(ctx context.Context) error
}

type IRedis interface {
	GetClient() *redis.Client
	Ping(ctx context.Context) error
	Close() error
}
