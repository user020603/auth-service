package repositories

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type ITokenRepository interface {
	SetRefreshToken(userID uint, token string, expiresIn time.Duration) error
	GetRefreshToken(userID uint) (string, error)
	DeleteRefreshToken(userID uint) error
}

type tokenRepository struct {
	rdb *redis.Client
}

func NewTokenRepository(rdb *redis.Client) ITokenRepository {
	return &tokenRepository{
		rdb: rdb,
	}
}

func (r *tokenRepository) SetRefreshToken(userID uint, token string, expiresIn time.Duration) error {
	key := "refresh_token:user" + strconv.FormatUint(uint64(userID), 10)
	return r.rdb.Set(context.Background(), key, token, expiresIn).Err()
}

func (r *tokenRepository) GetRefreshToken(userID uint) (string, error) {
	key := "refresh_token:user" + strconv.FormatUint(uint64(userID), 10)
	return r.rdb.Get(context.Background(), key).Result()
}

func (r *tokenRepository) DeleteRefreshToken(userID uint) error {
	key := "refresh_token:user" + strconv.FormatUint(uint64(userID), 10)
	return r.rdb.Del(context.Background(), key).Err()
}
