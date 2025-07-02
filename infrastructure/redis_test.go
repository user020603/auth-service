package infrastructure

import (
	"context"
	"errors"
	"os"
	"testing"

	"thanhnt208/vcs-sms/auth-service/config"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisWithClient(t *testing.T) {
	client, mock := redismock.NewClientMock()
	mock.ExpectPing().SetVal("PONG")
	r := NewRedisWithClient(client)
	assert.NotNil(t, r.GetClient())
	assert.NoError(t, r.Ping(context.Background()))
	assert.NoError(t, r.Close())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNewRedis_Success(t *testing.T) {
	os.Setenv("REDIS_ADDR", "localhost:6379")
	cfg := &config.Config{
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		RedisPassword: "",
	}
	// This test expects a running Redis at localhost:6379
	r, err := NewRedis(cfg)
	if err != nil {
		t.Skip("Redis server not available: ", err)
	}
	assert.NotNil(t, r)
	assert.NoError(t, r.Ping(context.Background()))
	assert.NoError(t, r.Close())
}

func TestNewRedis_Failure(t *testing.T) {
	cfg := &config.Config{
		RedisAddr:     "invalid:6379",
		RedisPassword: "",
	}
	r, err := NewRedis(cfg)
	assert.Nil(t, r)
	assert.Error(t, err)
}

func TestRedis_Ping_Error(t *testing.T) {
	client, mock := redismock.NewClientMock()
	mock.ExpectPing().SetErr(errors.New("ping error"))
	r := NewRedisWithClient(client)
	err := r.Ping(context.Background())
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// docker run --name redis-test -p 6379:6379 -d redis:7
