package infrastructure

import (
	"gorm.io/gorm"
	"github.com/redis/go-redis/v9"
	"context"
)

type IDatabase interface {
	GetDB() *gorm.DB
	Close() error
}

type IRedis interface {
	GetClient() *redis.Client
	Ping(ctx context.Context) error
	Close() error
}